package main

import (
	"encoding/json"
	"errors"
	"fmt"

	ciliumModels "github.com/cilium/cilium/api/v1/models"
	"github.com/containernetworking/cni/pkg/skel"
	cniTypes "github.com/containernetworking/cni/pkg/types"
	cniTypesV1 "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	ciliumClient "github.com/containernetworking/plugins/pkg/cilium-cni/client"
	ciliumConsts "github.com/containernetworking/plugins/pkg/cilium-cni/defaults"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/ovs-cni/debug"
	"github.com/containernetworking/plugins/pkg/ovs-cni/ovscmd"
	"github.com/containernetworking/plugins/pkg/ovs-cni/virtdev"
	"github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/vishvananda/netlink"
)

type PluginConf struct {
	cniTypes.NetConf
	BridgeName string `json:"bridge"`          // OVS bridge name, required
	PortName   string `json:"port"`            // OVS internal port name, required
	VlanTag    uint   `json:"vlan,omitempty"`  // OVS internal port vlan, default 0
	PortMac    string `json:"pmac,omitempty"`  // OVS internal port MAC address, default random	[Not used for now!!!]
	Mode       string `json:"mode,omitempty"`  // Cilium mode: default veth
	MTU        uint   `json:"mtu,omitempty"`   // OVS internal port and cilium netdev MTU, default bridge mtu
	QOS        int    `json:"qos,omitempty"`   // OVS number of qos rules for each container netdev
	ACL        int    `json:"acl,omitempty"`   // OVS number of acl rules for each container netdev
	UseRPC     bool   `json:"rpc,omitempty"`   // create ovs port using rpc, default false
	DoDebug    bool   `json:"debug,omitempty"` // debug logger enabler
}

var (
	CILIUM_VETH_MODE string = "veth"
	CILIUM_ABANDON   bool   = true
)

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, buildversion.BuildString("ovs-cilium"))
}

func cmdAdd(args *skel.CmdArgs) (err error) {
	ifName, err := virtdev.GetRandomIfNameWithPrefixWithMaxLength("ocilm", 11)
	hostVethName := ifName + "A"
	contVethName := ifName + "B"
	debug.Logf("\n****** Creating ovs-cilium CNI for %s(%s) in %s ******\n", args.IfName, ifName, args.Netns)

	conf, err := parseConfig(args.StdinData)
	if err != nil {
		debugLog(conf, ifName, err.Error())
		return err
	}
	client, err := ciliumClient.NewDefaultClientWithTimeout(ciliumConsts.ClientConnectTimeout)
	if err != nil {
		return fmt.Errorf("unable to connect to Cilium agent: %w", ciliumClient.Hint(err))
	}
	ciliumConf, err := getConfigFromCiliumAgent(client)
	if err != nil {
		return fmt.Errorf("unable to get conf from Cilium agent: %v", err)
	}
	debugLog(conf, ifName, fmt.Sprintf("cilium_host at %q, parseConfig ok", ciliumConf.Addressing.IPV4.IP))

	if CILIUM_ABANDON {
		return fmt.Errorf("cilium-ovs plugin implemented in a separate project originated from cilium project, abandon here!!!")
	}

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

	// Step 2 - create cilium/eBPF based veth pair
	var ciliumVirtDevIf *cniTypesV1.Interface
	switch conf.Mode {
	case CILIUM_VETH_MODE:
		// create veth pair and delete it if error exists
		ciliumVirtDevIf, _, err = virtdev.CreateVethPair(
			hostVethName,
			contVethName,
			containerNNS,
			conf.MTU,
			true,
			true,
		)
		if err != nil {
			return fmt.Errorf("[ovs-cilium] fail to create veth pair: %v", err)
		}
		debugLog(conf, ifName, "CreateVethPair ok")
	default:
		return fmt.Errorf("[ovs-cilium] mode %q does not supported yet!", conf.Mode)
	}
	if ciliumVirtDevIf == nil {
		return fmt.Errorf("[ovs-cilium] fail to create virdev in mode: %s", conf.Mode)
	}

	// Step 3 - IPAM and return results
	result := &cniTypesV1.Result{
		CNIVersion: cniTypesV1.ImplementedSpecVersion,
		Interfaces: []*cniTypesV1.Interface{ciliumVirtDevIf},
		DNS:        conf.DNS,
	}
	// Instead of using allocateIPsWithCiliumAgent from Cilium, we use the default ipam
	var ipamReponse *ciliumModels.IPAMResponse
	ipamReponse, err = allocateIPsWithCNIPlugin(conf, ciliumConf, args.StdinData)
	if err != nil {
		return fmt.Errorf("fail to allocate IP and build routes: %v", err)
	}
	if !ipv4IsEnabled(ipamReponse) {
		return errors.New("IPAM does not provide IPv4 address!")
	}
	if ipv6IsEnabled(ipamReponse) {
		return errors.New("IPAM does not support IPv6 address!")
	}

	// Step 4 - add acl and qos
	if conf.ACL > 0 {
		err = ovscmd.AddACLRules(conf.BridgeName, result.IPs[0].Address.String(), conf.ACL)
		if err != nil {
			return err
		}
		debugLog(conf, ifName, "Add ACL & QoS ok")
	}

	// other mask???
	// virtdev.MaskContainerdResultName(result, args.IfName, ifName)

	return cniTypes.PrintResult(result, conf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	return fmt.Errorf("[ovs-cilium] cmdDel not implemented")
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}

func parseConfig(stdin []byte) (*PluginConf, error) {
	conf := PluginConf{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse network configuration: %v", err)
	}

	// Config validation
	if conf.IPAM.Type == "" {
		return &conf, fmt.Errorf("ovs-cilium needs IPAM configuration! i.e., none empty conf.IPAM.Type!")
	}

	if conf.BridgeName == "" {
		return &conf, fmt.Errorf("ovs-cilium should be provided with the ovs bridge name!")
	}

	if conf.PortName == "" {
		return &conf, fmt.Errorf("ovs-cilium should be provided with the ovs internal port name!")
	}

	if conf.Mode == "" || conf.Mode == "l2" || conf.Mode == "l3" {
		conf.Mode = CILIUM_VETH_MODE
	}
	if conf.Mode != CILIUM_VETH_MODE {
		return &conf, fmt.Errorf("ovs-cilium only support veth mode, request = %q!", conf.Mode)
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

func getConfigFromCiliumAgent(client *ciliumClient.Client) (*ciliumModels.DaemonConfigurationStatus, error) {
	configResult, err := client.ConfigGet()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve configuration from Cilium agent: %w", err)
	}

	if configResult == nil || configResult.Status == nil {
		return nil, errors.New("received empty configuration object from Cilium agent")
	}

	return configResult.Status, nil
}

func allocateIPsWithCNIPlugin(conf *PluginConf, ciliumConf *ciliumModels.DaemonConfigurationStatus, stdin []byte) (*ciliumModels.IPAMResponse, error) {
	// run the IPAM plugin and get back the config to apply
	ipamRawResult, err := ipam.ExecAdd(conf.IPAM.Type, stdin)
	if err != nil {
		return nil, err
	}

	// Invoke ipam del if err to avoid ip leak
	defer func() {
		if err != nil {
			ipam.ExecDel(conf.IPAM.Type, stdin)
		}
	}()

	// Convert whatever the IPAM result was into the current Result type
	ipamResult, err := cniTypesV1.NewResultFromResult(ipamRawResult)
	if err != nil {
		return nil, err
	}

	if len(ipamResult.IPs) == 0 {
		return nil, errors.New("IPAM plugin returned missing IP config")
	}

	// Translate the IPAM result into the same format as a response from Cilium agent.
	ipamReponse := &ciliumModels.IPAMResponse{
		HostAddressing: ciliumConf.Addressing,
		Address:        &ciliumModels.AddressPair{},
	}

	// Safe to assume at most one IP per family. The K8s API docs say:
	// "Pods may be allocated at most 1 value for each of IPv4 and IPv6"
	// https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/
	for _, ipConfig := range ipamResult.IPs {
		ipNet := ipConfig.Address
		if ipv4 := ipNet.IP.To4(); ipv4 != nil {
			ipamReponse.Address.IPV4 = ipNet.String()
			ipamReponse.IPV4 = &ciliumModels.IPAMAddressResponse{IP: ipv4.String()}
		} else {
			return nil, errors.New("does not support IPv6 yet...")
		}
	}

	return ipamReponse, nil
}

func ipv4IsEnabled(ipamReponse *ciliumModels.IPAMResponse) bool {
	if ipamReponse == nil || ipamReponse.Address.IPV4 == "" {
		return false
	}

	if ipamReponse.HostAddressing != nil && ipamReponse.HostAddressing.IPV4 != nil {
		return ipamReponse.HostAddressing.IPV4.Enabled
	}

	return true
}

func ipv6IsEnabled(ipamReponse *ciliumModels.IPAMResponse) bool {
	if ipamReponse == nil || ipamReponse.Address.IPV6 == "" {
		return false
	}

	if ipamReponse.HostAddressing != nil && ipamReponse.HostAddressing.IPV6 != nil {
		return ipamReponse.HostAddressing.IPV6.Enabled
	}

	return true
}

func debugLog(conf *PluginConf, ifName string, message string) {
	if conf.DoDebug {
		debug.Logf("[%s:%s:%s]: %s\n", conf.BridgeName, conf.PortName, ifName, message)
	}
}
