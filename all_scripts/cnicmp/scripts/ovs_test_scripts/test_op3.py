import os
from ovs.db import idl as ovs_idl
from ovsdbapp.backend.ovs_idl import connection
from ovsdbapp.schema.open_vswitch import impl_idl

src_dir = "/home/cni/openvswitch-2.17.6/"
run_dir = os.getenv("OVS_RUNDIR", "/usr/local/var/run/openvswitch/")
schema_file = os.path.join(src_dir, "vswitchd", "vswitch.ovsschema")
db_sock = os.path.join(run_dir, "db.sock")
remote = f"unix:{db_sock}"

schema_helper = ovs_idl.SchemaHelper(schema_file)
schema_helper.register_all()
idl = ovs_idl.Idl(remote, schema_helper)
conn = connection.Connection(idl=idl, timeout=60)

api = impl_idl.OvsdbIdl(conn)


txn = api.create_transaction(check_error=True)
list_cmd = api.db_list("Open_vSwitch")
txn.add(list_cmd)
txn.commit()
results = list_cmd.result

print(results)


def create_port(br_name, port_name, port_type, tag=100):
    if_cmd = txn.add(api.db_create("Interface", name=port_name, type=port_type))
    port_cmd = txn.add(api.db_create("Port", name=port_name, tag=100, interfaces=if_cmd))
    txn.add(api.db_add("Bridge", br_name, "ports", port_cmd))
    txn.commit()

create_port("test-br", "test-tap6", "dpdkvhostuser")

exit()

with api.transaction(check_error=True) as txn:
    #br_cmd = txn.add(api.db_create("Bridge", name="test-br", datapath_type="netdev"))
    #txn.add(api.db_add("Open_vSwitch", ".", "bridges", br_cmd))
    if_cmd = txn.add(api.db_create("Interface", name="test-tap5", type="tap"))
    #port_cmd = txn.add(api.db_create("Port", name="test-tap1", tag=100, interfaces=if_cmd))
    #txn.add(api.db_add("Bridge", 'test-br', "ports", port_cmd))
    txn.add(api.db_add("Port",  "test-tap1", "interfaces", if_cmd))
    #print(br_cmd)
    txn.commit()
    #print(port_cmd.result)
    print(if_cmd.result)

