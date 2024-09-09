package virtdevcmd

import (
	"fmt"
	"os/exec"
	"path/filepath"

	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/ovs-cni/debug"
	"github.com/vishvananda/netlink"
)

func CreateSubDevIf(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, mac string, devTypeStr string, devModeStr string, doSetUp bool) (*current.Interface, error) {
	if mac != "" {
		return nil, fmt.Errorf("mac configuration is not supported yet!")
	}

	var err error
	containerNetnsName := GetNetNSName(containerNetns)

	if err = CreateSubLink(ifName, masterName, devTypeStr, devModeStr); err != nil {
		return nil, err
	}

	if err = MoveLinkNetNS(ifName, containerNetnsName, devTypeStr); err != nil {
		return nil, err
	}

	return SetupRefetchAndReturnVirtIf(ifName, devTypeStr, containerNetns, containerNetnsName, doSetUp)
}

func CreateMacvlan(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, mac string, doSetUp bool) (*current.Interface, error) {
	devModeStr := "bridge" // only bridge supported!
	return CreateSubDevIf(masterName, ifName, containerNetns, mtu, mac, "macvlan", devModeStr, doSetUp)
}

func CreateMacvtap(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, mac string, doSetUp bool) (*current.Interface, error) {
	devModeStr := "bridge" // only bridge supported!
	return CreateSubDevIf(masterName, ifName, containerNetns, mtu, mac, "macvtap", devModeStr, doSetUp)
}

// deprecated
func createMacvlanIf(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, mac string, isMacvtap bool, doSetUp bool) (*current.Interface, error) {
	var err error
	var devTypeStr string
	if isMacvtap {
		devTypeStr = "macvtap"
	} else {
		devTypeStr = "macvlan"
	}
	devModeStr := "bridge" // only bridge supported!

	containerNetnsName := GetNetNSName(containerNetns)

	if err = CreateSubLink(ifName, masterName, devTypeStr, devModeStr); err != nil {
		return nil, err
	}

	if err = MoveLinkNetNS(ifName, containerNetnsName, devTypeStr); err != nil {
		return nil, err
	}

	return SetupRefetchAndReturnVirtIf(ifName, devTypeStr, containerNetns, containerNetnsName, doSetUp)
}

func CreateIpvlan(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, devModeStr string, doSetUp bool) (*current.Interface, error) {
	return CreateSubDevIf(masterName, ifName, containerNetns, mtu, "", "ipvlan", devModeStr, doSetUp)
}

func CreateIpvtap(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, devModeStr string, doSetUp bool) (*current.Interface, error) {
	return CreateSubDevIf(masterName, ifName, containerNetns, mtu, "", "ipvtap", devModeStr, doSetUp)
}

// deprecated
func createIpvlanIf(masterName string, ifName string, containerNetns ns.NetNS, mtu uint, devModeStr string, isIpvtap bool, doSetUp bool) (*current.Interface, error) {
	var err error
	var devTypeStr string
	if isIpvtap {
		devTypeStr = "ipvtap"
	} else {
		devTypeStr = "ipvlan"
	}

	containerNetnsName := GetNetNSName(containerNetns)

	cmd := "ip"
	cmd_args := []string{"link", "add", ifName, "link", masterName, "type", devTypeStr, "mode", devModeStr}
	if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
		return nil, fmt.Errorf("fail to create %s (%q) on (%q): %v", devTypeStr, ifName, masterName, err)
	}

	cmd_args = []string{"link", "set", "dev", ifName, "netns", containerNetnsName}
	if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
		return nil, fmt.Errorf("fail to move %s (%q) to nns (%q): %v", devTypeStr, ifName, containerNetnsName, err)
	}

	if doSetUp {
		cmd_args = []string{"link", "set", "dev", ifName, "up"}
		if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
			return nil, fmt.Errorf("fail to set %s (%q) up: %v", devTypeStr, ifName, err)
		}
	}

	ipvlan := &current.Interface{}
	err = containerNetns.Do(func(_ ns.NetNS) error {
		// Re-fetch macvlan to get all properties/attributes
		contMacvlan, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("failed to refetch ipvlan %q: %v", ifName, err)
		}
		debug.Logf("macvlan dev (%q) created in (%q) [virtdevcmd]", ifName, containerNetns.Path())

		ipvlan.Name = ifName
		ipvlan.Mac = contMacvlan.Attrs().HardwareAddr.String()
		ipvlan.Sandbox = containerNetns.Path()

		return nil
	})

	return ipvlan, nil
}

func ConfigureIPv4andSetup(ifName string, ipAddr string) error {
	cmd := "ip"
	cmd_args := []string{"addr", "add", ipAddr, "dev", ifName}
	if err := exec.Command(cmd, cmd_args...).Run(); err != nil {
		return fmt.Errorf("fail to assign ip addr (%q) to (%q): %v", ipAddr, ifName, err)
	}

	cmd_args = []string{"link", "set", "dev", ifName, "up"}
	if err := exec.Command(cmd, cmd_args...).Run(); err != nil {
		return fmt.Errorf("fail to set (%q) up: %v", ifName, err)
	}

	cmd_args = []string{"link", "set", "dev", ifName, "arp", "on"}
	if err := exec.Command(cmd, cmd_args...).Run(); err != nil {
		return fmt.Errorf("fail to set (%q) arp on: %v", ifName, err)
	}

	return nil
}

func GetNetNSName(nns ns.NetNS) string {
	name := filepath.Base(nns.Path())
	return name
}

func CreateSubLink(ifName string, masterName string, devTypeStr string, devModeStr string) error {
	cmd_args := []string{"link", "add", ifName, "link", masterName, "type", devTypeStr, "mode", devModeStr}
	if err := exec.Command("ip", cmd_args...).Run(); err != nil {
		return fmt.Errorf("fail to create %s (%q) on (%q): %v", devTypeStr, ifName, masterName, err)
	}
	return nil
}

func MoveLinkNetNS(ifName string, netnsName string, devTypeStr string) error {
	cmd_args := []string{"link", "set", "dev", ifName, "netns", netnsName}
	if err := exec.Command("ip", cmd_args...).Run(); err != nil {
		return fmt.Errorf("fail to move %s (%q) to nns (%q): %v", devTypeStr, ifName, netnsName, err)
	}
	return nil
}

func SetLinkUp(ifName string, devTypeStr string) error {
	cmd_args := []string{"link", "set", "dev", ifName, "up"}
	if err := exec.Command("ip", cmd_args...).Run(); err != nil {
		return fmt.Errorf("fail to set %s (%q) up: %v", devTypeStr, ifName, err)
	}
	return nil
}

func SetupRefetchAndReturnVirtIf(ifName string, devTypeStr string, containerNetns ns.NetNS, containerNetnsName string, doSetUp bool) (*current.Interface, error) {
	virtIf := &current.Interface{}
	err := containerNetns.Do(func(_ ns.NetNS) error {
		if doSetUp {
			if _err := SetLinkUp(ifName, devTypeStr); _err != nil {
				return _err
			}
		}

		// Re-fetch virtIf to get all properties/attributes
		contVirtIf, _err := netlink.LinkByName(ifName)
		if _err != nil {
			return fmt.Errorf("failed to refetch %q(%q): %v", devTypeStr, ifName, _err)
		}
		debug.Logf("find %q(if=%q, mac=%q) created in (%q) [virtdevcmd]",
			devTypeStr,
			ifName,
			contVirtIf.Attrs().HardwareAddr,
			containerNetnsName,
		)

		virtIf.Name = ifName
		virtIf.Mac = contVirtIf.Attrs().HardwareAddr.String()
		virtIf.Sandbox = containerNetns.Path()

		return nil
	})
	if err != nil {
		return nil, err
	}
	return virtIf, nil
}
