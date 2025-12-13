package websocket

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/arabella/ai-studio-backend/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

// ClientMessage represents an incoming WebSocket message
type ClientMessage struct {
	Type   string          `json:"type"`
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

// SubscribeJobData represents data for subscribing to a job
type SubscribeJobData struct {
	JobID string `json:"job_id"`
}

// Handler handles WebSocket connections
type Handler struct {
	hub         *Hub
	authUseCase *usecase.AuthUseCase
	logger      *zap.Logger
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, authUseCase *usecase.AuthUseCase, logger *zap.Logger) *Handler {
	return &Handler{
		hub:         hub,
		authUseCase: authUseCase,
		logger:      logger,
	}
}

// HandleConnection handles a new WebSocket connection
func (h *Handler) HandleConnection(c *gin.Context) {
	// Get token from query parameter
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		return
	}

	// Validate token
	user, err := h.authUseCase.ValidateToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Create client
	client := &Client{
		ID:     uuid.New(),
		UserID: user.ID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h.hub,
	}

	// Register client
	h.hub.register <- client

	// Start goroutines for reading and writing
	go h.writePump(client)
	go h.readPump(client)
}

// HandleJobConnection handles a WebSocket connection for a specific job
func (h *Handler) HandleJobConnection(c *gin.Context) {
	// Get job ID from path
	jobIDStr := c.Param("id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	// Get token from query parameter
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		return
	}

	// Validate token
	user, err := h.authUseCase.ValidateToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection", zap.Error(err))
		return
	}

	// Create client with job subscription
	client := &Client{
		ID:     uuid.New(),
		UserID: user.ID,
		JobID:  &jobID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h.hub,
	}

	// Register client
	h.hub.register <- client

	// Start goroutines for reading and writing
	go h.writePump(client)
	go h.readPump(client)
}

// readPump pumps messages from the WebSocket connection to the hub
func (h *Handler) readPump(client *Client) {
	defer func() {
		h.hub.unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket error", zap.Error(err))
			}
			break
		}

		// Parse message
		var msg ClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			h.logger.Warn("Failed to parse message", zap.Error(err))
			continue
		}

		// Handle message
		h.handleMessage(client, msg)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (h *Handler) writePump(client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Channel closed
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming WebSocket messages
func (h *Handler) handleMessage(client *Client, msg ClientMessage) {
	switch msg.Action {
	case "subscribe_job":
		var data SubscribeJobData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			h.logger.Warn("Failed to parse subscribe data", zap.Error(err))
			return
		}

		jobID, err := uuid.Parse(data.JobID)
		if err != nil {
			h.logger.Warn("Invalid job ID", zap.String("job_id", data.JobID))
			return
		}

		h.hub.SubscribeToJob(client, jobID)

		// Send confirmation
		response, _ := json.Marshal(Message{
			Type:    "subscribed",
			Payload: map[string]string{"job_id": jobID.String()},
		})
		client.Send <- response

	case "ping":
		// Send pong
		response, _ := json.Marshal(Message{
			Type:    "pong",
			Payload: map[string]int64{"timestamp": time.Now().Unix()},
		})
		client.Send <- response

	default:
		h.logger.Warn("Unknown action", zap.String("action", msg.Action))
	}
}

