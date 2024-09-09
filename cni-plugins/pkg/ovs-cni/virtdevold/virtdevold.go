package virtdevold

import (
	"crypto/sha256"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/opencontainers/selinux/go-selinux"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ns"
)

// We want to share the parent process std{in|out|err} - fds 0 through 2.
// Since the FDs are inherited on fork / exec, we close on exec all others.
func closeFileDescriptorsOnExec() {
	minFDToCloseOnExec := 3
	maxFDToCloseOnExec := 256
	for fd := minFDToCloseOnExec; fd < maxFDToCloseOnExec; fd++ {
		syscall.CloseOnExec(fd)
	}
}

// Due to issues with the vishvananda/netlink library (fix pending) it is not possible to create an ownerless/groupless
// tap device. Until the issue is fixed, the workaround for creating a tap device with no owner/group is to use the iptool
func createTapWithIptool(tmpName string, mtu int, multiqueue bool, mac string, owner *uint32, group *uint32) error {
	closeFileDescriptorsOnExec()

	tapDeviceArgs := []string{"tuntap", "add", "mode", "tap", "name", tmpName}
	if multiqueue {
		tapDeviceArgs = append(tapDeviceArgs, "multi_queue")
	}

	if owner != nil {
		tapDeviceArgs = append(tapDeviceArgs, "user", fmt.Sprintf("%d", *owner))
	}
	if group != nil {
		tapDeviceArgs = append(tapDeviceArgs, "group", fmt.Sprintf("%d", *group))
	}
	output, err := exec.Command("ip", tapDeviceArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run command %s: %v", output, err)
	}

	tapDeviceArgs = []string{"link", "set", tmpName}
	if mtu != 0 {
		tapDeviceArgs = append(tapDeviceArgs, "mtu", strconv.Itoa(mtu))
	}
	if mac != "" {
		tapDeviceArgs = append(tapDeviceArgs, "address", mac)
	}
	output, err = exec.Command("ip", tapDeviceArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run command %s: %v", output, err)
	}
	return nil
}

func createLinkWithNetlink(tmpName string, mtu int, nsFd int, multiqueue bool, mac string, owner *uint32, group *uint32) error {
	linkAttrs := netlink.LinkAttrs{
		Name:      tmpName,
		Namespace: netlink.NsFd(nsFd),
	}
	if mtu != 0 {
		linkAttrs.MTU = mtu
	}
	mv := &netlink.Tuntap{
		LinkAttrs: linkAttrs,
		Mode:      netlink.TUNTAP_MODE_TAP,
	}

	if owner != nil {
		mv.Owner = *owner
	}
	if group != nil {
		mv.Group = *group
	}

	if mac != "" {
		addr, err := net.ParseMAC(mac)
		if err != nil {
			return fmt.Errorf("invalid args %v for MAC addr: %v", mac, err)
		}
		linkAttrs.HardwareAddr = addr
	}
	mv.Flags = netlink.TUNTAP_VNET_HDR | unix.IFF_TAP
	if multiqueue {
		mv.Flags = netlink.TUNTAP_MULTI_QUEUE_DEFAULTS | mv.Flags
	}
	if err := netlink.LinkAdd(mv); err != nil {
		return fmt.Errorf("failed to create tap: %v", err)
	}
	return nil
}

func createLink(tmpName string, selinuxContext string, mtu int, multiQueue bool, mac string, owner *uint32, group *uint32, netns ns.NetNS) error {
	switch {
	case selinuxContext != "":
		if err := selinux.SetExecLabel(selinuxContext); err != nil {
			return fmt.Errorf("failed set socket label: %v", err)
		}
		return createTapWithIptool(tmpName, mtu, multiQueue, mac, owner, group)
	case owner == nil || group == nil:
		return createTapWithIptool(tmpName, mtu, multiQueue, mac, owner, group)
	default:
		return createLinkWithNetlink(tmpName, mtu, int(netns.Fd()), multiQueue, mac, owner, group)
	}
}

func CreateTap(bridgeName string, selinuxContext string, mtu int, multiQueue bool, mac string, owner *uint32, group *uint32, ifName string, netns ns.NetNS) (*current.Interface, error) {
	tap := &current.Interface{}
	// due to kernel bug we have to create with tmpName or it might
	// collide with the name on the host and error out
	tmpName, err := ip.RandomVethName()
	if err != nil {
		return nil, err
	}

	err = netns.Do(func(_ ns.NetNS) error {
		err := createLink(tmpName, selinuxContext, mtu, multiQueue, mac, owner, group, netns)
		if err != nil {
			return err
		}

		if err = ip.RenameLink(tmpName, ifName); err != nil {
			link, err := netlink.LinkByName(tmpName)
			if err != nil {
				netlink.LinkDel(link)
				return fmt.Errorf("failed to rename tap to %q: %v", ifName, err)
			}
		}
		tap.Name = ifName

		// Re-fetch link to get all properties/attributes
		link, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("failed to refetch tap %q: %v", ifName, err)
		}

		if bridgeName != "" {
			bridge, err := netlink.LinkByName(bridgeName)
			if err != nil {
				return fmt.Errorf("failed to get bridge %s: %v", bridgeName, err)
			}

			tapDev := link
			if err := netlink.LinkSetMaster(tapDev, bridge); err != nil {
				return fmt.Errorf("failed to set tap %s as a port of bridge %s: %v", tap.Name, bridgeName, err)
			}
		}

		err = netlink.LinkSetUp(link)
		if err != nil {
			return fmt.Errorf("failed to set tap interface up: %v", err)
		}

		tap.Mac = link.Attrs().HardwareAddr.String()
		tap.Sandbox = netns.Path()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return tap, nil
}

func CreateTapInHostNetNS(bridgeName string, selinuxContext string, mtu int, multiQueue bool, mac string, owner *uint32, group *uint32, ifName string, netns ns.NetNS) (*current.Interface, error) {
	tap := &current.Interface{}
	// due to kernel bug we have to create with tmpName or it might
	// collide with the name on the host and error out
	tmpName, err := ip.RandomVethName()
	if err != nil {
		return nil, err
	}

	err = createLink(tmpName, selinuxContext, mtu, multiQueue, mac, owner, group, netns)
	if err != nil {
		return nil, err
	}

	if err = ip.RenameLink(tmpName, ifName); err != nil {
		link, err := netlink.LinkByName(tmpName)
		if err != nil {
			netlink.LinkDel(link)
			return nil, fmt.Errorf("failed to rename tap to %q: %v", ifName, err)
		}
	}
	tap.Name = ifName

	// Re-fetch link to get all properties/attributes
	link, err := netlink.LinkByName(ifName)
	if err != nil {
		return nil, fmt.Errorf("failed to refetch tap %q: %v", ifName, err)
	}

	if bridgeName != "" {
		bridge, err := netlink.LinkByName(bridgeName)
		if err != nil {
			return nil, fmt.Errorf("failed to get bridge %s: %v", bridgeName, err)
		}

		tapDev := link
		if err := netlink.LinkSetMaster(tapDev, bridge); err != nil {
			return nil, fmt.Errorf("failed to set tap %s as a port of bridge %s: %v", tap.Name, bridgeName, err)
		}
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		return nil, fmt.Errorf("failed to set tap interface up: %v", err)
	}

	tap.Mac = link.Attrs().HardwareAddr.String()
	tap.Sandbox = netns.Path()

	return tap, nil
}

func CreateMacvlan(isMasterInContNS bool, masterName string, mtu int, mac string, ifName string, netns ns.NetNS) (*current.Interface, error) {
	macvlan := &current.Interface{}

	var err error
	var masterDev netlink.Link

	if isMasterInContNS {
		err = netns.Do(func(_ ns.NetNS) error {
			masterDev, err = netlink.LinkByName(masterName)
			return err
		})
	} else {
		masterDev, err = netlink.LinkByName(masterName)
	}
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
		MTU:         mtu,
		Name:        tmpName,
		ParentIndex: masterDev.Attrs().Index,
		Namespace:   netlink.NsFd(int(netns.Fd())),
	}

	if mac != "" {
		addr, err := net.ParseMAC(mac)
		if err != nil {
			return nil, fmt.Errorf("invalid args %v for MAC addr: %v", mac, err)
		}
		linkAttrs.HardwareAddr = addr
	}

	// only bridge mode supported
	mv := &netlink.Macvlan{
		LinkAttrs: linkAttrs,
		Mode:      netlink.MACVLAN_MODE_BRIDGE,
	}

	// if isMasterInContNS {
	// 	err = netns.Do(func(_ ns.NetNS) error {
	// 		return netlink.LinkAdd(mv)
	// 	})
	// } else {
	// 	if err = netlink.LinkAdd(mv); err != nil {
	// 		return nil, fmt.Errorf("failed to create macvlan: %v", err)
	// 	}
	// }

	// Macvlan must be created in the containerNetNS
	err = netns.Do(func(_ ns.NetNS) error {
		return netlink.LinkAdd(mv)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create macvlan: %v", err)
	}

	err = netns.Do(func(_ ns.NetNS) error {
		err := ip.RenameLink(tmpName, ifName)
		if err != nil {
			_ = netlink.LinkDel(mv)
			return fmt.Errorf("failed to rename macvlan to %q: %v", ifName, err)
		}
		macvlan.Name = ifName

		// Re-fetch macvlan to get all properties/attributes
		contMacvlan, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("failed to refetch macvlan %q: %v", ifName, err)
		}
		macvlan.Mac = contMacvlan.Attrs().HardwareAddr.String()
		macvlan.Sandbox = netns.Path()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return macvlan, nil
}

// IPAddrToHWAddr takes the four octets of IPv4 address (aa.bb.cc.dd, for example) and uses them in creating
// a MAC address (0A:58:AA:BB:CC:DD).  For IPv6, create a hash from the IPv6 string and use that for MAC Address.
// Assumption: the caller will ensure that an empty net.IP{} will NOT be passed.
// This method is copied from https://github.com/ovn-org/ovn-kubernetes/blob/master/go-controller/pkg/util/net.go
func IPAddrToHWAddr(ip net.IP) net.HardwareAddr {
	// Ensure that for IPv4, we are always working with the IP in 4-byte form.
	ip4 := ip.To4()
	if ip4 != nil {
		// safe to use private MAC prefix: 0A:58
		return net.HardwareAddr{0x0A, 0x58, ip4[0], ip4[1], ip4[2], ip4[3]}
	}

	hash := sha256.Sum256([]byte(ip.String()))
	return net.HardwareAddr{0x0A, 0x58, hash[0], hash[1], hash[2], hash[3]}
}

func AssignMacToLink(link netlink.Link, mac net.HardwareAddr, name string) error {
	err := netlink.LinkSetHardwareAddr(link, mac)
	if err != nil {
		return fmt.Errorf("failed to set container iface %q MAC %q: %v", name, mac.String(), err)
	}
	return nil
}
