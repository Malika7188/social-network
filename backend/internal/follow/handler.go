package follow

import (
	"encoding/json"
	"net/http"

	"github.com/Athooh/social-network/internal/auth"
	"github.com/Athooh/social-network/pkg/httputil"
	"github.com/Athooh/social-network/pkg/logger"
)

// Handler handles HTTP requests for follow functionality
type Handler struct {
	service Service
	log     *logger.Logger
}

// NewHandler creates a new follow handler
func NewHandler(service Service, log *logger.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// FollowUser handles a request to follow a user
func (h *Handler) FollowUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	followerID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || followerID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var request struct {
		UserID string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.UserID == "" {
		h.sendError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Follow the user
	autoFollowed, err := h.service.FollowUser(followerID, request.UserID)
	if err != nil {
		h.log.Error("Failed to follow user: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return response
	response := struct {
		Success      bool `json:"success"`
		AutoFollowed bool `json:"autoFollowed"`
	}{
		Success:      true,
		AutoFollowed: autoFollowed,
	}

	h.sendJSON(w, http.StatusOK, response)
}

// UnfollowUser handles a request to unfollow a user
func (h *Handler) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	followerID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || followerID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var request struct {
		UserID string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.UserID == "" {
		h.sendError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Unfollow the user
	if err := h.service.UnfollowUser(followerID, request.UserID); err != nil {
		h.log.Error("Failed to unfollow user: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// AcceptFollowRequest handles a request to accept a follow request
func (h *Handler) AcceptFollowRequest(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	followingID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || followingID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var request struct {
		UserID string `json:"followerId"` // This is the followerID (user who sent the request)
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.UserID == "" {
		h.sendError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Accept the follow request
	if err := h.service.AcceptFollowRequest(request.UserID, followingID); err != nil {
		h.log.Error("Failed to accept follow request: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// DeclineFollowRequest handles a request to decline a follow request
func (h *Handler) DeclineFollowRequest(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	followingID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || followingID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var request struct {
		UserID string `json:"followerId"` // This is the followerID (user who sent the request)
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if request.UserID == "" {
		h.sendError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Decline the follow request
	if err := h.service.DeclineFollowRequest(request.UserID, followingID); err != nil {
		h.log.Error("Failed to decline follow request: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GetPendingFollowRequests handles a request to get all pending follow requests
func (h *Handler) GetPendingFollowRequests(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get pending follow requests
	requests, err := h.service.GetPendingFollowRequests(userID)
	if err != nil {
		h.log.Error("Failed to get pending follow requests: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the requests
	h.sendJSON(w, http.StatusOK, requests)
}

// GetFollowers handles a request to get all followers of a user
func (h *Handler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	profileID := r.URL.Query().Get("userId")
	if profileID == "" {
		profileID = userID
	}

	// Get followers
	followers, err := h.service.GetFollowers(profileID)
	if err != nil {
		h.log.Error("Failed to get followers: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the followers
	h.sendJSON(w, http.StatusOK, followers)
}

// GetFollowing handles a request to get all users a user is following
func (h *Handler) GetFollowing(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	profileID := r.URL.Query().Get("userId")
	if profileID == "" {
		profileID = userID
	}

	// Get following
	following, err := h.service.GetFollowing(profileID)
	if err != nil {
		h.log.Error("Failed to get following: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the following
	h.sendJSON(w, http.StatusOK, following)
}

// IsFollowing handles a request to check if a user is following another user
func (h *Handler) IsFollowing(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	followerID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || followerID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get target user ID from query parameters
	targetID := r.URL.Query().Get("userId")
	if targetID == "" {
		h.sendError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	// Check if following
	isFollowing, err := h.service.IsFollowing(followerID, targetID)
	if err != nil {
		h.log.Error("Failed to check if following: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the result
	h.sendJSON(w, http.StatusOK, map[string]bool{"isFollowing": isFollowing})
}

func (h *Handler) GetSuggestedFriends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	// Get suggested friends
	suggestions, err := h.service.GetSuggestedFriends(userID)
	if err != nil {
		h.log.Error("Failed to get suggested friends: %v", err)
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the suggestions
	h.sendJSON(w, http.StatusOK, map[string]interface{}{
		"suggestions": suggestions,
		"count":       len(suggestions),
	})
}

func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	httputil.SendJSON(w, status, data)
}

func (h *Handler) sendError(w http.ResponseWriter, status int, message string) {
	var isWarning bool = false
	if status >= 500 {
		isWarning = true
	}
	httputil.SendError(w, status, message, isWarning)
}
