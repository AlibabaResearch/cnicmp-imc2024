
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



# plot the figure
plt.figure(figsize=(6, 3.2))

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

bar_width = 0.8
bar_offset = 2.5

#plt.xlabel("CNIs", fontsize = 16)
plt.ylabel("Time cost(s)", fontsize = 16)
#plt.xlabel("Concurrency", fontsize = 16)
labels = ["no-net", "no-opt", "CC", "pre-c", "pre-c\n+CC"]
plt.xticks([i*bar_offset for i in range(len(labels))], [l for l in labels], fontsize = 16)
plt.yticks(fontsize=16)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])
plt.ylim(0, 30)
#plt.yticks([i * 100 for i in range(6)])


colors = plot_colors.colors#plot_colors.get_colors(len(data))
colors =[colors[0], "darkorange", plot_colors.colors_pack3[4]]
patterns = ["////", "\\\\\\\\", "---", "|||||", "//////", "\\\\\\", "---", "***"]


raw_data = [
        [10.3,10.5,10.5],
        [20.7, 21.0, 21.1],
        [16.8, 16.7, 17.5],
        [9.7, 11.0, 10.9],
        [8.6,  8.4, 8.7]
]

#data = [(np.array(no_opt) - np.array(opt)) / np.array(no_opt) * 100]

data = []
for i in range(len(raw_data)):
    data.append([np.mean(raw_data[i]), np.std(raw_data[i])])
data = [data]
# data = [
#         [np.mean(np.array(no_opt[i]) - np.array(opt[i])), np.mean(np.array(no_opt[i]) - np.array(opt[i]))] for i in range(len(opt))
# ]

for cni_id, cni_data in enumerate(data):
    for eid in range(len(cni_data)):
        plt.bar(bar_width * (-int(len(data)/2)+ cni_id) + eid * bar_offset, cni_data[eid][0],
                ec="white",
                color=colors[cni_id], alpha=1, width=bar_width, hatch=patterns[cni_id])
        plt.bar(bar_width * (-int(len(data)/2)+ cni_id) + eid * bar_offset, cni_data[eid][0],
                ec="black",
                color="none", alpha=1, width=bar_width)
        plt.errorbar(bar_width * (-int(len(data)/2) + cni_id)+ eid * bar_offset, cni_data[eid][0], yerr=cni_data[eid][1], fmt='-', ecolor='red',
                      linewidth=0.8, capsize=4)
        plt.text(bar_width * (-int(len(data)/2) - 0.6 + cni_id) + eid * bar_offset, cni_data[eid][0] + 1, str(round(cni_data[eid][0], 1)), fontsize=14)

#plt.axhline(np.mean(raw_data[0]), linewidth = 0.5, color = "darkorange", linestyle = "dashed")

#plt.text(-0.1, 1, "0%", fontsize = 14)
#plt.yscale("log")
plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
           mode="expand", borderaxespad=0, ncol=4, frameon=False,
           fontsize = 15.8, handletextpad=0.15)
#legend = plt.legend(frameon=True,fontsize=13)

plt.tight_layout()
plt.savefig("./figs_fixed/pipeline_optimize_combined_bar_camera.pdf")
plt.savefig("./figs_fixed/pipeline_optimize_combined_bar_camera.png")
plt.show()



