from bcc import BPF

# 定义eBPF程序
bpf_program = """
#include <uapi/linux/ptrace.h>
#include <linux/sched.h> // for TASK_COMM_LEN
#include <linux/kernfs.h>
#include <linux/string.h>

struct data_t {
    u32 pid;
    u32 count;
    u64 timestamp;
    u64 timecost;
    char comm[32];
    char name[128];
    char parentName[128];
};

BPF_HASH(start, u32, struct data_t);
BPF_PERF_OUTPUT(events);

int kprobe__rtnl_lock(struct pt_regs *ctx) {
    struct data_t data = {};
    bpf_get_current_comm(&data.comm, sizeof(data.comm));

    for (int i = 0; i < 14; i++) {
        if (data.comm[i] != "containerd-shim"[i]) {
            return 0;
        }
    }

    u32 pid = bpf_get_current_pid_tgid() >> 32;
    u64 ts = bpf_ktime_get_ns();

    data.timestamp = ts;
    data.pid = pid;
    start.update(&pid, &data);

    return 0;
}

int kretprobe__rtnl_unlock(struct pt_regs *ctx) {
    u32 pid = bpf_get_current_pid_tgid() >> 32;
    struct data_t *datap;
    u64 duration;

    datap = start.lookup(&pid);
    if (datap != 0) {
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
    print("spinlock() PID: %d Duration: %d ns Comm: %s Name: %s" % (event.pid, event.timestamp, event.comm, event.parentName))

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
