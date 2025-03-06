package usecases

import (
	"context"
	"testing"
	"time"

	"notification-service/internal/domain"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type NoOpPublisher struct{}

func (n *NoOpPublisher) PublishNotification(notif *domain.Notification) error {
	return nil
}

func TestProcessNotification_GeneratesDefaults(t *testing.T) {
	logger := zap.NewNop()
	uc := NewNotificationUseCase(logger, &NoOpPublisher{})

	notif := &domain.Notification{
		Type:    "test",
		Message: "hello world",
	}

	err := uc.ProcessNotification(context.Background(), notif)
	require.NoError(t, err)

	require.NotEmpty(t, notif.ID, "ID should be generated")
	require.False(t, notif.CreatedAt.IsZero(), "CreatedAt should be set")
	require.WithinDuration(t, time.Now(), notif.CreatedAt, 2*time.Second)
}

func TestProcessNotification_UsesExistingValues(t *testing.T) {
	logger := zap.NewNop()
	uc := NewNotificationUseCase(logger, &NoOpPublisher{})

	customID := "abc-123"
	customTime := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)
	notif := &domain.Notification{
		ID:        customID,
		Type:      "test",
		Message:   "hello world",
		CreatedAt: customTime,
	}

	err := uc.ProcessNotification(context.Background(), notif)
	require.NoError(t, err)

	require.Equal(t, customID, notif.ID, "ID should remain unchanged")
	require.Equal(t, customTime, notif.CreatedAt, "CreatedAt should remain unchanged")
}
