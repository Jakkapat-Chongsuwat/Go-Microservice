package grpc

import (
	"fmt"
	"inventory-service/internal/usecases"
	"net"

	"github.com/jakkapat-chongsuwat/go-microservice/proto/inventory_service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func StartGRPCServer(port string, useCase usecases.InventoryUseCase, logger *zap.Logger) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	grpcServer := grpc.NewServer()

	inventory_service.RegisterInventoryServiceServer(grpcServer, NewInventoryGRPCServer(useCase, logger))

	// for testing purpose
	reflection.Register(grpcServer)

	logger.Info("Starting InventoryService gRPC server", zap.String("port", port))
	return grpcServer.Serve(lis)
}
