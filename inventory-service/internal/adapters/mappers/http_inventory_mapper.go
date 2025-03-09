package mappers

import (
	"inventory-service/internal/adapters/models"
	"inventory-service/internal/domain"
)

func MapCreateProductRequestToProduct(dto models.CreateProductRequest) *domain.Product {
	return domain.NewProduct(dto.Name, dto.Quantity, dto.Price)
}

func MapProductToProductResponse(product *domain.Product) models.ProductResponse {
	return models.ProductResponse{
		ID:       product.ID,
		Name:     product.Name,
		Quantity: product.Quantity,
		Price:    product.Price,
	}
}

func MapUpdateProductMetadataRequestToProduct(dto models.UpdateProductMetadataRequest, product *domain.Product) {
	product.Name = dto.Name
	product.Price = dto.Price
}

func MapUpdateProductStockQuantityRequestToProduct(dto models.UpdateProductStockQuantityRequest, product *domain.Product) {
	product.Quantity += dto.QuantityChange
}
