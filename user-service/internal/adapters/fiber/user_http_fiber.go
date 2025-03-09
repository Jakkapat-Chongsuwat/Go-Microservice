package fiber_http

import (
	"user-service/internal/adapters/models"
	"user-service/internal/usecases"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type UserHTTPHandler struct {
	userUseCase usecases.UserUseCase
	logger      *zap.Logger
}

func NewUserHttpHandler(userUC usecases.UserUseCase, logger *zap.Logger) *UserHTTPHandler {
	return &UserHTTPHandler{
		userUseCase: userUC,
		logger:      logger,
	}
}

func (u *UserHTTPHandler) CreateUser(c *fiber.Ctx) error {
	u.logger.Info("CreateUser endpoint called")

	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		u.logger.Error("failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request payload",
		})
	}

	if req.Username == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username and email are required",
		})
	}

	ctx := c.UserContext()

	user, err := u.userUseCase.CreateUser(ctx, req.Username, req.Email)
	if err != nil {
		u.logger.Error("failed to create user", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func RegisterUserRoutes(app *fiber.App, userHandler *UserHTTPHandler) {
	api := app.Group("/api")
	api.Post("/users", userHandler.CreateUser)
}
