// Copyright 2018-2019 Red Hat, Inc.
// Copyright 2014 CNI authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Go version 1.10 or greater is required. Before that, switching namespaces in
// long running processes in go did not work in a reliable way.
//go:build go1.10
// +build go1.10

package plugin

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"time"

	"github.com/containernetworking/cni/pkg/skel"
	cnitypes "github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ip"
	"github.com/containernetworking/plugins/pkg/ipam"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"

	//"github.com/j-keck/arping"

	"github.com/k8snetworkplumbingwg/ovs-cni/pkg/config"
	"github.com/k8snetworkplumbingwg/ovs-cni/pkg/ovsdb"
	"github.com/k8snetworkplumbingwg/ovs-cni/pkg/types"
	"github.com/k8snetworkplumbingwg/ovs-cni/pkg/utils"

	"github.com/k8snetworkplumbingwg/ovs-cni/pkg/cnicmp/cnilogger"
	// "go.uber.org/zap"
	// "go.uber.org/zap/zapcore"
	// "gopkg.in/natefinch/lumberjack.v2"
)

// EnvArgs args containing common, desired mac and ovs port name
type EnvArgs struct {
	cnitypes.CommonArgs
	MAC     cnitypes.UnmarshallableString `json:"mac,omitempty"`
	OvnPort cnitypes.UnmarshallableString `json:"ovnPort,omitempty"`
}

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func logCall(command string, args *skel.CmdArgs) {
	log.Printf("CNI %s was called for container ID: %s, network namespace %s, interface name %s, configuration: %s",
		command, args.ContainerID, args.Netns, args.IfName, string(args.StdinData[:]))
}

func getEnvArgs(envArgsString string) (*EnvArgs, error) {
	if envArgsString != "" {
		e := EnvArgs{}
		err := cnitypes.LoadArgs(envArgsString, &e)
		if err != nil {
			return nil, err
		}
		return &e, nil
	}
	return nil, nil
}

func getHardwareAddr(ifName string) string {
	ifLink, err := netlink.LinkByName(ifName)
	if err != nil {
		return ""
	}
	return ifLink.Attrs().HardwareAddr.String()

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

func setupVeth(contNetns ns.NetNS, contIfaceName string, requestedMac string, mtu int) (*current.Interface, *current.Interface, error) {
	hostIface := &current.Interface{}
	contIface := &current.Interface{}

	// Enter container network namespace and create veth pair inside. Doing
	// this we will make sure that both ends of the veth pair will be removed
	// when the container is gone.
	err := contNetns.Do(func(hostNetns ns.NetNS) error {
		hostVeth, containerVeth, err := ip.SetupVeth(contIfaceName, mtu, requestedMac, hostNetns)
		if err != nil {
			return err
		}

		contIface.Name = containerVeth.Name
		contIface.Mac = containerVeth.HardwareAddr.String()
		contIface.Sandbox = contNetns.Path()
		hostIface.Name = hostVeth.Name
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// Refetch the hostIface since its MAC address may change during network namespace move.
	if err = refetchIface(hostIface); err != nil {
		return nil, nil, err
	}

	return hostIface, contIface, nil
}

func assignMacToLink(link netlink.Link, mac net.HardwareAddr, name string) error {
	err := netlink.LinkSetHardwareAddr(link, mac)
	if err != nil {
		return fmt.Errorf("failed to set container iface %q MAC %q: %v", name, mac.String(), err)
	}
	return nil
}

func getBridgeName(bridgeName, ovnPort string) (string, error) {
	if bridgeName != "" {
		return bridgeName, nil
	} else if bridgeName == "" && ovnPort != "" {
		return "br-int", nil
	}

	return "", fmt.Errorf("failed to get bridge name")
}

func attachIfaceToBridge(ovsDriver *ovsdb.OvsBridgeDriver, hostIfaceName string, contIfaceName string, ofportRequest uint, vlanTag uint, trunks []uint, portType string, intfType string, contNetnsPath string, ovnPortName string) error {
	err := ovsDriver.CreatePort(hostIfaceName, contNetnsPath, contIfaceName, ovnPortName, ofportRequest, vlanTag, trunks, portType, intfType)
	if err != nil {
		return err
	}

	hostLink, err := netlink.LinkByName(hostIfaceName)
	if err != nil {
		return err
	}

	if err := netlink.LinkSetUp(hostLink); err != nil {
		return err
	}

	return nil
}

func refetchIface(iface *current.Interface) error {
	iface.Mac = getHardwareAddr(iface.Name)
	return nil
}

func splitVlanIds(trunks []*types.Trunk) ([]uint, error) {
	vlans := make(map[uint]bool)
	for _, item := range trunks {
		var minID uint = 0
		var maxID uint = 0
		if item.MinID != nil {
			minID = *item.MinID
			if minID > 4096 {
				return nil, errors.New("incorrect trunk minID parameter")
			}
		}
		if item.MaxID != nil {
			maxID = *item.MaxID
			if maxID > 4096 {
				return nil, errors.New("incorrect trunk maxID parameter")
			}
			if maxID < minID {
				return nil, errors.New("minID is greater than maxID in trunk parameter")
			}
		}
		if minID > 0 && maxID > 0 {
			for v := minID; v <= maxID; v++ {
				vlans[v] = true
			}
		}
		var id uint
		if item.ID != nil {
			id = *item.ID
			if minID > 4096 {
				return nil, errors.New("incorrect trunk id parameter")
			}
			vlans[id] = true
		}
	}
	if len(vlans) == 0 {
		return nil, errors.New("trunk parameter is misconfigured")
	}
	vlanIds := make([]uint, 0, len(vlans))
	for k := range vlans {
		vlanIds = append(vlanIds, k)
	}
	sort.Slice(vlanIds, func(i, j int) bool { return vlanIds[i] < vlanIds[j] })
	return vlanIds, nil
}

// utils - genrate random ip
func randomIP() string {
	var ip string
	for i := 0; i < 4; i++ {
		ip += strconv.Itoa(rand.Intn(256))
		if i < 3 {
			ip += "."
		}
	}
	return ip
}
func ObtainStageLock(stageName string) error {
	url := "http://127.0.0.1:5005/get_lock?op=obtain&stage=" + stageName
	_, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("obtain stage lock-%s failed: %v", stageName, err)
	}
	return nil
}

func ReleaseStageLock(stageName string) error {
	url := "http://127.0.0.1:5005/get_lock?op=release&stage=" + stageName
	_, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("release stage lock-%s failed: %v", stageName, err)
	}
	return nil
}

// CmdAdd add handler for attaching container into network
func CmdAdd(args *skel.CmdArgs) error {
	err2 := ObtainStageLock("cniAdd")
	if err2 != nil {
		return err2
	}

	/**** cnicmp ****/
	id := args.ContainerID
	startTCNI := cnilogger.RecordStart()
	/**** cnicmp ****/

	// 获取当前时间戳
	t := time.Now().UnixNano()
	rand.Seed(t)

	logCall("ADD", args)

	envArgs, err := getEnvArgs(args.Args)
	if err != nil {
		return err
	}

	//var mac string
	var ovnPort string
	if envArgs != nil {
		//mac = string(envArgs.MAC)
		ovnPort = string(envArgs.OvnPort)
	}

	netconf, err := config.LoadConf(args.StdinData)
	if err != nil {
		return err
	}

	// var vlanTagNum uint = 0
	// trunks := make([]uint, 0)
	// portType := "access"
	// if netconf.VlanTag == nil || len(netconf.Trunk) > 0 {
	// 	portType = "trunk"
	// 	if len(netconf.Trunk) > 0 {
	// 		trunkVlanIds, err := splitVlanIds(netconf.Trunk)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		trunks = append(trunks, trunkVlanIds...)
	// 	}
	// } else if netconf.VlanTag != nil {
	// 	vlanTagNum = *netconf.VlanTag
	// }

	bridgeName, err := getBridgeName(netconf.BrName, ovnPort)
	if err != nil {
		return err
	}

	ovsDriver, err := ovsdb.NewOvsBridgeDriver(bridgeName, netconf.SocketFile)
	if err != nil {
		return err
	}

	// removes all ports whose interfaces have an error
	if err := cleanPorts(ovsDriver); err != nil {
		return err
	}

	contNetns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer contNetns.Close()

	// var origIfName string
	// if sriov.IsOvsHardwareOffloadEnabled(netconf.DeviceID) {
	// 	origIfName, err = sriov.GetVFLinkName(netconf.DeviceID)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// vhostuser starts here
	result := &current.Result{}
	result.Interfaces = []*current.Interface{{
		Name:    args.IfName,
		Sandbox: contNetns.Path(),
	}}
	// 1. call ipam to assign IP
	if netconf.IPAM.Type != "" {

		// run the IPAM plugin and get back the config to apply
		ipamResult, err := ipam.ExecAdd(netconf.IPAM.Type, args.StdinData)
		defer func() {
			if err != nil {
				if err := ipam.ExecDel(netconf.IPAM.Type, args.StdinData); err != nil {
					log.Printf("Failed best-effort cleanup IPAM configuration: %v", err)
				}
			}
		}()
		if err != nil {
			log.Printf("cmdAdd: IPAM ERROR - %v", err)
			return err
		}

		// Convert whatever the IPAM result was into the current Result type
		newResult, err := current.NewResultFromResult(ipamResult)
		if err != nil {
			// TBD: CLEAN-UP
			log.Printf("cmdAdd: IPAM Result ERROR - %v", err)
			return err
		}

		if len(newResult.IPs) == 0 {
			// TBD: CLEAN-UP
			err = fmt.Errorf("ERROR: Unable to get IP Address")
			log.Printf("cmdAdd: IPAM ERROR - %v", err)
			return err
		}

		newResult.Interfaces = result.Interfaces
		//newResult.Interfaces[0].Mac = macAddr

		// Clear out the Gateway if set by IPAM, not being used.
		for _, ip := range newResult.IPs {
			ip.Gateway = nil
		}

		result = newResult
	}

	// 2. create vhost-user interface

	/* create port with ovsdb api - todo */
	// err = ovsDriver.CreatePort(hostIfaceName, contNetnsPath, contIfaceName, ovnPortName, ofportRequest, vlanTag, trunks, portType, intfType)
	// if err != nil {
	// 	return err
	// }

	/**** cnicmp ****/
	startT := cnilogger.RecordStart()
	/**** cnicmp ****/
	var vhostName string
	sock_name := fmt.Sprintf("v%s", result.IPs[0].Address.IP.String())

	url := "http://127.0.0.1:5000/get_lock?op=obtain"
	_, err = http.Get(url)
	if err != nil {
		return fmt.Errorf("acl http-obtain failed: %v", err)
	}
	vhostName, err = createVhostPort("/usr/local/var/run/openvswitch/", sock_name, false, bridgeName)
	url = "http://127.0.0.1:5000/get_lock?op=release"
	_, err = http.Get(url)
	if err != nil {
		return fmt.Errorf("qos http-release failed: %v", err)
	}

	/**** cnicmp ****/
	cnilogger.RecordEnd(id, "CNI_create_port", startT)
	/**** cnicmp ****/

	// Cache NetConf for CmdDel
	if err = utils.SaveCache(config.GetCRef(args.ContainerID, args.IfName),
		&types.CachedNetConf{Netconf: netconf, OrigIfName: vhostName}); err != nil {
		return fmt.Errorf("error saving NetConf %q", err)
	}

	// 3. create dummy interface and assign IP to it
	err = contNetns.Do(func(_ ns.NetNS) error {

		/**** cnicmp ****/
		startT = cnilogger.RecordStart()
		/**** cnicmp ****/

		cmd := "ip"
		cmd_args := []string{"link", "add", vhostName, "type", "dummy", "arp", "on"}
		if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
			log.Printf("cmdAdd: failed in creating interface - %v", err)
			return err
		}

		cmd_args = []string{"addr", "add", result.IPs[0].Address.String(), "dev", vhostName}
		if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
			log.Printf("cmdAdd: failed in assigning IP to dummy interface - %v", err)
			return err
		}

		cmd_args = []string{"link", "set", "dev", vhostName, "up"}
		if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
			log.Printf("cmdAdd: failed in bringing up dummy - %v", err)
			return err
		}

		cmd_args = []string{"link", "set", "dev", vhostName, "arp", "on"}
		if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
			log.Printf("cmdAdd: failed in enabling dummy arp - %v", err)
			return err
		}

		/**** cnicmp ****/
		cnilogger.RecordEnd(id, "CNI_add_dummy", startT)
		/**** cnicmp ****/

		// // ip route add 10.88.0.0/16 dev v10.88.200.27 proto kernel scope link src 10.88.200.27
		// cmd_args = []string{"route", "add", "10.88.0.0/16", "dev", vhostName, "proto", "kernel", "scope", "link", "src", result.IPs[0].Address.IP.String()}
		// if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
		// 	log.Printf("cmdAdd: failed in adding routes1 - %v", err)
		// 	return err
		// }

		// // ip route  add default via 10.88.0.1 dev v10.88.200.28
		// cmd_args = []string{"route", "add", "default", "via", "10.88.0.1", "dev", vhostName}
		// if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
		// 	log.Printf("cmdAdd: failed in adding routes2  - %v", err)
		// 	return err
		// }

		// link, err := netlink.LinkByName(vhostName)
		// if err != nil {
		// 	return fmt.Errorf("failed to lookup %q: %v", vhostName, err)
		// }
		// var v4gw, v6gw net.IP
		// for _, ipc := range result.IPs {
		// 	// if ipc.Interface == nil {
		// 	// 	continue
		// 	// }
		// 	gwIsV4 := ipc.Gateway.To4() != nil
		// 	if gwIsV4 && v4gw == nil {
		// 		v4gw = ipc.Gateway
		// 	} else if !gwIsV4 && v6gw == nil {
		// 		v6gw = ipc.Gateway
		// 	}
		// }
		// for _, r := range result.Routes {
		// 	routeIsV4 := r.Dst.IP.To4() != nil
		// 	gw := r.GW
		// 	if gw == nil {
		// 		if routeIsV4 && v4gw != nil {
		// 			gw = v4gw
		// 		} else if !routeIsV4 && v6gw != nil {
		// 			gw = v6gw
		// 		}
		// 	}
		// 	if gw == nil {
		// 		return fmt.Errorf("failed to add route: no2 gw")
		// 	}
		// 	route := netlink.Route{
		// 		Dst:       &r.Dst,
		// 		LinkIndex: link.Attrs().Index,
		// 		Gw:        gw,
		// 	}

		// 	if err = netlink.RouteAddEcmp(&route); err != nil {
		// 		return fmt.Errorf("failed to add route '%v via %v dev %v': %v", r.Dst, gw, vhostName, err)
		// 	}
		// }

		/**** cnicmp ****/
		//startT_2 := cnilogger.RecordStart()
		/**** cnicmp ****/
		// err = ipam.ConfigureIface(vhostName, result)
		// if err != nil {
		// 	return err
		// }
		// contVeth, err := net.InterfaceByName(vhostName)
		// if err != nil {
		// 	return fmt.Errorf("failed to look up %q: %v", vhostName, err)
		// }
		// /**** cnicmp ****/
		// cnilogger.RecordEnd(id, "CNI_IPAM_assignIP-iface", startT_2)
		// /**** cnicmp ****/

		// /**** cnicmp ****/
		// startT_2 = cnilogger.RecordStart()
		// /**** cnicmp ****/
		// for _, ipc := range result.IPs {
		// 	// if ip address version is 4
		// 	if ipc.Address.IP.To4() != nil {
		// 		// send gratuitous arp for other ends to refresh its arp cache
		// 		err = arping.GratuitousArpOverIface(ipc.Address.IP, *contVeth)
		// 		if err != nil {
		// 			// ok to ignore returning this error
		// 			log.Printf("error sending garp for ip %s: %v", ipc.Address.IP.String(), err)
		// 		}
		// 	}
		// }
		// /**** cnicmp ****/
		// cnilogger.RecordEnd(id, "CNI_IPAM_assignIP-arp", startT_2)
		// /**** cnicmp ****/

		return nil
	})
	if err != nil {
		log.Printf("cmdAdd: dummy interface error - %v", err)
		return err
	}

	// 4. add ovs flow rules to connect the network
	/**** cnicmp ****/
	startT = cnilogger.RecordStart()
	/**** cnicmp ****/

	url = "http://127.0.0.1:5000/get_lock?op=obtain"
	_, err = http.Get(url)
	if err != nil {
		return fmt.Errorf("acl http-obtain failed: %v", err)
	}

	cmd := "ovs-ofctl"
	cmd_args := []string{"add-flow", bridgeName, "dl_type=0x0800,nw_dst=" + result.IPs[0].Address.IP.String() + ",actions=output:" + vhostName}
	if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
		log.Printf("cmdAdd: failed in adding ovs flow rules - %v", err)
		return err
	}

	// /**** cnicmp ****/
	// cnilogger.RecordEnd(id, "CNI_add_flow_rule", startT)
	// /**** cnicmp ****/

	// 5. add user rules

	//startT = cnilogger.RecordStart()

	if netconf.QOS > 0 {
		cmd := "ovs-vsctl"
		cmd_args := []string{"set", "port", vhostName, "qos=@newqos" + id, "--", "--id=@newqos" + id, "create", "qos", "type=egress-policer", "other-config:cir=46000000", "other-config:cbs=2048"}
		//cmd_args := []string{"/home/cni/cnicmp/scripts/add_qos.sh", hostIface.Name, id}
		if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
			return fmt.Errorf("cmdAdd: failed in adding ovs qos rules1:"+"2set port "+vhostName+" qos=@newqos"+id+" -- --id=@newqos"+id+" create qos type=egress-policer other-config:cir=46000000 other-config:cbs=2048"+" - %v", err)
			return err
		}
		cmd_args = []string{"set", "interface", vhostName, "ingress_policing_rate=368000", "ingress_policing_burst=1000"}
		if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
			return fmt.Errorf("cmdAdd: failed in adding ovs qos rules2 - %v", err)
			return err
		}
	}

	for acl_id := 0; acl_id < netconf.ACL; acl_id++ {
		rand_ip := randomIP()
		cmd := "ovs-ofctl"
		//cmd_args := []string{"add-flow", "{}", "in_port={}", "dl_src=00:00:00:00:00:{}{}/00:00:00:00:00:{}{}", "actions=output:{}"}
		//cmd_args := []string{"add-flow", netconf.BrName, "ip,nw_src=30.0.0." + strconv.Itoa(acl_id + 1) + "/32,nw_dst=" + randomIP() + "/32,actions=output:" + vhostName}
		cmd_args := []string{"add-flow", netconf.BrName, "ip,nw_src=30.0.0." + strconv.Itoa(acl_id+1) + "/32,nw_dst=" + rand_ip + "/32,actions=drop"}
		if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
			return fmt.Errorf("cmdAdd: failed in adding ovs acl rules%d - %v", acl_id, err)
			return err
		}
	}

	url = "http://127.0.0.1:5000/get_lock?op=release"
	_, err = http.Get(url)
	if err != nil {
		return fmt.Errorf("qos http-release failed: %v", err)
	}

	cnilogger.RecordEnd(id, "CNI_add_user_rule", startT)

	// 5. wait to ensure the socket is created
	// err = waitSocket("/usr/local/var/run/openvswitch/v" + result.IPs[0].Address.IP.String())
	// if err != nil {
	// 	return err
	// }

	/**** cnicmp ****/
	cnilogger.RecordEnd(id, "CNI", startTCNI)
	/**** cnicmp ****/

	cnilogger.Sync()

	err2 = ReleaseStageLock("cniAdd")
	if err2 != nil {
		return err2
	}

	return cnitypes.PrintResult(result, netconf.CNIVersion)
}

func waitSocket(vhost_sock_path string) error {
	checkInterval := time.Duration(10) * time.Millisecond
	for i := 1; i <= 1000; i++ {
		_, err := os.Stat(vhost_sock_path)
		if err == nil {
			break
		}
		if i == 1000 {
			return fmt.Errorf("Wait for vhost socket %s time out", vhost_sock_path)
		}
		time.Sleep(checkInterval)
	}
	return nil

}

func waitLinkUp(ovsDriver *ovsdb.OvsBridgeDriver, ofPortName string, retryCount, interval int) error {
	checkInterval := time.Duration(interval) * time.Millisecond
	for i := 1; i <= retryCount; i++ {
		portState, err := ovsDriver.GetOFPortOpState(ofPortName)
		if err != nil {
			log.Printf("error in retrieving port %s state: %v", ofPortName, err)
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

func getOvsPortForContIface(ovsDriver *ovsdb.OvsBridgeDriver, contIface string, contNetnsPath string) (string, bool, error) {
	// External IDs were set on the port during ADD call.
	return ovsDriver.GetOvsPortForContIface(contIface, contNetnsPath)
}

// cleanPorts removes all ports whose interfaces have an error.
func cleanPorts(ovsDriver *ovsdb.OvsBridgeDriver) error {
	ifaces, err := ovsDriver.FindInterfacesWithError()
	if err != nil {
		return fmt.Errorf("clean ports: %v", err)
	}
	for _, iface := range ifaces {
		log.Printf("Info: interface %s has error: removing corresponding port", iface)
		if err := ovsDriver.DeletePort(iface); err != nil {
			// Don't return an error here, just log its occurrence.
			// Something else may have removed the port already.
			log.Printf("Error: %v\n", err)
		}
	}
	return nil
}

func removeOvsPort(ovsDriver *ovsdb.OvsBridgeDriver, portName string) error {

	return ovsDriver.DeletePort(portName)
}

// CmdDel remove handler for deleting container from network
func CmdDel(args *skel.CmdArgs) error {
	// logCall("DEL", args)

	// cRef := config.GetCRef(args.ContainerID, args.IfName)
	// cache, err := config.LoadConfFromCache(cRef)
	// if err != nil {
	// 	// If cmdDel() fails, cached netconf is cleaned up by
	// 	// the followed defer call. However, subsequence calls
	// 	// of cmdDel() from kubelet fail in a dead loop due to
	// 	// cached netconf doesn't exist.
	// 	// Return nil when loadConfFromCache fails since the rest
	// 	// of cmdDel() code relies on netconf as input argument
	// 	// and there is no meaning to continue.
	// 	return nil
	// }

	// defer func() {
	// 	if err == nil {
	// 		if err := utils.CleanCache(cRef); err != nil {
	// 			log.Printf("Failed cleaning up cache: %v", err)
	// 		}
	// 	}
	// }()

	// envArgs, err := getEnvArgs(args.Args)
	// if err != nil {
	// 	return err
	// }

	// var ovnPort string
	// if envArgs != nil {
	// 	ovnPort = string(envArgs.OvnPort)
	// }

	// bridgeName, err := getBridgeName(cache.Netconf.BrName, ovnPort)
	// if err != nil {
	// 	return err
	// }

	// ovsDriver, err := ovsdb.NewOvsBridgeDriver(bridgeName, cache.Netconf.SocketFile)
	// if err != nil {
	// 	return err
	// }

	// if cache.Netconf.IPAM.Type != "" {
	// 	err = ipam.ExecDel(cache.Netconf.IPAM.Type, args.StdinData)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// if args.Netns == "" {
	// 	// The CNI_NETNS parameter may be empty according to version 0.4.0
	// 	// of the CNI spec (https://github.com/containernetworking/cni/blob/spec-v0.4.0/SPEC.md).
	// 	if sriov.IsOvsHardwareOffloadEnabled(cache.Netconf.DeviceID) {
	// 		// SR-IOV Case - The sriov device is moved into host network namespace when args.Netns is empty.
	// 		// This happens container is killed due to an error (example: CrashLoopBackOff, OOMKilled)
	// 		var rep string
	// 		if rep, err = sriov.GetNetRepresentor(cache.Netconf.DeviceID); err != nil {
	// 			return err
	// 		}
	// 		if err = removeOvsPort(ovsDriver, rep); err != nil {
	// 			// Don't throw err as delete can be called multiple times because of error in ResetVF and ovs
	// 			// port is already deleted in a previous invocation.
	// 			log.Printf("Error: %v\n", err)
	// 		}
	// 		if err = sriov.ResetVF(args, cache.Netconf.DeviceID, cache.OrigIfName); err != nil {
	// 			return err
	// 		}
	// 	} else {
	// 		// In accordance with the spec we clean up as many resources as possible.
	// 		if err := cleanPorts(ovsDriver); err != nil {
	// 			return err
	// 		}
	// 	}
	// 	return nil
	// }

	// // Unlike veth pair, OVS port will not be automatically removed when
	// // container namespace is gone. Find port matching DEL arguments and remove
	// // it explicitly.
	// portName := cache.OrigIfName
	// if err := removeOvsPort(ovsDriver, portName); err != nil {
	// 	return err
	// }

	// if sriov.IsOvsHardwareOffloadEnabled(cache.Netconf.DeviceID) {
	// 	err = sriov.ReleaseVF(args, cache.OrigIfName)
	// 	if err != nil {
	// 		// try to reset vf into original state as much as possible in case of error
	// 		if err := sriov.ResetVF(args, cache.Netconf.DeviceID, cache.OrigIfName); err != nil {
	// 			log.Printf("Failed best-effort cleanup of VF %s: %v", cache.OrigIfName, err)
	// 		}
	// 	}
	// } else {
	// 	err = ns.WithNetNSPath(args.Netns, func(ns.NetNS) error {
	// 		err = ip.DelLinkByName(portName)
	// 		return err
	// 	})
	// }

	// // removes all ports whose interfaces have an error
	// if err := cleanPorts(ovsDriver); err != nil {
	// 	return err
	// }

	return nil
}

// CmdCheck check handler to make sure networking is as expected.
func CmdCheck(args *skel.CmdArgs) error {
	// logCall("CHECK", args)

	// netconf, err := config.LoadConf(args.StdinData)
	// if err != nil {
	// 	return err
	// }

	// // run the IPAM plugin
	// if netconf.NetConf.IPAM.Type != "" {
	// 	err = ipam.ExecCheck(netconf.NetConf.IPAM.Type, args.StdinData)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to check with IPAM plugin type %q: %v", netconf.NetConf.IPAM.Type, err)
	// 	}
	// }

	// // check cache
	// cRef := config.GetCRef(args.ContainerID, args.IfName)
	// cache, err := config.LoadConfFromCache(cRef)
	// if err != nil {
	// 	return err
	// }
	// if err := validateCache(cache, netconf); err != nil {
	// 	return err
	// }

	// // Parse previous result.
	// if netconf.NetConf.RawPrevResult == nil {
	// 	return fmt.Errorf("Required prevResult missing")
	// }
	// if err := version.ParsePrevResult(&netconf.NetConf); err != nil {
	// 	return err
	// }
	// result, err := current.NewResultFromResult(netconf.NetConf.PrevResult)
	// if err != nil {
	// 	return err
	// }

	// var contIntf, hostIntf current.Interface
	// // Find interfaces
	// for _, intf := range result.Interfaces {
	// 	if args.IfName == intf.Name {
	// 		if args.Netns == intf.Sandbox {
	// 			contIntf = *intf
	// 		}
	// 	} else {
	// 		// Check prevResults for ips against values found in the host
	// 		if err := validateInterface(*intf, true); err != nil {
	// 			return err
	// 		}
	// 		hostIntf = *intf
	// 	}
	// }

	// // The namespace must be the same as what was configured
	// if args.Netns != contIntf.Sandbox {
	// 	return fmt.Errorf("Sandbox in prevResult %s doesn't match configured netns: %s",
	// 		contIntf.Sandbox, args.Netns)
	// }

	// netns, err := ns.GetNS(args.Netns)
	// if err != nil {
	// 	return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	// }
	// defer netns.Close()

	// // Check prevResults for ips and routes against values found in the container
	// if err := netns.Do(func(_ ns.NetNS) error {

	// 	// Check interface against values found in the container
	// 	err := validateInterface(contIntf, false)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = ip.ValidateExpectedInterfaceIPs(args.IfName, result.IPs)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = ip.ValidateExpectedRoute(result.Routes)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return nil
	// }); err != nil {
	// 	return err
	// }

	// // ovs specific check
	// if err := validateOvs(args, netconf, hostIntf.Name); err != nil {
	// 	return err
	// }

	return nil
}

func validateCache(cache *types.CachedNetConf, netconf *types.NetConf) error {
	if cache.Netconf.BrName != netconf.BrName {
		return fmt.Errorf("BrName mismatch. cache=%s,netconf=%s",
			cache.Netconf.BrName, netconf.BrName)
	}

	if cache.Netconf.SocketFile != netconf.SocketFile {
		return fmt.Errorf("SocketFile mismatch. cache=%s,netconf=%s",
			cache.Netconf.SocketFile, netconf.SocketFile)
	}

	if cache.Netconf.IPAM.Type != netconf.IPAM.Type {
		return fmt.Errorf("IPAM mismatch. cache=%s,netconf=%s",
			cache.Netconf.IPAM.Type, netconf.IPAM.Type)
	}

	if cache.Netconf.DeviceID != netconf.DeviceID {
		return fmt.Errorf("DeviceID mismatch. cache=%s,netconf=%s",
			cache.Netconf.DeviceID, netconf.DeviceID)
	}

	return nil
}

func validateInterface(intf current.Interface, isHost bool) error {
	var link netlink.Link
	var err error
	var iftype string
	if isHost {
		iftype = "Host"
	} else {
		iftype = "Container"
	}

	if intf.Name == "" {
		return fmt.Errorf("%s interface name missing in prevResult: %v", iftype, intf.Name)
	}
	link, err = netlink.LinkByName(intf.Name)
	if err != nil {
		return fmt.Errorf("Error: %s Interface name in prevResult: %s not found", iftype, intf.Name)
	}
	if !isHost && intf.Sandbox == "" {
		return fmt.Errorf("Error: %s interface %s should not be in host namespace", iftype, link.Attrs().Name)
	}

	_, isVeth := link.(*netlink.Veth)
	if !isVeth {
		return fmt.Errorf("Error: %s interface %s not of type veth/p2p", iftype, link.Attrs().Name)
	}

	if intf.Mac != "" && intf.Mac != link.Attrs().HardwareAddr.String() {
		return fmt.Errorf("Error: Interface %s Mac %s doesn't match %s Mac: %s", intf.Name, intf.Mac, iftype, link.Attrs().HardwareAddr)
	}

	return nil
}

func validateOvs(args *skel.CmdArgs, netconf *types.NetConf, hostIfname string) error {
	envArgs, err := getEnvArgs(args.Args)
	if err != nil {
		return err
	}
	var ovnPort string
	if envArgs != nil {
		ovnPort = string(envArgs.OvnPort)
	}

	bridgeName, err := getBridgeName(netconf.BrName, ovnPort)
	if err != nil {
		return err
	}

	ovsDriver, err := ovsdb.NewOvsBridgeDriver(bridgeName, netconf.SocketFile)
	if err != nil {
		return err
	}

	found, err := ovsDriver.IsBridgePresent(netconf.BrName)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("Error: bridge %s is not found in OVS", netconf.BrName)
	}

	ifaces, err := ovsDriver.FindInterfacesWithError()
	if err != nil {
		return err
	}
	if len(ifaces) > 0 {
		return fmt.Errorf("Error: There are some interfaces in error state: %v", ifaces)
	}

	vlanMode, tag, trunk, err := ovsDriver.GetOFPortVlanState(hostIfname)
	if err != nil {
		return fmt.Errorf("Error: Failed to retrieve port %s state: %v", hostIfname, err)
	}

	// check vlan tag
	if netconf.VlanTag == nil {
		if tag != nil {
			return fmt.Errorf("vlan tag mismatch. ovs=%d,netconf=nil", *tag)
		}
	} else {
		if tag == nil {
			return fmt.Errorf("vlan tag mismatch. ovs=nil,netconf=%d", *netconf.VlanTag)
		}
		if *tag != *netconf.VlanTag {
			return fmt.Errorf("vlan tag mismatch. ovs=%d,netconf=%d", *tag, *netconf.VlanTag)
		}
		if vlanMode != "access" {
			return fmt.Errorf("vlan mode mismatch. expected=access,real=%s", vlanMode)
		}
	}

	// check trunk
	netconfTrunks := make([]uint, 0)
	if len(netconf.Trunk) > 0 {
		trunkVlanIds, err := splitVlanIds(netconf.Trunk)
		if err != nil {
			return err
		}
		netconfTrunks = append(netconfTrunks, trunkVlanIds...)
	}
	if len(trunk) != len(netconfTrunks) {
		return fmt.Errorf("trunk mismatch. ovs=%v,netconf=%v", trunk, netconfTrunks)
	}
	if len(netconfTrunks) > 0 {
		for i := 0; i < len(trunk); i++ {
			if trunk[i] != netconfTrunks[i] {
				return fmt.Errorf("trunk mismatch. ovs=%v,netconf=%v", trunk, netconfTrunks)
			}
		}

		if vlanMode != "trunk" {
			return fmt.Errorf("vlan mode mismatch. expected=trunk,real=%s", vlanMode)
		}
	}

	return nil
}
