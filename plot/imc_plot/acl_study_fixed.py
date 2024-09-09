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

#plot
style = ["-", "--", "-.", "-", "-.", "--", ":", "-"]
markers = [".","x", "o","^", "*", ">", "<", "o"]
labels = ["veth", "vhostuser", "ipvlan", "ipvtap", "tc-routing"]
    
# processing raw data
acl_num = [0, 10, 50, 100]
veth = [24.3, 44.8, 87.0, 135.8]
vhostuser = [81.5, 83.8, 127.8, 221.8]
ipvlan = [23.6, 23.4, 25.6, 33.6]
ipvtap = [21.5, 19.0, 24.7, 33.4]
#macvlan = [33.0, 34.1, 34.5, 34.1]
#macvtap = [21.8, 20.5, 25.3, 34.0]
tc_routing = [24.8, 24.9, 26.5, 36.0]
#tc_routing_ovs = [52.4, 51.2, 53.6, 50.3] 

data = [veth, vhostuser, ipvlan, ipvtap, tc_routing] #, macvlan, macvtap, tc_routing]
#highlight_id = [d.index(min(d)) for d in data]

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
plt.xlabel("Security group rules/pod", fontsize = 14)
plt.ylabel("Time cost (s)", fontsize = 14)
plt.xticks([i for i in range(len(acl_num))], [num for num in acl_num], fontsize = 14)
plt.yticks(fontsize = 14)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])
#plt.ylim(0, 250)
#plt.yticks([i * 100 for i in range(6)])
colors = plot_colors.colors#plot_colors.get_colors(len(data))
#print(colors)
colors = plot_colors.colors_pack3#2
colors = plot_colors.sample_colors2(colors, 2)
#colors = plot_colors.sample_colors(colors, len(colors))
for line_id, line_data in enumerate(data):
   # print(line_id)
    plt.plot([i for i in range(len(line_data))], line_data, marker = markers[line_id], color = colors[line_id], label = labels[line_id], linestyle=style[line_id])
    #plt.scatter([highlight_id[line_id]], [line_data[highlight_id[line_id]]], color="red", marker="o", s=60)

#plt.text(1-0.05, tc_routing[-1]- 10, "X", color="red")

plt.yscale("log")
plt.yticks([10, 100])
plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
           mode="expand", borderaxespad=0, ncol=2, frameon=False,
           fontsize = 10.8, handletextpad=0.15)
plt.tight_layout()
plt.savefig("./figs_fixed/acl_study.pdf")
plt.savefig("./figs_fixed/acl_study.png")
plt.show()