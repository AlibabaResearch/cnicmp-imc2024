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
labels = ["0", "10", "50", "100"] # [ "ipvlan", "ipvtap"] #, "macvlan", "macvtap", "tc-routing"]

# processing raw data
data =[
    #[[7.35, 34.9], [11.9, 33.3], [19.8, 34.3], [30.4, 35.1]],
    #[[7.35, 24.8], [11.9, 24.9], [19.8, 26.5], [33.4, 36.0]],
    [[9.3, 21.5], [15.5, 20.5], [23.3, 24.7], [31.5, 33.4]],
    #[[7.9, 33.0], [12.3, 34.1],[19.7, 34.5],[30.9, 34.1]], 
    #[[9.0, 21.8],[13.9, 20.5],[22.5, 25.3],[32.7, 34.0]], #[21.8, 20.5, 25.3, 34.0]
    #[[10.9, 52.4], [21.2, 51.2], [37.9, 53.6], [43.7, 50.3]]  #[52.4, 51.2, 53.6, 50.3] 
]
# exit()
# plot the figure
plt.figure(figsize=(5.2 , 3.5 * 0.8))

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

bar_width = 0.3
bar_offset = 0.8

#plt.xlabel("CNIs", fontsize = 16)
plt.ylabel("SR/Pod", fontsize = 15)
plt.xlabel("Timeline (s)", fontsize = 15)
plt.yticks([i * bar_offset for i in range(len(labels))], [l for l in labels], fontsize = 15)
plt.xticks(fontsize=15)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])
plt.xlim(0, 50)
plt.ylim(-0.5, 3)
#plt.yticks([i * 100 for i in range(6)])


colors = plot_colors.colors#plot_colors.get_colors(len(data))
colors =[colors[0], "darkorange"]
patterns = ["//////", "\\\\\\", "---", "***", "//////", "\\\\\\", "---", "***"]
segment_name = labels
for cni_id, cni_data in enumerate(data):
    for eid in range(len(cni_data)):

        plt.barh(bar_width * (-0) + eid * bar_offset, cni_data[eid][0], label="AddCNI" if cni_id == 0 and eid == 0 else None,
                ec="white",
                color=colors[0], alpha=1, height=bar_width, hatch=patterns[0])
        plt.barh(bar_width * (-0) + eid * bar_offset,  cni_data[eid][1]-cni_data[eid][0], left=cni_data[eid][0],  label="Others" if cni_id == 0 and eid == 0 else None,
                ec="white",
                color=colors[1], alpha=1, height=bar_width, hatch=patterns[1])

        plt.barh(bar_width * (-0) + eid * bar_offset, cni_data[eid][1],
                ec="black",
                color="none", alpha=1, height=bar_width)
        #text = "SR/Pod=" + ["0", "10", "50", "100"][eid]
        #plt.text(40, 0.4 * (-int(len(cni_data)/2) + 0.25 +eid) + cni_id * bar_offset, text, fontsize=10)
        # plt.errorbar(bar_width * (-1 + line_id) + eid * bar_offset, avg_lat[eid], yerr=std[eid], fmt='-', ecolor='orangered',
        #              linewidth=0.8, capsize=4)

plt.plot([data[0][2][1], 29], [bar_width * 0 + 2 * bar_offset, 1.7], linewidth=0.6, color="red")
plt.plot([data[0][3][1], 35], [bar_width * 0 + 3 * bar_offset, 2.05], linewidth=0.6, color="red")
plt.text(30, 2 * 0.8, "AddCNI dominates", color="red", fontsize = 10)
#plt.plot([data[0][3][1], 36], [bar_width * (0*int(len(cni_data)/2) + 0.02 +2) + 0 * bar_offset, 1.7], linewidth=0.6, color="red")

plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
           mode="expand", borderaxespad=0, ncol=2, frameon=False,
           fontsize = 13, handletextpad=0.15)
#legend = plt.legend(frameon=True,fontsize=13)
# for handle in legend.legendHandles:
#         handle.set_edgecolor('black')
#         handle.set_linewidth(1)
plt.tight_layout()
plt.savefig("./figs_fixed/pipeline_acl_imc2.pdf")
plt.savefig("./figs_fixed/pipeline_acl_imc2.png")
plt.show()



