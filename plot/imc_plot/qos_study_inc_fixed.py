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
style = ["-", "--", "-.", ".-"]
markers = [".","x", "o","^"]
labels = ["veth", "vhostuser", "ipvlan",
          "ipvtap", 
          "tc-routing"]

# processing raw data
# one dimension for this figure
time_cost =[
    [47.5],
    [82.8], # vhost not available
    [34.9],
    [21.5],
    #[33.0],
    #[21.8],
    [52.9]
]

original =[
    [24.3, 25.5, 25.8],
    [83.1, 84.5, 83.7], 
    [23.6, 24.1, 24.2],
    [21.5, 19.6, 20.8 ],
    [24.8, 26.3, 25.5]
]

time_cost =[
    [35.5, 39.7, 37.0],
    [86.7, 86.6, 87.9], # vhost not available
    [23.9, 24.3, 24.4],
    [20.6, 20.8, 20.7],
    #[33.0],
    #[21.8],
    [25.8, 27.0, 25.9]
]




time_diff = (np.array(time_cost) - np.array(original)) / np.array(original) * 100
print(time_diff)
data = [[[np.abs(np.mean(i))] for i in time_diff]]
std = [np.std(i) for i in time_diff]

label_texts = [str(np.round(i, 1)[0]) for i in data[0]]
# exit()
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
plt.ylabel("Increased ratio (%)", fontsize = 14)
plt.yticks(fontsize = 14)
plt.xticks([i for i in range(len(labels))], [plot_colors.circled_num[i] for i in range(len(labels))], fontsize = 20)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])

plt.ylim(0, 60)

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
        # plt.errorbar(bar_width * (-int(len(data)/2)+ line_id)+ eid * bar_offset, line_data[eid], yerr=std[eid], fmt='-', ecolor='red',
        #               linewidth=1.8, capsize=4)

        plt.bar(bar_width * (-int(len(data)/2) + line_id) + eid * bar_offset, line_data[eid],
                ec="black",
                color="none", alpha=1, width=bar_width)
        text = label_texts[eid]
        if text is not None:
            print(text)
            plt.text(bar_width * (-int(len(data)/2) - 0.6) + eid * bar_offset, line_data[eid][0] + 2, text, fontsize=10)
        # plt.errorbar(bar_width * (-1 + line_id) + eid * bar_offset, avg_lat[eid], yerr=std[eid], fmt='-', ecolor='orangered',
        #              linewidth=0.8, capsize=4)

plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
           mode="expand", borderaxespad=0, ncol=2, frameon=False,
           fontsize = 11, handletextpad=0.15)
plt.tight_layout()
plt.savefig("./figs_fixed/qos_study_inc.png")
plt.savefig("./figs_fixed/qos_study_inc.pdf")
plt.show()