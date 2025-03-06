package websocket_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	ws "notification-service/internal/adapters/websocket"

	gws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestWebSocketServerIntegration(t *testing.T) {
	hub := ws.NewHub()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := gws.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		hub.Add(conn)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				hub.Remove(conn)
				break
			}
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	u.Scheme = "ws"
	u.Path = "/"

	wsConn, _, err := gws.DefaultDialer.Dial(u.String(), nil)
	require.NoError(t, err)
	defer wsConn.Close()

	time.Sleep(100 * time.Millisecond)

	testMessage := "hello from hub"
	err = hub.Broadcast(testMessage)
	require.NoError(t, err)

	_, data, err := wsConn.ReadMessage()
	require.NoError(t, err)
	require.Contains(t, string(data), testMessage)
}
