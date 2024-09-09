import os
import sys

marker = sys.argv[1]

if "tc" in marker:
    cmd_list_veth = "ifconfig | grep veth"
    interface_len = len("vetha0c5aa72")
    ifnames = []
    veths = os.popen(cmd_list_veth).readlines()
    for line in veths:
        ifname = line.split()[0][:-1]
        if len(ifname) == interface_len:
            print(ifname)
            ifnames.append(ifname)
    assert(len(ifnames) > 1)
    cmd_tc_arp = "tc filter add dev tap0 parent ffff: protocol arp u32 match u32 0 0 action mirred egress mirror dev {}".format(ifnames[0])
    for ifname in ifnames[1:]:
        cmd_tc_arp += " pipe action mirred egress mirror dev {}".format(ifname)
    print(cmd_tc_arp)
    os.system(cmd_tc_arp)
