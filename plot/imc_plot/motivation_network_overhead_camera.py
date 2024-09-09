import sys
import os
current_file_path = os.path.abspath(__file__)
current_dir = os.path.dirname(current_file_path)
sys.path.append(current_dir)
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
style = ["-", "--", ":", "-.", "-", "--", ":", "-."]
markers = ["o","o", "^","^", "x", "x", "o", "x"]
labels = ["Firecracker", "Firecracker-net", "Kata", "Kata-net","RunD", "RunD-net"] #"RunD(O)", "RunD(O)-net",

# processing raw data
concurrency = [1, 20, 40, 100, 200, 400]
#updates-0122
kata_qemu = [
    [0.48],
    [0.78],
    [1.1], 
    [2.2], 
    [4.5], 
    [9.0]
][:len(concurrency)]

rund = [
    [0.42],
    [0.55],
    [0.73],
    [1.2], 
    [2.1], 
    [4.4]
][:len(concurrency)]

rund_open = [
    [0.34],
    [0.54],
    [0.93],
    [2.4], 
    [5.5], 
    []
][:len(concurrency) - 1]

firecracker = [
    [0.33],
    [1.3],
    [2.4],
    [7.3], 
    [15.3], #concurrency-40 
    [42.7]  #concurrency-40 
][:len(concurrency)]

# ----------------------

kata_qemu_network = [
    [0.57],
    [1.3],
    [2.2], 
    [4.8], 
    [8.7], 
    [21.6]
][:len(concurrency)]

rund_network = [
    [0.50],
    [1.1],
    [1.5],
    [3.6], 
    [6.9], 
    [14.8]
][:len(concurrency)]

rund_open_network = [
    [0.52],
    [1.1],
    [2.0],
    [5.2], 
    [], 
    []
][:len(concurrency) - 2]

firecracker_network = [
    [0.38],
    [1.6],
    [2.9],
    [9.6], 
    [25.1], 
    [67.4]
][:len(concurrency)]

data = [firecracker, firecracker_network, kata_qemu, kata_qemu_network, rund, rund_network] #rund_open, rund_open_network,  

# plot the figure
plt.figure(figsize=(5.2, 4.5))

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
plt.xlabel("Concurrency", fontsize = 21)
plt.ylabel("Time cost (s)", fontsize = 21)
plt.xticks([i for i in range(len(concurrency))], [con for con in concurrency], fontsize = 21)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])
plt.ylim(0, 80)
plt.yticks(fontsize = 21)
#plt.yscale('log')
#plt.yticks([i * 100 for i in range(6)])
colors = plot_colors.colors#plot_colors.get_colors(len(data))
colors =[colors[0], "darkorange",  "y"] #"brown",
#print(colors)
for line_id, line_data in enumerate(data):
    #print(int(line_id/2))
    plt.plot([i for i in range(len(line_data))], line_data, marker = markers[line_id], color = colors[int(line_id/2)], label = labels[line_id], linestyle=style[line_id%2])

plt.legend(fontsize = 14)
# plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
#            mode="expand", borderaxespad=0, ncol=2, frameon=False,
#            fontsize = 13.6, handletextpad=0.15)
plt.tight_layout()
plt.savefig("./figs_fixed/motivation_network_overhead_imc_camera.png")
plt.savefig("./figs_fixed/motivation_network_overhead_imc_camera.pdf")
plt.show()