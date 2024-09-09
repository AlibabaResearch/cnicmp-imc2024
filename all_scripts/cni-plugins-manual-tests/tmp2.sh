tc qdisc add dev tc-tap handle ffff: ingress
tc filter add dev tc-tap parent ffff: protocol all u32 match ip dst 10.1.12.1 action mirred egress redirect dev veth-tap-a0
tc filter add dev tc-tap parent ffff: protocol all u32 match ip dst 10.1.12.2 action mirred egress redirect dev veth-tap-b0
tc filter add dev tc-tap parent ffff: protocol arp u32 match u32 0 0 action mirred egress mirror dev veth-tap-b0 pipe action mirred egress mirror dev veth-tap-a0 
tc qdisc add dev veth-tap-a0 handle ffff: ingress
tc filter add dev veth-tap-a0 parent ffff: protocol all u32 match u8 0 0 action mirred ingress redirect dev tc-tap
tc qdisc add dev veth-tap-b0 handle ffff: ingress
tc filter add dev veth-tap-b0 parent ffff: protocol all u32 match u8 0 0 action mirred ingress redirect dev tc-tap
#ip netns exec netns-tc-b arp -s 10.1.12.1 b2:1c:5a:14:8e:e2 -i veth-tap-b1
#ip netns exec netns-tc-a arp -s 10.1.12.2 a6:d8:07:70:f4:8a -i veth-tap-a1
