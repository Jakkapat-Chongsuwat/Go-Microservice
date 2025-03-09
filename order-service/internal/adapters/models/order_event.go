package models

import "time"

type OrderEvent struct {
	OrderID   string
	EventType string
	Timestamp time.Time
}
