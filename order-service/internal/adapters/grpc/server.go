// order-service/internal/adapters/grpc/server.go

package grpc

import (
	"fmt"
	"net"
	"order-service/internal/domain/interfaces"

	"github.com/jakkapat-chongsuwat/go-microservice/proto/order_service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func StartGRPCServer(port string, orderUseCase interfaces.IOrderUseCase, logger *zap.Logger) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	grpcServer := grpc.NewServer()

	order_service.RegisterOrderServiceServer(
		grpcServer,
		NewOrderGRPCServer(orderUseCase, logger),
	)

	// for testing purpose
	reflection.Register(grpcServer)

	logger.Info("Starting OrderService gRPC server", zap.String("port", port))
	return grpcServer.Serve(lis)
}
