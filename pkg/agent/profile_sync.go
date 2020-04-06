package agent

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	pb "github.com/moolen/secco/proto"
)

func (o *AgentServer) SyncProfile(ctx context.Context, req *pb.SyncProfileRequest) (*pb.SyncProfileResponse, error) {
	var errors []string
	profileBase, err := filepath.Abs(o.profileBasePath)
	if err != nil {
		return nil, err
	}
	for _, profile := range req.Profiles {
		dPath := path.Join(o.profileBasePath, profile.GetName())
		dDir, err := filepath.Abs(path.Dir(dPath))
		if err != nil {
			errors = append(errors, "invalid path configuration")
			continue

		}
		// check directory traversal attempt
		if !strings.HasPrefix(profileBase, dDir) {
			errors = append(errors, "dest dir not in base path")
			continue
		}
		err = os.MkdirAll(dDir, os.ModeDir)
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
