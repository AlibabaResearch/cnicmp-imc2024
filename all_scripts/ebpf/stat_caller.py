import sys
import numpy as np

file_path = sys.argv[1]
keyws = ["Caller"]

def check_keys(line, keyws):
    #return True
    matched = False
    for k in keyws:
        if k in line: 
            matched = True
            break
    return matched


pod_cost = {}
cost_t = []
caller_stats = {}
with open(file_path, "r") as f:
    cnt = 0
    for line in f.readlines():
        cnt += 1
        if cnt == 1: continue
        if not check_keys(line, keyws): continue
        #print(line)
        time_tag = line.split()[4].split("-")
        dur = int(time_tag[1]) / 1000000.0 #ms
        pid = int(line.split()[2])
        start_time = int(time_tag[0]) / 1000000.0

        caller = line.split()[-1]
        proc = line.split()[-5]

        if "kworker" in proc: continue
        if "rund" in proc: proc = "rund"

        if proc not in caller_stats.keys(): caller_stats[proc] = {}

        if caller in caller_stats[proc].keys():
            caller_stats[proc][caller] += 1
        else:
            caller_stats[proc][caller] = 1

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

for proc in caller_stats.keys():
    print(proc, ":", caller_stats[proc])
