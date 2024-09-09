import sys
import numpy as np

file_path = sys.argv[1]
keyws = ["kata", "k8s"]

def check_keys(line, keyws):
    return True
    matched = False
    for k in keyws:
        if k in line: 
            matched = True
            break
    return matched


pod_cost = {}
cost_t = []
with open(file_path, "r") as f:
    cnt = 0
    for line in f.readlines():
        cnt += 1
        if cnt == 1: continue
        if not check_keys(line, keyws): continue
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
print(np.cumsum(pod_cost_sum))
print(pod_cost_sum)
print("Pod avg time: {} ({} entries in total)".format(np.mean(pod_cost_sum), len(pod_cost_sum)))
print("Pod total time:{}".format(np.sum(pod_cost_sum)))
print("Single subsystem avg time: {} ( {} entries)".format(np.mean(cost_t), len(cost_t)))