ovs-ofctl add-flow br0 priority=100,dl_type=0x0806,actions=FLOOD
ovs-ofctl add-flow br0 dl_type=0x0800,nw_dst=$1,actions=output:$2
ovs-ofctl add-flow br0 dl_type=0x0800,nw_dst=$3,actions=output:$4
