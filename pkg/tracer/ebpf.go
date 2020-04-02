package tracer

const source string = `
#include <linux/bpf.h>
#include <linux/nsproxy.h>
#include <net/net_namespace.h>
#include <linux/ns_common.h>
#include <linux/sched.h>
#include <linux/tracepoint.h>

// Opens a custom BPF table to push data to user space via perf ring buffer
BPF_PERF_OUTPUT(events);

struct syscall_data {
    // PID of the process
    u32 pid;
    // the syscall number
	u32 id;
	// ns
    u32 ns;
};

struct syscall_enter_args {
	unsigned long		common_tp_fields;
	long				id;
	unsigned long		args[6];
};

// specification of args from /sys/kernel/debug/tracing/events/raw_syscalls/sys_enter/format
int enter_trace(struct syscall_enter_args* args)
{
    struct syscall_data data = {};
    struct task_struct* task;
    data.pid = bpf_get_current_pid_tgid();
    data.id = (int)args->id;
    task = (struct task_struct*)bpf_get_current_task();
    struct nsproxy* ns = task->nsproxy;
    data.ns = ns->net_ns->ns.inum;
	if(data.ns != $TRACE_NS){
		return 0;
	}
    events.perf_submit(args, &data, sizeof(data));
    return 0;
}
`
