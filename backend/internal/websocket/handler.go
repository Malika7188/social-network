package websocket

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Athooh/social-network/internal/auth"
	"github.com/Athooh/social-network/internal/user"

	"github.com/Athooh/social-network/pkg/httputil"
	"github.com/Athooh/social-network/pkg/logger"
	ws "github.com/Athooh/social-network/pkg/websocket"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Handler handles WebSocket connections
type Handler struct {
	hub           *ws.Hub
	log           *logger.Logger
	statusService *user.StatusService
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *ws.Hub, log *logger.Logger, statusService *user.StatusService) *Handler {
	return &Handler{
		hub:           hub,
		log:           log,
		statusService: statusService,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, you should check the origin
		return true
	},
}

// HandleConnection handles WebSocket connections
func (h *Handler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.SendError(w, http.StatusUnauthorized, "(WebSocket) Unauthorized", false)
		return
	}

	// Get tab ID from query parameters
	tabID := r.URL.Query().Get("tabId")
	if tabID == "" {
		// Generate a fallback ID if none provided
		tabID = uuid.New().String()
	}

	// Create a client ID that combines user and tab IDs
	clientID := fmt.Sprintf("%s:%s", userID, tabID)

	// Check if this tab already has an active connection
	if h.hub.HasClientWithID(clientID) {
		h.log.Info("Closing existing connection for client %s", clientID)
		h.hub.CloseClientWithID(clientID)
	}

	

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		httputil.SendError(w, http.StatusInternalServerError, "(WebSocket) Failed to upgrade connection", false)
		return
	}

	// Create a new client
	client := &ws.Client{
		ID:           clientID,
		UserID:       userID,
		TabID:        tabID,
		Conn:         conn,
		Hub:          h.hub,
		Send:         make(chan []byte, 256),
		IsActive:     true,
		Done:         make(chan struct{}),
		LastPingTime: time.Now(),
	}

	// Register client with hub
	h.hub.Register <- client

	// Mark user as online (if not already)
	if h.statusService != nil && !h.hub.HasActiveClient(userID) {
		go h.statusService.SetUserOnline(userID)
	}

	h.log.Info("New WebSocket connection established for user: %s, tab: %s", userID, tabID)

	// Start goroutines for reading and writing
	go client.ReadPump()
	go client.WritePump()

	// Add a deferred function to mark user as offline when all connections are closed
	go func() {
		// Wait for this client to be unregistered
		<-client.Done

		// Check if user has any remaining active connections
		if !h.hub.HasActiveClient(userID) && h.statusService != nil {
			h.statusService.SetUserOffline(userID)
		}
	}()
}
