package grpc

import (
	"context"
	"order-service/internal/domain"
	"order-service/internal/usecases"
	"order-service/proto/order_service"

	"go.uber.org/zap"
)

type OrderGRPCServer struct {
	order_service.UnimplementedOrderServiceServer
	orderUseCase usecases.OrderUseCase
	logger       *zap.Logger
}

func NewOrderGRPCServer(orderUC usecases.OrderUseCase, logger *zap.Logger) *OrderGRPCServer {
	return &OrderGRPCServer{
		orderUseCase: orderUC,
		logger:       logger,
	}
}

func (s *OrderGRPCServer) CreateOrder(ctx context.Context, req *order_service.CreateOrderRequest) (*order_service.CreateOrderResponse, error) {
	s.logger.Info("Received CreateOrder request", zap.String("userID", req.UserId))

	dOrder := &domain.Order{
		UserID: req.UserId,
	}

	dItems := make([]*domain.OrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		dItems = append(dItems, &domain.OrderItem{
			ProductID: it.ProductId,
			Quantity:  int(it.Quantity),
		})
	}

	createdOrder, err := s.orderUseCase.CreateOrderWithItems(ctx, dOrder, dItems)
	if err != nil {
		s.logger.Error("CreateOrderWithItems failed", zap.Error(err))
		return nil, err
	}

	protoItems := make([]*order_service.OrderItem, 0, len(createdOrder.Items))
	for _, di := range createdOrder.Items {
		protoItems = append(protoItems, &order_service.OrderItem{
			ProductId: di.ProductID,
			Quantity:  int32(di.Quantity),
		})
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

	protoItems := make([]*order_service.OrderItem, 0, len(dOrder.Items))
	for _, di := range dOrder.Items {
		protoItems = append(protoItems, &order_service.OrderItem{
			ProductId: di.ProductID,
			Quantity:  int32(di.Quantity),
		})
	}

	pbOrder := &order_service.Order{
		OrderId: dOrder.ID,
		UserId:  dOrder.UserID,
		Status:  string(dOrder.Status),
		Items:   protoItems,
	}

	return &order_service.GetOrderResponse{Order: pbOrder}, nil
}
