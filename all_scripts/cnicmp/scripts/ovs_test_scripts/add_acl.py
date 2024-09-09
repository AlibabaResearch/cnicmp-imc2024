import os
import sys

br_name = sys.argv[1]
acl_num = eval(sys.argv[2])

os.system("ovs-vsctl add-br {} -- set bridge {} datapath_type=netdev".format(br_name, br_name))
for acl_id in range(acl_num):
        ovs_pid = 3 #use a fixed port
        cmd = "ovs-ofctl add-flow {} in_port={},dl_src=00:00:00:00:{}{}:{}{}/00:00:00:00:00:{}{},actions=output:{}".format(br_name, ovs_pid, int(acl_id/100000)%10, int(acl_id/10000)%10, int(acl_id/1000)%10, int(acl_id/100)%10, int(acl_id/10)%10, acl_id%10, ovs_pid)
        cmd2 = "ovs-ofctl add-flow {} in_port={},dl_src=00:00:00:00:00:01/00:00:00:00:00:99,actions=output:{}".format(br_name, int(acl_id/10000), acl_id%10000)
        pid = int(acl_id/2)
        cmd3 = "ovs-ofctl add-flow {} in_port={},dl_src=00:00:00:00:00:{}1/00:00:00:00:00:99,actions=output:{}".format(br_name, pid, acl_id%2, pid)
        print(cmd3)
        os.system(cmd3)
