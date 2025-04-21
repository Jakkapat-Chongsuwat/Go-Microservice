package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"notification-service/internal/adapters/kafka"
	ws "notification-service/internal/adapters/websocket"
	"notification-service/internal/usecases"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Configuration structs
type Config struct {
	Kafka          KafkaConfig
	SchemaRegistry string
	WebSocket      WebSocketConfig
}

type KafkaConfig struct {
	Brokers []string
	GroupID string
	Topic   string
}

type WebSocketConfig struct {
	Port            string
	ReadBufferSize  int
	WriteBufferSize int
}

// WebSocket upgrader with configurable buffer sizes and permissive CORS
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
	// Add these options to better support proxy scenarios
	EnableCompression: true,
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	config := loadConfig()
	logger.Info("Configuration loaded",
		zap.Strings("kafka_brokers", config.Kafka.Brokers),
		zap.String("kafka_topic", config.Kafka.Topic),
		zap.String("ws_port", config.WebSocket.Port),
	)

	// Initialize WebSocket hub
	hub := ws.NewHub(logger)

	// Setup HTTP server with all routes
	server := setupHTTPServer(config.WebSocket.Port, hub, logger)

	// Setup notification use case
	notificationUseCase := usecases.NewNotificationUseCase(logger, hub)

	// Setup Kafka consumer
	consumerGroup, err := kafka.NewKafkaConsumerGroup(
		config.Kafka.Brokers,
		config.Kafka.GroupID,
		config.Kafka.Topic,
		config.SchemaRegistry,
		notificationUseCase,
		logger,
	)
	if err != nil {
		logger.Fatal("Failed to create Kafka consumer group", zap.Error(err))
	}

	// Start the HTTP server
	go func() {
		logger.Info("Starting WebSocket server", zap.String("port", config.WebSocket.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("WebSocket server failed", zap.Error(err))
		}
	}()

	// Start Kafka consumer in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		logger.Info("Starting Kafka consumer",
			zap.String("topic", config.Kafka.Topic),
			zap.String("group_id", config.Kafka.GroupID))
		if err := consumerGroup.Start(ctx); err != nil {
			logger.Error("Error during Kafka consumption", zap.Error(err))
		}
	}()

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a termination signal
	sig := <-sigChan
	logger.Info("Received termination signal", zap.String("signal", sig.String()))

	// Create a deadline for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Stop Kafka consumer
	cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	logger.Info("Notification service gracefully stopped")
}

// initLogger creates a production-ready zap logger
func initLogger() (*zap.Logger, error) {
	// Use development logger if in dev mode
	if os.Getenv("DEV_MODE") == "true" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

// loadConfig loads all configuration from environment variables
func loadConfig() Config {
	// Kafka configuration
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "localhost:9092"
	}
	brokers := strings.Split(brokersEnv, ",")

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "notification-consumer-group"
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "notifications"
	}

	// Schema Registry URL
	schemaRegistryURL := os.Getenv("SCHEMA_REGISTRY_URL")
	if schemaRegistryURL == "" {
		schemaRegistryURL = "http://localhost:8081"
	}

	// WebSocket configuration
	wsPort := os.Getenv("WS_PORT")
	if wsPort == "" {
		wsPort = "20052" // Default to 20052 as specified in your k8s config
	}

	return Config{
		Kafka: KafkaConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,
		},
		SchemaRegistry: schemaRegistryURL,
		WebSocket: WebSocketConfig{
			Port:            wsPort,
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// setupHTTPServer configures the HTTP server with all routes
func setupHTTPServer(port string, hub *ws.Hub, logger *zap.Logger) *http.Server {
	// Create router
	mux := http.NewServeMux()

	// Root handler to avoid 404s on the root path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Exact path matching for root
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Notification Service API\n")
		fmt.Fprintf(w, "Available endpoints:\n")
		fmt.Fprintf(w, "- /health: Service health check\n")
		fmt.Fprintf(w, "- /ws, /websocket, /socket: WebSocket connections\n")
		fmt.Fprintf(w, "- /debug: Debug information\n")
	})

	// Define your health route with exact matching
	mux.HandleFunc("/health", healthHandler())

	// WebSocket routes
	wsHandlerFunc := wsHandler(hub, logger)
	mux.HandleFunc("/ws", wsHandlerFunc)
	mux.HandleFunc("/websocket", wsHandlerFunc)
	mux.HandleFunc("/socket", wsHandlerFunc)

	// Debug routes
	debugRouter := http.NewServeMux()
	debugRouter.HandleFunc("/", debugHandler())
	debugRouter.HandleFunc("/headers", headersHandler())
	debugRouter.HandleFunc("/echo", echoHandler())
	debugRouter.HandleFunc("/time", timeHandler())
	debugRouter.HandleFunc("/test-ws", testWSHandler(logger))
	debugRouter.HandleFunc("/broadcast", broadcastHandler(hub, logger))

	// Mount the debug router
	mux.Handle("/debug/", http.StripPrefix("/debug", debugRouter))
	mux.HandleFunc("/debug", debugHandler()) // Serve the root debug path

	// Configure server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      loggingMiddleware(logger)(mux),
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  240 * time.Second,
	}

	return server
}

// customResponseWriter wraps http.ResponseWriter to capture the status code and implement http.Hijacker
type customResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before writing it
func (crw *customResponseWriter) WriteHeader(code int) {
	crw.statusCode = code
	crw.ResponseWriter.WriteHeader(code)
}

// Hijack implements the http.Hijacker interface for WebSocket support
func (crw *customResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := crw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("websocket: response does not implement http.Hijacker")
}

// Flush implements http.Flusher if the underlying response writer implements it
func (crw *customResponseWriter) Flush() {
	if flusher, ok := crw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Push implements http.Pusher if the underlying response writer implements it (for HTTP/2)
func (crw *customResponseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := crw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

// CloseNotify implements http.CloseNotifier if the underlying response writer implements it
// Note: CloseNotifier is deprecated but some older frameworks still use it
func (crw *customResponseWriter) CloseNotify() <-chan bool {
	if notifier, ok := crw.ResponseWriter.(http.CloseNotifier); ok {
		return notifier.CloseNotify()
	}
	// Create a channel that's never used
	ch := make(chan bool, 1)
	return ch
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Log request start
			logger.Info("Request received",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)

			// Check if this is a WebSocket upgrade request
			if websocketUpgrade := strings.ToLower(r.Header.Get("Upgrade")) == "websocket"; websocketUpgrade {
				logger.Debug("WebSocket upgrade request detected",
					zap.String("path", r.URL.Path),
					zap.String("connection", r.Header.Get("Connection")),
					zap.String("upgrade", r.Header.Get("Upgrade")),
				)
			}

			// Create a custom response writer with Hijacker support
			crw := &customResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Process the request
			next.ServeHTTP(crw, r)

			// Log request completion
			logger.Info("Request completed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", crw.statusCode),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

// wsHandler handles WebSocket connections
func wsHandler(hub *ws.Hub, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Log connection attempt with detailed headers
		logger.Info("WebSocket connection attempt",
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("host", r.Host),
			zap.String("origin", r.Header.Get("Origin")),
			zap.String("connection", r.Header.Get("Connection")),
			zap.String("upgrade", r.Header.Get("Upgrade")),
			zap.String("sec-websocket-key", r.Header.Get("Sec-WebSocket-Key")),
		)

		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("Failed to upgrade to WebSocket connection",
				zap.Error(err),
				zap.String("remote_addr", r.RemoteAddr),
				// Log more details for debugging
				zap.String("connection_header", r.Header.Get("Connection")),
				zap.String("upgrade_header", r.Header.Get("Upgrade")),
			)
			// Return a clearer error response
			http.Error(w, "Could not upgrade to WebSocket: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Log successful connection
		logger.Info("WebSocket connection established",
			zap.String("remote_addr", conn.RemoteAddr().String()),
			zap.String("path", r.URL.Path),
		)

		// Register connection with hub
		hub.Add(conn)
		defer hub.Remove(conn)

		// Handle incoming messages
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				closeErr, ok := err.(*websocket.CloseError)
				if ok {
					logger.Info("WebSocket connection closed by client",
						zap.Int("code", closeErr.Code),
						zap.String("text", closeErr.Text),
						zap.String("remote_addr", conn.RemoteAddr().String()),
					)
				} else {
					logger.Error("Error reading WebSocket message",
						zap.Error(err),
						zap.String("remote_addr", conn.RemoteAddr().String()),
					)
				}
				break
			}

			// Log received message
			logger.Debug("Received WebSocket message",
				zap.String("message", string(message)),
				zap.Int("type", messageType),
				zap.String("remote_addr", conn.RemoteAddr().String()),
			)

			// Echo the message back (for testing)
			if err := conn.WriteMessage(messageType, message); err != nil {
				logger.Error("Error writing WebSocket message",
					zap.Error(err),
					zap.String("remote_addr", conn.RemoteAddr().String()),
				)
				break
			}
		}
	}
}

// healthHandler returns the service health status
func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "notification-service",
		}

		json.NewEncoder(w).Encode(response)
	}
}

// debugHandler returns detailed information about the request
func debugHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		fmt.Fprintf(w, "Debug Information\n")
		fmt.Fprintf(w, "================\n\n")

		fmt.Fprintf(w, "Request Details:\n")
		fmt.Fprintf(w, "  Method:     %s\n", r.Method)
		fmt.Fprintf(w, "  URL:        %s\n", r.URL.String())
		fmt.Fprintf(w, "  Protocol:   %s\n", r.Proto)
		fmt.Fprintf(w, "  Host:       %s\n", r.Host)
		fmt.Fprintf(w, "  RemoteAddr: %s\n", r.RemoteAddr)
		fmt.Fprintf(w, "  RequestURI: %s\n", r.RequestURI)
		fmt.Fprintf(w, "  TLS:        %v\n\n", r.TLS != nil)

		fmt.Fprintf(w, "Headers:\n")
		for name, values := range r.Header {
			for _, value := range values {
				fmt.Fprintf(w, "  %s: %s\n", name, value)
			}
		}

		fmt.Fprintf(w, "\nEnvironment:\n")
		for _, env := range os.Environ() {
			fmt.Fprintf(w, "  %s\n", env)
		}

		fmt.Fprintf(w, "\nServer Time: %s\n", time.Now().Format(time.RFC3339))
	}
}

// headersHandler returns all request headers as JSON
func headersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		headers := make(map[string]string)
		for name, values := range r.Header {
			headers[name] = strings.Join(values, ", ")
		}

		json.NewEncoder(w).Encode(headers)
	}
}

// echoHandler echoes back the request body
func echoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Method not allowed. Use POST to send data.\n")
			return
		}

		var data interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error parsing JSON: %v\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}

// timeHandler returns the current server time
func timeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := map[string]interface{}{
			"time":     time.Now().Format(time.RFC3339),
			"unix":     time.Now().Unix(),
			"timezone": time.Now().Location().String(),
		}

		json.NewEncoder(w).Encode(response)
	}
}

// broadcastHandler sends a test message to all connected WebSocket clients
func broadcastHandler(hub *ws.Hub, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Method not allowed. Use POST to broadcast a message.\n")
			return
		}

		var message interface{}
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error parsing JSON: %v\n", err)
			return
		}

		// Send the message to all connected clients
		if err := hub.Broadcast(message); err != nil {
			logger.Error("Error broadcasting message", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Error broadcasting message: %v\n", err)
			return
		}

		// Send success response
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"status":  "success",
			"message": "Broadcast sent successfully",
		}

		json.NewEncoder(w).Encode(response)
	}
}

// testWSHandler serves an HTML page with an embedded WebSocket client
func testWSHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Serving WebSocket test client")

		html := `<!DOCTYPE html>
<html>
<head>
    <title>WebSocket Tester</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
            color: #333;
        }
        h1 {
            color: #2c3e50;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .connection-info {
            margin: 20px 0;
            padding: 15px;
            border-radius: 5px;
            background-color: #f8f9fa;
            border-left: 5px solid #6c757d;
        }
        .status {
            display: inline-block;
            padding: 5px 10px;
            border-radius: 4px;
            font-weight: bold;
        }
        .connected { background-color: #d4edda; color: #155724; }
        .disconnected { background-color: #f8d7da; color: #721c24; }
        .connecting { background-color: #fff3cd; color: #856404; }
        
        button {
            padding: 8px 16px;
            border: none;
            border-radius: 4px;
            background-color: #3498db;
            color: white;
            font-weight: bold;
            cursor: pointer;
            transition: background-color 0.3s;
            margin-right: 5px;
        }
        button:hover { background-color: #2980b9; }
        button:disabled { background-color: #95a5a6; cursor: not-allowed; }
        
        input[type="text"] {
            padding: 8px;
            border: 1px solid #ddd;
            border-radius: 4px;
            width: 70%;
            margin-right: 10px;
        }
        
        #log {
            height: 300px;
            overflow-y: auto;
            padding: 10px;
            background-color: #f8f9fa;
            border: 1px solid #ddd;
            border-radius: 4px;
            margin-top: 20px;
            font-family: monospace;
        }
        
        .log-entry {
            margin: 5px 0;
            padding: 5px;
            border-radius: 3px;
        }
        .log-entry.sent { background-color: #e8f4f8; border-left: 3px solid #3498db; }
        .log-entry.received { background-color: #eafaf1; border-left: 3px solid #2ecc71; }
        .log-entry.error { background-color: #fdedec; border-left: 3px solid #e74c3c; }
        .log-entry.info { background-color: #f5f5f5; border-left: 3px solid #95a5a6; }
        
        .input-group {
            display: flex;
            margin-top: 15px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>WebSocket Tester</h1>
        
        <div class="connection-info">
            <div>
                <strong>Status:</strong> 
                <span id="status" class="status disconnected">Disconnected</span>
            </div>
            <div>
                <strong>URL:</strong> <span id="wsUrl"></span>
            </div>
        </div>
        
        <div>
            <button id="connectBtn">Connect</button>
            <button id="disconnectBtn" disabled>Disconnect</button>
        </div>
        
        <div class="input-group">
            <input type="text" id="messageInput" placeholder="Enter message to send..." disabled />
            <button id="sendBtn" disabled>Send</button>
        </div>
        
        <div id="log"></div>
    </div>
    
    <script>
        const statusEl = document.getElementById('status');
        const wsUrlEl = document.getElementById('wsUrl');
        const logEl = document.getElementById('log');
        const connectBtn = document.getElementById('connectBtn');
        const disconnectBtn = document.getElementById('disconnectBtn');
        const messageInput = document.getElementById('messageInput');
        const sendBtn = document.getElementById('sendBtn');
        
        let socket = null;
        
        function updateStatus(isConnected, isConnecting = false) {
            if (isConnecting) {
                statusEl.textContent = 'Connecting...';
                statusEl.className = 'status connecting';
                connectBtn.disabled = true;
                disconnectBtn.disabled = true;
                messageInput.disabled = true;
                sendBtn.disabled = true;
            } else if (isConnected) {
                statusEl.textContent = 'Connected';
                statusEl.className = 'status connected';
                connectBtn.disabled = true;
                disconnectBtn.disabled = false;
                messageInput.disabled = false;
                sendBtn.disabled = false;
            } else {
                statusEl.textContent = 'Disconnected';
                statusEl.className = 'status disconnected';
                connectBtn.disabled = false;
                disconnectBtn.disabled = true;
                messageInput.disabled = true;
                sendBtn.disabled = true;
            }
        }
        
        function addLogEntry(message, type) {
            const timestamp = new Date().toISOString().substr(11, 8);
            const entry = document.createElement('div');
            entry.className = 'log-entry ' + type;
            entry.innerHTML = '<strong>' + timestamp + '</strong> ' + message;
            logEl.appendChild(entry);
            logEl.scrollTop = logEl.scrollHeight;
        }
        
        function getWebSocketUrl() {
            // Get current location and build WebSocket URL
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const host = window.location.host;
            const path = '/ws'; // You can customize this
            const url = protocol + '//' + host + path;
            
            wsUrlEl.textContent = url;
            return url;
        }
        
        connectBtn.addEventListener('click', function() {
            const wsUrl = getWebSocketUrl();
            updateStatus(false, true);
            addLogEntry('Connecting to ' + wsUrl + '...', 'info');
            
            try {
                socket = new WebSocket(wsUrl);
                
                socket.onopen = function(event) {
                    updateStatus(true);
                    addLogEntry('Connection established', 'info');
                };
                
                socket.onclose = function(event) {
                    updateStatus(false);
                    let reason = event.reason ? ' - ' + event.reason : '';
                    addLogEntry('Connection closed (Code: ' + event.code + ')' + reason, 'info');
                    socket = null;
                };
                
                socket.onerror = function(error) {
                    addLogEntry('WebSocket error occurred', 'error');
                    console.error('WebSocket error:', error);
                };
                
                socket.onmessage = function(event) {
                    try {
                        // Try to parse as JSON
                        const data = JSON.parse(event.data);
                        const formatted = JSON.stringify(data, null, 2);
                        addLogEntry('Received: <pre>' + formatted + '</pre>', 'received');
                    } catch(e) {
                        // If not JSON, display as plain text
                        addLogEntry('Received: ' + event.data, 'received');
                    }
                };
            } catch(error) {
                updateStatus(false);
                addLogEntry('Error creating WebSocket: ' + error.message, 'error');
            }
        });
        
        disconnectBtn.addEventListener('click', function() {
            if (socket) {
                socket.close(1000, 'User initiated disconnect');
                addLogEntry('Disconnecting...', 'info');
            }
        });
        
        sendBtn.addEventListener('click', function() {
            const message = messageInput.value.trim();
            if (!message || !socket) return;
            
            try {
                socket.send(message);
                addLogEntry('Sent: ' + message, 'sent');
                messageInput.value = '';
            } catch(error) {
                addLogEntry('Error sending message: ' + error.message, 'error');
            }
        });
        
        messageInput.addEventListener('keypress', function(event) {
            if (event.key === 'Enter' && !sendBtn.disabled) {
                sendBtn.click();
            }
        });
        
        // Initialize page
        getWebSocketUrl();
        updateStatus(false);
        addLogEntry('WebSocket test client ready. Click "Connect" to start.', 'info');
    </script>
</body>
</html>
`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}
}
