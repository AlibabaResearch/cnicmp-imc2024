// Copyright 2018-2020 Red Hat, Intel Corp.
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

package plugin

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	//"github.com/intel/userspace-cni-network-plugin/logging"
)

const defaultOvSSocketDir = "/usr/local/var/run/openvswitch/"

/*
OVS command execution handling and its public interface
*/

type ExecCommandInterface interface {
	execCommand(cmd string, args []string) ([]byte, error)
}

type realExecCommand struct{}

func (e *realExecCommand) execCommand(cmd string, args []string) ([]byte, error) {
	return exec.Command(cmd, args...).Output()
}

var ovsCommand ExecCommandInterface = &realExecCommand{}

func SetExecCommand(o ExecCommandInterface) {
	ovsCommand = o
}

func SetDefaultExecCommand() {
	ovsCommand = &realExecCommand{}
}

func execCommand(cmd string, args []string) ([]byte, error) {
	return ovsCommand.execCommand(cmd, args)
}

/*
Functions to control OVS by using the ovs-vsctl cmdline client.
*/

func createVhostPort(sock_dir string, sock_name string, client bool, bridge_name string) (string, error) {
	var err error

	type_str := "type=dpdkvhostuser"
	if client {
		type_str = "type=dpdkvhostuserclient"
	}

	// COMMAND: ovs-vsctl add-port <bridge_name> <sock_name> -- set Interface <sock_name> type=<dpdkvhostuser|dpdkvhostuserclient>
	cmd := "ovs-vsctl"
	args := []string{"add-port", bridge_name, sock_name, "--", "set", "Interface", sock_name, type_str}

	if client {
		socketarg := "options:vhost-server-path=" + filepath.Join(sock_dir, sock_name)
		//logging.Debugf("Additional string: %s", socketarg)

		args = append(args, socketarg)
	}

	if _, err = execCommand(cmd, args); err != nil {
		return "", err
	}

	if !client {
		// Determine the location OvS uses for Sockets. Default location can be
		// overwritten with environmental variable: OVS_SOCKDIR
		ovs_socket_dir, ok := os.LookupEnv("OVS_SOCKDIR")
		if !ok {
			ovs_socket_dir = defaultOvSSocketDir
		}

		// Move socket to defined dir for easier mounting
		err = os.Rename(filepath.Join(ovs_socket_dir, sock_name), filepath.Join(sock_dir, sock_name))
		if err != nil {
			//_ = logging.Errorf("Rename ERROR: %v", err)
			err = nil

			//deleteVhostPort(sock_name, bridge_name)
		}
	}

	return sock_name, err
}

func deleteVhostPort(sock_name string, bridge_name string) error {
	// COMMAND: ovs-vsctl del-port <bridge_name> <sock_name>
	cmd := "ovs-vsctl"
	args := []string{"--if-exists", "del-port", bridge_name, sock_name}
	_, err := execCommand(cmd, args)
	//logging.Verbosef("ovsctl.deleteVhostPort(): return=%v", err)
	return err
}

func createBridge(bridge_name string) error {
	// COMMAND: ovs-vsctl add-br <bridge_name> -- set bridge <bridge_name> datapath_type=netdev
	cmd := "ovs-vsctl"
	args := []string{"add-br", bridge_name, "--", "set", "bridge", bridge_name, "datapath_type=netdev"}
	_, err := execCommand(cmd, args)
	//logging.Verbosef("ovsctl.createBridge(): return=%v", err)
	return err
}

func configL2Bridge(bridge_name string) error {
	// COMMAND: ovs-ofctl add-flow <bridge_name> actions=NORMAL
	cmd := "ovs-ofctl"
	args := []string{"add-flow", bridge_name, "actions=NORMAL"}
	_, err := execCommand(cmd, args)
	//logging.Verbosef("ovsctl.configL2Bridge(): return=%v", err)
	return err
}

func deleteBridge(bridge_name string) error {
	// COMMAND: ovs-vsctl del-br <bridge_name>
	cmd := "ovs-vsctl"
	args := []string{"del-br", bridge_name}

	_, err := execCommand(cmd, args)
	//logging.Verbosef("ovsctl.deleteBridge(): return=%v", err)
	return err
}

func getVhostPortMac(sock_name string) (string, error) {
	// COMMAND: ovs-vsctl --bare --columns=mac find port name=<sock_name>
	cmd := "ovs-vsctl"
	args := []string{"--bare", "--columns=mac", "find", "port", "name=" + sock_name}
	if mac_b, err := execCommand(cmd, args); err != nil {
		return "", err
	} else {
		return strings.Replace(string(mac_b), "\n", "", -1), nil
	}
}

func findBridge(bridge_name string) bool {
	found := false

	// COMMAND: ovs-vsctl --bare --columns=name find bridge name=<bridge_name>
	cmd := "ovs-vsctl"
	args := []string{"--bare", "--columns=name", "find", "bridge", "name=" + bridge_name}
	//if name, err := execCommand(cmd, args); err != nil {
	name, err := execCommand(cmd, args)
	//logging.Verbosef("ovsctl.findBridge(): return  name=%v err=%v", name, err)
	if err == nil {
		if len(name) != 0 {
			found = true
		}
	}

	return found
}

func doesBridgeContainInterfaces(bridge_name string) bool {
	found := false

	// ovs-vsctl list-ports <bridge_name>
	cmd := "ovs-vsctl"
	args := []string{"list-ports", bridge_name}
	//if name, err := execCommand(cmd, args); err != nil {
	name, err := execCommand(cmd, args)
	//logging.Verbosef("ovsctl.doesBridgeContainInterfaces(): return  name=%v err=%v", name, err)
	if err == nil {
		if len(name) != 0 {
			found = true
		}
	}

	return found
}
