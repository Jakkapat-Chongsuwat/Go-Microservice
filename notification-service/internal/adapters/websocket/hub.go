package websocket

import (
	"encoding/json"
	"sync"

	"notification-service/internal/domain"
	"notification-service/internal/domain/interfaces"

	"github.com/gorilla/websocket"
)

type Hub struct {
	connections map[*websocket.Conn]bool
	mu          sync.RWMutex
}

var _ interfaces.NotificationPublisher = (*Hub)(nil)

func NewHub() *Hub {
	return &Hub{
		connections: make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) Add(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[conn] = true
}

func (h *Hub) Remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.connections, conn)
}

func (h *Hub) Broadcast(message interface{}) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	for conn := range h.connections {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		}
	}
	return nil
}

func (h *Hub) PublishNotification(notif *domain.Notification) error {
	return h.Broadcast(notif)
}
