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
style = ["-", "--", ":", "-.", "--"]
markers = [".","x", "o","^", "*"]
labels = ["Firecracker-net", "Kata-net", "RunD-net"] # "RunD(O)-net", 

# processing raw data
concurrency = [1, 20, 40, 100, 200, 400]
kata_qemu = [
    [0.5],
    [0.7],
    [1.2], # done
    [2.3], 
    [4.9], # todo
    [10.5]
][:len(concurrency)]


kata_qemu_network = [
    [0.6],
    [2.0],
    [3.3],
    [7.8], 
    [15.5],
    [34.0]
][:len(concurrency)]


rund = [
    [0.44],
    [0.55],
    [0.71],
    [1.3], 
    [2.3], 
    [4.6]
][:len(concurrency)]


rund_network = [
    [0.54],
    [1.87],
    [3.11],
    [7.4], 
    [14.3],
    [28.6]
][:len(concurrency)]


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
    [-1],
    [-1]
][:len(concurrency)]

firecracker = [
    [0.33],
    [1.3],
    [2.4],
    [7.3], 
    [15.3], 
    [42.7] 
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
    [-1], 
    [-1]
][:len(concurrency)]

firecracker_network = [
    [0.38],
    [1.6],
    [2.9],
    [9.6], 
    [25.1], 
    [67.4]
][:len(concurrency)]

data = [ (-np.array(firecracker) + np.array(firecracker_network)) / np.array(firecracker) * 100,
    (-np.array(kata_qemu) + np.array(kata_qemu_network)) / np.array(kata_qemu) * 100,
       (-np.array(rund) + np.array(rund_network)) / np.array(rund) * 100,
        ] #(-np.array(rund_open) + np.array(rund_open_network)) / np.array(rund_open) * 100,

print(np.max(data))

# plot the figure
plt.figure(figsize=(5.38, 4.5))

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

bar_width = 0.28
bar_offset = 1.6
#plt.xlabel("CNIs", fontsize = 16)
plt.xlabel("Concurrency", fontsize = 21)
plt.ylabel("Increase (%)", fontsize = 21)
plt.xticks([i*bar_offset for i in range(len(concurrency))], [con for con in concurrency], fontsize = 21)
plt.yticks(fontsize = 21)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])
#plt.ylim(0, 40)
#plt.yticks([0, 20, 40, 60, 80],fontsize = 16)
#plt.yticks([i * 100 for i in range(6)])
colors = plot_colors.colors_pack3#plot_colors.get_colors(len(data))
colors = plot_colors.sample_colors(colors, 2)
colors = plot_colors.colors#plot_colors.get_colors(len(data))
colors =[colors[0], "darkorange", "y"]#"brown", "y"]
#print(colors)



patterns = ["///", "\\\\\\", "---", "***", "//////", "\\\\\\", "---", "***"]
segment_name = labels
for line_id, line_data in enumerate(data):
    for eid in range(len(line_data)):
        if eid == 0:
            # plt.bar(bar_width * (-int(len(data)/2)+ line_id) + eid * bar_offset, line_data[eid], label=segment_name[line_id],
            #         ec=colors[line_id],
            #         color="none", alpha=1, width=bar_width, hatch=patterns[line_id])
            plt.bar(bar_width * (-int(len(data)/2)+ line_id) + eid * bar_offset, line_data[eid], label=segment_name[line_id],
                    ec="white",
                    color=colors[line_id], alpha=1, width=bar_width, hatch=patterns[line_id])
        else:
            # plt.bar(bar_width * (-int(len(data)/2) + line_id) + eid * bar_offset, line_data[eid],
            #         ec=colors[line_id],
            #         color="none", alpha=1, width=bar_width, hatch=patterns[line_id])
            plt.bar(bar_width * (-int(len(data)/2)+ line_id) + eid * bar_offset, line_data[eid], 
                    ec="white",
                    color=colors[line_id], alpha=1, width=bar_width, hatch=patterns[line_id])
        plt.bar(bar_width * (-int(len(data)/2) + line_id) + eid * bar_offset, line_data[eid],
                ec="black",
                color="none", alpha=1, width=bar_width)
        
        # plt.errorbar(bar_width * (-1 + line_id) + eid * bar_offset, avg_lat[eid], yerr=std[eid], fmt='-', ecolor='orangered',
        #              linewidth=0.8, capsize=4)

#plt.text(bar_width * (-int(len(data)/2)+ 1.4) + 4 * bar_offset, 0, "x", color="red", fontsize = 14)
#plt.text(bar_width * (-int(len(data)/2)+ 1.4) + 5 * bar_offset, 0, "x", color="red", fontsize = 14)
#plt.scatter(bar_width * (-int(len(data)/2)+ 2) + 4 * bar_offset, 5, marker="x", color="red", label="Fails to work")
#plt.scatter(bar_width * (-int(len(data)/2)+ 2) + 5 * bar_offset, 5, marker="x", color="red")

handles, labels = plt.gca().get_legend_handles_labels()

#legend = plt.legend([handles[idx] for idx in [1, 2, 3, 4, 0]], [labels[idx] for idx in [1, 2, 3, 4, 0]],fontsize = 14)
plt.legend(fontsize=14)
# for handle in legend.legendHandles:
#         handle.set_edgecolor('black')
#         handle.set_linewidth(1)
# plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
#            mode="expand", borderaxespad=0, ncol=2, frameon=False,
#            fontsize = 15.6, handletextpad=0.15)
plt.tight_layout()
plt.savefig("./figs_fixed/motivation_network_overhead_bar_imc_camera.png")
plt.savefig("./figs_fixed/motivation_network_overhead_bar_imc_camera.pdf")
plt.show()