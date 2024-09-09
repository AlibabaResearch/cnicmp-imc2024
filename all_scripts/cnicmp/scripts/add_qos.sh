ovs-vsctl set port $1 qos=@newqos$2 -- --id=@newqos$2 create qos type=egress-policer other-config:cir=46000000 other-config:cbs=2048 >> /home/cni/cnicmp/log 2>&1
