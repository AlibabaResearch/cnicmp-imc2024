cat <<EOF | CNI_COMMAND=ADD CNI_CONTAINERID=ns1 CNI_NETNS=/var/run/netns/ns1 CNI_IFNAME=eth2 CNI_PATH=/opt/cni/bin ./cmd/plugin/plugin
{
  "cniVersion": "0.3.1",
  "name": "containerd-net",
   "type": "ovs",
   "bridge": "br1",
   "vlan": 100,
   "ipam": {
     "type": "host-local",
     "ranges": [
       [{
         "subnet": "10.88.0.0/16"
       }],
       [{
         "subnet": "2001:4860:4860::8888/32"
       }]
     ],
     "routes": [
       { "dst": "0.0.0.0/0" },
       { "dst": "::/0" }
     ]
   }
}
EOF
