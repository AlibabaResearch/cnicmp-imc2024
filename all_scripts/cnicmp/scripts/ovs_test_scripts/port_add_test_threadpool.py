import os
import sys
import threading
import threadpool
import time
import ovsdbMT
import random

logf = ">tmp"

br_name = sys.argv[1]
port_type = sys.argv[2]
port_num = eval(sys.argv[3])
thread_num = eval(sys.argv[4])
allocate_ip = eval(sys.argv[5])
qos_num = eval(sys.argv[6])
acl_num = eval(sys.argv[7])
cmd = sys.argv[8]
use_batch = sys.argv[9]
use_file = sys.argv[10]

assert use_batch in ["batched", "nobatch"]
if use_batch == "batched": batched = True
else: batched = False

assert cmd in ["use_cmd", "no_cmd"]
if cmd == "no_cmd": ovsdbMT.set_up_connections(thread_num, batched = batched)

assert use_file in ["nofile", "flowfile"]
if use_file == "flowfile": flowfile = True
else: flowfile = False

def generate_random_ip():
    ip_parts = [str(random.randint(0, 255)) for _ in range(4)]
    ip_address = '.'.join(ip_parts)
    return ip_address


def get_ovs_pid(pid):
    msg = eval(os.popen("ovs-ofctl show {} | grep port{}".format(br_name, pid)).readlines()[0].split('(')[0])
    return msg

def add_port(pid):
    #print("\n\n {} {} {}".format(br_name, "port{}".format(pid), port_type))
    if cmd == "use_cmd": os.system("ovs-vsctl add-port {} port{} -- set Interface port{} type={}".format(br_name, pid, pid, port_type))
    else: ovsdbMT.create_port(br_name,"port{}".format(pid), port_type, batched = batched)
    if allocate_ip: os.system("ip addr add 192.168.11.{}/24 dev port{}".format(pid%200 + 1, pid))
    
    if not flowfile:
        for acl_id in range(acl_num):
            src_ip = generate_random_ip()
            dst_ip = generate_random_ip()
            os.system("ovs-ofctl add-flow {} ip,nw_src={}/32,nw_dst={}/32,actions=output:{}".format(br_name, src_ip, dst_ip, "port{}".format(pid)))
            continue
            pass
            ovs_pid = pid#get_ovs_pid(pid)
            #os.system("ovs-ofctl add-flow {} in_port={},dl_src=00:00:00:00:00:{}{}/00:00:00:00:00:{}{},actions=output:{}".format(br_name, ovs_pid, int(acl_id/1000)%10, int(acl_id/100)%10, int(acl_id/10)%10, acl_id%10, ovs_pid))
            os.system("ovs-ofctl add-flow {} ip,nw_src=30.0.0.{}/32,nw_dst=30.0.{}.{}/32,actions=output:{}".format(br_name, acl_id + 1, int(ovs_pid / 255) + 1,  ovs_pid % 255 + 1, "port{}".format(ovs_pid)))
    else:
        # write to a flow file first
        with open("./flows-{}".format(pid), "w") as f:
            for acl_id in range(acl_num):
                src_ip = generate_random_ip()
                dst_ip = generate_random_ip()
                f.write("ip,nw_src={}/32,nw_dst={}/32,actions=output:{}\n".format(src_ip, dst_ip, "port{}".format(pid)))
            f.flush()
        os.system("ovs-ofctl add-flows {} {}".format(br_name, "./flows-{}".format(pid)))

    
    for i in range(qos_num):
        #os.system("ovs-vsctl set port port{} qos=@newqos{} -- --id=@newqos{} create qos type=linux-htb other-config:max-rate=2000000 queues=123=@q1{},456=@q2{} -- --id=@q1{} create queue other-config:min-rate=1000000 -- --id=@q2{} create queue other-config:min-rate=400000 other-config:min-rate=2000000 > tmp".format(pid,pid,pid,pid,pid,pid,pid))
        os.system("ovs-vsctl set port port{} qos=@newqos{} -- --id=@newqos{} create qos type=egress-policer other-config:cir=46000000 other-config:cbs=2048 {}".format(pid, pid, pid, logf))
        os.system("ovs-vsctl set interface port{} ingress_policing_rate=368000 ingress_policing_burst=1000 {}".format(pid, logf))

def callback(req, res):
    pass

os.system("ovs-vsctl del-br {}".format(br_name))
os.system("ovs-vsctl -- --all destroy Qos -- --all destroy queue")
os.system("ovs-vsctl add-br {} -- set bridge {} datapath_type=netdev".format(br_name, br_name))
mem_usage_before = eval(os.popen("free | grep Mem | awk '{print $3}'").readlines()[0])
pool = threadpool.ThreadPool(thread_num)
requests_ = threadpool.makeRequests(add_port, range(port_num), callback) 

start_t = time.time()
for req in requests_:
    pool.putRequest(req)
pool.wait()
if batched: ovsdbMT.batched_commit()
ovsdbMT.check_creation(port_num, port_type)
end_t = time.time()
mem_usage_after = eval(os.popen("free | grep Mem | awk '{print $3}'").readlines()[0])
created_port_num = os.popen("ovs-vsctl show | grep Port | wc -l").readlines()
print(created_port_num[0].strip() + " ports created.")
print("cleaning up ...")
#os.system("ovs-vsctl del-br {}".format(br_name))
#os.system("ovs-vsctl -- --all destroy Qos -- --all destroy queue")
end_t2 = time.time()

print("Creating {} {} ports takes {}s in total.".format(port_num, port_type, end_t - start_t))
print("Total mem usage: {}MB".format((mem_usage_after - mem_usage_before) / 1024))
print("Cleanup takes {}s in total.".format(end_t2 - end_t))
