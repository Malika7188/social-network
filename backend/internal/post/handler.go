package post

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Athooh/social-network/internal/auth"
	"github.com/Athooh/social-network/pkg/httputil"
	"github.com/Athooh/social-network/pkg/logger"
	models "github.com/Athooh/social-network/pkg/models/dbTables"
)

// Handler handles HTTP requests for posts
type Handler struct {
	service Service
	log     *logger.Logger
}

// NewHandler creates a new post handler
func NewHandler(service Service, log *logger.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// CreatePostRequest represents the request to create a post
type CreatePostRequest struct {
	Content string `json:"content"`
	Privacy string `json:"privacy"`
}

// PostResponse represents the response for a post
type PostResponse struct {
	ID         int64                `json:"id"`
	UserID     string               `json:"userId"`
	Content    string               `json:"content"`
	ImageURL   string               `json:"imageUrl,omitempty"`
	VideoURL   string               `json:"videoUrl,omitempty"`
	Privacy    string               `json:"privacy"`
	LikesCount int                  `json:"likesCount"`
	Comments   []CommentResponse    `json:"comments"`
	CreatedAt  string               `json:"createdAt"`
	UpdatedAt  string               `json:"updatedAt"`
	UserData   *models.PostUserData `json:"userData"`
}

// CommentResponse represents the response for a comment
type CommentResponse struct {
	ID        int64                `json:"id"`
	PostID    int64                `json:"postId"`
	UserID    string               `json:"userId"`
	Content   string               `json:"content"`
	ImageURL  string               `json:"imageUrl,omitempty"`
	CreatedAt string               `json:"createdAt"`
	UpdatedAt string               `json:"updatedAt"`
	UserData  *models.PostUserData `json:"userData"`
}

// PostWithCommentsResponse represents the response for a post with its comments
type PostWithCommentsResponse struct {
	ID         int64                `json:"id"`
	UserID     string               `json:"userId"`
	Content    string               `json:"content"`
	ImageURL   string               `json:"imageUrl,omitempty"`
	VideoURL   string               `json:"videoUrl,omitempty"`
	Privacy    string               `json:"privacy"`
	CreatedAt  string               `json:"createdAt"`
	UpdatedAt  string               `json:"updatedAt"`
	LikesCount int                  `json:"likesCount"`
	Comments   []CommentResponse    `json:"comments"`
	UserData   *models.PostUserData `json:"userData"`
}

// CreatePost handles the creation of a new post
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(20 << 20); err != nil { // 20 MB max for videos
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get form values
	content := r.FormValue("content")
	privacy := r.FormValue("privacy")
	viewers := r.Form["viewers"]

	if privacy == "" {
		h.sendError(w, http.StatusBadRequest, "Privacy field is missing")
		return
	}

	// Get image file if provided
	var imageFile *multipart.FileHeader
	if file, header, err := r.FormFile("image"); err == nil {
		defer file.Close()
		imageFile = header
	}

	// Get video file if provided
	var videoFile *multipart.FileHeader
	if file, header, err := r.FormFile("video"); err == nil {
		defer file.Close()
		videoFile = header
	}

	// Validate required fields
	if content == "" && imageFile == nil && videoFile == nil {
		h.sendError(w, http.StatusBadRequest, "Content, image, or video field is required")
		return
	}

	// Create post
	post, err := h.service.CreatePost(userID, content, privacy, imageFile, videoFile)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(viewers) > 0 {
		if err := h.service.SetPostViewers(post.ID, userID, viewers); err != nil {
			h.sendError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Prepare response
	response := PostResponse{
		ID:        post.ID,
		UserID:    post.UserID,
		Content:   post.Content,
		Privacy:   post.Privacy,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
		UpdatedAt: post.UpdatedAt.Format(time.RFC3339),
	}

	if post.ImagePath.String != "" {
		response.ImageURL = "/uploads/" + post.ImagePath.String
	}

	if post.VideoPath.String != "" {
		response.VideoURL = "/uploads/" + post.VideoPath.String
	}

	// Return response
	h.sendJSON(w, http.StatusCreated, response)
}

// GetPost handles retrieving a post by ID
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get post ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.sendError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	postID, err := strconv.ParseInt(pathParts[len(pathParts)-1], 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get post with comments
	post, comments, err := h.service.GetPostWithComments(postID, userID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare response
	response := PostWithCommentsResponse{
		ID:         post.ID,
		UserID:     post.UserID,
		Content:    post.Content,
		Privacy:    post.Privacy,
		LikesCount: int(post.LikesCount),
		CreatedAt:  post.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  post.UpdatedAt.Format(time.RFC3339),
		Comments:   make([]CommentResponse, 0, len(comments)),
		UserData:   post.UserData,
	}

	if post.ImagePath.String != "" {
		response.ImageURL = "/uploads/" + post.ImagePath.String
	}

	if post.VideoPath.String != "" {
		response.VideoURL = "/uploads/" + post.VideoPath.String
	}

	// Add comments to response
	for _, comment := range comments {
		commentResp := CommentResponse{
			ID:        comment.ID,
			PostID:    comment.PostID,
			UserID:    comment.UserID,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
			UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
		}
		if comment.ImagePath.String != "" {
			commentResp.ImageURL = "/uploads/" + comment.ImagePath.String
		}
		response.Comments = append(response.Comments, commentResp)
	}

	// Return response
	h.sendJSON(w, http.StatusOK, response)
}

// GetUserPosts handles retrieving all posts by a user
func (h *Handler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	viewerID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || viewerID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get target user ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.sendError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	targetID := pathParts[len(pathParts)-1]
	if viewerID == "" {
		h.sendError(w, http.StatusBadRequest, "Invalid Viewer ID")
		return
	}

	// Get posts
	posts, err := h.service.GetUserPosts(targetID, viewerID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare response
	var response []PostResponse
	for _, post := range posts {
		// Get comments for each post
		comments, err := h.service.GetPostComments(post.ID, viewerID)
		if err != nil {
			h.log.Error("Failed to get comments for post %d: %v", post.ID, err)
			continue
		}
		postResp := PostResponse{
			ID:         post.ID,
			UserID:     post.UserID,
			Content:    post.Content,
			Privacy:    post.Privacy,
			LikesCount: int(post.LikesCount),
			CreatedAt:  post.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  post.UpdatedAt.Format(time.RFC3339),
			Comments:   make([]CommentResponse, 0, len(comments)),
			UserData:   post.UserData,
		}

		if post.ImagePath.String != "" {
			postResp.ImageURL = "/uploads/" + post.ImagePath.String
		}

		if post.VideoPath.String != "" {
			postResp.VideoURL = "/uploads/" + post.VideoPath.String
		}
		if post.UserData.Avatar != "" {
			postResp.UserData.Avatar = "/uploads/" + postResp.UserData.Avatar
		}
		// Add comments to response
		for _, comment := range comments {
			commentResp := CommentResponse{
				ID:        comment.ID,
				PostID:    comment.PostID,
				UserID:    comment.UserID,
				Content:   comment.Content,
				CreatedAt: comment.CreatedAt.Format(time.RFC3339),
				UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
				UserData:  comment.UserData,
			}
			if comment.ImagePath.String != "" {
				commentResp.ImageURL = "/uploads/" + comment.ImagePath.String
			}
			postResp.Comments = append(postResp.Comments, commentResp)
		}

		response = append(response, postResp)
	}

	// Return response
	h.sendJSON(w, http.StatusOK, response)
}

// GetUserPosts handles retrieving all posts by a user
func (h *Handler) GetUserPhotos(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	viewerID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || viewerID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get target user ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.sendError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	targetID := pathParts[len(pathParts)-1]
	if viewerID == "" {
		h.sendError(w, http.StatusBadRequest, "Invalid Viewer ID")
		return
	}

	// Get posts
	posts, err := h.service.GetUserPosts(targetID, viewerID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var response []PostResponse
	for _, post := range posts {
		if post.ImagePath.String == "" {
			continue
		}
		postResp := PostResponse{
			ID:        post.ID,
			UserID:    post.UserID,
			Privacy:   post.Privacy,
			CreatedAt: post.CreatedAt.Format(time.RFC3339),
			UpdatedAt: post.UpdatedAt.Format(time.RFC3339),
			UserData:  post.UserData,
		}
		if post.ImagePath.String != "" {
			postResp.ImageURL = "/uploads/" + post.ImagePath.String
		}
		response = append(response, postResp)
	}
	h.sendJSON(w, http.StatusOK, response)
}

// GetPublicPosts handles retrieving public posts with pagination
func (h *Handler) GetPublicPosts(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default limit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
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

	// Get posts
	posts, err := h.service.GetPublicPosts(limit, offset)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare response
	var response []PostWithCommentsResponse
	for _, post := range posts {
		// Get comments for each post
		comments, err := h.service.GetPostComments(post.ID, userID)
		if err != nil {
			h.log.Error("Failed to get comments for post %d: %v", post.ID, err)
			continue
		}

		postResp := PostWithCommentsResponse{
			ID:         post.ID,
			UserID:     post.UserID,
			Content:    post.Content,
			Privacy:    post.Privacy,
			LikesCount: int(post.LikesCount),
			CreatedAt:  post.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  post.UpdatedAt.Format(time.RFC3339),
			Comments:   make([]CommentResponse, 0, len(comments)),
			UserData:   post.UserData,
		}

		if post.ImagePath.String != "" {
			postResp.ImageURL = "/uploads/" + post.ImagePath.String
		}

		if post.VideoPath.String != "" {
			postResp.VideoURL = "/uploads/" + post.VideoPath.String
		}

		// Add comments to response
		for _, comment := range comments {
			commentResp := CommentResponse{
				ID:        comment.ID,
				PostID:    comment.PostID,
				UserID:    comment.UserID,
				Content:   comment.Content,
				CreatedAt: comment.CreatedAt.Format(time.RFC3339),
				UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
			}
			if comment.ImagePath.String != "" {
				commentResp.ImageURL = "/uploads/" + comment.ImagePath.String
			}
			postResp.Comments = append(postResp.Comments, commentResp)
		}

		response = append(response, postResp)
	}

	// Return response
	h.sendJSON(w, http.StatusOK, response)
}

// UpdatePost handles updating an existing post
func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get post ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.sendError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	postID, err := strconv.ParseInt(pathParts[len(pathParts)-1], 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(20 << 20); err != nil { // 20 MB max for videos
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get form values
	content := r.FormValue("content")
	privacy := r.FormValue("privacy")

	// Validate required fields
	if content == "" {
		h.sendError(w, http.StatusBadRequest, "Content field is missing")
		return
	}

	// Get image file if provided
	var imageFile *multipart.FileHeader
	if file, header, err := r.FormFile("image"); err == nil {
		defer file.Close()
		imageFile = header
	}

	// Get video file if provided
	var videoFile *multipart.FileHeader
	if file, header, err := r.FormFile("video"); err == nil {
		defer file.Close()
		videoFile = header
	}

	// Update post
	post, err := h.service.UpdatePost(postID, userID, content, privacy, imageFile, videoFile)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare response
	response := PostResponse{
		ID:        post.ID,
		UserID:    post.UserID,
		Content:   post.Content,
		Privacy:   post.Privacy,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
		UpdatedAt: post.UpdatedAt.Format(time.RFC3339),
	}

	if post.ImagePath.String != "" {
		response.ImageURL = "/uploads/" + post.ImagePath.String
	}

	if post.VideoPath.String != "" {
		response.VideoURL = "/uploads/" + post.VideoPath.String
	}

	// Return response
	h.sendJSON(w, http.StatusOK, response)
}

// DeletePost handles deleting a post
func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body to get post ID
	var request struct {
		PostID int64 `json:"postId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if request.PostID <= 0 {
		h.sendError(w, http.StatusBadRequest, "Invalid or missing Post ID")
		return
	}

	// Delete post
	if err := h.service.DeletePost(request.PostID, userID); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// SetPostViewers handles setting the users who can view a private post
func (h *Handler) SetPostViewers(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get post ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.sendError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	postID, err := strconv.ParseInt(pathParts[len(pathParts)-1], 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Parse request body
	var request struct {
		ViewerIDs []string `json:"viewerIds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Set post viewers
	if err := h.service.SetPostViewers(postID, userID, request.ViewerIDs); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// CreateComment handles creating a new comment on a post
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get post ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.sendError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	postID, err := strconv.ParseInt(pathParts[len(pathParts)-1], 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get form values
	content := r.FormValue("content")

	// Get image file if provided
	var imageFile *multipart.FileHeader
	if file, header, err := r.FormFile("image"); err == nil {
		defer file.Close()
		imageFile = header
	}

	// Validate required fields
	if content == "" && imageFile == nil {
		h.sendError(w, http.StatusBadRequest, "Content or image field is missing please provide one")
		return
	}

	// Create comment
	comment, err := h.service.CreateComment(postID, userID, content, imageFile)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare response
	response := CommentResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt.Format(time.RFC3339),
		UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
		UserData:  comment.UserData,
	}

	if comment.ImagePath.String != "" {
		response.ImageURL = "/uploads/" + comment.ImagePath.String
	}

	// Return response
	h.sendJSON(w, http.StatusCreated, response)
}

// GetPostComments handles retrieving all comments for a post
func (h *Handler) GetPostComments(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get post ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.sendError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	postID, err := strconv.ParseInt(pathParts[len(pathParts)-1], 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get comments
	comments, err := h.service.GetPostComments(postID, userID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare response
	var response []CommentResponse
	for _, comment := range comments {
		commentResp := CommentResponse{
			ID:        comment.ID,
			PostID:    comment.PostID,
			UserID:    comment.UserID,
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
			UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
			UserData:  comment.UserData,
		}

		if comment.ImagePath.String != "" {
			commentResp.ImageURL = "/uploads/" + comment.ImagePath.String
		}

		response = append(response, commentResp)
	}

	// Return response
	h.sendJSON(w, http.StatusOK, response)
}

// DeleteComment handles deleting a comment
func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	postid := r.URL.Query().Get("postid")

	// Get comment ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.sendError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	commentID, err := strconv.ParseInt(pathParts[len(pathParts)-1], 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Delete comment
	if err := h.service.DeleteComment(commentID, userID, postid); err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// HandleComments handles comments for a post
func (h *Handler) HandleComments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.CreateComment(w, r)
	case http.MethodGet:
		h.GetPostComments(w, r)
	case http.MethodDelete:
		h.DeleteComment(w, r)
	default:
		h.sendError(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method not allowed: %s", r.Method))
	}
}

// GetFeedPosts handles retrieving posts from the user's feed
func (h *Handler) GetFeedPosts(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get pagination parameters
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 10 // Default page size
	}

	// Get feed posts
	posts, err := h.service.GetFeedPosts(userID, page, pageSize)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare response
	var response []PostWithCommentsResponse
	for _, post := range posts {
		// Get comments for each post
		comments, err := h.service.GetPostComments(post.ID, userID)
		if err != nil {
			h.log.Error("Failed to get comments for post %d: %v", post.ID, err)
			continue
		}

		postResp := PostWithCommentsResponse{
			ID:         post.ID,
			UserID:     post.UserID,
			Content:    post.Content,
			Privacy:    post.Privacy,
			LikesCount: int(post.LikesCount),
			CreatedAt:  post.CreatedAt.Format(time.RFC3339),
			UpdatedAt:  post.UpdatedAt.Format(time.RFC3339),
			Comments:   make([]CommentResponse, 0, len(comments)),
			UserData:   post.UserData,
		}

		if post.ImagePath.String != "" {
			postResp.ImageURL = "/uploads/" + post.ImagePath.String
		}

		if post.VideoPath.String != "" {
			postResp.VideoURL = "/uploads/" + post.VideoPath.String
		}

		if postResp.UserData.Avatar != "" {
			postResp.UserData.Avatar = "/uploads/" + postResp.UserData.Avatar
		}

		// Add comments to response
		for _, comment := range comments {
			commentResp := CommentResponse{
				ID:        comment.ID,
				PostID:    comment.PostID,
				UserID:    comment.UserID,
				Content:   comment.Content,
				CreatedAt: comment.CreatedAt.Format(time.RFC3339),
				UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
				UserData:  comment.UserData,
			}
			if comment.ImagePath.String != "" {
				commentResp.ImageURL = "/uploads/" + comment.ImagePath.String
			}
			postResp.Comments = append(postResp.Comments, commentResp)
		}

		response = append(response, postResp)
	}

	// Return response
	h.sendJSON(w, http.StatusOK, response)
}

// LikePost handles liking or unliking a post
func (h *Handler) LikePost(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		h.sendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get post ID from URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		h.sendError(w, http.StatusBadRequest, "Invalid URL")
		return
	}
	postID, err := strconv.ParseInt(pathParts[len(pathParts)-1], 10, 64)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Toggle like status
	isLiked, err := h.service.LikePost(postID, userID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get updated likes count
	post, err := h.service.GetPost(postID, userID)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return response
	response := struct {
		PostID     int64 `json:"postId"`
		LikesCount int   `json:"likesCount"`
		IsLiked    bool  `json:"isLiked"`
	}{
		PostID:     post.ID,
		LikesCount: int(post.LikesCount),
		IsLiked:    isLiked,
	}

	h.sendJSON(w, http.StatusOK, response)
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
