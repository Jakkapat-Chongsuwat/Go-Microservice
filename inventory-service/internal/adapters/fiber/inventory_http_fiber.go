package fiber_http

import (
	"inventory-service/internal/domain"
	"inventory-service/internal/usecases"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type CreateProductRequest struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type UpdateProductMetadataRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type UpdateProductStockQuantityRequest struct {
	QuantityChange int `json:"quantityChange"`
}

type InventoryHTTPHandler struct {
	inventoryUseCase usecases.InventoryUseCase
	logger           *zap.Logger
}

func NewInventoryHTTPHandler(inventoryUC usecases.InventoryUseCase, logger *zap.Logger) *InventoryHTTPHandler {
	return &InventoryHTTPHandler{
		inventoryUseCase: inventoryUC,
		logger:           logger,
	}
}

func (h *InventoryHTTPHandler) CreateProduct(c *fiber.Ctx) error {
	h.logger.Info("CreateProduct endpoint called")
	var req CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("failed to parse request", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}
	if req.Name == "" || req.Quantity < 0 || req.Price < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid product data"})
	}

	product := domain.NewProduct(req.Name, req.Quantity, req.Price)
	created, err := h.inventoryUseCase.CreateProduct(c.Context(), product)
	if err != nil {
		h.logger.Error("failed to create product", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create product"})
	}
	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *InventoryHTTPHandler) GetProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Product ID is required"})
	}
	product, err := h.inventoryUseCase.GetProduct(c.Context(), id)
	if err != nil {
		h.logger.Error("failed to get product", zap.String("id", id), zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}
	return c.JSON(product)
}

func (h *InventoryHTTPHandler) UpdateProductMetadata(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Product ID is required"})
	}
	var req UpdateProductMetadataRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("failed to parse request", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}
	product, err := h.inventoryUseCase.GetProduct(c.Context(), id)
	if err != nil {
		h.logger.Error("failed to get product", zap.String("id", id), zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Product not found"})
	}
	product.Name = req.Name
	product.Price = req.Price
	updated, err := h.inventoryUseCase.UpdateProductMetadata(c.Context(), product)
	if err != nil {
		h.logger.Error("failed to update product metadata", zap.String("id", id), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update product metadata"})
	}
	return c.JSON(updated)
}

func (h *InventoryHTTPHandler) UpdateProductStockQuantity(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Product ID is required"})
	}
	var req UpdateProductStockQuantityRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("failed to parse request", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}
	updated, err := h.inventoryUseCase.UpdateProductStockQuantity(c.Context(), id, req.QuantityChange)
	if err != nil {
		h.logger.Error("failed to update product stock quantity", zap.String("id", id), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update product stock quantity"})
	}
	return c.JSON(updated)
}

func (h *InventoryHTTPHandler) ListProducts(c *fiber.Ctx) error {
	products, err := h.inventoryUseCase.ListProducts(c.Context())
	if err != nil {
		h.logger.Error("failed to list products", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list products"})
	}
	return c.JSON(products)
}

func RegisterInventoryRoutes(app *fiber.App, handler *InventoryHTTPHandler) {
	api := app.Group("/api")
	api.Post("/products", handler.CreateProduct)
	api.Get("/products", handler.ListProducts)
	api.Get("/products/:id", handler.GetProduct)
	api.Put("/products/:id/metadata", handler.UpdateProductMetadata)
	api.Patch("/products/:id/quantity", handler.UpdateProductStockQuantity)
}
