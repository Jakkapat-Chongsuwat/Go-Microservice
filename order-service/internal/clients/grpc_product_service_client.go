// internal/clients/product_service_client.go
package clients

import (
	"context"
)

// DummyProductServiceClient implements usecases.ProductServiceClient
// and always returns success.
type DummyProductServiceClient struct{}

func (d DummyProductServiceClient) VerifyProduct(ctx context.Context, productID string) error {
	// In a real scenario, you might add some logic to simulate failures.
	return nil
}
