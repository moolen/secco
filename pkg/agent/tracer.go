package agent

import (
	"context"
	"time"

	"github.com/moolen/secco/pkg/tracer"
	pb "github.com/moolen/secco/proto"
	log "github.com/sirupsen/logrus"
)

// RunTrace ..
func (o *AgentServer) RunTrace(req *pb.RunTraceRequest, gfs pb.Agent_RunTraceServer) error {
	reqID := req.GetId()
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	callChan, err := tracer.StartForDockerID(reqID, ctx)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		log.Infof("reading callChan")
		for calls := range callChan {
			log.Infof("sending calls go server")
			err := gfs.Send(&pb.RunTraceResponse{
				Syscalls: calls,
			})
			if err != nil {
				log.Error(err)
			}
		}
		log.Infof("call chan returned")
	}()
	for {
		select {
		case <-gfs.Context().Done():
			cancel()
			return gfs.Context().Err()
		default:
		}
	}
}
