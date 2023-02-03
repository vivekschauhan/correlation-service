package server

import (
	"fmt"
	"net"

	"github.com/Axway/agent-sdk/pkg/amplify/agent/correlation"
	"github.com/sirupsen/logrus"
	"github.com/vivekschauhan/correlation-service/pkg/config"
	"google.golang.org/grpc"
)

type Server interface {
	Start()
	Stop()
}

type server struct {
	logger     *logrus.Logger
	service    correlation.CorrelationServiceServer
	listener   net.Listener
	grpcServer *grpc.Server
}

func NewServer(cfg *config.Config, log *logrus.Logger, service correlation.CorrelationServiceServer) (Server, error) {
	grpcServer := grpc.NewServer()
	correlation.RegisterCorrelationServiceServer(grpcServer, service)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, err
	}

	svr := &server{
		logger:     log,
		service:    service,
		grpcServer: grpcServer,
		listener:   listener,
	}
	return svr, nil
}

func (s *server) Start() {
	if err := s.grpcServer.Serve(s.listener); err != nil {
		s.logger.Fatal("unable to start gRPC server", err)
	}
}

func (s *server) Stop() {

}
