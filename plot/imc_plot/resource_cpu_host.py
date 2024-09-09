import sys
sys.path.append("/home/cni/cnicmp/analysis/nsdi_plot/")
from scapy.all import *
import seaborn as sns
import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
import random
#plt.rcParams['font.sans-serif'] = ['SimHei']
import matplotlib
import plot_colors
matplotlib.rcParams['pdf.fonttype'] = 42
matplotlib.rcParams['ps.fonttype'] = 42
plt.rcParams['font.sans-serif'] = ['Times New Roman']
sns.set(style="whitegrid", font_scale=1.5)


# statistic tools
def _stat_cpu(ts, utils, marker):
    total_cpu_s = 0
    for i in range(1, len(ts)):
        total_cpu_s += (ts[i] - ts[i-1]) * utils[i] / 100.0
    print("{} - cpu time: {} s".format(marker, total_cpu_s))

def _stat_mem():
    pass

def get_cpu_utils(id, marker = "unknown"):
    log_path = "/home/cni/cnicmp_logs/mem_cpu_usage-{}.log".format(id)
    with open(log_path, "r") as f:
        lines = f.readlines()
        ts = []
        utils = []
        for i, line in enumerate(lines):
            # print(line)
            if i == 0 or len(lines) < 2: continue
            tss = int(line.split()[0]) / 1000.0 # in second
            cpu_util = eval(line.split()[1])
            mem = eval(line.split()[2]) * 500
            ts.append(tss)
            utils.append(cpu_util)
    _stat_cpu(ts, utils, marker)
    return np.array(ts) - min(ts), utils
        

#plot
style = ["-", "--", ":", "-.", "--", "-", "--", ":", "-.", "--"]
markers = [".","x", "o","^", "*", ".","x", "o","^", "*",]
labels = ["no-net", "veth", "vhostuser", "ipvlan", "ipvtap",  "tc-routing"]#"macvlan", "macvtap", "tc-routing"]
log_ids = [20230908041002, 20230908021859, 20230907225839, 20230908023350, 20230908023033,  20230908032647] #20230908023642, 20230908032427, 20230908032647]
#log_ids2 = [20230907073151, 20230918064632, , 20230918052714, 20230918051500,20230918045423]
#log_ids3 = [20230907072859, 20230918070334, , 20230918053029, 20230918051741,20230918045423]
# processing raw data
data = []
for i, lid in enumerate(log_ids):
    ts, utils = get_cpu_utils(lid, labels[i])
    data.append([ts, utils])

#data = [ipvlan, ipvtap, macvlan, macvtap, tc_routing]

# plot the figure
plt.figure(figsize=(3.6, 3.2))

edge_width = 1
ax = plt.gca()
ax.spines['top'].set_linewidth(edge_width)
ax.spines['bottom'].set_linewidth(edge_width)
ax.spines['left'].set_linewidth(edge_width)
ax.spines['right'].set_linewidth(edge_width)
ax.spines['top'].set_color("black")
ax.spines['bottom'].set_color("black")
ax.spines['left'].set_color("black")
ax.spines['right'].set_color("black")

#plt.xlabel("CNIs", fontsize = 16)
plt.ylabel("CPU Consumption (%)", fontsize = 14)
plt.xlabel("Time (s)", fontsize = 14)
#plt.xticks([i for i in range(len(user_nums))], [num for num in user_nums], fontsize = 16)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])
#plt.ylim(0, 80)
plt.yticks(fontsize = 14)
plt.xticks(fontsize = 14)
#plt.yticks([0, 20, 40, 60, 80],fontsize = 16)
#plt.yticks([i * 100 for i in range(6)])
colors = list(plot_colors.colors_pack3)#2
colors.insert(1, colors[1])
colors.insert(1, colors[1])
#print(len(colors))
colors = plot_colors.sample_colors2(colors, 2)
#print(colors)
for line_id, line_data in enumerate(data):
   # print(line_id)
    plt.plot(line_data[0], line_data[1], color = colors[line_id], label = labels[line_id], linestyle=style[line_id])#,  marker = markers[line_id], markersize=5)

plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
           mode="expand", borderaxespad=0, ncol=2, frameon=False,
           fontsize = 10.8, handletextpad=0.15)
plt.tight_layout()
plt.savefig("./figs/resource_cpu_host_camera.pdf")
plt.show()