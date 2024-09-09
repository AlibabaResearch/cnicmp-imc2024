import os
from ovs.db import idl as ovs_idl
from ovsdbapp.backend.ovs_idl import connection
from ovsdbapp.schema.open_vswitch import impl_idl
import sys
import random
from pyroute2 import IPRoute
import time

src_dir = "/home/cni/openvswitch-2.17.6/"
run_dir = os.getenv("OVS_RUNDIR", "/usr/local/var/run/openvswitch/")
schema_file = os.path.join(src_dir, "vswitchd", "vswitch.ovsschema")
db_sock = os.path.join(run_dir, "db.sock")
remote = f"unix:{db_sock}"

schema_helper = ovs_idl.SchemaHelper(schema_file)
schema_helper.register_all()
idl = ovs_idl.Idl(remote, schema_helper)
apis = []
txns = []

def set_up_connections(con_num, batched = False):
    for cid in range(con_num):
        conn = connection.Connection(idl=idl, timeout=60)
        api = impl_idl.OvsdbIdl(conn)
        apis.append(api)
        if batched:
             txn = api.create_transaction(check_error=True)
             txns.append(txn)

def create_port(br_name, port_name, port_type, tag=100, batched = False):
    assert len(apis) != 0
    cid = random.randint(0, len(apis) - 1)
    #print(cid)
    api = apis[cid]
    if batched: txn = txns[cid]
    else: txn = api.create_transaction(check_error=True)
    if_cmd = txn.add(api.db_create("Interface", name=port_name, type=port_type))
    port_cmd = txn.add(api.db_create("Port", name=port_name, tag=100, interfaces=if_cmd))
    txn.add(api.db_add("Bridge", br_name, "ports", port_cmd))
    if not batched: txn.commit()

def batched_commit():
    assert len(txns) != 0
    for txn in txns:
        txn.commit()


def check_creation(device_num, port_type):
    if port_type == "tap" or port_type == "internal":
        _check_tap(device_num)
    elif port_type == "dpdkvhotuser":
        _check_dpdkvhostuser(device_num)

def get_all_if_names():
    ip = IPRoute()
    interfaces = ip.get_links()
    names = []
    for interface in interfaces:
        name = interface.get_attr("IFLA_IFNAME") 
        if name and len(name) >0: names.append(name)
    return names

def _check_tap(device_num):
    port_exits = [0 for i in range(device_num)]
    while sum(port_exits) < len(port_exits):
        ifs = get_all_if_names()
        for pid in range(device_num):
            if "port{}".format(pid) in ifs: port_exits[pid] = 1
        print(sum(port_exits))
        time.sleep(0.1)
        

def _check_dpdkvhostuser():
    pass

if __name__ == "__main__":
    # list_cmd = api.db_list("Open_vSwitch")
    # txn.add(list_cmd)
    # txn.commit()
    # results = list_cmd.result
    # print(results)
    br_name = sys.argv[1]
    port_name = sys.argv[2]
    port_type = sys.argv[3]
    create_port(br_name, port_name, port_type)
    create_port(br_name, port_name+"2", port_type)

    # with api.transaction(check_error=True) as txn:
    #     #br_cmd = txn.add(api.db_create("Bridge", name="test-br", datapath_type="netdev"))
    #     #txn.add(api.db_add("Open_vSwitch", ".", "bridges", br_cmd))
    #     if_cmd = txn.add(api.db_create("Interface", name="test-tap5", type="tap"))
    #     #port_cmd = txn.add(api.db_create("Port", name="test-tap1", tag=100, interfaces=if_cmd))
    #     #txn.add(api.db_add("Bridge", 'test-br', "ports", port_cmd))
    #     txn.add(api.db_add("Port",  "test-tap1", "interfaces", if_cmd))
    #     #print(br_cmd)
    #     txn.commit()
    #     #print(port_cmd.result)
    #     print(if_cmd.result)

