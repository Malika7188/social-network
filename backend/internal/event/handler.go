package event

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/Athooh/social-network/internal/auth"
	"github.com/Athooh/social-network/pkg/httputil"
	"github.com/Athooh/social-network/pkg/logger"
)

// Handler handles HTTP requests for event operations
type Handler struct {
	service Service
	log     *logger.Logger
}

// NewHandler creates a new event handler
func NewHandler(service Service, log *logger.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// CreateEvent handles creating an event in a group
func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	groupID := r.FormValue("groupId")
	title := r.FormValue("title")
	description := r.FormValue("description")
	eventDateStr := r.FormValue("eventDate")
	response := r.FormValue("attendance")

	if groupID == "" || title == "" || eventDateStr == "" {
		http.Error(w, "Group ID, title, and event date are required", http.StatusBadRequest)
		return
	}

	// Parse event date
	eventDate, err := time.Parse(time.RFC3339, eventDateStr)
	if err != nil {
		http.Error(w, "Invalid event date format", http.StatusBadRequest)
		return
	}

	// Get banner file
	var banner *multipart.FileHeader
	if files := r.MultipartForm.File["banner"]; len(files) > 0 {
		banner = files[0]
	}

	// Create event
	event, err := h.service.CreateEvent(groupID, userID, title, description, eventDate, banner, response)
	if err != nil {
		h.log.Error("Failed to create event: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusCreated, event)
}

// GetEvent handles getting an event by ID
func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get event ID from query
	eventID := r.URL.Query().Get("eventId")
	if eventID == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	// Get event
	event, err := h.service.GetEvent(eventID, userID)
	if err != nil {
		h.log.Error("Failed to get event: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusOK, event)
}

// GetGroupEvents handles getting all events in a group
func (h *Handler) GetGroupEvents(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get group ID from query
	groupID := r.URL.Query().Get("groupId")
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Get events
	events, err := h.service.GetGroupEvents(groupID, userID)
	if err != nil {
		h.log.Error("Failed to get group events: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusOK, events)
}

// UpdateEvent handles updating an event
func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	// Only allow PUT method
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	eventID := r.FormValue("eventId")
	title := r.FormValue("title")
	description := r.FormValue("description")
	eventDateStr := r.FormValue("eventDate")

	if eventID == "" || title == "" || eventDateStr == "" {
		http.Error(w, "Event ID, title, and event date are required", http.StatusBadRequest)
		return
	}

	// Parse event date
	eventDate, err := time.Parse(time.RFC3339, eventDateStr)
	if err != nil {
		http.Error(w, "Invalid event date format", http.StatusBadRequest)
		return
	}

	// Get banner file
	var banner *multipart.FileHeader
	if files := r.MultipartForm.File["banner"]; len(files) > 0 {
		banner = files[0]
	}

	// Update event
	event, err := h.service.UpdateEvent(eventID, userID, title, description, eventDate, banner)
	if err != nil {
		h.log.Error("Failed to update event: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusOK, event)
}

// DeleteEvent handles deleting an event
func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	// Only allow DELETE method
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get event ID from query
	eventID := r.URL.Query().Get("eventId")
	if eventID == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	// Delete event
	if err := h.service.DeleteEvent(eventID, userID); err != nil {
		h.log.Error("Failed to delete event: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Event deleted successfully"})
}

// RespondToEvent handles responding to an event
func (h *Handler) RespondToEvent(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var request struct {
		EventID  string `json:"eventId"`
		Response string `json:"response"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.EventID == "" || request.Response == "" {
		http.Error(w, "Event ID and response are required", http.StatusBadRequest)
		return
	}

	// Respond to event
	if err := h.service.RespondToEvent(request.EventID, userID, request.Response); err != nil {
		h.log.Error("Failed to respond to event: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Response recorded successfully"})
}

// GetEventResponses handles getting responses to an event
func (h *Handler) GetEventResponses(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get query parameters
	eventID := r.URL.Query().Get("eventId")
	responseType := r.URL.Query().Get("responseType")

	if eventID == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	// Get responses
	responses, err := h.service.GetEventResponses(eventID, userID, responseType)
	if err != nil {
		h.log.Error("Failed to get event responses: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusOK, responses)
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
