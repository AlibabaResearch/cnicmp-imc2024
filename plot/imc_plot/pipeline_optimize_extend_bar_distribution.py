
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
plt.ylabel("Time reduction (%)", fontsize = 16)
plt.xlabel("Concurrency", fontsize = 16)
labels = ["40", "100", "200", "400", "600", "800"]
plt.xticks([i*bar_offset for i in range(len(labels))], [l for l in labels], fontsize = 16)
plt.yticks(fontsize=16)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])
plt.ylim(0, 40)
#plt.yticks([i * 100 for i in range(6)])


colors = plot_colors.colors#plot_colors.get_colors(len(data))
colors =[colors[0], "darkorange", plot_colors.colors_pack3[4]]
patterns = ["////", "\\\\\\\\", "---", "|||||", "//////", "\\\\\\", "---", "***"]

target_names = ["No-opt", "Opt"]
no_opt= [2.0, 4.7, 9.2, 20.7, 42.3, 65.4]
opt = [2.0, 4.4, 8.4, 16.8, 28.6, 43.4]

no_opt= [
        [2.0, 2.0], 
        [4.7, 4.8, 4.7], 
        [8.9, 9.2, 9.3], 
        [20.7, 21.0, 21.1],
        [42.3, 39.4, 39.4], 
        [65.4, 63.1, 69.6]
]

opt_four_hundred = [
        [2.0, 2.0], 
        [4.4, 4.5, 4.5], 
        [8.4, 8.3, 8.4], 
        [16.8, 17.1, 17.1], 
        [28.6, 29.6, 28.5], 
        [43.4, 45.0, 48.0]
]

opt_train = [
        [2.0, 2.0], 
        [4.2, 4.5, 4.4], 
        [8.1, 8.2, 8.1], 
        [16.8, 17.1, 17.1], 
        [29.4, 29.5, 29.7], 
        [46.8, 50.1, 48.3]
]

#data = [(np.array(no_opt) - np.array(opt)) / np.array(no_opt) * 100]


data1 = []
for i in range(len(opt)):
    no_opt[i] = np.array(no_opt[i])[np.argsort(no_opt[i])]
    opt_four_hundred[i] = np.array(opt_four_hundred[i])[np.argsort(opt_four_hundred[i])]
    improve = (np.array(no_opt[i]) - np.array(opt_four_hundred[i])) / np.array(no_opt[i]) * 100
    sum_improve = (np.sum(no_opt[i]) - np.sum(opt_four_hundred[i])) / np.sum(no_opt[i]) * 100
    #data.append([np.mean(improve), np.std(improve)])
    data1.append([sum_improve, np.std(improve)])

data2 = []
for i in range(len(opt_train)):
    opt_train[i] = np.array(opt_train[i])[np.argsort(opt_train[i])]
    improve = (np.array(no_opt[i]) - np.array(opt_train[i])) / np.array(no_opt[i]) * 100
    sum_improve = (np.sum(no_opt[i]) - np.sum(opt_train[i])) / np.sum(no_opt[i]) * 100
    #data.append([np.mean(improve), np.std(improve)])
    data2.append([sum_improve, np.std(improve)])


data = [data2, data1]
# data = [
#         [np.mean(np.array(no_opt[i]) - np.array(opt[i])), np.mean(np.array(no_opt[i]) - np.array(opt[i]))] for i in range(len(opt))
# ]

for cni_id, cni_data in enumerate(data):
    for eid in range(len(cni_data)):
        if cni_id == 1 and eid == 3: continue
        if eid == 3: offsett = 0.5
        else: offsett = 0

        if cni_id == 0: ll = "Customize"
        else: ll = "Generalize-400"


        if eid == 0:
            plt.bar(bar_width * (-int(len(data)/2)+ 0.5 + offsett + cni_id) + eid * bar_offset, cni_data[eid][0],
                ec="white",
                color=colors[cni_id], alpha=1, width=bar_width, hatch=patterns[cni_id], label=ll)
        else:
            plt.bar(bar_width * (-int(len(data)/2)+ 0.5 + offsett + cni_id) + eid * bar_offset, cni_data[eid][0],
                ec="white",
                color=colors[cni_id], alpha=1, width=bar_width, hatch=patterns[cni_id])
        
        plt.bar(bar_width * (-int(len(data)/2)+ 0.5 + offsett + cni_id) + eid * bar_offset, cni_data[eid][0],
                ec="black",
                color="none", alpha=1, width=bar_width)
        plt.errorbar(bar_width * (-int(len(data)/2) + 0.5 + offsett + cni_id)+ eid * bar_offset, cni_data[eid][0], yerr=cni_data[eid][1], fmt='-', ecolor='red',
                      linewidth=0.8, capsize=4)
        
        extra_set = [
                [0, 0, 0, 0, -2, -2],
                [0, 0, 0, 0, 0, 0]]
        plt.text(bar_width * (-int(len(data)/2) - 0.1 + extra_set[cni_id][eid] * 0.1  + [-1.4, 0.8][cni_id]*0.2 + offsett + cni_id) + eid * bar_offset, cni_data[eid][0] + 3, str(round(cni_data[eid][0], 1)), fontsize=12)

#plt.text(-0.1, 1, "0%", fontsize = 14)
#plt.yscale("log")
plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
           mode="expand", borderaxespad=0, ncol=4, frameon=False,
           fontsize = 14, handletextpad=0.15)
#legend = plt.legend(frameon=True,fontsize=13)

plt.tight_layout()
plt.savefig("./figs/pipeline_optimize_extend_bar_distribution.pdf")
plt.show()



