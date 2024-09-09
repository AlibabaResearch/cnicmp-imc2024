import sys
import numpy as np
import matplotlib.pyplot as plt

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


durs = []
pids = []
start_times = []

with open(file_path, "r") as f:
    cnt = 0
    for line in f.readlines():
        cnt += 1
        if cnt == 1: continue
        if not check_keys(line, keyws): continue
        if "kworker" in line: continue
        if "PID" not in line: continue
        #print(line)
        time_tag = line.split()[4].split("-")
        if line.split()[0] != "spinlock()": 
            print(line)
            continue
        dur = int(time_tag[1]) / 1000000.0 #ms
        pid = int(line.split()[2])
        start_time = int(time_tag[0]) / 1000000.0

        durs.append(dur)
        pids.append(pid)
        start_times.append(start_time)

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
exit()

#plot the figure
start_times = np.array(start_times) - np.min(start_times)
sort_ids = np.argsort(start_times)
start_times = start_times[sort_ids]
pids = np.array(pids)[sort_ids]
durs = np.array(durs)[sort_ids]

plt.figure()
for pid, st in enumerate(start_times):
    plt.plot([st, st+durs[pid]], [1,1])#[pids[pid], pids[pid]])
plt.xlabel("timeline (ms)")
plt.ylabel("pid")
plt.savefig("images/rtnl-timeline.png")


#calculate the overall time
time_entries = []
for pid, st in enumerate(start_times):
    time_entries.append([st, st+durs[pid]])
merged_entries = []
entry = None
for tid in range(len(time_entries)):
    entry = time_entries[tid]
    merged_to_pop = []
    for eid, existing_entry in enumerate(merged_entries):
        if not (entry[1] < existing_entry[0] or entry[0] > existing_entry[1]):
            entry = [min(entry[0], existing_entry[0]), max(entry[1], existing_entry[1])]
            merged_to_pop.append(eid)
    if len(merged_to_pop) > 0:
        merged_entries[merged_to_pop[0]] = entry
        merged_to_pop = merged_to_pop[1:]
        np.delete(merged_entries, merged_to_pop)
    else:
        merged_entries.append(entry)
#print(merged_entries)
sum_t = 0
for entry in merged_entries:
    sum_t += (entry[1] - entry[0])
print(sum_t)
