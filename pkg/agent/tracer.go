package agent

import (
	pb "github.com/moolen/secco/proto"
)

// RunTrace ..
func (o *AgentServer) RunTrace(req *pb.RunTraceRequest, gfs pb.Agent_RunTraceServer) error {

	for {
		select {
		case <-gfs.Context().Done():
			return gfs.Context().Err()
		default:
		}

		err := gfs.Send(&pb.RunTraceResponse{})
		if err != nil {
			return err
		}
	}
}
