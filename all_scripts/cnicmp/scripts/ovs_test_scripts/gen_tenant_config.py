import json
 
cni_conf_fmt = {
  "cniVersion": "0.3.1",
  "name": "containerd-net",
  "plugins": [
    {
      "type": "ovs-vhost",
      "bridge": "br0",
      "acl": 0,
      "qos": 0,
      "port": "tap0",
      "vlan": 100,
      "mode": "l2",
      "debug": "true",
      "cmd": "true",
      "ipam": {
        "type": "host-local",
        "ranges": [
          [{
            "subnet": "10.77.0.0/16"
          }]
        ],
        "routes": [
          { "dst": "0.0.0.0/0" },
          { "dst": "::/0" }
        ]
      }
    }
  ]
}

pod_conf_fmt = """metadata:
  name: sandbox-{}
  namespace: test
  uid: busybox-sandbox
  attempt: 1
log_directory: /tmp/pod_test
linux:
  security_context:
    namespace_options:
      network: {}"""


def gen_cni_conf(user_id, cni_type, acl=0, qos=0):
    cni_conf_fmt["plugins"][0]["type"] = cni_type
    cni_conf_fmt["plugins"][0]["port"] = "tap{}".format(user_id)
    cni_conf_fmt["plugins"][0]["acl"] = acl
    cni_conf_fmt["plugins"][0]["acl"] = qos
    json_str = json.dumps(cni_conf_fmt)
    file_path = "/etc/cni/{}-containerd-net.conflist".format(user_id + 1)
    # with open("data.json", "w") as json_file:
    #     json_file.write(json_str)
    with open(file_path, "w") as json_file:
        json.dump(cni_conf_fmt, json_file, indent=2)

def gen_pod_config(user_id, conf_id): # means conf file with conf_id corresponds to user_id
    pass

if __name__ == "__main__":
    # test
    gen_cni_conf(10, "ovs-vhost")