import matplotlib.pyplot as plt
import numpy as np
import os, sys
import dateutil.parser as dp
import plot_colors
#import matplotlib.patches as patches

# 400 current dockers
#[0, 1, 2, 3, 6, -5, -6, -3, -2, -1 ]
setid = int(sys.argv[1])
lim = 60

setlabel = ["no-net", "net-veth", "net-ipvlan", "net-ipvtap",  "net-vhost", "net-macvtap", 
"net-tc-routing", "net-tc-routing-no-cmd", "net-veth2", "net-ipvtap-acl10", "net-ipvtap-acl50",
"net-ipvlan-acl10", "net-ipvlan-acl50",  "net-veth-acl5", "no-net2", "net-ebpf", "debug", "no-opt", "opt", "net-ipvtap-pool"]

if setid in [0, -6]: 
    hi = 2.8
else:
    hi = 3

if setid == 6: lim = 30
elif setid == 4: lim = 100
elif setid in [0, 1, 2, 3]: lim = 30
elif setid == -5: lim = 200
elif setid in [-3, -2]: lim = 22
elif setid in [-1, -6]: lim = 12

log_id =[20230907073151,    20240122011910, 20240122015716,        20240818022115,     
 20240122014205,      20240122015716,  20240122015931, 20230828231019,    20230828231422,
 20230829014113,         20230829015151,  20230827072048, 20230907073151, 20230904140820, 
 20230912054104, 20230906235229, 20230906235000, 20230912034906][setid] #20230904135332, 20230904140145
containerd_log_path = "/home/cni/cnicmp_logs/containerd_logs/cnicmp-{}.log".format(log_id)
kata_log_path = "/home/cni/cnicmp_logs/kata_logs/cnicmp-{}.log".format(log_id)
ovs_log_path = "/home/cni/cnicmp_logs/cni_logs/ovs_cni/ovs-{}.log".format(log_id)#-400-for-plot.log"
tmp_log_path = "./tmp.log"

if not os.path.exists(containerd_log_path):
    os.system("scp 22.22.22.170:{} {}".format(containerd_log_path, containerd_log_path))
# if not os.path.exists(ovs_log_path):
#     os.system("scp 22.22.22.174:{} {}".format(ovs_log_path, ovs_log_path))
if not os.path.exists(kata_log_path):
    os.system("scp 22.22.22.170:{} {}".format(kata_log_path, kata_log_path))

if not os.path.exists(containerd_log_path):
    print("Error: log file exists neither in cni1 or cni2!")
    exit()

time_scaleup_us = 1000.0 # ms

def get_time_stamp(ts):
    parsed_t = dp.parse(ts)
    t_in_us = parsed_t.strftime('%s%f')
    return eval(t_in_us) / time_scaleup_us

# aggregate the logs
os.system("cat {} > {}".format(containerd_log_path, tmp_log_path))
os.system("cat {} >> {}".format(kata_log_path, tmp_log_path))
os.system("cat {} >> {}".format(ovs_log_path, tmp_log_path))

target_markers = ["CNI_nns", "CNI_md_add_lo", "CNI_md_add_cni", "CNI_ad"]
target_labels = ["containerd-nns", "add-lo", "add-cni", "kata-interface"]

# ovs veth/rund
target_markers = ["CNI_nns", "CNI_md_add_lo", "CNI_setupVeth",  "CNI_IPAM_waitlinkip", "CNI_IPAM_assignIP", "CNI_create_pod", "CNI_setup_task"]
target_labels = ["containerd-nns", "add-lo", "setVeth", "ovs-port", "assignIP", "prePod",   "startQEMU"]

# ovs tc routing
#target_markers =  ["CNI_nns", "CNI_md_add_lo", "CNI_createTap", "CNI_createVeths",  "CNI_add_tc", "CNI_setIP", "CNI_create_pod",  "CNI_setup_task"]
#target_labels = ["containerd-nns", "add-lo", "ovs-port", "setVeth", "add-tc-rule", "assignIP", "pre_pod",   "start_pod"]

# "veth"
target_markers = ["CNI_nns", "CNI_md_add_lo", "CNI_setupVeth",  "CNI_IPAM_waitlinkip", "CNI_IPAM_assignIP",  "CNI_kata_create_sandbox2",  "CNI_kata_set_cgroup", "CNI_kata_start_vm", "constrainHypervisor", "CNI_ad"]
target_labels = ["containerd-nns", "add-lo", "setVeth", "ovs-port", "assignIP",   "resourceController", "setupResourceController",   "start_vm",  "resourceController-vCPU", "kata-interface"]

# "ipvlan"
target_markers = ["CNI_nns", "CNI_md_add_cni", "CNI_kata_create_sandbox",  "CNI_kata_set_cgroup", "CNI_ad", "CNI_kata_start_vm"]
target_labels = ["NetNS", "AddCNI",  "Cgroup", "Cgroup",  "AttachIf",  "StartVM"]
skip_id = 3

# target_markers = ["CNI_nns", "CNI_md_add_lo", "CNI_md_add_cni", "CNI_kata_create_sandbox2",  "CNI_kata_start_vm", "CNI_ad"]
# target_labels = ["containerd-nns", "add-lo", "add-cni",  "resourceController",  "start_vm", "kata-interface"]

# "ebpf"
# target_markers = ["CNI_nns", "CNI_md_add_lo", "CNI_md_add_cni", "CNI_kata_create_sandbox2",  "CNI_kata_set_cgroup", "CiliumAgent_Create_Postprocess_Sync", "CNI_kata_start_vm", "CNI_ad"]
# target_labels = ["createNNS", "addLO", "addCNI",  "cgroupOP1", "cgroupOP2",  "eBPFRecompile", "startVM", "attachInterface"]

# rund
#target_markers = ["CNI_nns", "CNI_md_add_lo", "CNI_md_add_cni" ,  "CNI_create_pod", "CNI_setup_task"]
#target_labels = ["containerd-nns", "add-lo",  "addCNI", "prePod",   "startQEMU"]

container_ids = []
container_start_t = []
container_end_t = []
log_data = {}
order_entry =   "CNI_kata_start_vm"  #["constrainHypervisor", "CNI_ad"][setid]
# parse the log
lid = 0
with open(tmp_log_path, "r") as log:
    line = log.readline()
    #print(line)
    while line:
        #lid += 1
        if len(line) > 2 and len(line.split()) > 7:
            sline = line.split()
            # print(lid)
            # print(sline)
            # print(''.join(sline[-6:]))
            try:
                data = eval(''.join(sline[-6:]))
            except:
                data = eval(''.join(sline[-8:]))
                # lid += 1
                # print(line)
                # print("error:", lid)
                # line = log.readline()
                # continue
            marker = sline[2]
            if marker not in target_markers:
                line = log.readline()
                #print("!!!",marker)
                continue
            id = data['record_id']
            start_t = get_time_stamp(data['start_t'])
            
            n_flag = False
            if id not in container_ids: 
                n_flag = True
                container_ids.append(id)
                container_start_t.append(start_t)        
            if start_t < container_start_t[container_ids.index(id)]: 
                container_start_t[container_ids.index(id)] = start_t
            if id not in log_data.keys(): log_data[id] = {}
            if 'elapsed_ns' in data.keys():
                log_data[id][marker] = [start_t, data['elapsed_ns'] / 1000000.0]
            else:
                log_data[id][marker] = [start_t, data['elapsed_ms']]
            #print(log_data[id].keys())
            
            if n_flag: 
                container_end_t.append(log_data[id][marker][0] + log_data[id][marker][1])
            if (log_data[id][marker][0] + log_data[id][marker][1]) > container_end_t[container_ids.index(id)] and marker == order_entry: 
                container_end_t[container_ids.index(id)] = (log_data[id][marker][0] + log_data[id][marker][1])
        line = log.readline()


# plot the results
# veth
colors = ["purple", "grey", "blue",  "brown", "grey", "pink", "green",  "black","darkorange",  "red"]
# macvlan
colors = ["purple", "grey", "blue", "pink", "green",  "black","darkorange",  "red"]
#colors = ["purple", "grey", "blue", "pink", "black", "red"]
# --- colors = ["purple", "grey", "blue", "pink", "green", "brown", "black","darkorange",  "red"]
# ovs veth
#colors = ["purple", "grey", "blue", "darkorange",  "red",  "pink",  "black"]
# ovs tc routing
#colors = ["purple", "grey", "blue", "darkorange", "green", "red",  "pink",   "black"]

colors = plot_colors.colors_pack1#2
colors = plot_colors.colors_pack3#2
colors = plot_colors.sample_colors2(colors, 2)
colors[0] = "grey"
tmp = colors[-1]
colors[0] = "grey"
# colors[-1] = colors[-2]
# colors[-2] = tmp
colors[3] = "brown"
colors[4] = "RoyalBlue"#"CornflowerBlue"#
#colors = plot_colors.sample_colors(colors, len(colors))
alphas = [0.2, 1, 0.8, 1, 1, 1]
#print(colors)

#exit()

sample_interval = 1 # sample the info of the containers
sampled_ids = []
sampled_end_t = []
for i, id in enumerate(container_ids):
    if i % sample_interval == 0: 
        sampled_ids.append(id)
        sampled_end_t.append(container_end_t[i])


# reorder the containers basing on start_t

base_t = min(container_start_t)
t_args = np.argsort(np.array(sampled_end_t))
sampled_ids = np.array(sampled_ids)[t_args]

sampled_end_t = np.array(sampled_end_t)[t_args]
print("{} last end: {}".format(order_entry, (sampled_end_t[-1] - base_t) / 1000.0))

#plt.figure(figsize = (3.8, 2.8))
style = ["-", "--", ":", "-.", "--", "-", "--", ":", "-.", "--"]
patterns = ["//", "\\\\", "--", "||",  "*", "\\\\\\", "---", "***"]
patterns = [i * 2 for i in patterns]
plt.figure(figsize = (3.8, hi))
last_end_t = -1 
for i, id in enumerate(sampled_ids):
    for mi, marker in enumerate(target_markers):
        #print(log_data[id])
        if marker not in log_data[id].keys(): 
            #print(marker)
            #exit()
            continue
        start_tt = (log_data[id][marker][0] - base_t) / 1000.0
        end_tt = (start_tt + log_data[id][marker][1] / 1000.0)
        if end_tt > last_end_t: last_end_t = end_tt
        #print(start_tt)
        if mi >= skip_id: offset = 1
        else: offset = 0
        
        # if i == 0 and mi!=skip_id: 
        #     plt.plot([start_tt, end_tt], [i, i], label = target_labels[target_markers.index(marker)], color = colors[mi - offset], alpha=[0.5, 1][mi%2])
        #     print(target_labels[target_markers.index(marker)], [start_tt, end_tt])
        # else: plt.plot([start_tt, end_tt], [i, i], color = colors[mi - offset] , alpha=[0.5, 1][mi%2])

        # print(colors)
        # print(mi - offset)
        # exit()


        color_name = colors[mi - offset]

        if i == 0 and mi!=skip_id: 
            # plt.barh(i, end_tt - start_tt, left = start_tt, label = target_labels[target_markers.index(marker)],
            #         ec="dimgrey",
            #         color=colors[mi - offset], alpha=1, height=4, linewidth= 0.002, hatch=patterns[mi-offset])
            plt.barh(i, end_tt - start_tt, left = start_tt, label = target_labels[target_markers.index(marker)],
                    color=colors[mi - offset], alpha=alphas[mi - offset], height=4)
            print(target_labels[target_markers.index(marker)], [start_tt, end_tt])
        else: 
            # plt.barh(i, end_tt - start_tt, left = start_tt, 
            #         ec="dimgrey",
            #         color=colors[mi - offset], alpha=1, height=4, linewidth=0.002, hatch=patterns[mi-offset])
            plt.barh(i, end_tt - start_tt, left = start_tt, 
                    color=colors[mi - offset], alpha=alphas[mi - offset], height=4)

        # if i == 0 and mi!=skip_id: 
        #     plt.barh(i, end_tt - start_tt, left = start_tt, label = target_labels[target_markers.index(marker)],
        #             ec="white",
        #             color=colors[mi - offset], alpha=1, height=1, linewidth= 0.002, hatch=patterns[mi-offset])
        #     print(target_labels[target_markers.index(marker)], [start_tt, end_tt])
        # else: 
        #     plt.barh(i, end_tt - start_tt, left = start_tt, 
        #             ec=colors[mi - offset],
        #             color="none", alpha=1, height=1, linewidth=0.002, hatch=patterns[mi-offset])

        # if i == 399:
        #     print("{} - {} - {}".format(marker, start_tt, end_tt))


points = [(0.2, -10), (1.5, -10), (1.5, 410), (0.2, 410), (0.2, -10)]  # 闭合矩形
# 拆分 x 和 y 坐标
x, y = zip(*points)
# 绘制矩形框
plt.plot(x, y, color='black', linestyle="--", linewidth = 1)  # 添加标记 'o' 可视化点



plt.xlabel("Timeline (s)", fontsize = 14)
plt.ylabel("Container ID ", fontsize = 14)
plt.yticks(fontsize = 14)
plt.xticks(fontsize = 14)
plt.xlim(0, lim)

# plot a verticle line for readline
plt.axvline(last_end_t + 0.2, linewidth = 1, color = "grey", linestyle = "--")
#plt.text(last_end_t - lim / 8 , 2, str(round(last_end_t, 1)) )

#plt.legend(fontsize = 8)

plt.legend(bbox_to_anchor = (0,1,1,0), loc="lower right",
           mode="expand", borderaxespad=0, ncol=3, frameon=False,
           fontsize = 9.5, handletextpad=0.15)


plt.tight_layout()
#plt.savefig(["./e2e-no.png", "./e2e.png"][setid])
fig_path = "./rotate_figs_fixed_poisson/" + ["./kata-no-net-2.png", "./kata-net-veth-2.png", "./kata-net-ipvlan-2.png", "./kata-net-ipvtap-poisson.png", "./kata-net-vhost.png",
 "./kata-net-macvtap.png", "./kata-net-tc-routing-2.png", "./kata-net-tc-routing-no-cmd-2.png", "./kata-net-veth2",
 "./kata-net-ipvtap-acl10.png", "./kata-net-ipvtap-acl50.png", "./kata-net-ipvlan-acl10.png", "./kata-net-ipvlan-acl50.png",
 "./kata-net-veth-acl5.png", "./kata-no-net2-2.png", "./kata-net-ebpf-2.png", "./debug-acl50-80con.png", "./no-opt.png", "./opt.png", "./kata-net-ipvtap-pool.png"][setid]
plt.savefig(fig_path)
fig_path = fig_path[:-3] 
fig_path += "pdf" 
plt.savefig(fig_path)
# clear the files
#os.system("rm {}".format(tmp_log_path))