package clients

import (
	"context"
	"fmt"

	"github.com/jakkapat-chongsuwat/go-microservice/proto/user_service"

	"google.golang.org/grpc"
)

type GRPCUserServiceClient struct {
	client user_service.UserServiceClient
}

func NewGRPCUserServiceClient(conn *grpc.ClientConn) *GRPCUserServiceClient {
	return &GRPCUserServiceClient{
		client: user_service.NewUserServiceClient(conn),
	}
}

func (c *GRPCUserServiceClient) VerifyUser(ctx context.Context, userID string) error {
	req := &user_service.GetUserRequest{Id: userID}
	resp, err := c.client.GetUserByID(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}
	if resp == nil || resp.Id == "" {
		return fmt.Errorf("user not found")
	}
	return nil
}
