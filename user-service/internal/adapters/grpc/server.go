// user-service\internal\adapters\grpc\server.go

package grpc

import (
	"fmt"
	"net"
	"user-service/internal/usecases"

	"github.com/jakkapat-chongsuwat/go-microservice/proto/user_service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func StartGRPCServer(port string, userUseCase usecases.UserUseCase, logger *zap.Logger) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()

	user_service.RegisterUserServiceServer(grpcServer, NewUserGRPCServer(userUseCase, logger))

	logger.Info("Starting UserService gRPC server", zap.String("port", port))
	return grpcServer.Serve(lis)
}
