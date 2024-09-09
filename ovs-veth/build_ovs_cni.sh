docker run -it --privileged \
    -v ${PWD}/ovs-cni:/go/src/github.com/ovs-cni \
    -e GOPATH=/go \
    -w /go/src/github.com/ovs-cni containerd/build-ovs-cni sh -c 'echo "**** Building ovs-cni ****" && make build-plugin && echo "**** Complete! ****" && exit'

echo "**** Installing ovs-cni ****"
cp ./ovs-cni/cmd/plugin/plugin /opt/cni/bin/ovs
chmod 777 /opt/cni/bin/ovs
#systemctl stop containerd
#cp ./kata-containers/src/runtime/containerd-shim-kata-v2 /opt/kata25/bin/containerd-shim-kata-v2
#systemctl start containerd
echo "**** Complete! ****"
