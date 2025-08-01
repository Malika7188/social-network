package chat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Athooh/social-network/internal/auth"
	"github.com/Athooh/social-network/pkg/httputil"
	"github.com/Athooh/social-network/pkg/logger"
)

// Handler handles HTTP requests for chat functionality
type Handler struct {
	service Service
	log     *logger.Logger
}

// NewHandler creates a new chat handler
func NewHandler(service Service, log *logger.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// SendMessage handles sending a private message
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method not allowed %s", r.Method))
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var request struct {
		ReceiverID string `json:"receiverId"`
		Content    string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.ReceiverID == "" || request.Content == "" {
		h.sendError(w, http.StatusBadRequest, "Receiver ID and content are required")
		return
	}

	// Send message
	message, err := h.service.SendMessage(userID, request.ReceiverID, request.Content)
	if err != nil {
		h.log.Error("Failed to send message: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Mark messages from receiver as read when sending a message
	// This ensures that when you send a message, you've implicitly read all previous messages
	if err := h.service.MarkAsRead(request.ReceiverID, userID); err != nil {
		// Log the error but don't fail the request since the message was sent successfully
		h.log.Error("Failed to mark messages as read after sending: %v", err)
	}

	// Return response
	h.sendJSON(w, http.StatusCreated, message)
}

// GetMessages handles getting messages between two users
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method not allowed %s", r.Method))
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get query parameters
	otherUserID := r.URL.Query().Get("userId")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if otherUserID == "" {
		h.sendError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	limit := 100
	offset := 0

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get messages
	messages, err := h.service.GetMessages(userID, otherUserID, limit, offset)
	if err != nil {
		h.log.Error("Failed to get messages: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return response
	h.sendJSON(w, http.StatusOK, messages)
}

// MarkAsRead handles marking messages as read
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method not allowed %s", r.Method))
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var request struct {
		SenderID string `json:"senderId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.SenderID == "" {
		h.sendError(w, http.StatusBadRequest, "Sender ID is required")
		return
	}

	// Mark messages as read
	if err := h.service.MarkAsRead(request.SenderID, userID); err != nil {
		h.log.Error("Failed to mark messages as read: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GetContacts handles getting chat contacts
func (h *Handler) GetContacts(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method %s not allowed", r.Method))
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get contacts
	contacts, err := h.service.GetContacts(userID)
	if err != nil {
		h.log.Error("Failed to get contacts: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return response
	h.sendJSON(w, http.StatusOK, contacts)
}

// SendTypingIndicator handles sending a typing indicator
func (h *Handler) SendTypingIndicator(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method not allowed %s", r.Method))
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var request struct {
		ReceiverID string `json:"receiverId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.ReceiverID == "" {
		h.sendError(w, http.StatusBadRequest, "Receiver ID is required")
		return
	}

	// Send typing indicator via WebSocket
	if err := h.service.SendTypingIndicator(userID, request.ReceiverID); err != nil {
		h.log.Error("Failed to send typing indicator: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// SearchMessages handles searching for messages
func (h *Handler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	h.log.Info("SearchMessages endpoint called")

	// Only allow GET method
	if r.Method != http.MethodGet {
		h.log.Warn("Invalid method for search: %s", r.Method)
		h.sendError(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method not allowed %s", r.Method))
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.log.Warn("Search request without valid user ID")
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get query parameters
	query := r.URL.Query().Get("q")
	if query == "" {
		h.log.Warn("Search request without query parameter")
		h.sendError(w, http.StatusBadRequest, "Search query is required")
		return
	}

	// Get optional otherUserId parameter for conversation-specific search
	otherUserID := r.URL.Query().Get("otherUserId")

	h.log.Info("Searching messages for user %s with query: %s, otherUserId: %s", userID, query, otherUserID)

	// Get optional limit parameter
	limit := 100 // default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	// Search messages
	messages, err := h.service.SearchMessages(userID, query, otherUserID, limit)
	if err != nil {
		h.log.Error("Failed to search messages: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.log.Info("Search completed successfully. Found %d messages", len(messages))

	// Return messages
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"messages": messages,
		"query":    query,
		"count":    len(messages),
	})
}

// Helper method to send JSON responses
func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	httputil.SendJSON(w, status, data)
}

// Helper method to send error responses
func (h *Handler) sendError(w http.ResponseWriter, status int, message string) {
	var isWarning bool = false
	if status >= 500 {
		isWarning = true
	}
	httputil.SendError(w, status, message, isWarning)
}
