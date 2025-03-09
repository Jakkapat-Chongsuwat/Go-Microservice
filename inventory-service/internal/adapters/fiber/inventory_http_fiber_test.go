package fiber_http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"inventory-service/internal/adapters/models"
	"inventory-service/internal/domain"
	"inventory-service/internal/usecases"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// FakeInventoryUseCase implements the InventoryUseCase interface for testing.
type FakeInventoryUseCase struct {
	CreateProductFunc              func(ctx context.Context, product *domain.Product) (*domain.Product, error)
	GetProductFunc                 func(ctx context.Context, productID string) (*domain.Product, error)
	UpdateProductMetadataFunc      func(ctx context.Context, product *domain.Product) (*domain.Product, error)
	UpdateProductStockQuantityFunc func(ctx context.Context, productID string, quantityChange int) (*domain.Product, error)
	ListProductsFunc               func(ctx context.Context) ([]*domain.Product, error)
}

var _ usecases.InventoryUseCase = (*FakeInventoryUseCase)(nil)

func (f *FakeInventoryUseCase) CreateProduct(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	return f.CreateProductFunc(ctx, product)
}

func (f *FakeInventoryUseCase) GetProduct(ctx context.Context, productID string) (*domain.Product, error) {
	return f.GetProductFunc(ctx, productID)
}

func (f *FakeInventoryUseCase) UpdateProductMetadata(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	return f.UpdateProductMetadataFunc(ctx, product)
}

func (f *FakeInventoryUseCase) UpdateProductStockQuantity(ctx context.Context, productID string, quantityChange int) (*domain.Product, error) {
	return f.UpdateProductStockQuantityFunc(ctx, productID, quantityChange)
}

func (f *FakeInventoryUseCase) ListProducts(ctx context.Context) ([]*domain.Product, error) {
	return f.ListProductsFunc(ctx)
}

func TestCreateProduct_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeInventoryUseCase{
		CreateProductFunc: func(ctx context.Context, product *domain.Product) (*domain.Product, error) {
			product.ID = "prod123"
			return product, nil
		},
	}
	app := fiber.New()
	handler := NewInventoryHTTPHandler(fakeUC, logger)
	RegisterInventoryRoutes(app, handler)

	reqPayload := models.CreateProductRequest{
		Name:     "Widget",
		Quantity: 100,
		Price:    9.99,
	}
	body, err := json.Marshal(reqPayload)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/products", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var productResp models.ProductResponse
	err = json.NewDecoder(resp.Body).Decode(&productResp)
	assert.NoError(t, err)
	assert.Equal(t, "prod123", productResp.ID)
	assert.Equal(t, reqPayload.Name, productResp.Name)
	assert.Equal(t, reqPayload.Quantity, productResp.Quantity)
	assert.Equal(t, reqPayload.Price, productResp.Price)
}

func TestCreateProduct_InvalidPayload(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeInventoryUseCase{
		CreateProductFunc: func(ctx context.Context, product *domain.Product) (*domain.Product, error) {
			return nil, nil
		},
	}
	app := fiber.New()
	handler := NewInventoryHTTPHandler(fakeUC, logger)
	RegisterInventoryRoutes(app, handler)

	req := httptest.NewRequest("POST", "/api/products", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetProduct_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeInventoryUseCase{
		GetProductFunc: func(ctx context.Context, productID string) (*domain.Product, error) {
			return &domain.Product{
				ID:       productID,
				Name:     "Widget",
				Quantity: 100,
				Price:    9.99,
			}, nil
		},
	}
	app := fiber.New()
	handler := NewInventoryHTTPHandler(fakeUC, logger)
	RegisterInventoryRoutes(app, handler)

	req := httptest.NewRequest("GET", "/api/products/prod123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var productResp models.ProductResponse
	err = json.NewDecoder(resp.Body).Decode(&productResp)
	assert.NoError(t, err)
	assert.Equal(t, "prod123", productResp.ID)
	assert.Equal(t, "Widget", productResp.Name)
}

func TestGetProduct_NotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeInventoryUseCase{
		GetProductFunc: func(ctx context.Context, productID string) (*domain.Product, error) {
			return nil, errors.New("product not found")
		},
	}
	app := fiber.New()
	handler := NewInventoryHTTPHandler(fakeUC, logger)
	RegisterInventoryRoutes(app, handler)

	req := httptest.NewRequest("GET", "/api/products/nonexistent", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestUpdateProductMetadata_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeInventoryUseCase{
		GetProductFunc: func(ctx context.Context, productID string) (*domain.Product, error) {
			return &domain.Product{
				ID:       productID,
				Name:     "Widget",
				Quantity: 100,
				Price:    9.99,
			}, nil
		},
		UpdateProductMetadataFunc: func(ctx context.Context, product *domain.Product) (*domain.Product, error) {
			return product, nil
		},
	}
	app := fiber.New()
	handler := NewInventoryHTTPHandler(fakeUC, logger)
	RegisterInventoryRoutes(app, handler)

	reqPayload := models.UpdateProductMetadataRequest{
		Name:  "Updated Widget",
		Price: 12.99,
	}
	body, err := json.Marshal(reqPayload)
	assert.NoError(t, err)

	req := httptest.NewRequest("PUT", "/api/products/prod123/metadata", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedResp models.ProductResponse
	err = json.NewDecoder(resp.Body).Decode(&updatedResp)
	assert.NoError(t, err)
	assert.Equal(t, "prod123", updatedResp.ID)
	assert.Equal(t, reqPayload.Name, updatedResp.Name)
	assert.Equal(t, reqPayload.Price, updatedResp.Price)
}

func TestUpdateProductMetadata_Failure(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeInventoryUseCase{
		GetProductFunc: func(ctx context.Context, productID string) (*domain.Product, error) {
			return &domain.Product{
				ID:       productID,
				Name:     "Widget",
				Quantity: 100,
				Price:    9.99,
			}, nil
		},
		UpdateProductMetadataFunc: func(ctx context.Context, product *domain.Product) (*domain.Product, error) {
			return nil, errors.New("update failed")
		},
	}
	app := fiber.New()
	handler := NewInventoryHTTPHandler(fakeUC, logger)
	RegisterInventoryRoutes(app, handler)

	reqPayload := models.UpdateProductMetadataRequest{
		Name:  "Updated Widget",
		Price: 12.99,
	}
	body, err := json.Marshal(reqPayload)
	assert.NoError(t, err)

	req := httptest.NewRequest("PUT", "/api/products/prod123/metadata", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestUpdateProductStockQuantity_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeInventoryUseCase{
		UpdateProductStockQuantityFunc: func(ctx context.Context, productID string, quantityChange int) (*domain.Product, error) {
			return &domain.Product{
				ID:       productID,
				Name:     "Widget",
				Quantity: 90,
				Price:    9.99,
			}, nil
		},
	}
	app := fiber.New()
	handler := NewInventoryHTTPHandler(fakeUC, logger)
	RegisterInventoryRoutes(app, handler)

	payload := models.UpdateProductStockQuantityRequest{
		QuantityChange: -10,
	}
	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	req := httptest.NewRequest("PATCH", "/api/products/prod123/quantity", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updatedResp models.ProductResponse
	err = json.NewDecoder(resp.Body).Decode(&updatedResp)
	assert.NoError(t, err)
	assert.Equal(t, "prod123", updatedResp.ID)
	assert.Equal(t, 90, updatedResp.Quantity)
}

func TestUpdateProductStockQuantity_Failure(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeInventoryUseCase{
		UpdateProductStockQuantityFunc: func(ctx context.Context, productID string, quantityChange int) (*domain.Product, error) {
			return nil, errors.New("stock update failed")
		},
	}
	app := fiber.New()
	handler := NewInventoryHTTPHandler(fakeUC, logger)
	RegisterInventoryRoutes(app, handler)

	payload := models.UpdateProductStockQuantityRequest{
		QuantityChange: -10,
	}
	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	req := httptest.NewRequest("PATCH", "/api/products/prod123/quantity", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestListProducts_Success(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeInventoryUseCase{
		ListProductsFunc: func(ctx context.Context) ([]*domain.Product, error) {
			return []*domain.Product{
				{ID: "prod1", Name: "Widget", Quantity: 100, Price: 9.99},
				{ID: "prod2", Name: "Gadget", Quantity: 50, Price: 19.99},
			}, nil
		},
	}
	app := fiber.New()
	handler := NewInventoryHTTPHandler(fakeUC, logger)
	RegisterInventoryRoutes(app, handler)

	req := httptest.NewRequest("GET", "/api/products", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var productsResp []models.ProductResponse
	err = json.NewDecoder(resp.Body).Decode(&productsResp)
	assert.NoError(t, err)
	assert.Len(t, productsResp, 2)
	assert.Equal(t, "prod1", productsResp[0].ID)
	assert.Equal(t, "prod2", productsResp[1].ID)
}

func TestListProducts_Failure(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	fakeUC := &FakeInventoryUseCase{
		ListProductsFunc: func(ctx context.Context) ([]*domain.Product, error) {
			return nil, errors.New("failed to list products")
		},
	}
	app := fiber.New()
	handler := NewInventoryHTTPHandler(fakeUC, logger)
	RegisterInventoryRoutes(app, handler)

	req := httptest.NewRequest("GET", "/api/products", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
