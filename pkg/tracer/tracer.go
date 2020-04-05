package tracer

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"syscall"

	"github.com/iovisor/gobpf/bcc"
	"github.com/pkg/errors"
	seccomp "github.com/seccomp/libseccomp-golang"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"
)

// event struct used to read data from the perf ring buffer
type event struct {
	// PID of the process making the syscall
	Pid uint32
	// syscall number
	ID uint32
	// NS id
	NS uint32
}

// StartForDockerID starts a tracer for the specified docker container id
func StartForDockerID(id string, ctx context.Context) (chan string, error) {
	ns, err := netns.GetFromDocker(id)
	if err != nil {
		return nil, fmt.Errorf("could not get docker id ns: %s", err)
	}
	var s syscall.Stat_t
	err = syscall.Fstat(int(ns), &s)
	if err != nil {

		return nil, fmt.Errorf("could not fstat ns: %s", err)
	}
	netnsid := ns.UniqueId()
	log.Debugf("found netns: %s for %s", netnsid, id)
	return Start(s.Ino, ctx)
}

// Start starts a trace for the provided netns id
// get nets from `sudo lsns -t net` or using vishvananda/netns
// see nsproxy.h, net_namespace.h and ns_common.h in kernel headers
func Start(nsid uint64, ctx context.Context) (chan string, error) {
	log.Infof("starting tracer for %d", nsid)
	out := make(chan string)
	syscalls := make(map[string]int64, 303)
	src := strings.Replace(source, "$TRACE_NS", strconv.FormatUint(nsid, 10), -1)
	m := bcc.NewModule(src, []string{})
	enterTrace, err := m.LoadTracepoint("enter_trace")
	if err != nil {
		return out, errors.Wrap(err, "error loading tracepoint")
	}
	if err := m.AttachTracepoint("raw_syscalls:sys_enter", enterTrace); err != nil {
		return out, errors.Wrap(err, "error attaching to tracepoint")
	}
	table := bcc.NewTable(m.TableId("events"), m)
	channel := make(chan []byte)
	perfMap, err := bcc.InitPerfMap(table, channel)
	if err != nil {
		return out, errors.Wrap(err, "error initializing perf map")
	}
	perfMap.Start()

	go func() {
		var e event
	readLoop:
		for {
			select {
			case <-ctx.Done():
				log.Debugf("got stop signal, returning from trace")
				close(out)
				m.Close()
				break readLoop
			case data := <-channel:
				err := binary.Read(bytes.NewBuffer(data), binary.LittleEndian, &e)
				if err != nil {
					log.Errorf("failed to decode received data '%s': %s\n", data, err)
					continue
				}
				name, err := seccomp.ScmpSyscall(e.ID).GetName()
				if err != nil {
					log.Errorf("error getting the name for syscall ID %d", e.ID)
					continue
				}
				syscalls[name]++
				out <- name
				log.Debugf("wrote evt %v", e)
			}
		}
		perfMap.Stop()
		log.Debugf("stopped perf map")
	}()

	return out, nil
}
