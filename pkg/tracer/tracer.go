package tracer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

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
func StartForDockerID(id string, stop <-chan struct{}, out chan map[string]int64) (map[string]int64, error) {
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
	log.Infof("found netns: %s", netnsid)
	return Start(s.Ino, stop, out)
}

// Start starts a trace for the provided netns id
// get nets from `sudo lsns -t net` or using vishvananda/netns
// see nsproxy.h, net_namespace.h and ns_common.h in kernel headers
func Start(nsid uint64, stop <-chan struct{}, out chan map[string]int64) (map[string]int64, error) {
	log.Infof("Started")
	syscalls := make(map[string]int64, 303)
	mx := &sync.Mutex{}
	src := strings.Replace(source, "$TRACE_NS", strconv.FormatUint(nsid, 10), -1)
	m := bcc.NewModule(src, []string{})
	defer m.Close()
	enterTrace, err := m.LoadTracepoint("enter_trace")
	if err != nil {
		return syscalls, errors.Wrap(err, "error loading tracepoint")
	}
	if err := m.AttachTracepoint("raw_syscalls:sys_enter", enterTrace); err != nil {
		return syscalls, errors.Wrap(err, "error attaching to tracepoint")
	}
	table := bcc.NewTable(m.TableId("events"), m)
	channel := make(chan []byte)

	perfMap, err := bcc.InitPerfMap(table, channel)
	if err != nil {
		return syscalls, errors.Wrap(err, "error initializing perf map")
	}
	go perfMap.Start()
	var e event

	go func() {
		for {
			mx.Lock()
			out <- syscalls
			mx.Unlock()
			<-time.After(time.Second * 2)
		}
	}()

readLoop:
	for {
		select {
		case <-stop:
			log.Infof("got stop signal, returning from trace")
			close(out)
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
			mx.Lock()
			syscalls[name]++
			mx.Unlock()
			log.Debugf("recorded evt %v", e)
		}
	}

	perfMap.Stop()
	return syscalls, nil
}
