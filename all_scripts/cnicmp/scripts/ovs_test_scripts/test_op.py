from ovs import jsonrpc
from ovs.db import idl
import socket

ss = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
ss.connect("/usr/local/var/run/openvswitch/db.sock")

# 连接到ovsdb服务器

conn = jsonrpc.Connection(ss)#"/usr/local/var/run/openvswitch/db.sock")#('unix:/var/run/openvswitch/db.sock')

# 创建一个IDL对象并设置OVS数据库
idl = idl.Idl(conn)

# 获取“Open_vSwitch”行
row = idl.tables['Open_vSwitch'].rows[uuid.UUID(int=0)]

# 创建一个新的Bridge行
new_bridge = idl.tables['Bridge'].rows.add()
new_bridge.uuid = idl.gen_next_uuid()
new_bridge.name = 'py-test'  # 设置桥名称
new_bridge.datapath_type = 'netdev'  # 设置datapath类型为netdev

# 创建一个新的Port行
new_port = idl.tables['Port'].rows.add()
new_port.uuid = idl.gen_next_uuid()
new_port.name = 'py-tap0'  # 设置端口名称
new_port._parent = new_bridge

# 创建一个新的Interface行
new_iface = idl.tables['Interface'].rows.add()
new_iface.uuid = idl.gen_next_uuid()
new_iface.name = 'py-tap0'  # 设置接口名称
new_iface.type = 'internal'  # 设置接口类型为internal
new_iface._parent = new_port

# 将新的Bridge行与“Open_vSwitch”行关联
new_bridge._parent = row
row.bridges.append(new_bridge)

# 提交更改
idl.commit()
