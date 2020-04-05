package agent

import (
	"context"
	"sync"
	"time"

	"github.com/moolen/secco/pkg/tracer"
	pb "github.com/moolen/secco/proto"
	log "github.com/sirupsen/logrus"
)

// RunTrace starts a trace for the provided docker container id
func (o *AgentServer) RunTrace(req *pb.RunTraceRequest, gfs pb.Agent_RunTraceServer) error {
	reqID := req.GetId()
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	callChan, err := tracer.StartForDockerID(reqID, ctx)
	if err != nil {
		log.Fatal(err)
	}
	syscalls := make(map[string]int64)
	mu := sync.RWMutex{}

	// continuously push data to the client
	go func() {
		for {
			select {
			case <-time.After(time.Second * 2):
				mu.RLock()
				err := gfs.Send(&pb.RunTraceResponse{
					Syscalls: syscalls,
				})
				if err != nil {
					log.Error(err)
				}
				mu.RUnlock()
			case <-gfs.Context().Done():
				log.Infof("stopping push timer")
				return
			}
		}
	}()

	// read incoming syscalls from the tracer chan
	go func() {
		for calls := range callChan {
			mu.Lock()
			syscalls[calls]++
			mu.Unlock()
		}
	}()

	for {
		select {
		case <-gfs.Context().Done():
			cancel()
			return gfs.Context().Err()
		}
	}
}
