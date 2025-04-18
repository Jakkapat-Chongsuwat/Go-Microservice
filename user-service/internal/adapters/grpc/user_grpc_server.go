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

func (s *UserGRPcServer) CreateUser(ctx context.Context, req *user_service.CreateUserRequest) (*user_service.CreateUserResponse, error) {
	s.logger.Info("Received CreateUser request", zap.String("username", req.Username), zap.String("email", req.Email))
	user, err := s.userUseCase.CreateUser(ctx, req.Username, req.Email)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	return &user_service.CreateUserResponse{
		Id:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}
