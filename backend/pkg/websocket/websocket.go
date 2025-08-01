package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/Athooh/social-network/pkg/logger"
	"github.com/gorilla/websocket"
)

// Client represents a connected WebSocket client
type Client struct {
	ID           string
	UserID       string
	TabID        string // New field to identify browser tab
	Conn         *websocket.Conn
	Hub          *Hub
	Send         chan []byte
	Mu           sync.Mutex
	IsActive     bool
	Done         chan struct{} // New channel to signal when client disconnects
	LastPingTime time.Time
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	Clients map[*Client]bool

	// User ID to clients mapping for targeted messages
	UserClients map[string][]*Client

	// Inbound messages from clients
	Broadcast chan []byte

	// Register requests from clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Mutex for concurrent access to maps
	Mu sync.RWMutex

	// Logger
	log *logger.Logger

	// Status updater for online/offline notifications
	statusUpdater StatusUpdater

	heartbeatCheckInterval time.Duration
	heartbeatTimeout       time.Duration
}

// Message represents a WebSocket message
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// NewHub creates a new Hub instance
func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		Broadcast:              make(chan []byte),
		Register:               make(chan *Client),
		Unregister:             make(chan *Client),
		Clients:                make(map[*Client]bool),
		UserClients:            make(map[string][]*Client),
		Mu:                     sync.RWMutex{},
		log:                    log,
		heartbeatCheckInterval: 60 * time.Second,
		heartbeatTimeout:       120 * time.Second,
	}
}

// Run starts the Hub
func (h *Hub) Run() {
	heartbeatTicker := time.NewTicker(h.heartbeatCheckInterval)
	pingTicker := time.NewTicker(30 * time.Second) // Send ping every 30 seconds
	defer func() {
		heartbeatTicker.Stop()
		pingTicker.Stop()
	}()

	for {
		select {
		case client := <-h.Register:
			h.Mu.Lock()

			// Close ALL existing connections for this user
			// This ensures only one connection per user
			if clients, exists := h.UserClients[client.UserID]; exists {
				for _, existingClient := range clients {
					h.log.Debug("Closing existing connection for user: %s (client ID: %s)", client.UserID, existingClient.ID)
					delete(h.Clients, existingClient)
					close(existingClient.Send)

					// Force close the connection
					existingClient.Conn.Close()
				}
				// Clear all existing clients for this user
				h.UserClients[client.UserID] = nil
			}

			// Register the new client
			h.Clients[client] = true

			// Set this as the only client for this user
			h.UserClients[client.UserID] = []*Client{client}
			h.Mu.Unlock()

			h.log.Debug("Client registered: %s (User: %s, Tab: %s)", client.ID, client.UserID, client.TabID)

		case client := <-h.Unregister:
			h.Mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)

				// Remove from user-specific clients map
				clients := h.UserClients[client.UserID]
				for i, c := range clients {
					if c.ID == client.ID {
						h.UserClients[client.UserID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}

				// Clean up empty user entries
				if len(h.UserClients[client.UserID]) == 0 {
					delete(h.UserClients, client.UserID)
				}

				h.log.Debug("Client unregistered: %s (User: %s)", client.ID, client.UserID)
			}
			h.Mu.Unlock()

		case message := <-h.Broadcast:
			h.Mu.RLock()
			for client := range h.Clients {
				logger.Warn("Sending message to client %s", client.UserID)
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.Mu.RUnlock()

		case <-heartbeatTicker.C:
			h.checkHeartbeats()

		case <-pingTicker.C:
			// Send ping to all clients
			h.sendPingToAllClients()
		}
	}
}

// BroadcastToAll sends a message to all connected clients
// func (h *Hub) BroadcastToAll(message interface{}) {
// 	msgType, payload := prepareMessage("broadcast", message)
// 	h.Broadcast <- payload
// 	h.log.Debug("Broadcasting message type: %s to all clients", msgType)
// }

// BroadcastToUser sends a message to a specific user's clients
func (h *Hub) BroadcastToUser(userID string, message interface{}) {
	// _, payload := prepareMessage("user", message)
	payload, err := json.Marshal(message)
	if err != nil {
		h.log.Error("Error marshaling message: %v", err)
		return
	}

	h.Mu.RLock()
	clients, exists := h.UserClients[userID]
	h.Mu.RUnlock()

	if !exists {
		return
	}

	for _, client := range clients {
		client.Mu.Lock()
		if client.IsActive {
			select {
			case client.Send <- payload:
				h.log.Info("Sent message to user %s client %s", userID, client.ID)
			default:
				h.log.Info("Failed to send message to user %s client %s", userID, client.ID)
			}
		}
		client.Mu.Unlock()
	}
}

// BroadcastToFollowers sends a message to all followers of a user
// func (h *Hub) BroadcastToFollowers(userID string, followerIDs []string, message interface{}) {
// 	_, payload := prepareMessage("followers", message)

// 	// Send to the user themselves
// 	h.BroadcastToUser(userID, message)

// 	// Send to all followers
// 	for _, followerID := range followerIDs {
// 		h.BroadcastToUser(followerID, message)
// 	}
// }

// prepareMessage formats a message for sending
// func prepareMessage(msgType string, payload interface{}) (string, []byte) {
// 	msg := Message{
// 		Type:    msgType,
// 		Payload: payload,
// 	}

// 	data, _ := json.Marshal(msg)
// 	return msgType, data
// }

// HasActiveClient checks if a user already has an active client connection
func (h *Hub) HasActiveClient(userID string) bool {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	// Check if user has any active connections
	clients, exists := h.UserClients[userID]
	if !exists {
		return false
	}

	// Check if any of the user's clients are active
	for _, client := range clients {
		if client.IsActive {
			return true
		}
	}
	return false
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
		close(c.Done) // Signal that this client has disconnected
	}()

	c.Conn.SetReadLimit(512 * 1024) // 512KB max message size
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.log.Error("WebSocket read error: %v", err)
			}
			break
		}

		// Handle ping/pong
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			c.Hub.log.Error("Error unmarshaling message: %v", err)
			continue
		}

		if msg.Type == "ping" {
			c.handlePing()
			continue
		}

		if msg.Type == "user_away" {

			// Mark this client as inactive
			c.IsActive = false

			// Check if this was the user's last active connection
			hasOtherActiveConnections := false

			c.Hub.Mu.RLock()
			clients, exists := c.Hub.UserClients[c.UserID]
			if exists {
				for _, otherClient := range clients {
					if otherClient.ID != c.ID && otherClient.IsActive {
						hasOtherActiveConnections = true
						break
					}
				}
			}
			c.Hub.Mu.RUnlock()

			// If no other active connections, mark user as offline
			if !hasOtherActiveConnections {
				userID := c.UserID
				go func(userID string) {
					c.Hub.SetUserOffline(userID)
				}(userID)
			}

			continue
		}

		if msg.Type == "user_active" {

			// Mark this client as active
			c.Mu.Lock()
			c.IsActive = true
			c.Mu.Unlock()

			// Check if this is the first active connection for this user
			c.Hub.Mu.Lock() // Changed from RLock to Lock since we'll be modifying
			wasOffline := true
			clients, exists := c.Hub.UserClients[c.UserID]
			if exists {
				// Close any other active connections for this user
				for _, otherClient := range clients {
					if otherClient.ID != c.ID {
						if otherClient.IsActive {
							wasOffline = false
							// Close the other connection
							c.Hub.log.Info("Closing duplicate connection for user %s (client ID: %s)", c.UserID, otherClient.ID)
							otherClient.IsActive = false
							otherClient.Conn.Close()
						}
					}
				}

				// Clean up the user clients list to only include this client
				newClients := []*Client{}
				for _, otherClient := range clients {
					if otherClient.ID == c.ID {
						newClients = append(newClients, otherClient)
					}
				}
				c.Hub.UserClients[c.UserID] = newClients
			}
			c.Hub.Mu.Unlock()

			// If this is the first active connection, mark user as online
			if wasOffline {
				userID := c.UserID
				go func(userID string) {
					c.Hub.SetUserOnline(userID)
				}(userID)
			}

			continue
		}

		// Process incoming messages if needed
		// For now, we're just handling server -> client communication
		c.Hub.log.Debug("Received message from client %s: %s", c.ID, string(message))
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Hub) GetActiveConnectionCount(userID string) int {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	count := 0
	for client := range h.Clients {
		if client.UserID == userID && client.IsActive {
			count++
		}
	}
	return count
}

func (h *Hub) CloseUserConnections(userID string) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	if clients, exists := h.UserClients[userID]; exists {
		for _, client := range clients {
			if client.IsActive {
				client.IsActive = false
				client.Conn.Close()
				h.Unregister <- client
			}
		}
		// Clear the user's client list
		delete(h.UserClients, userID)
	}
}

func (c *Client) handlePing() {
	c.LastPingTime = time.Now()

	pong := Message{
		Type: "pong",
	}

	data, err := json.Marshal(pong)
	if err != nil {
		c.Hub.log.Error("Error marshaling pong message: %v", err)
		return
	}

	c.Hub.log.Debug("Sending pong to client %s", c.ID)
	c.Send <- data
}

// SetUserOnline marks a user as online and notifies their followers
func (h *Hub) SetUserOnline(userID string) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	// Check if user already has active clients
	clients, exists := h.UserClients[userID]
	if exists && len(clients) > 0 {
		// User is already online, no need to broadcast again
		return
	}

	// Notify status service
	if h.statusUpdater != nil {
		go func() {
			if err := h.statusUpdater.SetUserOnline(userID); err != nil {
				h.log.Error("Failed to set user online in status service: %v", err)
			}
		}()
	}
}

// SetUserOffline marks a user as offline and notifies their followers
func (h *Hub) SetUserOffline(userID string) {
	h.Mu.Lock()

	// Check if user still has any active clients
	clients, exists := h.UserClients[userID]
	if exists {
		for _, client := range clients {
			if client.IsActive {
				// Close the connection
				client.IsActive = false
				client.Conn.Close()
			}
		}
	}

	// Remove all clients for this user
	if exists {
		delete(h.UserClients, userID)
	}

	// Double check if user has any remaining connections
	// This is a safety check to ensure we don't mark a user as offline
	// if they have other active connections we might have missed
	clients, exists = h.UserClients[userID]
	if exists && len(clients) > 0 {
		for _, client := range clients {
			if client.IsActive {
				break
			}
		}
	}

	h.Mu.Unlock()

	// Notify status service
	if h.statusUpdater != nil {
		go func() {
			if err := h.statusUpdater.SetUserOffline(userID); err != nil {
				h.log.Error("Failed to set user offline in status service: %v", err)
			}
		}()
	}
}

// BroadcastUserStatus sends a user's online status to specified recipients
func (h *Hub) BroadcastUserStatus(userID string, isOnline bool, recipientIDs []string) {
	for _, recipientID := range recipientIDs {
		// Don't send status update to the user themselves
		if recipientID == userID {
			continue
		}

		h.BroadcastToUser(recipientID, map[string]interface{}{
			"type": "user_status_update",
			"payload": map[string]interface{}{
				"userId":    userID,
				"isOnline":  isOnline,
				"timestamp": time.Now().Unix(),
			},
		})
	}
}

func (h *Hub) checkHeartbeats() {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	now := time.Now()
	for client := range h.Clients {
		if client.IsActive && now.Sub(client.LastPingTime) > h.heartbeatTimeout {
			h.log.Info("Client %s (User: %s) timed out due to missed heartbeats", client.ID, client.UserID)

			// Mark client as inactive
			client.IsActive = false

			// Close connection
			client.Conn.Close()

			// Remove from active clients
			delete(h.Clients, client)
			close(client.Send)
			close(client.Done)

			// Update user-specific clients map
			clients := h.UserClients[client.UserID]
			for i, c := range clients {
				if c.ID == client.ID {
					h.UserClients[client.UserID] = append(clients[:i], clients[i+1:]...)
					break
				}
			}

			// Clean up empty user entries
			if len(h.UserClients[client.UserID]) == 0 {
				delete(h.UserClients, client.UserID)

				// Mark user as offline if this was their last connection
				userID := client.UserID
				go func(userID string) {
					// Call SetUserOffline outside of the lock
					h.SetUserOffline(userID)
				}(userID)
			}
		}
	}
}

// SetStatusUpdater sets the status updater for the hub
func (h *Hub) SetStatusUpdater(updater StatusUpdater) {
	h.statusUpdater = updater
}

// HasClientWithID checks if a specific client ID exists and is active
func (h *Hub) HasClientWithID(clientID string) bool {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	for client := range h.Clients {
		if client.ID == clientID && client.IsActive {
			return true
		}
	}
	return false
}

// CloseClientWithID closes a specific client connection
func (h *Hub) CloseClientWithID(clientID string) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	for client := range h.Clients {
		if client.ID == clientID {
			client.IsActive = false
			client.Conn.Close()
			// Don't unregister here - let the ReadPump handle that
			// when the connection actually closes
			break
		}
	}
}

// sendPingToAllClients sends a ping message to all connected clients
func (h *Hub) sendPingToAllClients() {
	pingMessage := []byte(`{"type":"ping"}`)

	h.Mu.RLock()
	for client := range h.Clients {
		if client.IsActive {
			select {
			case client.Send <- pingMessage:
				// Successfully sent ping
			default:
				// Client's send buffer is full, close the connection
				close(client.Send)
				delete(h.Clients, client)
			}
		}
	}
	h.Mu.RUnlock()
}
