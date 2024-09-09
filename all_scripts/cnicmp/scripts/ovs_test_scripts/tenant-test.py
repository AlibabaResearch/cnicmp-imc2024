import os 
import sys
import time
current_file_path = os.path.abspath(__file__)
root_dir = os.path.dirname(current_file_path) + "/../../"


cni_type = sys.argv[1]
pod_num = int(sys.argv[2])
user_num = int(sys.argv[3])
acl_num = int(sys.argv[4])
qos = int(sys.argv[5])
rtype = sys.argv[6]

# try:
#     total_lock = int(sys.argv[6])
# except:
#     total_lock = 400

# cni conf format
# 0. clean up
os.system(f"bash {root_dir}/add_br0_netdev.sh")
os.system(f"bash {root_dir}/scripts/clean.sh test")
os.system(f"bash {root_dir}/del_netns.sh")
os.system(f"bash {root_dir}/scripts/clean.sh test")
# 1. generate cni config file
if cni_type == "ovs-tc-routing":
    use_cmd = "false"
else:
    use_cmd = "true"
#conf = conf_fmt.format(cni_type, acl_num, qos, use_cmd)

conf = """{
  "cniVersion": "0.3.1",
  "name": "containerd-net",
  "plugins": [
    {
      "type": \"""" + cni_type +"""\",
      "bridge": "br0",
      "acl": """ + str(acl_num) +""",
      "qos": """ + str(qos) +""",
      "port": "tap0",
      "vlan": 100,
      "mode": "l2",
      "debug": true,
      "cmd": """ + use_cmd +""",
      "rpc": true,
      "tcOvsFlow": false,
      "poolType": "ipvtap",
      "ipam": {
        "type": "host-local",
        "ranges": [
          [{
            "subnet": "10.20.0.0/16"
          }]
        ],
        "routes": [
          { "dst": "0.0.0.0/0" },
          { "dst": "::/0" }
        ]
      }
    }
  ]
}"""
conf_path = "/etc/cni/net.d/10-containerd-net.conflist"

with open(conf_path, "w") as f:
    f.write(conf)

time_test_conf = """
concurency=({})
test_iters=1
""".format(pod_num)

time_test_conf_path = f"bash {root_dir}/scripts/time_test.conf"
with open(time_test_conf_path, "w") as f:
    f.write(time_test_conf)


# 2. start flask&go servers
os.system(f"nohup python3 {root_dir}/scripts/ovs_test_scripts/ovs_server.py {pod_num} {user_num} > flask.log &")
time.sleep(10)
os.system(f"nohup {root_dir}/bin/ovs-server > ovs-server.log &")

# 3. start tests
os.system(f"bash {root_dir}/scripts/time_{rtype}_test_net.sh")

# 4. clean up
os.system(f"bash {root_dir}/add_br0_netdev.sh")
os.system(f"bash {root_dir}/scripts/clean.sh test")
os.system(f"bash {root_dir}/del_netns.sh")
os.system(f"bash {root_dir}/scripts/clean.sh test")
os.system("ovs-vsctl -- --all destroy qos")

# 5. stop servers
os.system("pkill -9 ovs-server")
#os.system("pkill -9 python3")
for i in range(2):
    os.system("kill -9 `ps -ef | grep ovs_server | head -n 1 |  awk '{print $2}'`")
os.system("ifconfig lo 127.0.0.1 up")
