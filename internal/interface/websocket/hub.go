package websocket

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Client represents a WebSocket client
type Client struct {
	ID     uuid.UUID
	UserID uuid.UUID
	JobID  *uuid.UUID
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Clients by user ID
	userClients map[uuid.UUID]map[*Client]bool

	// Clients by job ID
	jobClients map[uuid.UUID]map[*Client]bool

	// Register requests
	register chan *Client

	// Unregister requests
	unregister chan *Client

	// Broadcast to all
	broadcast chan []byte

	// Logger
	logger *zap.Logger

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHub creates a new Hub
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		userClients: make(map[uuid.UUID]map[*Client]bool),
		jobClients:  make(map[uuid.UUID]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan []byte),
		logger:      logger,
	}
}

// Run starts the hub's event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToAll(message)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true

	// Add to user clients
	if _, ok := h.userClients[client.UserID]; !ok {
		h.userClients[client.UserID] = make(map[*Client]bool)
	}
	h.userClients[client.UserID][client] = true

	// Add to job clients if subscribed
	if client.JobID != nil {
		if _, ok := h.jobClients[*client.JobID]; !ok {
			h.jobClients[*client.JobID] = make(map[*Client]bool)
		}
		h.jobClients[*client.JobID][client] = true
	}

	h.logger.Debug("Client registered",
		zap.String("client_id", client.ID.String()),
		zap.String("user_id", client.UserID.String()),
	)
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)

		// Remove from user clients
		if clients, ok := h.userClients[client.UserID]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.userClients, client.UserID)
			}
		}

		// Remove from job clients
		if client.JobID != nil {
			if clients, ok := h.jobClients[*client.JobID]; ok {
				delete(clients, client)
				if len(clients) == 0 {
					delete(h.jobClients, *client.JobID)
				}
			}
		}

		h.logger.Debug("Client unregistered",
			zap.String("client_id", client.ID.String()),
		)
	}
}

// broadcastToAll broadcasts a message to all clients
func (h *Hub) broadcastToAll(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.Send <- message:
		default:
			h.unregister <- client
		}
	}
}

// BroadcastToJob broadcasts a message to all clients subscribed to a job
func (h *Hub) BroadcastToJob(jobID uuid.UUID, eventType string, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.jobClients[jobID]
	if !ok {
		return
	}

	message, err := json.Marshal(Message{
		Type:    eventType,
		Payload: payload,
	})
	if err != nil {
		h.logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	for client := range clients {
		select {
		case client.Send <- message:
		default:
			h.unregister <- client
		}
	}
}

// BroadcastToUser broadcasts a message to all clients for a user
func (h *Hub) BroadcastToUser(userID uuid.UUID, eventType string, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.userClients[userID]
	if !ok {
		return
	}

	message, err := json.Marshal(Message{
		Type:    eventType,
		Payload: payload,
	})
	if err != nil {
		h.logger.Error("Failed to marshal message", zap.Error(err))
		return
	}

	for client := range clients {
		select {
		case client.Send <- message:
		default:
			h.unregister <- client
		}
	}
}

// SubscribeToJob subscribes a client to job updates
func (h *Hub) SubscribeToJob(client *Client, jobID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Unsubscribe from previous job if any
	if client.JobID != nil {
		if clients, ok := h.jobClients[*client.JobID]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.jobClients, *client.JobID)
			}
		}
	}

	// Subscribe to new job
	client.JobID = &jobID
	if _, ok := h.jobClients[jobID]; !ok {
		h.jobClients[jobID] = make(map[*Client]bool)
	}
	h.jobClients[jobID][client] = true
}

// GetActiveConnections returns the number of active connections
func (h *Hub) GetActiveConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetUserConnections returns the number of connections for a user
func (h *Hub) GetUserConnections(userID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.userClients[userID]; ok {
		return len(clients)
	}
	return 0
}

