ip tuntap add tap0 mode tap
ip addr add 10.88.0.1/16 dev tap0
ip link set tap0 up
