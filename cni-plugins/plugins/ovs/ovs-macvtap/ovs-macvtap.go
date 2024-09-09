package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/ovs-cni/debug"
	"github.com/containernetworking/plugins/pkg/ovs-cni/ovscmd"
	"github.com/containernetworking/plugins/pkg/ovs-cni/virtdev"
	"github.com/containernetworking/plugins/pkg/ovs-cni/virtdevcmd"
	"github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/containernetworking/plugins/pkg/utils/sysctl"
	"github.com/vishvananda/netlink"
)

type PluginConf struct {
	types.NetConf
	BridgeName string `json:"bridge"`          // OVS bridge name, required
	PortName   string `json:"port"`            // OVS internal port name, required
	VlanTag    uint   `json:"vlan,omitempty"`  // OVS internal port vlan, default 0
	MTU        uint   `json:"mtu,omitempty"`   // OVS internal port and macvtap MTU, default bridge mtu
	QOS        int    `json:"qos,omitempty"`   // OVS number of qos rules for each container netdev
	ACL        int    `json:"acl,omitempty"`   // OVS number of acl rules for each container netdev
	PortMac    string `json:"pmac,omitempty"`  // OVS internal port MAC address, default random	[Not used for now!!!]
	DevMac     string `json:"dmac,omitempty"`  // macvtap MAC address, default random
	UseRPC     bool   `json:"rpc,omitempty"`   // create ovs port using rpc, default false
	UseCmd     bool   `json:"cmd,omitempty"`   // create macvlan using cmd, default false
	DoDebug    bool   `json:"debug,omitempty"` // debug logger enabler
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, buildversion.BuildString("ovs-macvtap"))
}

func cmdAdd(args *skel.CmdArgs) error {
	ifName, err := virtdev.GetRandomIfNameWithPrefix("omcvt")
	debug.Logf("\n****** Creating ovs-macvtap CNI for %s(%s) in %s ******\n", args.IfName, ifName, args.Netns)

	conf, err := parseConfig(args.StdinData)
	if err != nil {
		debugLog(conf, ifName, err.Error())
		return err
	}
	debugLog(conf, ifName, "parseConfig ok")

	// Step 0 - check container network namespace
	containerNNS, err := ns.GetNS(args.Netns)
	if err != nil {
		debugLog(conf, ifName, err.Error())
		return fmt.Errorf("failed to open container netns %q: %v", args.Netns, err)
	}
	defer containerNNS.Close()
	debugLog(conf, ifName, "GetNS ok")

	// Step 1 - create port if needed
	if conf.UseRPC {
		err = ovscmd.CreateInternalVportRPC(conf.BridgeName, conf.PortName, conf.PortMac, conf.VlanTag, conf.MTU, true)
	} else {
		err = ovscmd.CreateInternalVport(conf.BridgeName, conf.PortName, conf.PortMac, conf.VlanTag, conf.MTU, true)
	}
	if err != nil {
		debugLog(conf, ifName, err.Error())
		return err
	}
	debugLog(conf, ifName, "CreateInternalVport ok")

	// Step 2 - create macvtap in container nns with master of the ovs port
	var macvtapInterface *current.Interface
	if conf.UseCmd {
		macvtapInterface, err = virtdevcmd.CreateMacvtap(
			conf.PortName,
			ifName,
			containerNNS,
			conf.MTU,
			conf.DevMac,
			false,
		)
	} else {
		macvtapInterface, err = virtdev.CreateMacvtap(
			conf.PortName,
			ifName,
			containerNNS,
			conf.MTU,
			conf.DevMac,
			false,
		)
	}
	if err != nil {
		debugLog(conf, ifName, err.Error())
		return err
	}
	defer func() {
		if err != nil {
			debugLog(conf, ifName, err.Error())
			containerNNS.Do(func(_ ns.NetNS) error {
				return ip.DelLinkByName(ifName)
			})
		}
	}()
	debugLog(conf, ifName, "CreateMacvtap ok")

	// Step 3 - IPAM and return results
	result := &current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		Interfaces: []*current.Interface{macvtapInterface},
		DNS:        conf.DNS,
	}
	err = doIPAM(conf, args.StdinData, result, ifName, containerNNS)
	if err != nil {
		debugLog(conf, ifName, err.Error())
		return err
	}
	debugLog(conf, ifName, "doIPAM ok")

	// Step 4 - add acl and qos
	if conf.ACL > 0 {
		err = ovscmd.AddACLRules(conf.BridgeName, result.IPs[0].Address.String(), conf.ACL)
		if err != nil {
			return err
		}
		debugLog(conf, ifName, "Add ACL & QoS ok")
	}

	virtdev.MaskContainerdResultName(result, args.IfName, ifName)
	return types.PrintResult(result, conf.CNIVersion)
}

func cmdCheck(_ *skel.CmdArgs) error {
	return nil
}

func cmdDel(_ *skel.CmdArgs) error {
	return nil
}

func parseConfig(stdin []byte) (*PluginConf, error) {
	conf := PluginConf{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse network configuration: %v", err)
	}

	// Config validation
	if conf.IPAM.Type == "" {
		return &conf, fmt.Errorf("ovs-macvtap needs IPAM configuration! i.e., none empty conf.IPAM.Type!")
	}

	if conf.BridgeName == "" {
		return &conf, fmt.Errorf("ovs-macvtap should be provided with the ovs bridge name!")
	}

	if conf.PortName == "" {
		return &conf, fmt.Errorf("ovs-macvtap should be provided with the ovs internal port name!")
	}

	// Check bridge existence
	ovsBridge, err := netlink.LinkByName(conf.BridgeName)
	if err != nil {
		return &conf, fmt.Errorf("failed to lookup ovs bridge %q: %v", conf.BridgeName, err)
	}
	brMTU := uint(ovsBridge.Attrs().MTU)
	if conf.MTU == 0 {
		conf.MTU = brMTU
	} else if conf.MTU > brMTU {
		return &conf, fmt.Errorf("invalid MTU %d, must be [0, ovs bridge MTU(%d)]", conf.MTU, brMTU)
	}

	return &conf, nil
}

func debugLog(conf *PluginConf, ifName string, message string) {
	if conf.DoDebug {
		debug.Logf("[%s:%s:%s]: %s\n", conf.BridgeName, conf.PortName, ifName, message)
	}
}

func doIPAM(conf *PluginConf, stdin []byte, result *current.Result, ifName string, containerNNS ns.NetNS) error {
	// run the IPAM plugin and get back the config to apply
	r, err := ipam.ExecAdd(conf.IPAM.Type, stdin)
	if err != nil {
		return err
	}

	// Invoke ipam del if err to avoid ip leak
	defer func() {
		if err != nil {
			ipam.ExecDel(conf.IPAM.Type, stdin)
		}
	}()

	// Convert whatever the IPAM result was into the current Result type
	ipamResult, err := current.NewResultFromResult(r)
	if err != nil {
		return err
	}

	if len(ipamResult.IPs) == 0 {
		return errors.New("IPAM plugin returned missing IP config")
	}

	result.IPs = ipamResult.IPs
	result.Routes = ipamResult.Routes

	for _, ipc := range ipamResult.IPs {
		// All addresses apply to the container macvtap interface
		ipc.Interface = current.Int(0)
	}

	err = containerNNS.Do(func(_ ns.NetNS) error {
		_, _ = sysctl.Sysctl(fmt.Sprintf("net/ipv4/conf/%s/arp_notify", ifName), "1")

		if conf.UseCmd {
			return virtdevcmd.ConfigureIPv4andSetup(ifName, result.IPs[0].Address.String())
		} else {
			return virtdevcmd.ConfigureIPv4andSetup(ifName, result.IPs[0].Address.String())
		}
	})
	if err != nil {
		return err
	}

	return nil
}
