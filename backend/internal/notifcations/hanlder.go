package notifications

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Athooh/social-network/internal/auth"
	"github.com/Athooh/social-network/pkg/httputil"
	"github.com/Athooh/social-network/pkg/logger"
)

// Handler handles HTTP requests for notifications
type Handler struct {
	service Service
	log     *logger.Logger
}

// NewHandler creates a new notification handler
func NewHandler(service Service, log *logger.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// NotificationResponse represents the response for a notification
type NotificationResponse struct {
	ID            int64  `json:"id"`
	UserID        string `json:"userId"`
	SenderID      string `json:"senderId,omitempty"`
	Type          string `json:"type"`
	Message       string `json:"message"`
	IsRead        bool   `json:"isRead"`
	CreatedAt     string `json:"createdAt"`
	TargetGroupID string `json:"targetGroupId,omitempty"`
	TargetEventID string `json:"targetEventId,omitempty"`
	SenderName    string `json:"senderName,omitempty"`
	SenderAvatar  string `json:"senderAvatar,omitempty"`
}

// GetNotifications handles retrieving notifications for a user
func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default limit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 && parsedLimit <= 100 { // Max limit to prevent abuse
			limit = parsedLimit
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get notifications
	notifications, err := h.service.GetNotifications(userID, limit, offset)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare response
	response := make([]NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		resp := NotificationResponse{
			ID:            notification.ID,
			UserID:        notification.UserID,
			Type:          notification.Type,
			Message:       notification.Message,
			IsRead:        notification.IsRead,
			CreatedAt:     notification.CreatedAt.Format(time.RFC3339),
			TargetGroupID: notification.TargetGroupID.String,
			TargetEventID: notification.TargetEventID.String,
			SenderName:    notification.SenderName,
			SenderAvatar:  notification.SenderAvatar,
		}

		if notification.SenderID.Valid && *&notification.SenderID.String != "" {
			resp.SenderID = *&notification.SenderID.String
		}

		if notification.TargetGroupID.Valid {
			resp.TargetGroupID = *&notification.TargetGroupID.String
		}

		if notification.TargetEventID.Valid {
			resp.TargetEventID = *&notification.TargetEventID.String
		}

		response = append(response, resp)
	}

	// Return response
	h.sendJSON(w, http.StatusOK, response)
}

// MarkNotificationAsRead handles marking a single notification as read
func (h *Handler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	notificationID, err := strconv.ParseInt(r.URL.Query().Get("notificationId"), 10, 64)
	if err != nil || notificationID <= 0 {
		h.sendError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	// Mark notification as read
	if err := h.service.MarkNotificationAsRead(notificationID); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// MarkAllNotificationsAsRead handles marking all notifications as read for a user
func (h *Handler) MarkAllNotificationsAsRead(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Mark all notifications as read
	if err := h.service.MarkAllNotificationsAsRead(userID); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// ClearAllNotifications handles clearing all notifications for a user
func (h *Handler) ClearAllNotifications(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Clear all notifications
	if err := h.service.ClearAllNotifications(userID); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	notificationID, err := strconv.ParseInt(r.URL.Query().Get("notificationId"), 10, 64)
	if err != nil || notificationID <= 0 {
		h.sendError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	if err := h.service.DeleteNotification(notificationID); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
}

// sendJSON sends a JSON response
func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	httputil.SendJSON(w, status, data)
}

// sendError sends an error response
func (h *Handler) sendError(w http.ResponseWriter, status int, message string) {
	isWarning := status >= 500
	httputil.SendError(w, status, message, isWarning)
}
