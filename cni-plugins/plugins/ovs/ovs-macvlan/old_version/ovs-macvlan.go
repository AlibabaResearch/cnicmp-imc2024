package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	// "runtime"

	"github.com/vishvananda/netlink"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"

	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"

	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/containernetworking/plugins/pkg/utils/sysctl"

	"github.com/containernetworking/plugins/pkg/ovs-cni/ovsdb"
	vd "github.com/containernetworking/plugins/pkg/ovs-cni/virtdevold"
)

const (
	linkstateCheckRetries  = 5
	linkStateCheckInterval = 600 // in milliseconds
)

const (
	defaultOVSIfType  = "tap"    // OVS Interface Type (tap only)
	defaultOfPortType = "access" // OpenFlow Port Type (copy from ovs-cni)
	defaultOfPortNum  = 0        // Static OpenFlow Port Numer (default no_specification)
	defaultOVNPort    = ""       // OVN Port (default empty)
)

const doDelTap = false

type PluginConf struct {
	// The based struct type "types.NetConf" embeds the standard NetConf structure which allows your plugin
	// to more easily parse standard fields like Name, Type, CNIVersion, and PrevResult.
	types.NetConf
	OVSBridge              string  `json:"ovsBridge"`                        // OVS conf: required
	OVSSocketFile          string  `json:"ovsSocketFile"`                    // OVS conf: required
	OVSVlanTag             uint    `json:"vlan,omitempty"`                   // OVS conf
	MTU                    int     `json:"mtu"`                              // VirtDev common conf: required
	MvMac                  string  `json:"macvlanMac,omitempty"`             // VirtDev macvlan conf
	TapIfName              string  `json:"tapIfName"`                        // VirtDev tap conf: required
	TapMultiQueue          bool    `json:"tapMultiQueue"`                    // VirtDev tap conf: required
	TapMac                 string  `json:"tapMac"`                           // VirtDev tap conf: required
	TapOwner               *uint32 `json:"tapOwner,omitempty"`               // VirtDev tap conf
	TapGroup               *uint32 `json:"tapGroup,omitempty"`               // VirtDev tap conf
	TapSelinuxContext      string  `json:"tapSelinuxContext,omitempty"`      // VirtDev tap conf
	LinkStateCheckRetries  int     `json:"linkStateCheckRetries,omitempty"`  // VirtDev tap conf
	LinkStateCheckInterval int     `json:"linkStateCheckInterval,omitempty"` // VirtDev tap conf
	RuntimeConfig          struct {
		TapMac string `json:"tapMac,omitempty"`
		MvMac  string `json:"macvlanMac,omitempty"`
	} `json:"runtimeConfig,omitempty"`
}

// standard CNI interface registration and implementation
func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("ovs-macvlan"))
}

func cmdAdd(args *skel.CmdArgs) error {
	// Step 1 - Preparation Stage
	// 1.1 load conf
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return err
	}

	// 1.2 open ovsdb driver
	ovsDriver, err := initOVSDriver(conf)
	if err != nil {
		return fmt.Errorf("failed to open ovs driver: %v", err)
	}

	// 1.3 check container/host nns
	containerNetns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer containerNetns.Close()
	hostNetns, err := ns.GetCurrentNS()
	if err != nil {
		return fmt.Errorf("failed to open host netns: %v", err)
	}
	defer hostNetns.Close()

	// Step 2 - create tap device and bind it to OVS vport if needed
	// 2.1 create tapInterface
	tapInterface, err := findOrCreateTapMaster(conf, hostNetns, ovsDriver)
	defer func() {
		if err != nil {
			// Unlike veth pair, OVS port will not be automatically removed
			// if the following IPAM configuration fails and netns gets removed.
			portName, portFound, err := ovsDriver.GetOvsPortForContIface(conf.TapIfName, args.Netns) // ToDo: bug here!!!!!!
			if err != nil {
				fmt.Printf("Failed best-effort cleanup: %v", err)
			}
			if portFound {
				if err := ovsDriver.DeletePort(portName); err != nil {
					fmt.Printf("Failed best-effort cleanup: %v", err)
				}
			}
		}
	}()
	// 2.2 wait tapInterface up
	err = waitLinkUp(ovsDriver, conf.TapIfName, conf.LinkStateCheckRetries, conf.LinkStateCheckInterval)
	if err != nil {
		return err
	}

	// Step 3 - create macvlan device with the tap device as its master/parent
	// 3.1 create macvlan in ContainerNetns
	macvlanInterface, err := vd.CreateMacvlan(
		true,
		tapInterface.Name,
		conf.MTU,
		conf.MvMac,
		args.IfName,
		containerNetns,
	)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			containerNetns.Do(func(_ ns.NetNS) error {
				return ip.DelLinkByName(args.IfName)
			})
		}
	}()

	// 3.2 IP configuration  & return result
	ipamResult, err := doIPAM(conf, args.StdinData)
	if err != nil {
		return err
	}

	result := &current.Result{
		CNIVersion: current.ImplementedSpecVersion,
		Interfaces: []*current.Interface{macvlanInterface},
		IPs:        ipamResult.IPs,
		Routes:     ipamResult.Routes,
		DNS:        conf.DNS,
	}

	for _, ipc := range result.IPs {
		// All addresses apply to the container macvlan interface
		ipc.Interface = current.Int(0)
	}

	err = containerNetns.Do(func(_ ns.NetNS) error {
		_, _ = sysctl.Sysctl(fmt.Sprintf("net/ipv4/conf/%s/arp_notify", args.IfName), "1")

		if conf.MvMac == "" && len(result.IPs) >= 1 {
			containerMac := vd.IPAddrToHWAddr(result.IPs[0].Address.IP)
			containerLink, err := netlink.LinkByName(args.IfName)
			if err != nil {
				return fmt.Errorf("failed to lookup container interface %q: %v", args.IfName, err)
			}
			err = vd.AssignMacToLink(containerLink, containerMac, args.IfName)
			if err != nil {
				return err
			}
			result.Interfaces[0].Mac = containerMac.String()
		}

		err := ipam.ConfigureIface(args.IfName, result)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Pass through the result for the next plugin
	return types.PrintResult(result, conf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	conf, err := parseConfig(args.StdinData)
	if err != nil {
		return err
	}

	// TODO: When to del Tap??????

	// Step 1 - del OVS interface/port
	ovsDriver, err := initOVSDriver(conf)
	if err != nil {
		return fmt.Errorf("failed to open ovs driver: %v", err)
	}

	if args.Netns != "" {
		// Unlike veth pair, OVS port will not be automatically removed when
		// container namespace is gone. Find port matching DEL arguments and remove
		// it explicitly.
		portName, portFound, err := ovsDriver.GetOvsPortForContIface(conf.TapIfName, args.Netns) // TODO: bug here, tap not in args.Netns
		if err != nil {
			return fmt.Errorf("Failed to obtain OVS port for given connection: %v", err)
		}

		// Do not return an error if the port was not found, it may have been
		// already removed by someone.
		if portFound {
			if err := ovsDriver.DeletePort(portName); err != nil {
				return err
			}
		}

		err = ns.WithNetNSPath(args.Netns, func(ns.NetNS) error {
			err = ip.DelLinkByName(args.IfName)
			return err
		})

		// do the following as per cni spec (i.e. Plugins should generally complete a DEL action
		// without error even if some resources are missing)
		if _, ok := err.(ns.NSPathNotExistErr); ok || err == ip.ErrLinkNotFound {
			if portFound {
				if err := ip.DelLinkByName(portName); err != nil {
					fmt.Printf("Failed best-effort cleanup of %s: %v", portName, err)
				}
			}
			return nil
		}

		// finally removes all ports whose interfaces have an error
		if err := ovsdb.CleanPorts(ovsDriver); err != nil {
			return err
		}
	}

	// Step2 - del tap
	// TODO....

	// Step 3 - del macvlan
	err = ipam.ExecDel(conf.IPAM.Type, args.StdinData)
	if err != nil {
		return err
	}

	// There is a netns so try to clean up. Delete can be called multiple times
	// so don't return an error if the device is already removed.
	// TODO: when to delete netns ?????
	err = ns.WithNetNSPath(args.Netns, func(_ ns.NetNS) error {
		if err := ip.DelLinkByName(args.IfName); err != nil {
			if err != ip.ErrLinkNotFound {
				return err
			}
		}
		return nil
	})

	if err != nil {
		//  if NetNs is passed down by the Cloud Orchestration Engine, or if it called multiple times
		// so don't return an error if the device is already removed.
		// https://github.com/kubernetes/kubernetes/issues/43014#issuecomment-287164444
		_, ok := err.(ns.NSPathNotExistErr)
		if ok {
			return nil
		}
		return err
	}

	return err
}

func cmdCheck(args *skel.CmdArgs) error {
	return fmt.Errorf("not implemented")
}

// parseConfig parses the supplied configuration (and prevResult) from stdin.
func parseConfig(stdin []byte) (*PluginConf, error) {
	conf := PluginConf{}

	if err := json.Unmarshal(stdin, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse CNI configuration: %v", err)
	}

	// Parse previous result: not used here
	// if err := version.ParsePrevResult(&conf.NetConf); err != nil {
	// 	return nil, fmt.Errorf("could not parse prevResult: %v", err)
	// }

	// Do any validation here
	// (1) validate conf.IPAM.Type != "" (enable Layer 3)
	if conf.IPAM.Type == "" {
		return nil, fmt.Errorf("ovs-macvlan only supports L3 mode! i.e., none empty conf.IPAM.Type")
	}

	// (2) validate Parameters of Tap Master
	if conf.TapIfName == "" {
		return nil, fmt.Errorf("ovs-macvlan should be provided with the specific name of its tap master")
	}
	if conf.TapMac == "" {
		return nil, fmt.Errorf("ovs-macvlan should be provided with the mac address of its tap master")
	}

	// (3) overwrite RuntimeConfig MAC if needed
	if conf.RuntimeConfig.TapMac != "" {
		conf.TapMac = conf.RuntimeConfig.TapMac
	}
	if conf.RuntimeConfig.MvMac != "" {
		conf.MvMac = conf.RuntimeConfig.MvMac
	}

	// (4) configure Retry Time
	if conf.LinkStateCheckRetries == 0 {
		conf.LinkStateCheckRetries = linkstateCheckRetries
	}
	if conf.LinkStateCheckInterval == 0 {
		conf.LinkStateCheckInterval = linkStateCheckInterval
	}

	// ignore envArgs for now

	return &conf, nil
}

func initOVSDriver(conf *PluginConf) (*ovsdb.OvsBridgeDriver, error) {
	ovsDriver, err := ovsdb.NewOvsBridgeDriver(conf.OVSBridge, conf.OVSSocketFile)
	if err != nil {
		return nil, err
	}
	if err := ovsdb.CleanPorts(ovsDriver); err != nil {
		return nil, err
	}
	return ovsDriver, nil
}

func findOrCreateTapMaster(conf *PluginConf, hostNetns ns.NetNS, ovsDriver *ovsdb.OvsBridgeDriver) (*current.Interface, error) {
	tapIfLink, err := netlink.LinkByName(conf.TapIfName)
	if err != nil {
		// step 1: create tap interface
		tapInterface, err := vd.CreateTapInHostNetNS(
			"",
			conf.TapSelinuxContext,
			conf.MTU,
			conf.TapMultiQueue,
			conf.TapMac,
			conf.TapOwner,
			conf.TapGroup,
			conf.TapIfName,
			hostNetns, // tap in hostNetNS rather than containerNetNS
		)
		if err != nil {
			return nil, err
		}
		// Delete link if err to avoid link leak in this ns
		defer func() {
			if err != nil {
				ip.DelLinkByName(conf.TapIfName)
			}
		}()

		// step 2: create OVS vport with hardcoded parameters
		err = ovsDriver.CreateTapPort(
			conf.TapIfName,    // Tap Interface Name
			defaultOVSIfType,  // OVS Interface Type (tap only)
			defaultOfPortType, // OpenFlow Port Type (copy from ovs-cni)
			defaultOfPortNum,  // Static OpenFlow Port Numer (default no_specification)
			conf.OVSVlanTag,   // OVS Vlan ID (default 0)
			nil,               // Trunk (?) (default nil)
			defaultOVNPort,    // OVN Port (default empty)
		)
		if err != nil {
			return nil, err
		}

		// step 3: change tap interface status to up
		tapIfLink, err := netlink.LinkByName(conf.TapIfName)
		if err != nil {
			return nil, fmt.Errorf("failed to find the newly created tap interface name %q: %v", conf.TapIfName, err)
		}
		if err := netlink.LinkSetUp(tapIfLink); err != nil {
			return nil, fmt.Errorf("failed to set %q UP: %v", conf.TapIfName, err)
		}

		return tapInterface, nil
	} else {
		tapInterface := &current.Interface{
			Name:    conf.TapIfName,
			Mac:     tapIfLink.Attrs().HardwareAddr.String(),
			Sandbox: hostNetns.Path(),
		}
		return tapInterface, nil
	}
}

func waitLinkUp(ovsDriver *ovsdb.OvsBridgeDriver, ofPortName string, retryCount, interval int) error {
	checkInterval := time.Duration(interval) * time.Millisecond
	for i := 1; i <= retryCount; i++ {
		portState, err := ovsDriver.GetOFPortOpState(ofPortName)
		if err != nil {
			fmt.Printf("error in retrieving port %s state: %v", ofPortName, err)
		} else {
			if portState == "up" {
				break
			}
		}
		if i == retryCount {
			return fmt.Errorf("The OF port %s state is not up, try increasing number of retries/interval config parameter", ofPortName)
		}
		time.Sleep(checkInterval)
	}
	return nil
}

func doIPAM(conf *PluginConf, stdin []byte) (*current.Result, error) {
	r, err := ipam.ExecAdd(conf.IPAM.Type, stdin)
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
	var ipamResult *current.Result
	ipamResult, err = current.NewResultFromResult(r)
	if err != nil {
		return nil, err
	}

	if len(ipamResult.IPs) == 0 {
		return nil, errors.New("IPAM plugin returned missing IP config")
	}

	return ipamResult, nil
}
