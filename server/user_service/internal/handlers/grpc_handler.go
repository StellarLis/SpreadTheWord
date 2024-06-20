package handlers

import (
	"fmt"
	"net"
	"user_service/internal/grpc_service"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type GrpcHandler struct {
	gRPCServer *grpc.Server
	port       int
}

func NewGrpcHandler(port int, jwtSecretKey string) *GrpcHandler {
	gRPCServer := grpc.NewServer()

	grpc_service.Register(gRPCServer, jwtSecretKey)

	return &GrpcHandler{
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (g *GrpcHandler) MustRun() {
	if err := g.Run(); err != nil {
		panic(err)
	}
}

func (g *GrpcHandler) Run() error {
	const op = "handlers.grpc_handler.Run"

	log := logrus.WithFields(logrus.Fields{
		"op":   op,
		"port": g.port,
	})

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", g.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("gRPC server is running")

	if err := g.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (g *GrpcHandler) Stop() {
	const op = "handlers.grpc_handler.Stop"

	logrus.WithField("op", op).Info("stopping gRPC server")
	g.gRPCServer.GracefulStop()
}
