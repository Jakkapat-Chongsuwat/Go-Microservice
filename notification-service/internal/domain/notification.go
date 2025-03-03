package domain

import (
	"time"
)

type Notification struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func NewNotificationWithID(id, notificationType, message string) *Notification {
	return &Notification{
		ID:        id,
		Type:      notificationType,
		Message:   message,
		CreatedAt: time.Now(),
	}
}
