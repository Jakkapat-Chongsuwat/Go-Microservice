package domain

import "time"

type OrderEvent struct {
	OrderID   string
	EventType string
	Timestamp time.Time
}
