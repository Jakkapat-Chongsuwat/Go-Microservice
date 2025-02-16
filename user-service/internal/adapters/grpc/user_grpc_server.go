// user-service\internal\adapters\grpc\user_grpc_server.go

package grpc

import (
	"context"
	"user-service/internal/usecases"

	"github.com/jakkapat-chongsuwat/go-microservice/proto/user_service"
	"go.uber.org/zap"
)

type UserGRPcServer struct {
	user_service.UnimplementedUserServiceServer
	userUseCase usecases.UserUseCase
	logger      *zap.Logger
}

func NewUserGRPCServer(u usecases.UserUseCase, logger *zap.Logger) *UserGRPcServer {
	return &UserGRPcServer{
		userUseCase: u,
		logger:      logger,
	}
}

func (s *UserGRPcServer) GetUserByID(ctx context.Context, req *user_service.GetUserRequest) (*user_service.GetUserResponse, error) {
	s.logger.Info("Received GetUserById request", zap.String("id", req.Id))

	user, err := s.userUseCase.GetUserInParallel(ctx, []string{req.Id})

	if err != nil {
		s.logger.Error("Failed to get user", zap.String("id", req.Id), zap.Error(err))
		return nil, err
	}

	return &user_service.GetUserResponse{
		Id:       user[0].ID,
		Username: user[0].Username,
		Email:    user[0].Email,
	}, nil
}
