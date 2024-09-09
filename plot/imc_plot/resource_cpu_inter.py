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


def get_cpu_utils(name):
    log_num = 2
    raw_data = {}
    for i in range(log_num):
        log_file = "/home/cni/cnicmp_logs/dataplane_logs_inter_server/{}-res-{}.server".format(name, i+1)
        data = []
        with open(log_file, "r") as f:
            lines = f.readlines()#[:20]
            base_t = int(lines[0].split()[0])
            for line in lines:
                sp = line.split()
                assert len(sp) == 2 or len(sp) == 3
                ts = int(sp[0]) - base_t
                ts_in_sec = int(ts/1000.0)
                if ts_in_sec not in raw_data:
                    raw_data[ts_in_sec] = []
                raw_data[ts_in_sec].append(eval(sp[-2]))
                #data.append(eval(sp[-2]))
        #raw_data.append(list(data))
    res = []
    keys = list(raw_data.keys())
    keys.sort()
    assert len(keys) == keys[-1] + 1
    for k in keys:
        res.append(np.mean(raw_data[k]))
    return keys, res#np.mean(raw_data, axis=0)
        

#plot
style = ["-", "--", ":", "-.", "--", "-", "--", ":", "-.", "--"]
markers = [".","x", "o","^", "*", ".","x", "o","^", "*",]
labels = ["veth", "vhostuser", "ipvlan", "ipvtap", "tc-routing"]
names  = ["veth-kernel", "vhost", "ipvlan-kernel", "ipvlan-kernel",  "tc-kernel"] #"macvlan", "macvtap",
# processing raw data
data = []
for name in names:
    sec, line = get_cpu_utils(name)
    for i in range(13 - len(sec)):
        sec.append(len(sec) + i)
        line.append(0)
    print(name, sec[-1], line[-1])
    data.append([sec, line])
#data.append([])
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
plt.yticks(fontsize = 14)
plt.xticks(fontsize = 14)
#plt.xticks([i for i in range(len(user_nums))], [num for num in user_nums], fontsize = 16)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])
plt.ylim(0, 22)
#plt.yticks([0, 20, 40, 60, 80],fontsize = 16)
#plt.yticks([i * 100 for i in range(6)])
colors = plot_colors.colors_pack3#2
colors = plot_colors.sample_colors2(colors, 2)
#print(colors)
for line_id, line_data in enumerate(data):
    print(line_id)
    plt.plot(line_data[0], line_data[1], marker = markers[line_id], color = colors[line_id], label = labels[line_id], linestyle=style[line_id])

plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
           mode="expand", borderaxespad=0, ncol=2, frameon=False,
           fontsize = 11.5, handletextpad=0.15)
plt.tight_layout()
plt.savefig("./figs/resource_cpu_inter_camera.pdf")
plt.show()