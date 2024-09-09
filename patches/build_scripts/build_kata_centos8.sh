docker run -it --privileged \
    -v ${PWD}/kata-containers:/go/src/github.com/kata-containers \
    -e GOPATH=/go \
    -w /go/src/github.com/kata-containers/src/runtime containerd/build-kata2 sh -c 'echo "**** Building Kata ****" && make && echo "**** Complete! ****" && exit'

