package mappers

import (
	"fmt"
	"notification-service/internal/domain"
)

func MapRawToNotification(raw map[string]interface{}) (*domain.Notification, error) {
	var id, typ, msg string

	if v, ok := raw["id"]; ok && v != nil {
		id = fmt.Sprintf("%v", v)
	} else if v, ok := raw["order_id"]; ok && v != nil {
		id = fmt.Sprintf("%v", v)
	} else if v, ok := raw["user_id"]; ok && v != nil {
		id = fmt.Sprintf("%v", v)
	}

	if v, ok := raw["type"]; ok && v != nil {
		typ = fmt.Sprintf("%v", v)
	} else if v, ok := raw["event_type"]; ok && v != nil {
		typ = fmt.Sprintf("%v", v)
	}

	if v, ok := raw["message"]; ok && v != nil {
		msg = fmt.Sprintf("%v", v)
	}

	if id == "" || typ == "" {
		return nil, fmt.Errorf("missing required fields: id=%q, type=%q", id, typ)
	}

	// Create and return a well-formed Notification.
	notif := domain.NewNotificationWithID(id, typ, msg)
	return notif, nil
}
