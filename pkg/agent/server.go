package agent

import (
	"context"
	"fmt"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	pb "github.com/moolen/secco/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// AgentServer ..
type AgentServer struct {
	profileBasePath string
	listener        net.Listener
	server          *grpc.Server
}

// NewAgentServer ..
func NewAgentServer(profileBasePath string) (*AgentServer, error) {

	as := &AgentServer{
		profileBasePath: profileBasePath,
		server: grpc.NewServer(
			grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
			grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		),
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", 3000))
	if err != nil {
		return nil, err
	}
	as.listener = listener
	pb.RegisterAgentServer(as.server, as)
	return as, nil
}

// Serve ..
func (srv *AgentServer) Serve(ctx context.Context) {
	log.Infof("serve")
	log.Infof("grpc listening on :%d", 3000)
	srv.server.Serve(srv.listener)
}

// Stop ..
func (srv *AgentServer) Stop() {
	log.Infof("stop")
	srv.server.GracefulStop()
}
