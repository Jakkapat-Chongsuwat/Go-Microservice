package grpc

import (
	"context"
	"inventory-service/internal/domain"
	"inventory-service/internal/usecases"

	"github.com/jakkapat-chongsuwat/go-microservice/proto/inventory_service"
	"go.uber.org/zap"
)

type InventoryGRPCServer struct {
	inventory_service.UnimplementedInventoryServiceServer
	inventoryUseCase usecases.InventoryUseCase
	logger           *zap.Logger
}

func NewInventoryGRPCServer(inventoryUC usecases.InventoryUseCase, logger *zap.Logger) *InventoryGRPCServer {
	return &InventoryGRPCServer{
		inventoryUseCase: inventoryUC,
		logger:           logger,
	}
}

func (s *InventoryGRPCServer) CreateProduct(ctx context.Context, req *inventory_service.CreateProductRequest) (*inventory_service.CreateProductResponse, error) {
	s.logger.Info("Received CreateProduct request", zap.String("name", req.GetName()))
	product := domain.NewProduct(req.GetName(), int(req.GetQuantity()), req.GetPrice())
	created, err := s.inventoryUseCase.CreateProduct(ctx, product)
	if err != nil {
		s.logger.Error("Failed to create product", zap.Error(err))
		return nil, err
	}

	return &inventory_service.CreateProductResponse{
		Product: &inventory_service.Product{
			Id:       created.ID,
			Name:     created.Name,
			Quantity: int32(created.Quantity),
			Price:    created.Price,
		},
	}, nil
}

func (s *InventoryGRPCServer) GetProduct(ctx context.Context, req *inventory_service.GetProductRequest) (*inventory_service.GetProductResponse, error) {
	s.logger.Info("Received GetProduct request", zap.String("id", req.GetId()))
	product, err := s.inventoryUseCase.GetProduct(ctx, req.GetId())
	if err != nil {
		s.logger.Error("Failed to get product", zap.String("id", req.GetId()), zap.Error(err))
		return nil, err
	}

	return &inventory_service.GetProductResponse{
		Product: &inventory_service.Product{
			Id:       product.ID,
			Name:     product.Name,
			Quantity: int32(product.Quantity),
			Price:    product.Price,
		},
	}, nil
}

func (s *InventoryGRPCServer) UpdateProductMetadata(ctx context.Context, req *inventory_service.UpdateProductMetadataRequest) (*inventory_service.UpdateProductMetadataResponse, error) {
	s.logger.Info("Received UpdateProductMetadata request", zap.String("id", req.GetId()))
	product, err := s.inventoryUseCase.GetProduct(ctx, req.GetId())
	if err != nil {
		s.logger.Error("Failed to get product", zap.String("id", req.GetId()), zap.Error(err))
		return nil, err
	}
	product.Name = req.GetName()
	product.Price = req.GetPrice()
	updated, err := s.inventoryUseCase.UpdateProductMetadata(ctx, product)
	if err != nil {
		s.logger.Error("Failed to update product metadata", zap.String("id", req.GetId()), zap.Error(err))
		return nil, err
	}
	return &inventory_service.UpdateProductMetadataResponse{
		Product: &inventory_service.Product{
			Id:       updated.ID,
			Name:     updated.Name,
			Quantity: int32(updated.Quantity),
			Price:    updated.Price,
		},
	}, nil
}

func (s *InventoryGRPCServer) UpdateProductStockQuantity(ctx context.Context, req *inventory_service.UpdateProductStockQuantityRequest) (*inventory_service.UpdateProductStockQuantityResponse, error) {
	s.logger.Info("Received AdjustInventory request", zap.String("id", req.GetId()))
	updated, err := s.inventoryUseCase.UpdateProductStockQuantity(ctx, req.GetId(), int(req.GetQuantityChange()))
	if err != nil {
		s.logger.Error("Failed to adjust inventory", zap.String("id", req.GetId()), zap.Error(err))
		return nil, err
	}
	return &inventory_service.UpdateProductStockQuantityResponse{
		Product: &inventory_service.Product{
			Id:       updated.ID,
			Name:     updated.Name,
			Quantity: int32(updated.Quantity),
			Price:    updated.Price,
		},
	}, nil
}

func (s *InventoryGRPCServer) ListProducts(ctx context.Context, req *inventory_service.ListProductsRequest) (*inventory_service.ListProductsResponse, error) {
	s.logger.Info("Received ListProducts request")
	products, err := s.inventoryUseCase.ListProducts(ctx)
	if err != nil {
		s.logger.Error("Failed to list products", zap.Error(err))
		return nil, err
	}

	var prodResponses []*inventory_service.Product
	for _, product := range products {
		prodResponses = append(prodResponses, &inventory_service.Product{
			Id:       product.ID,
			Name:     product.Name,
			Quantity: int32(product.Quantity),
			Price:    product.Price,
		})
	}
	return &inventory_service.ListProductsResponse{
		Products: prodResponses,
	}, nil
}
