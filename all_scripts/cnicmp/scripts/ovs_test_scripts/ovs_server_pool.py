import sys
sys.path.append("/home/cni/cnicmp/scripts/ovs_test_scripts") 
from flask import Flask, request
import threading
import random
import os
import time
import sys
import logging
import atexit

app = Flask(__name__)
app.logger.setLevel(logging.ERROR)

# global lock for limiting the number of calls of ovs-vsctl
THREAD_NUM = 60
lock_num = THREAD_NUM
global_lock1 = threading.Lock()
global_lock2 = threading.Lock()
locks = [threading.Lock() for ti in range(THREAD_NUM)]
lock_flag = [0 for i in range(THREAD_NUM)]
release_count = 0
obtain_count = 0
# locks for enforcing a certain level of sequential launch (deprecated, moved to lock server)
SEQ_LOCK_MAX = 100
seq_lock_num = SEQ_LOCK_MAX
seq_lock1 = threading.Lock()
seq_lock2 = threading.Lock()
# barrier for ovs-eth port attach
BARRIER_MAX = 1
barr_num = 0
barr_lock = threading.Lock()

# tenant-level cni test
pod_num = int(sys.argv[1])
user_num = int(sys.argv[2])
user_pod_num = []
ifname_lock = threading.Lock() 
#virtdev_pool_type = sys.argv[3] #vhost ot tap 

def _init_user_assignments():
    global user_pod_num
    user_pod_num = []
    for uid in range(user_num):
        if uid < (pod_num % user_num): user_pod_num.append(int(pod_num / user_num) + 1)
        else: user_pod_num.append(int(pod_num / user_num))
    assert sum(user_pod_num) == pod_num

_init_user_assignments()

def generate_random_ip():
    ip_parts = [str(random.randint(0, 255)) for _ in range(4)]
    ip_address = '.'.join(ip_parts)
    return ip_address

def ovs_add_flow(br_name, port_name, flow_num):
    start_t = time.time() * 1000
    for i in range(flow_num):
        src_ip = generate_random_ip()
        dst_ip = generate_random_ip()
        cmd = "ovs-ofctl add-flow {} ip,nw_src={}/32,nw_dst={}/32,actions=output:{}".format(br_name, src_ip, dst_ip, port_name)
        #print(cmd)
        os.system(cmd)
    wait_t = time.time() * 1000 - start_t
    #print("add flow - {}ms".format(wait_t))
    return "success"

def ovs_add_qos(port_name):
    cmd = "ovs-vsctl set port {} qos=@newqos{} -- --id=@newqos{} create qos type=egress-policer other-config:cir=46000000 other-config:cbs=2048".format(port_name, port_name, port_name)
    os.system(cmd)
    cmd = "ovs-vsctl set interface {} ingress_policing_rate=368000 ingress_policing_burst=1000".format(port_name)
    os.system(cmd)
    return "success"

# wget http://127.0.0.1:5000/?op=flowAdd&num=25&port_name=veth
@app.route('/')
def get_params():
    #return "success"
    start_t = time.time() * 1000
    op_type = request.args.get('op')
    port_name = request.args.get('port_name')
    br_name = request.args.get('br_name')
    # wait for locks
    lock_id = -1
    for lid, lock in enumerate(locks):
        if not lock.locked():
            lock_id = lid
            break
    if lock_id == -1: 
        #print("Oops, collided!")
        lock_id = random.randint(0, THREAD_NUM - 1)
    # print("try obtaining lock ..")
    locks[lock_id].acquire()
    wait_t = time.time() * 1000 - start_t
    #print("wait {} ms".format(wait_t))
    #print("lock sum{} - id{}".format(sum(lock_flag), lock_id))
    lock_flag[lock_id] = 1

    if op_type == "flowAdd":
        flow_num = int(request.args.get('num'))
        info = ovs_add_flow(br_name, port_name, flow_num)
    elif op_type == "qosAdd":
        info = ovs_add_flow(port_name)
    else:
        info = "Unsupported op!"
    #time.sleep(1)
    lock_flag[lock_id] = 0
    #print("release {}".format(lock_id))
    locks[lock_id].release()
    #print(info)
    return "success"


# wget http://127.0.0.1:5000/get_lock?op=obtain
@app.route('/get_lock')
def get_lock():
    global lock_num
    global obtain_count
    global release_count
    op_type = request.args.get('op')
    if op_type == "obtain":
        global_lock1.acquire()
        while lock_num == 0:
            time.sleep(0.1)
        global_lock2.acquire()
        lock_num -= 1
        obtain_count += 1
        print("obtain lock:{} - {}".format(lock_num, obtain_count))
        global_lock2.release()
        global_lock1.release()
    elif op_type == "release":
        global_lock2.acquire()
        lock_num += 1
        release_count += 1
        print("release lock:{} - {}".format(lock_num,release_count))
        global_lock2.release()
    else:
        raise Exception("Unsupported lock op")
    return "success"

# wget http://127.0.0.1:5000/seq_lock?op=obtain
@app.route('/seq_lock')
def seq_lock():
    global seq_lock_num
    global obtain_count
    global release_count
    op_type = request.args.get('op')
    if op_type == "obtain":
        seq_lock1.acquire()
        while seq_lock_num == 0:
            time.sleep(0.1)
        seq_lock2.acquire()
        seq_lock_num-= 1
        obtain_count += 1
        #print("obtain lock:{} - {}".format(seq_lock_num, obtain_count))
        seq_lock2.release()
        seq_lock1.release()
    elif op_type == "release":
        seq_lock2.acquire()
        seq_lock_num += 1
        release_count += 1
        #print("release lock:{} - {}".format(seq_lock_num,release_count))
        seq_lock2.release()
    else:
        raise Exception("Unsupported lock op")
    return "success"

# wget http://127.0.0.1:5000/barrier
@app.route('/barrier')
def barrier():
    global barr_num
    barr_lock.acquire()
    barr_num += 1
    print("barrier num:", barr_num)
    barr_lock.release()
    while barr_num % BARRIER_MAX != 0:
        time.sleep(0.05)
    return "success"

# wget the name for the shared interface
@app.route('/get_ifname')
def get_tap_name():
    # allocate and return an interface name
    name_fmt = "tap{}"
    ifname_lock.acquire()
    for uid, remain_num in enumerate(user_pod_num):
        if remain_num > 0:
            user_pod_num[uid] = remain_num - 1
            result = name_fmt.format(uid)
            #print("Allocate: {}".format(result))
            break
    ifname_lock.release()
    if sum(user_pod_num) == 0:
        #print("Reset user assignments.")
        _init_user_assignments()
    return result


flow_list = []
batch_size = 1000
batch_op_count = 0
flow_list_lock = threading.Lock()
def _ovs_add_flows(flows, br_name, batch_op_count):
    # write all flows to a file
    with open("./flowfile{}".format(batch_op_count), "w") as f:
        for flow in flows:
            f.write(flow)
        f.flush()
    os.system("ovs-ofctl add-flows {} {}".format(br_name, "./flowfile{}".format(batch_op_count)))

# add acl in batch
# http://127.0.0.1:5000/add_acl_inbatch?br_name=bridgeName&podIP=0.0.0.0&num=100
@app.route('/add_acl_inbatch')
def add_acl_inbatch():
    global flow_list
    global batch_op_count
    # allocate and return an interface name
    pod_ip = request.args.get('podIP')
    acl_num = int(request.args.get('num'))
    br_name = request.args.get('br_name')
    #port_name = request.args.get('port_name')
    add_list = None
    flow_list_lock.acquire()
    for fid in range(acl_num):
        src_ip = generate_random_ip()
        flow_list.append("ip,nw_src={}/32,nw_dst={}/32,actions=drop\n".format(src_ip, pod_ip))
    if len(flow_list) > batch_size:
        add_list = list(flow_list)
        flow_list = []
        batch_op_count += 1       
    flow_list_lock.release()
    if add_list is not None:
        print("Add flows-op{}: {}".format(batch_op_count, len(add_list)))
        _ovs_add_flows(add_list, br_name, batch_op_count)
    return "success"



# precreation/pooling of network devices
pool_size = 410
cid_to_id = {} # maintains a mapping from container id to device id
netnss = []
veth_pairs = []
taps = []
ipvtaps = []
vhosts = []
#ips = []
br_name = "br0"
ovs_tap = "tap0"
count = 0
ip_pool = []

def _init_ip_pools():
    global ip_pool
    print("Initing IP pool...")
    ip_fmt = "10.13.{}.{}/16"
    for i in range(pool_size):
        sub1 = int(i/256)
        sub2 = i % 256 + 1
        ip = ip_fmt.format(sub1, sub2)
        ip_pool.append(ip)
    print("Done.")



####
def _init_netns():
    global netnss
    print("Creating netns pool...")
    netns_name_fmt = "pool-netns-{}"
    for i in range(pool_size):
        netns_name = netns_name_fmt.format(i)
        os.system("ip netns add {}".format(netns_name))
        netnss.append(netns_name)
    print("Done.")

def _del_netns():
    global netnss
    print("Deleting netns pool...")
    netns_name_fmt = "pool-netns-{}"
    for i in range(pool_size):
        netns_name = netns_name_fmt.format(i)
        os.system("ip netns delete {}".format(netns_name))
    netnss = []
    print("Done.")
    
def _init_veth(): # used for crossing namespaces
    global veth_pairs
    print("Creating veth pairs...")
    veth_name_fmt = "pool-veth-{}" 
    veth_name_fmt_peer = "pool-veth-{}-peer" 
    for i in range(pool_size):
        veth_name = veth_name_fmt.format(i)
        veth_name_peer = veth_name_fmt_peer.format(i)
        os.system("ip link add {} type veth peer name {}".format(veth_name, veth_name_peer))
        os.system("ip link set {} up".format(veth_name))
        os.system("ip link set {} up".format(veth_name_peer))
        veth_pairs.append((veth_name, veth_name_peer))
    print("Done.")

def _del_veth():
    global veth_pairs
    print("Deleting veth pairs...")
    veth_name_fmt = "pool-veth-{}" 
    veth_name_fmt_peer = "pool-veth-{}-peer" 
    for i in range(pool_size):
        veth_name = veth_name_fmt.format(i)
        veth_name_peer = veth_name_fmt_peer.format(i)
        os.system("ip link delete {}".format(veth_name)) # the peer should be automatically removed with the netns
    veth_pairs = []
    print("Done.")

def _init_taps():
    global taps
    print("Creating taps...")
    tap_name_fmt = "pool-tap{}"
    for i in range(pool_size):
        tap_name = tap_name_fmt.format(i)
        os.system("ip tuntap add {} mode tap".format(tap_name))
        os.system("ip link set {} up".format(tap_name))
        taps.append(tap_name)
    print("Done.")

def _del_taps():
    global taps
    print("Deleting taps...")
    tap_name_fmt = "pool-tap{}"
    for i in range(pool_size):
        tap_name = tap_name_fmt.format(i)
        os.system("ip tuntap del {} mode tap".format(tap_name))
    taps = []
    print("Done.")


### only ipvtaps is used for test for now
def _init_ipvtaps():
    global ipvtaps
    print("Creating ipvtaps...")
    ipvtap_name_fmt = "pool-ipvtap{}"
    for i in range(pool_size):
        ipvtap_name = ipvtap_name_fmt.format(i)
        ovs_tap_name = "tap{}".format(int(i/800))
        os.system("ip link add {} link {} type ipvtap mode l2".format(ipvtap_name, ovs_tap_name))
        os.system("ip link set {} up".format(ipvtap_name))
        ipvtaps.append(ipvtap_name)
    print("Done.")

def _del_ipvtaps():
    global ipvtaps
    print("Deleting ipvtaps...")
    ipvtap_name_fmt = "pool-ipvtap{}"
    for i in range(pool_size):
        ipvtap_name = ipvtap_name_fmt.format(i)
        os.system("ip link delete {}".format(ipvtap_name))
    ipvtaps = []
    print("Done.")


def _init_taps_ovs(): # creating ports on ovs - type: internal tap
    print("Adding tap on ovs...")
    # recreate the bridge
    os.system("ovs-vsctl del-br {}".format(br_name))
    os.system("ovs-vsctl add-br {} -- set bridge {} datapath_type=netdev".format(br_name, br_name))
    for i in range(int(pool_size/800) + 1):
        ovs_tap_name = "tap{}".format(i)
        os.system("ovs-vsctl add-port {} {} -- set Interface {} type=internal".format(br_name, ovs_tap_name, ovs_tap_name))
    print("Done.")
    
# todo: not used for now
def _connect_tc_filter(): # connecting ovs-port and taps using tc filter - only
    tap_name_fmt = "pool-tap{}"
    for i in range(pool_size):
        # config ovs-tap port
        pass

def init_dev_pools():
    # only ipvtaps for now
    _init_ip_pools()
    _init_taps_ovs()
    #_init_ipvtaps()
    _init_taps()

def clean_dev_pools():
    #_del_ipvtaps()
    _del_taps()
    print("clean bridge...")
    os.system("ovs-vsctl del-br {}".format(br_name))

# wget the name for the shared interface
# http://127.0.0.1:5000/get_virtdev?role=fromCNI/kata&cid=xxx
@app.route('/get_virtdev')
def get_virtdev():
    global count
    #virtdev_type = request.args.get('type')
    cid = request.args.get('cid')
    role = request.args.get('role')
    #assert virtdev_type == virtdev_pool_type
    if role == "fromCNI":
        assert cid not in cid_to_id.keys()
        #ip = ip_pool[count]
        cid_to_id[cid] = count
        # if virtdev_type == "tap":
        #     dev_name = taps[count]
        # else:
        #     dev_name = vhosts[count] #dummy interfaces for vhost-user
        dev_name = taps[count]
        #os.system("ip addr add {} dev {}".format(ip, dev_name))
        count += 1
        print("return dev {} to cni".format(dev_name))
    elif role == "fromKATA":
        assert cid in cid_to_id.keys()
        dev_name = taps[cid_to_id[cid]]
        print("return dev {} to kata".format(dev_name))
    return "{},{}".format(dev_name, "tctap")


if __name__ == '__main__':
    init_dev_pools()
    atexit.register(clean_dev_pools)
    app.run(debug=False, threaded=True, processes=1)
