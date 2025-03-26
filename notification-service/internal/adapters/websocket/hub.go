package websocket

import (
	"encoding/json"
	"sync"

	"notification-service/internal/domain"
	"notification-service/internal/domain/interfaces"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Hub struct {
	connections map[*websocket.Conn]bool
	mu          sync.RWMutex
	logger      *zap.Logger
}

var _ interfaces.NotificationPublisher = (*Hub)(nil)

func NewHub(logger *zap.Logger) *Hub {
	logger.Info("Initializing WebSocket Hub")
	return &Hub{
		connections: make(map[*websocket.Conn]bool),
		logger:      logger,
	}
}

func (h *Hub) Add(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[conn] = true
	h.logger.Info("WebSocket connection added", zap.String("remoteAddr", conn.RemoteAddr().String()))
}

func (h *Hub) Remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.connections, conn)
	h.logger.Info("WebSocket connection removed", zap.String("remoteAddr", conn.RemoteAddr().String()))
}

func (h *Hub) Broadcast(message interface{}) error {
	h.mu.RLock()
	defer h.mu.RUnlock()
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal message", zap.Error(err))
		return err
	}
	h.logger.Info("Broadcasting message", zap.Int("connectionCount", len(h.connections)))
	for conn := range h.connections {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			h.logger.Error("Failed to write message to connection", zap.String("remoteAddr", conn.RemoteAddr().String()), zap.Error(err))
		}
	}
	return nil
}

func (h *Hub) PublishNotification(notif *domain.Notification) error {
	h.logger.Info("Publishing notification", zap.String("notificationID", notif.ID))
	return h.Broadcast(notif)
}
