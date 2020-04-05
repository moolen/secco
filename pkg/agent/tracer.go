package agent

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/moolen/secco/pkg/tracer"
	pb "github.com/moolen/secco/proto"
	log "github.com/sirupsen/logrus"
)

// RunTrace starts a trace for the provided docker container id
func (o *AgentServer) RunTrace(req *pb.RunTraceRequest, gfs pb.Agent_RunTraceServer) error {
	reqID := req.GetId()
	log.Infof("req: %v", time.Duration(req.Duration))
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Duration))
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
			case <-ctx.Done():
				return
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
		case <-ctx.Done():
			cancel()
			return nil
		case <-gfs.Context().Done():
			cancel()
			return gfs.Context().Err()
		}
	}
}

func (o *AgentServer) SyncProfile(ctx context.Context, req *pb.SyncProfileRequest) (*pb.SyncProfileResponse, error) {
	var errors []string
	for _, profile := range req.Profiles {
		dPath := path.Join(o.profileBasePath, profile.GetName())
		dDir := path.Dir(dPath)
		// check directory traversal attempt
		if !strings.HasPrefix(o.profileBasePath, dDir) {
			errors = append(errors, "dest dir not in base path")
			continue
		}
		err := os.MkdirAll(dDir, os.ModeDir)
		if err != nil {
			errors = append(errors, "failed to mkdir")
			continue
		}
		err = ioutil.WriteFile(dPath, profile.Profile, os.ModePerm)
		if err != nil {
			errors = append(errors, "failed to write file")
		}
	}

	if len(errors) > 0 {
		return &pb.SyncProfileResponse{
			Success: false,
			Error:   strings.Join(errors, ", "),
		}, nil
	}

	return &pb.SyncProfileResponse{
		Success: true,
	}, nil
}
