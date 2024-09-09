import sys
import numpy as np

file_path = sys.argv[1]
keyw = "k8s"

pod_cost = {}
cost_t = []
with open(file_path, "r") as f:
    for line in f.readlines():
        if keyw not in line and "overhead" not in line: continue
        #print(line)
        dur = int(line.split()[4]) / 1000000.0 #ms
        pid = int(line.split()[2])
        if pid not in pod_cost.keys():
            pod_cost[pid] = []
        pod_cost[pid].append(dur)
        cost_t.append(dur)

pod_cost_sum = []
for k in pod_cost.keys():
    pod_cost_sum.append(sum(pod_cost[k]))
pod_cost_sum.sort()
print(pod_cost_sum)
print("Pod avg time: {}").format(np.mean(pod_cost_sum))
print("Single subsystem avg time: {} ( {} entries)".format(np.mean(cost_t), len(cost_t)))
