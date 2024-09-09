from bcc import BPF

# 定义eBPF程序
bpf_program = """
#include <uapi/linux/ptrace.h>
#include <linux/sched.h> // for TASK_COMM_LEN
#include <linux/kernfs.h>

struct data_t {
    u32 pid;
    u64 timestamp;
    char comm[32];
    char name[128];
    char parentName[128];
};

BPF_HASH(start, u32, struct data_t);
BPF_PERF_OUTPUT(events);

int kprobe__cgroup_apply_control_enable(struct pt_regs *ctx) {
    struct data_t data = {};
    u32 pid = bpf_get_current_pid_tgid() >> 32;
    u64 ts = bpf_ktime_get_ns();
    //bpf_probe_read_str(&data.name, 128, (void *)PT_REGS_PARM2(ctx));

    struct kernfs_node *parent_kn = (struct kernfs_node *)PT_REGS_PARM1(ctx);
    //struct kernfs_node *parent_kn2 = kernfs_get_parent(parent_kn);
    //bpf_probe_read(&parent_kn2, sizeof(parent_kn2), &(parent_kn->parent));
    bpf_probe_read_str(&data.parentName, 128, (void *)(parent_kn->name));

    data.timestamp = ts;
    data.pid = pid;
    bpf_get_current_comm(&data.comm, sizeof(data.comm));
    start.update(&pid, &data);
    return 0;
}

int kretprobe__cgroup_apply_control_enable(struct pt_regs *ctx) {
    u32 pid = bpf_get_current_pid_tgid() >> 32;
    struct data_t *datap;
    u64 duration;

    datap = start.lookup(&pid);
    if (datap != 0) {
        struct kernfs_node *parent_kn = (struct kernfs_node *)PT_REGS_PARM1(ctx);
        bpf_probe_read_str(&(datap->parentName), 128, (void *)(parent_kn->name));
        bpf_get_current_comm(&(datap->comm), sizeof(datap->comm));

        duration = bpf_ktime_get_ns() - datap->timestamp;
        datap->timestamp = duration;
        events.perf_submit(ctx, datap, sizeof(struct data_t));
        start.delete(&pid);
    }
    return 0;
}
"""

# 初始化BPF
b = BPF(text=bpf_program)

# 定义处理事件的回调函数
def print_event(cpu, data, size):
    event = b["events"].event(data)
    print("cgroup_mkdir() PID: %d Duration: %d ns Comm: %s Name: %s" % (event.pid, event.timestamp, event.comm, event.parentName))

# 打开性能事件缓冲区
b["events"].open_perf_buffer(print_event)

# 打印信息提示
print("Tracing ... Hit Ctrl-C to end.")

# 轮询事件
while True:
    try:
        b.perf_buffer_poll()
    except KeyboardInterrupt:
        break
