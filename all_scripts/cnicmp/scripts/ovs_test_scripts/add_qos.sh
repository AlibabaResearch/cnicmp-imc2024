#ovs-vsctl set port tap0 qos=@newqos -- --id=@newqos create qos type=egress-policer other-config:cir=46000000 other-config:cbs=2048 
#ovs-vsctl set interface tap0 ingress_policing_rate=368000 ingress_policing_burst=1000

ovs-vsctl set port tap0 qos=@newqos -- --id=@newqos create qos type=linux-htb other-config:max-rate=2000000 queues=123=@q1,456=@q2 -- --id=@q1 create queue other-config:min-rate=1000000 -- --id=@q2 create queue other-config:min-rate=400000 other-config:min-rate=2000000
