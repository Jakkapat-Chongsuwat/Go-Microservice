package usecases

import (
	"context"
	"fmt"
	"notification-service/internal/domain"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NotificationUseCase interface {
	ProcessNotification(ctx context.Context, notif *domain.Notification) error
}

type notificationUseCaseImpl struct {
	logger *zap.Logger
}

func NewNotificationUseCase(logger *zap.Logger) NotificationUseCase {
	return &notificationUseCaseImpl{
		logger: logger,
	}
}

func (uc *notificationUseCaseImpl) ProcessNotification(ctx context.Context, notif *domain.Notification) error {
	if notif.ID == "" || notif.CreatedAt.IsZero() {
		uc.logger.Warn("notification missing id or created_at; generating defaults")
		if notif.ID == "" {
			notif.ID = uuid.NewString()
		}
		if notif.CreatedAt.IsZero() {
			notif.CreatedAt = time.Now()
		}
	}

	uc.logger.Info("Processing notification",
		zap.String("id", notif.ID),
		zap.String("type", notif.Type),
		zap.String("message", notif.Message))
	fmt.Printf("Processed notification: %+v\n", notif)
	return nil
}
