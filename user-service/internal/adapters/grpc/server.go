package grpc

import (
	"context"
	"fmt"
	"net"
	"user-service/internal/usecases"
	"user-service/proto/helloworld"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	helloworld.UnimplementedGreeterServer
	userUsecase usecases.UserUseCase
	logger      *zap.Logger
}

func NewServer(userUsecase usecases.UserUseCase, logger *zap.Logger) *Server {
	return &Server{
		userUsecase: userUsecase,
		logger:      logger,
	}
}

func (s *Server) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	s.logger.Info("SayHello called", zap.String("name", req.Name))
	reply := &helloworld.HelloReply{
		Message: fmt.Sprintf("Hello %s", req.Name),
	}
	s.logger.Debug("reply prepared", zap.String("message", reply.Message))
	return reply, nil
}

func StartGRPCServer(port string, userUsecase usecases.UserUseCase, logger *zap.Logger) error {
	address := ":" + port
	lis, err := net.Listen("tcp", address)
	if err != nil {
		logger.Error("failed to listen", zap.String("address", address), zap.Error(err))
		return fmt.Errorf("failed to listen: %w", err)
	}
	grpcServer := grpc.NewServer()
	helloworld.RegisterGreeterServer(grpcServer, NewServer(userUsecase, logger))
	logger.Info("Starting gRPC server on port", zap.String("port", port))
	return grpcServer.Serve(lis)
}
