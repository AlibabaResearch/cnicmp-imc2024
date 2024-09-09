from ovsapp import OVSBridge

# 创建一个ovs bridge对象
bridge = OVSBridge('br0')

# 向bridge添加flow规则
bridge.add_flow(priority=100, in_port=1, actions='output:2')
bridge.add_flow(priority=100, in_port=2, actions='output:1')

# 查看bridge中的flow规则
flow_rules = bridge.dump_flows()
for flow in flow_rules:
    print(flow)
