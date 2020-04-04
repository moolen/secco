package agent

import (
	"github.com/moolen/secco/pkg/tracer"
	pb "github.com/moolen/secco/proto"
	log "github.com/sirupsen/logrus"
)

// RunTrace ..
func (o *AgentServer) RunTrace(req *pb.RunTraceRequest, gfs pb.Agent_RunTraceServer) error {
	reqID := req.GetId()
	stop := make(chan struct{})
	out := make(chan map[string]int64)
	calls, err := tracer.StartForDockerID(reqID, stop, out)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("calls: %v", calls)

	go func() {
		for calls := range out {
			err := gfs.Send(&pb.RunTraceResponse{
				Syscalls: calls,
			})
			if err != nil {
				log.Error(err)
			}
		}
	}()

	for {
		select {
		case <-gfs.Context().Done():
			stop <- struct{}{}
			return gfs.Context().Err()
		default:
		}
	}
}
