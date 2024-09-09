from ovsdbapp.backend.ovs_idl import connection
from ovsdbapp.backend.ovs_idl import idlutils
from ovsdbapp.schema.ovn_northbound import impl_idl as nb_impl
from ovsdbapp.schema.ovn_southbound import impl_idl as sb_impl

# Connect to OVSDB
ovsdb_connection = connection.OvsdbIdl.from_server("unix:/usr/local/var/run/openvswitch/db.sock", "Open_vSwitch")

# Create a new bridge
bridge_name = "py-br"
bridge = ovsdb_connection.tables['Bridge'].rows
bridge.insert(name=bridge_name, datapath_type = 'netdev')
ovsdb_connection.commit()
