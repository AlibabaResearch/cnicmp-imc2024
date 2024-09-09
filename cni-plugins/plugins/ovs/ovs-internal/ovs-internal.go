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
	"github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/containernetworking/plugins/pkg/utils/sysctl"
	"github.com/vishvananda/netlink"
)

type PluginConf struct {
	types.NetConf
	BridgeName string `json:"bridge"`          // OVS bridge name, required
	VlanTag    uint   `json:"vlan,omitempty"`  // OVS internal port vlan, default 0
	MTU        uint   `json:"mtu,omitempty"`   // OVS internal port MTU, default bridge mtu
	DoDebug    bool   `json:"debug,omitempty"` // debug logger enabler
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, buildversion.BuildString("ovs-internal"))
}

func cmdAdd(args *skel.CmdArgs) error {
	portName, err := virtdev.GetRandomIfNameWithPrefix("itap")
	if err != nil {
		return err
	}
	debug.Logf("\n****** Creating ovs-internal CNI for %s in %s ******\n", portName, args.Netns)

	conf, err := parseConfig(args.StdinData)
	if err != nil {
		debugLog(conf, portName, err.Error())
		return err
	}
	debugLog(conf, portName, "parseConfig ok")

	// Step 0 - check container network namespace
	containerNNS, err := ns.GetNS(args.Netns)
	if err != nil {
		debugLog(conf, portName, err.Error())
		return fmt.Errorf("failed to open container netns %q: %v", args.Netns, err)
	}
	debugLog(conf, portName, "GetNS ok")

	// Step 1 - create port if needed
	err = ovscmd.CreateInternalVport(conf.BridgeName, portName, "", conf.VlanTag, conf.MTU, false)
	if err != nil {
		debugLog(conf, portName, err.Error())
		return err
	}
	debugLog(conf, portName, "CreateInternalVport ok")

	// Step 2 - create internal tap interface and move nns
	itapInterface, err := virtdev.CreateInternalTap(
		portName,
		containerNNS,
		false,
	)
	if err != nil {
		debugLog(conf, portName, err.Error())
		return err
	}
	defer func() {
		if err != nil {
			debugLog(conf, portName, err.Error())
			containerNNS.Do(func(_ ns.NetNS) error {
				return ip.DelLinkByName(portName)
			})
		}
	}()
	debugLog(conf, portName, "CreateInternalTap ok")

	// Step 3 - IPAM and return results
	result := &current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		Interfaces: []*current.Interface{itapInterface},
		DNS:        conf.DNS,
	}
	err = doIPAM(conf, args.StdinData, result, portName, containerNNS)
	if err != nil {
		debugLog(conf, portName, err.Error())
		return err
	}
	debugLog(conf, portName, "doIPAM ok")

	virtdev.MaskContainerdResultName(result, args.IfName, portName)
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
		return &conf, fmt.Errorf("ovs-internal needs IPAM configuration! i.e., none empty conf.IPAM.Type!")
	}

	if conf.BridgeName == "" {
		return &conf, fmt.Errorf("ovs-internal should be provided with the ovs bridge name!")
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
		debug.Logf("[%s:%s]: %s\n", conf.BridgeName, ifName, message)
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
		// All addresses apply to the container macvlan interface
		ipc.Interface = current.Int(0)
	}

	err = containerNNS.Do(func(_ ns.NetNS) error {
		_, _ = sysctl.Sysctl(fmt.Sprintf("net/ipv4/conf/%s/arp_notify", ifName), "1")

		return ipam.ConfigureIface(ifName, result)
	})
	if err != nil {
		return err
	}
	debugLog(conf, ifName, fmt.Sprintf("assigned ip addr: %q", result.IPs[0].Address.IP.String()))

	return nil
}
