import sys
#sys.path.append("/home/cni/cnicmp/analysis/nsdi_plot/")
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
style = ["-", "--", ":", "-.", "--"]
markers = [".","x", "o","^", "*"]
labels = ["ipvlan", "ipvtap",  "tc-routing"] #"macvlan", "macvtap",

# processing raw data
user_nums = [1, 50, 100, 200]
ipvlan = [
    [23.6],
    [23.7],
    [23.3], # done
    [26.7], 
    [0] # todo
][:len(user_nums)]


ipvtap = [
    [21.5],
    [20.7],
    [21.3],
    [26.7], 
    [0]
][:len(user_nums)]


macvlan = [
    [33.0],
    [35.8],
    [39.0],
    [47.1], 
    [0] # todo
][:len(user_nums)]


macvtap = [
    [21.8],
    [24.3],
    [26.1],
    [32.8], 
    [0]
][:len(user_nums)]


tc_routing = [
    [24.8],
    [24.5],
    [25.1],
    [25.0], 
    [0]
][:len(user_nums)]

data = [ipvlan, ipvtap, tc_routing] # macvlan, macvtap, 

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
plt.xlabel("Number of tenants", fontsize = 14)
plt.ylabel("Time cost (s)", fontsize = 14)
plt.xticks([i for i in range(len(user_nums))], [num for num in user_nums], fontsize = 14)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])
plt.ylim(15, 30)
#plt.yticks([0, 20, 40, 60, 80],fontsize = 14)
plt.xticks(fontsize = 14)
plt.yticks(fontsize = 14)
#plt.yticks([i * 100 for i in range(6)])
colors = plot_colors.colors#plot_colors.get_colors(len(data))
colors = plot_colors.colors_pack3
colors = plot_colors.sample_colors2(colors, 2)
new_colors = [colors[0], colors[3], colors[4]]

#print(colors)
for line_id, line_data in enumerate(data):
   # print(line_id)
    plt.plot([i for i in range(len(line_data))], line_data, marker = markers[line_id], color = new_colors[line_id], label = labels[line_id], linestyle=style[line_id])

plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
           mode="expand", borderaxespad=0, ncol=2, frameon=False,
           fontsize = 12.5, handletextpad=0.15)
plt.tight_layout()
plt.savefig("./figs_fixed/tenant_level_study_qos.pdf")
plt.savefig("./figs_fixed/tenant_level_study_qos.png")
plt.show()