package usecases

import (
	"context"
	"fmt"
	"notification-service/internal/domain"
	"notification-service/internal/domain/interfaces"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type NotificationUseCase interface {
	ProcessNotification(ctx context.Context, notif *domain.Notification) error
}

type notificationUseCaseImpl struct {
	logger    *zap.Logger
	publisher interfaces.NotificationPublisher
}

var _ NotificationUseCase = (*notificationUseCaseImpl)(nil)

func NewNotificationUseCase(logger *zap.Logger, publisher interfaces.NotificationPublisher) NotificationUseCase {
	return &notificationUseCaseImpl{
		logger:    logger,
		publisher: publisher,
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

	if err := uc.publisher.PublishNotification(notif); err != nil {
		uc.logger.Error("failed to publish notification", zap.Error(err))
	}

	return nil
}
