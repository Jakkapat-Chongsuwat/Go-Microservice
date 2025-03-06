package fiber_http

import (
	"order-service/internal/adapters/models"
	"order-service/internal/domain"
	"order-service/internal/domain/interfaces"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type OrderHTTPHandler struct {
	orderUseCase interfaces.IOrderUseCase
	logger       *zap.Logger
}

func NewOrderHTTPHandler(orderUC interfaces.IOrderUseCase, logger *zap.Logger) *OrderHTTPHandler {
	return &OrderHTTPHandler{
		orderUseCase: orderUC,
		logger:       logger,
	}
}

func (h *OrderHTTPHandler) CreateOrder(c *fiber.Ctx) error {
	h.logger.Info("CreateOrder endpoint called")

	var req models.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId is required",
		})
	}

	order := domain.NewOrder(req.UserID)

	var orderItems []*domain.OrderItem
	for _, item := range req.Items {
		if item.ProductID == "" || item.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid item data",
			})
		}
		orderItem := domain.NewOrderItem(item.ProductID, item.Quantity)
		orderItems = append(orderItems, orderItem)
	}

	order.Items = orderItems

	ctx := c.UserContext()

	createdOrder, err := h.orderUseCase.CreateOrderWithItems(ctx, order, orderItems)
	if err != nil {
		h.logger.Error("CreateOrderWithItems failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var itemsResp []models.OrderItemResponse
	for _, item := range createdOrder.Items {
		itemsResp = append(itemsResp, models.OrderItemResponse{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	resp := models.CreateOrderResponse{
		OrderID: createdOrder.ID,
		UserID:  createdOrder.UserID,
		Status:  string(createdOrder.Status),
		Items:   itemsResp,
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func RegisterOrderRoutes(app *fiber.App, orderHandler *OrderHTTPHandler) {
	api := app.Group("/api")
	api.Post("/orders", orderHandler.CreateOrder)
}
