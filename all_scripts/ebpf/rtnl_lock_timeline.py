from bcc import BPF

bpf_program = """
#include <uapi/linux/ptrace.h>
#include <linux/sched.h> // for TASK_COMM_LEN
#include <linux/kernfs.h>
#include <linux/string.h>
//#include <linux/stacktrace.h>

struct data_t {
    u32 pid;
    u32 count;
    u64 timestamp;
    u64 timecost;
    char comm[32];
    char name[128];
    char parentName[128];
    int kernel_stack_id;
};

BPF_HASH(start, u32, struct data_t);
BPF_STACK_TRACE(stack_traces, 4096); // Increase stack trace storage capacity
BPF_PERF_OUTPUT(events);

int kprobe__rtnl_lock(struct pt_regs *ctx) {
    struct data_t data = {};
    bpf_get_current_comm(&data.comm, sizeof(data.comm));

    /*
    for (int i = 0; i < 14; i++) {
        if (data.comm[i] != "containerd-shim"[i]) {
            return 0;
        }
    }
    */

    u32 pid = bpf_get_current_pid_tgid() >> 32;
    u64 ts = bpf_ktime_get_ns();

    // Getting the stack trace
    data.kernel_stack_id = stack_traces.get_stackid(ctx, BPF_F_REUSE_STACKID);

    data.timestamp = ts;
    data.pid = pid;
    start.update(&pid, &data);

    return 0;
}

int kretprobe__rtnl_lock(struct pt_regs *ctx) {
    u32 pid = bpf_get_current_pid_tgid() >> 32;
    struct data_t *datap;
    u64 duration;

    datap = start.lookup(&pid);
    if (datap != 0) {
        duration = bpf_ktime_get_ns() - datap->timestamp;
        datap->timecost = duration;
        events.perf_submit(ctx, datap, sizeof(struct data_t));
        start.delete(&pid);
    }
    return 0;
}
"""

b = BPF(text=bpf_program)

def print_event(cpu, data, size):
    event = b["events"].event(data)
    if event.kernel_stack_id >= 0:
        # Retrieve stack trace
        stack_trace = list(b["stack_traces"].walk(event.kernel_stack_id))
        # Symbolize the stack trace
        stack_symbols = [b.ksym(addr) for addr in stack_trace]
        parent_function = stack_symbols[1] if len(stack_symbols) > 1 else "[unknown]"
        print("rtnllock() PID: %d Duration: %d-%d ns Comm: %s Name: %s Caller: %s" % (event.pid, event.timestamp, event.timecost, event.comm, event.parentName, parent_function))
        print(stack_symbols)
    else:
        print("rtnllock() PID: %d Duration: %d-%d ns Comm: %s Name: %s" % (event.pid, event.timestamp, event.timecost, event.comm, event.parentName))

    #print("rtnllock() PID: %d Duration: %d-%d ns Comm: %s Name: %s" % (event.pid, event.timestamp, event.timecost, event.comm, event.parentName))

b["events"].open_perf_buffer(print_event)

print("Tracing ... Hit Ctrl-C to end.")

while True:
    try:
        b.perf_buffer_poll()
    except KeyboardInterrupt:
        break
