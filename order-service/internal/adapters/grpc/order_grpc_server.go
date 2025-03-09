package grpc

import (
	"context"
	"order-service/internal/adapters/mappers"
	"order-service/internal/adapters/models"
	"order-service/internal/domain/interfaces"

	"github.com/jakkapat-chongsuwat/go-microservice/proto/order_service"
	"go.uber.org/zap"
)

type OrderGRPCServer struct {
	order_service.UnimplementedOrderServiceServer
	orderUseCase interfaces.IOrderUseCase
	logger       *zap.Logger
}

func NewOrderGRPCServer(orderUC interfaces.IOrderUseCase, logger *zap.Logger) *OrderGRPCServer {
	return &OrderGRPCServer{
		orderUseCase: orderUC,
		logger:       logger,
	}
}

func (s *OrderGRPCServer) CreateOrder(ctx context.Context, req *order_service.CreateOrderRequest) (*order_service.CreateOrderResponse, error) {
	s.logger.Info("Received CreateOrder request", zap.String("userID", req.UserId))

	var itemReqs []models.OrderItemRequest
	for _, protoItem := range req.Items {
		itemReq := mappers.ProtoOrderItemToModel(protoItem)
		itemReqs = append(itemReqs, itemReq)
	}

	createdOrder, err := s.orderUseCase.CreateOrder(ctx, req.UserId, itemReqs)
	if err != nil {
		s.logger.Error("CreateOrder failed", zap.Error(err))
		return nil, err
	}

	var protoItems []*order_service.OrderItem
	for _, di := range createdOrder.Items {
		protoItem := mappers.DomainOrderItemToProto(*di)
		protoItems = append(protoItems, protoItem)
	}

	resp := &order_service.CreateOrderResponse{
		OrderId: createdOrder.ID,
		UserId:  createdOrder.UserID,
		Status:  string(createdOrder.Status),
		Items:   protoItems,
	}

	return resp, nil
}

func (s *OrderGRPCServer) GetOrder(ctx context.Context, req *order_service.GetOrderRequest) (*order_service.GetOrderResponse, error) {
	s.logger.Info("Received GetOrder request", zap.String("orderID", req.OrderId))

	dOrder, err := s.orderUseCase.GetOrder(ctx, req.OrderId)
	if err != nil {
		s.logger.Error("GetOrder failed", zap.Error(err))
		return nil, err
	}

	var protoItems []*order_service.OrderItem
	for _, di := range dOrder.Items {
		protoItem := mappers.DomainOrderItemToProto(*di)
		protoItems = append(protoItems, protoItem)
	}

	pbOrder := &order_service.Order{
		OrderId: dOrder.ID,
		UserId:  dOrder.UserID,
		Status:  string(dOrder.Status),
		Items:   protoItems,
	}

	return &order_service.GetOrderResponse{Order: pbOrder}, nil
}
