package ovscmd

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/ovs-cni/debug"
	ovs "github.com/containernetworking/plugins/pkg/ovs-cni/server/service"
	"github.com/containernetworking/plugins/pkg/ovs-cni/utils"
	"github.com/vishvananda/netlink"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func execOVSCommand(cmd string, args []string) error {
	stdout, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return fmt.Errorf("CMD err: %s, stdout: %v", err, stdout)
	}
	return nil
}

func CreateInternalVportRPC(brName string, portName string, portMac string, vlanTag uint, mtu uint, doSetUp bool) error {
	conn, err := grpc.Dial("localhost:50505", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("[OVSPortService] fail to connect grpc server: %v", err)
	}
	defer conn.Close()

	client := ovs.NewOVSPortServiceClient(conn)
	request := &ovs.PortRequest{
		BrName:   brName,
		PortName: portName,
		PortMac:  portMac,
		VlanTag:  uint32(vlanTag),
		Mtu:      uint32(mtu),
		DoSetUp:  doSetUp,
	}

	response, err := client.CreateInternalPort(context.Background(), request)
	if err != nil {
		return fmt.Errorf("[OVSPortService] fail to create vport: %v", err)
	} else if !response.Success {
		return fmt.Errorf("[OVSPortService] fail to create vport: reponse false")
	}

	return nil
}

func CreateInternalVport(brName string, portName string, portMac string, vlanTag uint, mtu uint, doSetUp bool) error {
	if portMac != "" {
		return fmt.Errorf("none empty portMac is not supported in CreateInternalVport yet!")
	}
	doPortExist, err := CheckInterfaceExistence(portName, false, "")
	if doPortExist {
		debug.Logf("already find port named %q", portName)
		return nil
	} else {
		if err != nil && strings.Compare(err.Error(), "Link not found") != 0 {
			return fmt.Errorf("failed to check interface existence: %v", err)
		}

		// CMD: ovs-vsctl add-port <brName> <portName> tag=<vlanTag> -- set interface jcovstap0 type=internal mtu_request=<mtu>
		cmd := "ovs-vsctl"
		var args []string
		if vlanTag != 0 {
			args = []string{"add-port", brName, portName, fmt.Sprintf("tag=%d", vlanTag), "--", "set", "interface", portName, "type=internal", fmt.Sprintf("mtu_request=%d", mtu)}
		} else {
			args = []string{"add-port", brName, portName, "--", "set", "interface", portName, "type=internal", fmt.Sprintf("mtu_request=%d", mtu)}
		}

		err = execOVSCommand(cmd, args)
		if err != nil {
			return fmt.Errorf("failed to create ovs internal port: %v", err)
		}
		debug.Logf("create port named %q ok", portName)
	}

	if doSetUp {
		err = SetInterfaceUp(portName, false, "")
		if err != nil {
			return fmt.Errorf("failed to set internal port up: %v", err)
		}
		// TODO: wait port up
	}

	return nil
}

func CheckInterfaceExistence(ifName string, inNNS bool, nns string) (bool, error) {
	var err error
	err = nil
	if inNNS {
		var netns ns.NetNS
		netns, err = ns.GetNS(nns)
		if err != nil {
			return false, fmt.Errorf("failed to open netns %q: %v", netns, err)
		}
		defer netns.Close()

		err = netns.Do(func(_ ns.NetNS) error {
			_, err = netlink.LinkByName(ifName)
			return err
		})
	} else {
		_, err = netlink.LinkByName(ifName)
	}
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func SetInterfaceUp(ifName string, inNNS bool, nns string) error {
	var err error
	cmd := "ip"
	var args []string
	if inNNS {
		args = []string{"netns", "exec", nns, "ip", "link", "set", ifName, "up"}
	} else {
		args = []string{"link", "set", ifName, "up"}
	}

	err = execOVSCommand(cmd, args)
	if err != nil {
		return err
	}

	return nil
}

func AddACLRules(bridgeName string, containerIP string, numACLs int) error {
	randIP := utils.RandIPAddr()

	url := "http://127.0.0.1:5000/get_lock?op=obtain&rand_ip=" + randIP
	_, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("acl http-obtain failed: %v", err)
	}

	for acl_id := 0; acl_id < numACLs; acl_id++ {
		// rand_ip2 := randomIP()
		/*
			example ovs rule:
				ovs-ofctl add-flow br0 ip,nw_src=30.0.0.1/32,nw_dst=30.0.0.2/32,aciton=output:222
		*/
		cmd := "ovs-ofctl"
		cmd_args := []string{"add-flow", bridgeName, "ip,nw_src=" + containerIP + "/32,nw_dst=" + randIP + "/32,actions=drop"}
		if err = exec.Command(cmd, cmd_args...).Run(); err != nil {
			return fmt.Errorf("cmdAdd: failed in adding ovs acl rules%d - %v", acl_id, err)
		}
	}

	url = "http://127.0.0.1:5000/get_lock?op=release"
	_, err = http.Get(url)
	if err != nil {
		return fmt.Errorf("qos http-release failed: %v", err)
	}

	return nil
}
