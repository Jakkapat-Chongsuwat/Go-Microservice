package domain

import "time"

type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}

type OrderEvent struct {
	OrderID   string
	Type      string
	Timestamp time.Time
}

func (e OrderEvent) EventType() string     { return e.Type }
func (e OrderEvent) OccurredAt() time.Time { return e.Timestamp }

func NewOrderEvent(orderID, eventType string) OrderEvent {
	return OrderEvent{
		OrderID:   orderID,
		Type:      eventType,
		Timestamp: Clock.Now(),
	}
}

type UserEvent struct {
	UserID    string
	Type      string
	Timestamp time.Time
}

func (e UserEvent) EventType() string     { return e.Type }
func (e UserEvent) OccurredAt() time.Time { return e.Timestamp }

func NewUserEvent(userID, eventType string) UserEvent {
	return UserEvent{
		UserID:    userID,
		Type:      eventType,
		Timestamp: Clock.Now(),
	}
}
