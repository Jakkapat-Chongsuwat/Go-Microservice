package interfaces

import "notification-service/internal/domain"

type NotificationPublisher interface {
	PublishNotification(notif *domain.Notification) error
}
