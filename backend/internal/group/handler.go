package group

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/Athooh/social-network/internal/auth"
	"github.com/Athooh/social-network/pkg/httputil"
	"github.com/Athooh/social-network/pkg/logger"
)

// Handler handles HTTP requests for group operations
type Handler struct {
	service Service
	log     *logger.Logger
}

// NewHandler creates a new group handler
func NewHandler(service Service, log *logger.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// CreateGroup handles the creation of a new group
func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
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
	name := r.FormValue("name")
	description := r.FormValue("description")
	isPublicStr := r.FormValue("privacy")
	isPublic := isPublicStr == "public"

	// Get file uploads
	var banner, profilePic *multipart.FileHeader
	if files := r.MultipartForm.File["banner"]; len(files) > 0 {
		banner = files[0]
	}
	if files := r.MultipartForm.File["profilePic"]; len(files) > 0 {
		profilePic = files[0]
	}

	// Create group
	group, err := h.service.CreateGroup(userID, name, description, isPublic, banner, profilePic)
	if err != nil {
		h.log.Error("Failed to create group: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusCreated, group)
}

// GetGroup handles getting a group by ID
func (h *Handler) GetGroup(w http.ResponseWriter, r *http.Request) {
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
	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Get group
	group, err := h.service.GetGroup(groupID, userID)
	if err != nil {
		h.log.Error("Failed to get group: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusOK, group)
}

// GetUserGroups handles getting all groups a user is a member of
func (h *Handler) GetUserGroups(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check for userId query parameter
	userID := r.URL.Query().Get("userId")
	viewerID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// If query parameter is empty, fallback to context
	if userID == "" {
		userID = viewerID
	}

	// Get groups
	groups, err := h.service.GetUserGroups(userID, viewerID)
	if err != nil {
		h.log.Error("Failed to get user groups: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusOK, groups)
}

// GetAllGroups handles getting all groups
func (h *Handler) GetAllGroups(w http.ResponseWriter, r *http.Request) {
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

	// Get pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
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

	// Get groups
	groups, err := h.service.GetAllGroups(userID, limit, offset)
	if err != nil {
		h.log.Error("Failed to get all groups: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusOK, groups)
}

// UpdateGroup handles updating a group
func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
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
	groupID := r.FormValue("id")
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	isPublicStr := r.FormValue("isPublic")
	isPublic := isPublicStr == "true"

	// Get file uploads
	var banner, profilePic *multipart.FileHeader
	if files := r.MultipartForm.File["banner"]; len(files) > 0 {
		banner = files[0]
	}
	if files := r.MultipartForm.File["profilePic"]; len(files) > 0 {
		profilePic = files[0]
	}

	// Update group
	group, err := h.service.UpdateGroup(groupID, userID, name, description, isPublic, banner, profilePic)
	if err != nil {
		h.log.Error("Failed to update group: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusOK, group)
}

// DeleteGroup handles deleting a group
func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
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

	// Get group ID from query
	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Delete group
	if err := h.service.DeleteGroup(groupID, userID); err != nil {
		h.log.Error("Failed to delete group: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Group deleted successfully"})
}

// InviteToGroup handles inviting a user to a group
func (h *Handler) InviteToGroup(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		GroupID   string `json:"groupId"`
		InviteeID string `json:"inviteeId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Invite user
	if err := h.service.InviteToGroup(req.GroupID, userID, req.InviteeID); err != nil {
		h.log.Error("Failed to invite user to group: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Invitation sent successfully"})
}

// JoinGroup handles a user requesting to join a group
func (h *Handler) JoinGroup(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		GroupID string `json:"groupId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Join group
	if err := h.service.JoinGroup(req.GroupID, userID); err != nil {
		h.log.Error("Failed to join group: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Join request sent successfully"})
}

// LeaveGroup handles a user leaving a group
func (h *Handler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		GroupID string `json:"groupId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Leave group
	if err := h.service.LeaveGroup(req.GroupID, userID); err != nil {
		h.log.Error("Failed to leave group: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Left group successfully"})
}

// AcceptInvitation handles accepting a group invitation
func (h *Handler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		GroupID string `json:"groupId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Accept invitation
	if err := h.service.AcceptInvitation(req.GroupID, userID); err != nil {
		h.log.Error("Failed to accept invitation: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Invitation accepted successfully"})
}

// RejectInvitation handles rejecting a group invitation
func (h *Handler) RejectInvitation(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		GroupID string `json:"groupId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Reject invitation
	if err := h.service.RejectInvitation(req.GroupID, userID); err != nil {
		h.log.Error("Failed to reject invitation: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Invitation rejected successfully"})
}

// AcceptJoinRequest handles accepting a join request
func (h *Handler) AcceptJoinRequest(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		GroupID string `json:"groupId"`
		UserID  string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Accept join request
	if err := h.service.AcceptJoinRequest(req.GroupID, userID, req.UserID); err != nil {
		h.log.Error("Failed to accept join request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Join request accepted successfully"})
}

// RejectJoinRequest handles rejecting a join request
func (h *Handler) RejectJoinRequest(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		GroupID string `json:"groupId"`
		UserID  string `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Reject join request
	if err := h.service.RejectJoinRequest(req.GroupID, userID, req.UserID); err != nil {
		h.log.Error("Failed to reject join request: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Join request rejected successfully"})
}

// UpdateMemberRole handles updating a member's role
func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
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

	// Parse request body
	var req struct {
		GroupID string `json:"groupId"`
		UserID  string `json:"userId"`
		Role    string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update member role
	if err := h.service.UpdateMemberRole(req.GroupID, userID, req.UserID, req.Role); err != nil {
		h.log.Error("Failed to update member role: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Member role updated successfully"})
}

// RemoveMember handles removing a member from a group
func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
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

	// Get query parameters
	groupID := r.URL.Query().Get("groupId")
	memberID := r.URL.Query().Get("userId")

	if groupID == "" || memberID == "" {
		http.Error(w, "Group ID and User ID are required", http.StatusBadRequest)
		return
	}

	// Remove member
	if err := h.service.RemoveMember(groupID, userID, memberID); err != nil {
		h.log.Error("Failed to remove member: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Member removed successfully"})
}

// GetGroupMembers handles getting all members of a group
func (h *Handler) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
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
	groupID := r.URL.Query().Get("groupId")
	status := r.URL.Query().Get("status")

	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Get members
	members, err := h.service.GetGroupMembers(groupID, userID, status)
	if err != nil {
		h.log.Error("Failed to get group members: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusOK, members)
}

// CreateGroupPost handles creating a post in a group
func (h *Handler) CreateGroupPost(w http.ResponseWriter, r *http.Request) {
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
	content := r.FormValue("content")

	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Get file uploads
	var image, video *multipart.FileHeader
	if files := r.MultipartForm.File["image"]; len(files) > 0 {
		image = files[0]
	}
	if files := r.MultipartForm.File["video"]; len(files) > 0 {
		video = files[0]
	}

	// Create post
	post, err := h.service.CreateGroupPost(groupID, userID, content, image, video)
	if err != nil {
		h.log.Error("Failed to create group post: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusCreated, post)
}

// GetGroupPosts handles getting all posts in a group
func (h *Handler) GetGroupPosts(w http.ResponseWriter, r *http.Request) {
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
	groupID := r.URL.Query().Get("groupId")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	limit := 10
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

	// Get posts
	posts, err := h.service.GetGroupPosts(groupID, userID, limit, offset)
	if err != nil {
		h.log.Error("Failed to get group posts: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusOK, posts)
}

// DeleteGroupPost handles deleting a post from a group
func (h *Handler) DeleteGroupPost(w http.ResponseWriter, r *http.Request) {
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

	// Get post ID from query
	postIDStr := r.URL.Query().Get("postId")
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Delete post
	if err := h.service.DeleteGroupPost(postID, userID); err != nil {
		h.log.Error("Failed to delete group post: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success response
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Post deleted successfully"})
}

// SendChatMessage handles sending a message to a group chat
func (h *Handler) SendChatMessage(w http.ResponseWriter, r *http.Request) {
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
		GroupID string `json:"groupId"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.GroupID == "" || request.Content == "" {
		http.Error(w, "Group ID and content are required", http.StatusBadRequest)
		return
	}

	// Send message
	message, err := h.service.SendChatMessage(request.GroupID, userID, request.Content)
	if err != nil {
		h.log.Error("Failed to send chat message: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusCreated, message)
}

// GetGroupChatMessages handles getting messages from a group chat
func (h *Handler) GetGroupChatMessages(w http.ResponseWriter, r *http.Request) {
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
	groupID := r.URL.Query().Get("groupId")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	limit := 50
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
	messages, err := h.service.GetGroupChatMessages(groupID, userID, limit, offset)
	if err != nil {
		h.log.Error("Failed to get group chat messages: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return response
	h.sendJSON(w, http.StatusOK, messages)
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
