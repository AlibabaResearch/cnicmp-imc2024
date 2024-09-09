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


# statistic tools
def _stat_cpu(ts, utils, marker):
    total_cpu_s = 0
    for i in range(1, len(ts)):
        total_cpu_s += (ts[i] - ts[i-1]) * utils[i] #/ 100.0
    print("{} - cpu time: {} s".format(marker, total_cpu_s))
    return total_cpu_s

def _stat_mem():
    pass

def get_cpu_utils(id, marker = "unknown"):
    log_path = "/home/cni/cnicmp_logs/mem_cpu_usage-{}.log".format(id)
    with open(log_path, "r") as f:
        lines = f.readlines()
        ts = []
        utils = []
        for i, line in enumerate(lines):
            # print(line)
            if i == 0 or len(lines) < 2: continue
            tss = int(line.split()[0]) / 1000.0 # in second
            cpu_util = eval(line.split()[1])
            mem = eval(line.split()[2]) * 500
            ts.append(tss)
            utils.append(cpu_util)
    cpu_s = _stat_cpu(ts, utils, marker)
    return np.array(ts) - min(ts), utils, cpu_s
        

#plot
style = ["-", "--", ":", "-.", "--", "-", "--", ":", "-.", "--"]
markers = [".","x", "o","^", "*", ".","x", "o","^", "*",]
labels = ["no-net", "veth", "vhostuser", "ipvlan", "ipvtap",  "tc-routing"] #"macvlan", "macvtap",
# log_ids = [20230908041002, 20230908021859, 20230907225839, 20230908023350, 20230908023033, 20230908032647] #20230908023642, 20230908032427, 
# log_ids2 = [20230907073151, 20230918064632, 20230918073716, 20230918052714, 20230918051741,20230918045429]
# log_ids3 = [20230907072859, 20230918070334, 20230918073716, 20230918053029, 20230918051741,20230918045429]

raw = ["20230908041002   20230918082118  20230918082540", "20230908021859   20230918082850  20230918082850",
"20230907225839   20230918073716  20230918074911", "20230908023350   20230918090902  20230918090902",
"20230908023033		20230918084737  20230918085225", "20230908032647   20230918091143  20230918091654"]
log_ids = [eval(i.split()[0]) for i in raw] #20230908023642, 20230908032427, 
log_ids2 = [eval(i.split()[1]) for i in raw]
log_ids3 = [eval(i.split()[2]) for i in raw]
# processing raw data
data = []
std = []
for i, lid in enumerate(log_ids):
    ts, utils, cpu_s = get_cpu_utils(lid, labels[i])
    ts, utils, cpu_s2 = get_cpu_utils(log_ids2[i], labels[i])
    ts, utils, cpu_s3 = get_cpu_utils(log_ids3[i], labels[i])
    data.append(np.mean([cpu_s, cpu_s2, cpu_s3]))#([ts, utils])
    std.append(np.std([cpu_s, cpu_s2, cpu_s3]))
data = [list(data)]

#data = [ipvlan, ipvtap, macvlan, macvtap, tc_routing]

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
plt.ylabel("Total CPU cost (%*s)", fontsize = 14)
plt.yticks(fontsize = 14)
plt.xticks([i for i in range(len(labels))], [plot_colors.circled_num[i] for i in range(len(labels))], fontsize = 20)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])
plt.ylim(0, 800)
#plt.yticks([i * 10 for i in range(9)], ["0", "10", "20", "30", "40", "50", "...", "180", "190"])
# plt.yscale("log")
# plt.yticks([10, 100])

bar_width = 0.4
bar_offset = 1
colors = list(plot_colors.colors_pack3)#2
colors.insert(1, colors[1])
colors.insert(1, colors[1])
#print(len(colors))
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

        plt.errorbar(bar_width * (-int(len(data)/2)+ line_id)+ eid * bar_offset, line_data[eid], yerr=std[eid], fmt='-', ecolor='red',
                      linewidth=1.8, capsize=4)

        plt.bar(bar_width * (-int(len(data)/2) + line_id) + eid * bar_offset, line_data[eid],
                ec="black",
                color="none", alpha=1, width=bar_width)


plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
           mode="expand", borderaxespad=0, ncol=2, frameon=False,
           fontsize = 11, handletextpad=0.15)
plt.tight_layout()
plt.savefig("./figs/resource_cpu_host_bar2.pdf")
plt.savefig("./figs/resource_cpu_host_bar2.png")
plt.show()

exit()

#plt.xlabel("CNIs", fontsize = 16)
plt.ylabel("CPU Utilization (%*s)", fontsize = 14)
plt.xlabel("Time (s)", fontsize = 14)
#plt.xticks([i for i in range(len(user_nums))], [num for num in user_nums], fontsize = 16)
#plt.xticks([int(len(time_cost)/2) - 1], ["CNIs"])
#plt.xticks([],[])
#plt.ylim(0, 80)
#plt.yticks([0, 20, 40, 60, 80],fontsize = 16)
#plt.yticks([i * 100 for i in range(6)])
colors = plot_colors.get_colors2() #plot_colors.colors#plot_colors.get_colors(len(data))
#print(colors)
for line_id, line_data in enumerate(data):
   # print(line_id)
    plt.plot(line_data[0], line_data[1], color = colors[line_id], label = labels[line_id], linestyle=style[line_id])#,  marker = markers[line_id], markersize=5)

plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
           mode="expand", borderaxespad=0, ncol=3, frameon=False,
           fontsize = 11.5, handletextpad=0.25)
plt.tight_layout()
print(1)
plt.savefig("./figs/resource_cpu_host_bar2.pdf")
plt.savefig("./figs/resource_cpu_host_bar2.png")
plt.show()