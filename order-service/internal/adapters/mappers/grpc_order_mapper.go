package mappers

import (
	"order-service/internal/adapters/models"
	"order-service/internal/domain"

	"github.com/jakkapat-chongsuwat/go-microservice/proto/order_service"
)

func ProtoOrderItemToModel(protoItem *order_service.OrderItem) models.OrderItemRequest {
	return models.OrderItemRequest{
		ProductID: protoItem.ProductId,
		Quantity:  int(protoItem.Quantity),
	}
}

func DomainOrderItemToProto(item domain.OrderItem) *order_service.OrderItem {
	return &order_service.OrderItem{
		ProductId: item.ProductID,
		Quantity:  int32(item.Quantity),
	}
}
