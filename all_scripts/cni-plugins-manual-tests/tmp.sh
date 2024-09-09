ovs-vsctl add-br br-tc -- set bridge br-tc datapath_type=netdev
ovs-vsctl add-port br-tc tc-tap -- set Interface tc-tap type=tap
ip link add veth-tap-a0 type veth peer name veth-tap-a1
ip link add veth-tap-b0 type veth peer name veth-tap-b1
ip netns add netns-tc-a
ip netns add netns-tc-b
ip link set veth-tap-a1 netns netns-tc-a
ip link set veth-tap-b1 netns netns-tc-b
ip netns exec netns-tc-a ip addr add 10.1.12.1/24 dev veth-tap-a1
ip netns exec netns-tc-b ip addr add 10.1.12.2/24 dev veth-tap-b1
ip netns exec netns-tc-a ip link set veth-tap-a1 up
ip netns exec netns-tc-b ip link set veth-tap-b1 up
ip link set veth-tap-a0 up
ip link set veth-tap-b0 up
ip link set tc-tap up
ip addr add 10.1.12.3/24 dev tc-tap
