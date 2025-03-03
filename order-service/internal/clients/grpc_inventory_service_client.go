package clients

import (
	"context"
	"fmt"

	"github.com/jakkapat-chongsuwat/go-microservice/proto/inventory_service"
	"google.golang.org/grpc"
)

type GRPCInventoryServiceClient struct {
	client inventory_service.InventoryServiceClient
}

func NewGRPCInventoryServiceClient(conn *grpc.ClientConn) *GRPCInventoryServiceClient {
	return &GRPCInventoryServiceClient{
		client: inventory_service.NewInventoryServiceClient(conn),
	}
}

func (c *GRPCInventoryServiceClient) VerifyInventory(ctx context.Context, productID string, requiredQuantity int) error {
	req := &inventory_service.GetProductRequest{Id: productID}
	resp, err := c.client.GetProduct(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to verify inventory: %w", err)
	}
	if resp == nil || resp.Product == nil || resp.Product.Id == "" {
		return fmt.Errorf("product not found")
	}
	if int(resp.Product.Quantity) < requiredQuantity {
		return fmt.Errorf("insufficient stock: available %d, required %d", resp.Product.Quantity, requiredQuantity)
	}
	return nil
}
