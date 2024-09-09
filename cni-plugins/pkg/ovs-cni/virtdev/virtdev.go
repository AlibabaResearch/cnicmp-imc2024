package virtdev

import (
	"fmt"
	"net"

	current "github.com/containernetworking/cni/pkg/types/100"
	csysctl "github.com/containernetworking/plugins/pkg/cilium-cni/sysctl"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/ovs-cni/debug"
	"github.com/containernetworking/plugins/pkg/ovs-cni/utils"
	"github.com/vishvananda/netlink"
)

func CreateMacvlan(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, mac string, doSetUp bool) (*current.Interface, error) {
	return createMacvlanIf(masterName, ifName, containerNetns, mtu, mac, false, doSetUp)
}

func CreateMacvtap(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, mac string, doSetUp bool) (*current.Interface, error) {
	return createMacvlanIf(masterName, ifName, containerNetns, mtu, mac, true, doSetUp)
}

func createMacvlanIf(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, mac string, isMacvtap bool, doSetUp bool) (*current.Interface, error) {
	macvlan := &current.Interface{}

	var devStr string
	if isMacvtap {
		devStr = "macvtap"
	} else {
		devStr = "macvlan"
	}

	var masterLink netlink.Link
	var err error
	masterLink, err = netlink.LinkByName(masterName)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup master %q: %v", masterName, err)
	}

	// due to kernel bug we have to create with tmpName or it might
	// collide with the name on the host and error out
	tmpName, err := ip.RandomVethName()
	if err != nil {
		return nil, err
	}

	linkAttrs := netlink.LinkAttrs{
		MTU:         int(mtu),
		Name:        tmpName,
		ParentIndex: masterLink.Attrs().Index,
		Namespace:   netlink.NsFd(int(containerNetns.Fd())),
	}

	if mac != "" {
		addr, err := net.ParseMAC(mac)
		if err != nil {
			return nil, fmt.Errorf("invalid args %v for MAC addr: %v", mac, err)
		}
		linkAttrs.HardwareAddr = addr
	}

	macvlanLink, err := createMacvlanLink(&linkAttrs, netlink.MACVLAN_MODE_BRIDGE, isMacvtap)
	if err != nil {
		return nil, err
	}

	if err = netlink.LinkAdd(macvlanLink); err != nil {
		return nil, fmt.Errorf("failed to create %s: %v", devStr, err)
	}

	err = containerNetns.Do(func(_ ns.NetNS) error {
		err := ip.RenameLink(tmpName, ifName)
		if err != nil {
			_ = netlink.LinkDel(macvlanLink)
			return fmt.Errorf("failed to rename %s to %q: %v", devStr, ifName, err)
		}
		macvlan.Name = ifName

		// Re-fetch macvlan to get all properties/attributes
		contMacvlan, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("failed to refetch %s named %q: %v", devStr, ifName, err)
		}
		debug.Logf("find %s(if=%q, mac=%q) created in (%q) [virtdev]",
			devStr,
			ifName,
			contMacvlan.Attrs().HardwareAddr,
			containerNetns.Path(),
		)

		if doSetUp {
			if err := netlink.LinkSetUp(contMacvlan); err != nil {
				return fmt.Errorf("failed to set %q UP: %v", ifName, err)
			}
		}

		macvlan.Mac = contMacvlan.Attrs().HardwareAddr.String()
		macvlan.Sandbox = containerNetns.Path()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return macvlan, nil
}

// for debugging the netlink & cmd versions of macvtap/ipvtap
func createMacvlanIfTest(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, mac string, isMacvtap bool, doSetUp bool) (*current.Interface, error) {
	macvlan := &current.Interface{}

	var devStr string
	if isMacvtap {
		devStr = "macvtap"
	} else {
		devStr = "macvlan"
	}

	var masterLink netlink.Link
	var err error
	masterLink, err = netlink.LinkByName(masterName)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup master %q: %v", masterName, err)
	}

	linkAttrs := netlink.LinkAttrs{
		MTU:         int(mtu),
		Name:        ifName,
		ParentIndex: masterLink.Attrs().Index,
	}

	macvlanLink, err := createMacvlanLink(&linkAttrs, netlink.MACVLAN_MODE_BRIDGE, isMacvtap)
	if err != nil {
		return nil, err
	}

	if err = netlink.LinkAdd(macvlanLink); err != nil {
		return nil, fmt.Errorf("failed to create %s: %v", devStr, err)
	}

	err = netlink.LinkSetNsFd(macvlanLink, int(containerNetns.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to re-nns %s named %q: %v", devStr, ifName, err)
	}

	err = containerNetns.Do(func(_ ns.NetNS) error {
		// Re-fetch macvlan to get all properties/attributes
		contMacvlan, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("failed to refetch %s named %q: %v", devStr, ifName, err)
		}
		debug.Logf("find %s(if=%q, mac=%q) created in (%q) [virtdev]",
			devStr,
			ifName,
			contMacvlan.Attrs().HardwareAddr,
			containerNetns.Path(),
		)

		if doSetUp {
			if err := netlink.LinkSetUp(contMacvlan); err != nil {
				return fmt.Errorf("failed to set %q UP: %v", ifName, err)
			}
		}

		macvlan.Name = ifName
		macvlan.Mac = contMacvlan.Attrs().HardwareAddr.String()
		macvlan.Sandbox = containerNetns.Path()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return macvlan, nil
}

func createMacvlanLink(linkAttrs *netlink.LinkAttrs, mode netlink.MacvlanMode, isMacvtap bool) (netlink.Link, error) {
	if linkAttrs == nil {
		return nil, fmt.Errorf("createMacvlanLink: linkAttrs can not be nil")
	}

	macvlan := netlink.Macvlan{
		LinkAttrs: *linkAttrs,
		Mode:      mode,
	}

	if isMacvtap {
		return &netlink.Macvtap{
			Macvlan: macvlan,
		}, nil
	} else {
		return &macvlan, nil
	}
}

func CreateIpvlan(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, modeStr string, doSetUp bool) (*current.Interface, error) {
	return createIpvlanIf(masterName, ifName, containerNetns, mtu, modeStr, false, doSetUp)
}

func CreateIpvtap(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, modeStr string, doSetUp bool) (*current.Interface, error) {
	return createIpvlanIf(masterName, ifName, containerNetns, mtu, modeStr, true, doSetUp)
}

func createIpvlanIf(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, modeStr string, isIpvtap bool, doSetUp bool) (*current.Interface, error) {
	ipvlan := &current.Interface{}

	var devStr string
	if isIpvtap {
		devStr = "ipvtap"
	} else {
		devStr = "ipvlan"
	}

	mode, err := GetIPvlanModeFromString(modeStr)
	if err != nil {
		return nil, err
	}

	var masterLink netlink.Link
	masterLink, err = netlink.LinkByName(masterName)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup master %q: %v", masterName, err)
	}

	// due to kernel bug we have to create with tmpName or it might
	// collide with the name on the host and error out
	tmpName, err := ip.RandomVethName()
	if err != nil {
		return nil, err
	}

	linkAttrs := netlink.LinkAttrs{
		MTU:         int(mtu),
		Name:        tmpName,
		ParentIndex: masterLink.Attrs().Index,
		Namespace:   netlink.NsFd(int(containerNetns.Fd())),
	}

	ipvlanLink, err := createIpvlanLink(&linkAttrs, mode, isIpvtap)
	if err != nil {
		return nil, err
	}

	if err = netlink.LinkAdd(ipvlanLink); err != nil {
		return nil, fmt.Errorf("failed to create %s: %v", devStr, err)
	}

	err = containerNetns.Do(func(_ ns.NetNS) error {
		err := ip.RenameLink(tmpName, ifName)
		if err != nil {
			_ = netlink.LinkDel(ipvlanLink)
			return fmt.Errorf("failed to rename %s to %q: %v", devStr, ifName, err)
		}
		ipvlan.Name = ifName

		// Re-fetch ipvlan to get all properties/attributes
		contIpvlan, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("failed to refetch %s named %q: %v", devStr, ifName, err)
		}
		debug.Logf("find %s(if=%q, mac=%q) created in (%q) [virtdev]",
			devStr,
			ifName,
			contIpvlan.Attrs().HardwareAddr,
			containerNetns.Path(),
		)

		if doSetUp {
			if err := netlink.LinkSetUp(contIpvlan); err != nil {
				return fmt.Errorf("failed to set %q UP: %v", ifName, err)
			}
		}

		ipvlan.Mac = contIpvlan.Attrs().HardwareAddr.String()
		ipvlan.Sandbox = containerNetns.Path()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return ipvlan, nil
}

func createIpvlanLink(linkAttrs *netlink.LinkAttrs, mode netlink.IPVlanMode, isIpvtap bool) (netlink.Link, error) {
	if linkAttrs == nil {
		return nil, fmt.Errorf("createIpvlanLink: linkAttrs can not be nil")
	}

	ipvlan := netlink.IPVlan{
		LinkAttrs: *linkAttrs,
		Mode:      mode,
	}

	if isIpvtap {
		return &netlink.IPVtap{
			IPVlan: ipvlan,
		}, nil
	} else {
		return &ipvlan, nil
	}
}

func CreateInternalTap(ifName string, containerNetns ns.NetNS, doSetUp bool) (*current.Interface, error) {
	itapIf := &current.Interface{}
	itapLink, err := netlink.LinkByName(ifName)

	if err != nil {
		return nil, fmt.Errorf("failed to lookup internal tap %q: %v", ifName, err)
	}

	err = netlink.LinkSetNsFd(itapLink, int(containerNetns.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to re-nns internal tap %q: %v", ifName, err)
	}

	err = containerNetns.Do(func(_ ns.NetNS) error {
		// Re-fetch internal tap to get all properties/attributes
		contItap, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("failed to refetch internal tap %q: %v", ifName, err)
		}
		debug.Logf("itap dev (%q) typed (%s) created in (%q)", ifName, contItap.Type(), containerNetns.Path())

		if doSetUp {
			if err := netlink.LinkSetUp(contItap); err != nil {
				return fmt.Errorf("failed to set %q UP: %v", ifName, err)
			}
		}

		itapIf.Name = ifName
		itapIf.Mac = contItap.Attrs().HardwareAddr.String()
		itapIf.Sandbox = containerNetns.Path()
		return nil
	})

	return itapIf, nil
}

func CreateVethPair(hostVethName string, contVethName string, containerNetns ns.NetNS, mtu uint, doDisableRpFilter bool, doSetUp bool) (*current.Interface, *current.Interface, error) {
	hostVethIf := &current.Interface{}
	contVethIf := &current.Interface{}

	hostVeth, contVeth, err := createVethPairLink(hostVethName, contVethName, mtu, doDisableRpFilter)
	if err != nil {
		return nil, nil, fmt.Errorf("fail to create veth pair in host netns from netlink: %v", err)
	}
	defer func() {
		if err != nil {
			if delErr := netlink.LinkDel(hostVeth); delErr != nil {
				debug.Logf("fail to delete incomplete created veth pair %q: %v", hostVethName, delErr)
			}
		}
	}()

	hostVethIf.Name = hostVeth.Attrs().Name
	hostVethIf.Mac = hostVeth.Attrs().HardwareAddr.String()

	if err = netlink.LinkSetNsFd(contVeth, int(containerNetns.Fd())); err != nil {
		return nil, nil, fmt.Errorf("unable to move container veth %q to netns %q: %v", contVeth, containerNetns.Path(), err)
	}

	if doSetUp {
		if err = netlink.LinkSetUp(hostVeth); err != nil {
			return nil, nil, fmt.Errorf("unable to bring up veth pair: %v", err)
		}
	}

	err = containerNetns.Do(func(_ ns.NetNS) error {
		contVethInContNetns, _err := netlink.LinkByName(contVethName)
		if _err != nil {
			return fmt.Errorf("failed to refetch container veth named %q: %v", contVethName, err)
		}
		debug.Logf("find veth(if=%q, mac=%q) created in (%q) [virtdev]",
			contVethName,
			contVethInContNetns.Attrs().HardwareAddr,
			containerNetns.Path(),
		)

		if doSetUp {
			if err := netlink.LinkSetUp(contVethInContNetns); err != nil {
				return fmt.Errorf("failed to set %q UP: %v", contVethName, err)
			}
		}

		contVethIf.Name = contVethInContNetns.Attrs().Name
		contVethIf.Mac = contVethInContNetns.Attrs().HardwareAddr.String()
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	contVethIf.Sandbox = containerNetns.Path()

	return hostVethIf, contVethIf, nil
}

func createVethPairLink(hostVethName string, contVethName string, mtu uint, doDisableRpFilter bool) (netlink.Link, netlink.Link, error) {
	hostVethMac, err := utils.RandMACAddr()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate random host veth mac addr: %s", err)
	}
	contVethMac, err := utils.RandMACAddr()
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate random host container mac addr: %s", err)
	}
	hostVeth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name:         hostVethName,
			MTU:          int(mtu),
			HardwareAddr: net.HardwareAddr(hostVethMac),
		},
		PeerName:         contVethName,
		PeerHardwareAddr: net.HardwareAddr(contVethMac),
	}

	if err := netlink.LinkAdd(hostVeth); err != nil {
		return nil, nil, fmt.Errorf("unable to create veth pair: %s", err)
	}
	defer func() {
		if err != nil {
			if err = netlink.LinkDel(hostVeth); err != nil {
				debug.Logf("unable to clean failed veth pair: %s", err)
			}
		}
	}()

	if doDisableRpFilter {
		/*
			Cilium operation:
				Disable reverse path filter on the host side veth peer to allow
				container addresses to be used as source address when the linux
				stack performs routing.
		*/
		err = disableRpFilter(hostVethName)
		if err != nil {
			return nil, nil, err
		}
	}

	hostVethFind, err := netlink.LinkByName(hostVethName)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to lookup host veth just created in host namespace: %s", err)
	}

	contVethFind, err := netlink.LinkByName(contVethName)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to lookup container veth just created in host namespace: %s", err)
	}

	return hostVethFind, contVethFind, nil
}

// disableRpFilter tries to disable rpfilter on specified interface
func disableRpFilter(ifName string) error {
	return csysctl.Disable(fmt.Sprintf("net.ipv4.conf.%s.rp_filter", ifName))
}

func GetIPvlanModeFromString(s string) (netlink.IPVlanMode, error) {
	switch s {
	case "", "l2":
		return netlink.IPVLAN_MODE_L2, nil
	case "l3":
		return netlink.IPVLAN_MODE_L3, nil
	case "l3s":
		return netlink.IPVLAN_MODE_L3S, nil
	default:
		return 0, fmt.Errorf("unknown ipvlan mode: %q", s)
	}
}

func GetIPvlanModeString(mode netlink.IPVlanMode) (string, error) {
	switch mode {
	case netlink.IPVLAN_MODE_L2:
		return "l2", nil
	case netlink.IPVLAN_MODE_L3:
		return "l3", nil
	case netlink.IPVLAN_MODE_L3S:
		return "l3s", nil
	default:
		return "", fmt.Errorf("unknown ipvlan mode: %d", mode)
	}
}

var (
	randomIfNameMaxLength int = 12
)

func GetRandomIfNameWithPrefix(prefix string) (string, error) {
	if len(prefix) >= randomIfNameMaxLength {
		return "", fmt.Errorf("Generate random ifname with prefix (%q) longer than limit (%d)", prefix, randomIfNameMaxLength)
	}
	return prefix + utils.RandStringBytesRmndr(randomIfNameMaxLength-len(prefix)), nil
}

func GetRandomIfNameWithPrefixWithMaxLength(prefix string, inputMaxLength int) (string, error) {
	if inputMaxLength > randomIfNameMaxLength {
		return "", fmt.Errorf("Generate random ifname with input length (%d) longer than limit (%d)", inputMaxLength, randomIfNameMaxLength)
	}

	if len(prefix) >= inputMaxLength {
		return "", fmt.Errorf("Generate random ifname with prefix (%q) longer than limit (%d)", prefix, randomIfNameMaxLength)
	}
	return prefix + utils.RandStringBytesRmndr(inputMaxLength-len(prefix)), nil
}

func MaskContainerdResultName(result *current.Result, containerdIfName string, ifName string) {
	/*
		Note: Change interface name in result to args.IfName to avoid following error:
		------------------------------------------------------------------
		FATA[0000] run pod sandbox failed: rpc error: code = Unknown desc = failed to setup network for sandbox "94b455ee35ba24884045833fa959606570604cc7ab70868517aefe32b28c62ce": failed to find network info for sandbox "94b455ee35ba24884045833fa959606570604cc7ab70868517aefe32b28c62ce"
		Incorrect Usage: flag provided but not defined: -auth

		NAME:
		crictl start - Start one or more created containers

		USAGE:
		crictl start CONTAINER-ID [CONTAINER-ID...]
		FATA[0000] flag provided but not defined: -auth
		------------------------------------------------------------------
	*/
	for _, resultIf := range result.Interfaces {
		resultIf.Name = containerdIfName
	}
	debug.Logf("result.Interfaces[%d].Name = %q rename from %q ok", len(result.Interfaces), result.Interfaces[0].Name, ifName)
}
