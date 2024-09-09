import os
import sys

br_name = sys.argv[1]

os.system("ovs-vsctl del-br {}".format(br_name))
os.system("ovs-vsctl -- --all destroy Qos -- --all destroy queue")
os.system("ovs-vsctl add-br {} -- set bridge {} datapath_type=netdev".format(br_name, br_name))
