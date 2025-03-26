package domain

import "time"

type OrderEvent struct {
	OrderID   string
	EventType string
	Message   string
	Timestamp time.Time
}
