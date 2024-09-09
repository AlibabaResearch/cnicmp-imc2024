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
style = ["-", "--", "-.", ".-"]
markers = [".","x", "o","^"]
labels = ["no-net", "veth", "vhostuser", "ipvlan",
          "ipvtap", "macvlan", "macvtap",
          "tc-routing"]#, "eBPF"]

# processing raw data
# one dimension for this figure


time_cost =[
    [10.3, 10.5, 10.3],
    [24.3, 25.5, 25.8],
    [83.1, 84.5, 83.7], 
    [23.6, 24.1, 24.2],
    [21.5, 19.6, 20.8 ],
    [22.9, 22.3, 21.9],
    [21.8, 20.8, 20.3],
    [24.8, 26.3, 25.5],
    #[104.9, 115.1, 110.1]  # 184.9 , -80 for displaying
]



data = [[[np.mean(i)] for i in time_cost]]
std = [np.std(i) for i in time_cost]
# exit()
# plot the figure
#plt.figure(figsize=(6.2, 5))

plt.figure(figsize=(7.6, 5))

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
plt.ylabel("Time cost (s)", fontsize = 18)
plt.xticks([i for i in range(len(labels))], [plot_colors.circled_num[i] for i in range(len(labels))], fontsize = 26)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])
plt.ylim(0, 90)
#plt.yticks([i * 10 for i in range(12)], ["0", "10", "20", "30", "40", "50", "60", "70", "80",  "...", "180", "190"])
#plt.yticks([i * 10 for i in range(13)], ["0", "", "20", "", "40", "", "60", "", "80",  "...", "180", "", "200"], fontsize=18)
#plt.yticks([i * 20 for i in range(7)], ["0",  "20",  "40",  "60",  "80",  "...", "180"])
# plt.yscale("log")
# plt.yticks([10, 100])

bar_width = 0.5
bar_offset = 1
colors = plot_colors.colors_pack3#2
#colors = plot_colors.sample_colors(colors, len(colors))
patterns = plot_colors.hatch_patterns#["//////", "\\\\\\", "---", "***", "//////", "\\\\\\", "---", "***"]
segment_name = labels
for line_id, line_data in enumerate(data):
    for eid in range(len(line_data)):
        if True:
            # plt.bar(bar_width * (-int(len(data)/2)+ line_id) + eid * bar_offset, line_data[eid], label=segment_name[eid],
            #         ec=colors[eid],
            #         color="none", alpha=1, width=bar_width, hatch=patterns[eid])
            plt.bar(bar_width * (-int(len(data)/2)+ line_id) + eid * bar_offset, line_data[eid], label=segment_name[eid],
                    ec="white",
                    color=colors[eid], alpha=1, width=bar_width, hatch=patterns[eid])
        else:
            # plt.bar(bar_width * (-int(len(data)/2) + line_id) + eid * bar_offset, line_data[eid],
            #         ec=colors[eid],
            #         color="none", alpha=1, width=bar_width, hatch=patterns[eid])
            plt.bar(bar_width * (-int(len(data)/2)+ line_id) + eid * bar_offset, line_data[eid], 
                    ec="white",
                    color=colors[eid], alpha=1, width=bar_width, hatch=patterns[eid])

        plt.errorbar(bar_width * (-int(len(data)/2)+ line_id)+ eid * bar_offset, line_data[eid], yerr=std[eid], fmt='-', ecolor='red',
                      linewidth=1.8, capsize=4)

        plt.bar(bar_width * (-int(len(data)/2) + line_id) + eid * bar_offset, line_data[eid],
                ec="black",
                color="none", alpha=1, width=bar_width)
        # plt.errorbar(bar_width * (-1 + line_id) + eid * bar_offset, avg_lat[eid], yerr=std[eid], fmt='-', ecolor='orangered',
        #              linewidth=0.8, capsize=4)

plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
           mode="expand", borderaxespad=0, ncol=3, frameon=False,
           fontsize = 17, handletextpad=0.25)
plt.tight_layout()
plt.savefig("./figs_fixed/concurrency_overall_cr.pdf")
plt.savefig("./figs_fixed/concurrency_overall_cr.png")
plt.show()