docker run -it --privileged \
    -v /var/lib/containerd \
    -v ${PWD}/containerd:/home/gopath/src/github.com/containerd/containerd \
    -e GOPATH=/go \
    -w /home/gopath/src/github.com/containerd/containerd containerd/build2 bash -c 'echo "**** Building Containerd ****" && export GO111MODULE=off && GOROOT=/usr/go && GOPATH=/home/gopath && export PATH=$GOROOT/bin:$GOPATH/bin:$PATH && make && echo "**** Complete! ****" && exit'

echo "**** Installing Containerd ****"
#systemctl stop containerd
#pkill -9 shim
#cp ./containerd/bin/* /usr/local/bin/
#systemctl start containerd
echo "**** Complete! ****"
