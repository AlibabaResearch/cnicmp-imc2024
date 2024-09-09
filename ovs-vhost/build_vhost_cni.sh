docker run -it --privileged \
    -v ${PWD}/ovs-vhost:/go/src/github.com/ovs-vhost \
    -e GOPATH=/go \
    -w /go/src/github.com/ovs-vhost containerd/build2 sh -c 'echo "**** Building ovs-cni ****" && make build-plugin && echo "**** Complete! ****" && exit'

echo "**** Installing ovs-vhost ****"
cp ./ovs-vhost/cmd/plugin/plugin /opt/cni/bin/ovs-vhost
chmod 777 /opt/cni/bin/ovs-vhost
#systemctl stop containerd
#cp ./kata-containers/src/runtime/containerd-shim-kata-v2 /opt/kata25/bin/containerd-shim-kata-v2
#systemctl start containerd
echo "**** Complete! ****"
