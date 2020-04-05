package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	client "github.com/moolen/secco/pkg/client"
	pb "github.com/moolen/secco/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Server struct {
	listener net.Listener
	server   *grpc.Server
	agent    *client.AgentClient
}

func New(target string, port int, syncInterval time.Duration, bufferSize int) (*Server, error) {
	agent, err := client.New(target)
	if err != nil {
		return nil, err
	}
	server := &Server{
		agent: agent,
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server.listener = listener
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	server.server = grpcServer
	return server, nil
}

func (srv *Server) Serve(ctx context.Context) {
	log.Infof("serve")

	go func() {
		// testing
		c, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
		rc, err := srv.agent.RunTrace(c, &pb.RunTraceRequest{
			Id: os.Getenv("DID"),
		})
		if err != nil {
			log.Fatal(err)
		}
		for {
			val, err := rc.Recv()
			if err != nil {
				log.Error(err)
				return
			}
			log.Infof("received: %v", val)
		}
	}()

	err := srv.server.Serve(srv.listener)
	if err != nil {
		log.Fatal(err)
	}

}

func (srv *Server) Stop() {
	log.Infof("stop")
	srv.server.GracefulStop()
}
