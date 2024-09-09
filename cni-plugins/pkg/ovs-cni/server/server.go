package main

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/containernetworking/plugins/pkg/ovs-cni/debug"
	"github.com/containernetworking/plugins/pkg/ovs-cni/ovscmd"
	ovs "github.com/containernetworking/plugins/pkg/ovs-cni/server/service"
	"google.golang.org/grpc"
)

/*
	run the following cmd at this directory to generate grpc file:
		protoc --go_out=plugins=grpc:. service.proto
*/

type OVSPortService struct {
	ovs.UnimplementedOVSPortServiceServer
	portRequests      map[string]bool
	portRequestLock   sync.Mutex
	portCreationLocks map[string]*sync.Mutex
	addr              string
}

func (s *OVSPortService) CreateInternalPort(ctx context.Context, req *ovs.PortRequest) (*ovs.PortResponse, error) {
	/*
		For now, only one bridge is supported...
	*/
	brName := req.GetBrName()
	portName := req.GetPortName()
	portMac := req.GetPortMac()
	vlanTag := uint(req.GetVlanTag())
	mtu := uint(req.GetMtu())
	doSetUp := req.GetDoSetUp()

	creationSuccess := false
	creationAuthority := s.acquirePortCreationAuthority(portName)

	if creationAuthority {
		/* has authority: create internal port [portName] under Lock protection */
		err := ovscmd.CreateInternalVport(brName, portName, portMac, vlanTag, mtu, doSetUp)
		if err == nil {
			creationSuccess = true
			debug.Logf("[OVSPortService] creating port (%q) on bridge (%q) ok", portName, brName)
		} else {
			debug.Logf("[OVSPortService] creating port (%q) on bridge (%q) fails", portName, brName)
		}
		s.portCreationLocks[portName].Unlock()
	} else {
		/* no authority: wait until internal port [portName] is created... */
		s.portCreationLocks[portName].Lock()
		creationSuccess = true
		debug.Logf("[OVSPortService] using created port (%q) on bridge (%q) ok", portName, brName)
		s.portCreationLocks[portName].Unlock()
	}

	return &ovs.PortResponse{Success: creationSuccess}, nil
}

func (s *OVSPortService) acquirePortCreationAuthority(portName string) bool {
	s.portRequestLock.Lock()
	defer s.portRequestLock.Unlock()
	if _, exists := s.portRequests[portName]; exists {
		debug.Logf("[OVSPortService] port (%q) is being creating or created...", portName)
		return false
	} else {
		debug.Logf("[OVSPortService] port (%q) creation request authorized...", portName)
		s.portRequests[portName] = true
		s.portCreationLocks[portName] = &sync.Mutex{}
		s.portCreationLocks[portName].Lock()
		return true
	}
}

func (s *OVSPortService) startListen() (net.Listener, error) {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		debug.Logf("[OVSPortService] fails to listen: %v", err)
	}
	return listener, err
}

func main() {
	portService := &OVSPortService{
		portRequests:      make(map[string]bool),
		portCreationLocks: make(map[string]*sync.Mutex),
		addr:              "localhost:50505",
	}

	listener, err := portService.startListen()
	if err == nil {
		grpcServer := grpc.NewServer()
		ovs.RegisterOVSPortServiceServer(grpcServer, portService)
		debug.Logf("[OVSPortService] listening at %q...", portService.addr)
		fmt.Printf("[OVSPortService] listening at %q...\n", portService.addr)
		if err = grpcServer.Serve(listener); err != nil {
			debug.Logf("[OVSPortService] fails to serve grpc: %v", err)
		}
	} else {
		debug.Logf("[OVSPortService] fails to listen at %q: %v", portService.addr, err)
	}
}
