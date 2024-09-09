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
style = ["-", "--", "-.", ".-"]
markers = [".","x", "o","^"]
labels = ["veth", "vhostuser", "ipvlan", "ipvtap", 
        #"macvlan", "macvtap",
          "tc-routing"]

# processing raw data
# one dimension for this figure
throughput =[
    [3.8, 3.55, 3.72, 3.71, 3.64, 3.65],
    [1.26, 1.26, 1.34, 1.34, 1.32, 1.34], # vhost not available
    [4.37, 4.31, 3.74, 3.96, 4.06, 4.31],
    [3.68, 4.34, 4.34, 4.37, 3.45, 3.96],
    #[4.01, 4.33, 4.13, 3.69, 3.34, 4.39],
    #[3.7,  3.63,  4.11,  4.22, 4.34, 4.36],
    #[3.6, 3.8, 4.1, 3.8, 3.9, 4.1]
    [4.2, 3.6, 3.5, 3.6, 3.2, 4.2]
]
latency = [
    [16.7, 16.3, 16.9, 16.7, 16.7, 16.5],
    [9.0, 9.0, 9.1, 8.9, 8.8, 8.9],
    [17.5, 16.8, 16, 16.4, 16.2, 17.8],
    #[15.9, 16.1, 16, 16.5, 16, 16.1],
    #[16.1, 17.8, 16.3, 16.3, 15.9, 16.7],
    #[16.1, 17.8, 16.3, 16.3, 15.9, 16.7]
    [17.7, 16.4, 16.6, 16.7, 16.3, 16.5]
] #todo


pps = [
    #[504, 506, 507, 507, 505]
    [622, 619, 614, 603],
    [597, 569, 590, 595, 596],
    [601, 625, 622, 629, 625],
    [604, 626, 593, 609, 611],
    [615, 596, 617, 633, 605],
    [609, 604, 618, 610, 613],
    [627, 634, 608, 607, 604]
]
throughput =[
    #[27.3, 28.3, 26.3, 28.3, 25.5],
    [26.7, 26.9, 27.8],
    [18.5, 17.9, 18.2], # vhost not available
    [28.2, 28.8, 28.8],
    [29.9, 27.7, 28.5],
    #[31.8, 31.9, 29.6, 30.2, 27.8],
    #[29.1, 29.5, 32.4, 31.4],
    [27.7, 26.5, 28.8]
]
latency = [
    [67.0, 3.5],
    [29.2, 5.3],
    [74.8, 3.6],
    [70.5, 4.3],
    #[32.8, 3.2],
    #[32.2, 2.8],
    [72.7, 5.1]
] #todo

processed_tp = []
stds = []
for d in throughput:
    processed_tp.append(np.mean(d) )
    stds.append(np.std(d) )

data = [processed_tp]
# exit()
# plot the figure#
#plt.figure(figsize=(6.2, 4))
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
plt.ylabel("Throughput (Gb/s)", fontsize = 14)
plt.yticks(fontsize = 14)
plt.xticks([i for i in range(len(labels))], [plot_colors.circled_num[i] for i in range(len(labels))], fontsize = 20)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])
plt.ylim(0, 50)
#plt.yticks([i * 100 for i in range(6)])

bar_width = 0.4
bar_offset = 1
colors = plot_colors.colors_pack3#2
colors = plot_colors.sample_colors2(colors, 2)
patterns = plot_colors.hatch_patterns#["//////", "\\\\\\", "---", "***", "//////", "\\\\\\", "---", "***"]
segment_name = labels
for line_id, line_data in enumerate(data):
    for eid in range(len(line_data)):
        if True:
            plt.bar(bar_width * (-int(len(data)/2)+ line_id) + eid * bar_offset, line_data[eid], label=segment_name[eid],
                    ec="white",
                    color=colors[eid], alpha=1, width=bar_width, hatch=patterns[eid])
        else:
            plt.bar(bar_width * (-int(len(data)/2) + line_id) + eid * bar_offset, line_data[eid],
                    ec="white",
                    color=colors[eid], alpha=1, width=bar_width, hatch=patterns[eid])
        plt.bar(bar_width * (-int(len(data)/2) + line_id) + eid * bar_offset, line_data[eid],
                ec="black",
                color="none", alpha=1, width=bar_width)
        plt.errorbar(bar_width * (-int(len(data)/2)+ line_id)+ eid * bar_offset, line_data[eid], yerr=stds[eid], fmt='-', ecolor='orangered',
                      linewidth=0.8, capsize=4)

plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
           mode="expand", borderaxespad=0, ncol=2, frameon=False,
           fontsize = 11.5, handletextpad=0.2)
plt.tight_layout()
plt.savefig("./figs/throughput_overall_inter.pdf")
plt.show()