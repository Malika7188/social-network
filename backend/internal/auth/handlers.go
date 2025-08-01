package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Athooh/social-network/pkg/filestore"
	"github.com/Athooh/social-network/pkg/httputil"
	"github.com/Athooh/social-network/pkg/logger"
	models "github.com/Athooh/social-network/pkg/models/authModels"
)

// Handler provides HTTP handlers for authentication
type Handler struct {
	service   *Service
	fileStore *filestore.FileStore
}

// NewHandler creates a new authentication handler
func NewHandler(service *Service, fileStore *filestore.FileStore) *Handler {
	return &Handler{
		service:   service,
		fileStore: fileStore,
	}
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method not allowed: %s", r.Method), false)
		return
	}

	// Parse multipart form with 10MB max memory
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse form data: %s", err.Error()), false)
		return
	}

	// Get form fields
	req := models.RegisterRequest{
		Email:       r.FormValue("email"),
		Password:    r.FormValue("password"),
		FirstName:   r.FormValue("firstName"),
		LastName:    r.FormValue("lastName"),
		DateOfBirth: r.FormValue("dateOfBirth"),
		Nickname:    r.FormValue("nickname"),
		AboutMe:     r.FormValue("aboutMe"),
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" ||
		req.FirstName == "" || req.LastName == "" || req.DateOfBirth == "" {
		h.sendError(w, http.StatusBadRequest,
			fmt.Sprintf("Missing required fields: email: %s, password: %s, firstName: %s, lastName: %s, dateOfBirth: %s",
				req.Email, req.Password, req.FirstName, req.LastName, req.DateOfBirth), true)
		return
	}

	// Handle avatar upload if present
	if file, header, err := r.FormFile("avatar"); err == nil {
		defer file.Close()

		filename, err := h.fileStore.SaveFile(header, "avatars")
		if err != nil {
			h.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save avatar: %s", err.Error()), false)
			return
		}
		req.Avatar = filename
	}

	// Register the user
	tokenResponse, err := h.service.Register(req)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("Failed to register user: %s", err.Error()), false)
		return
	}

	// Create a session
	if err := h.service.sessionManager.CreateSession(w, tokenResponse.User.ID); err != nil {
		h.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create session: %s", err.Error()), false)
		return
	}

	logger.Info("User registered successfully: %s %s (%s)",
		req.FirstName, req.LastName, req.Email)

	// Return success response
	h.sendJSON(w, http.StatusCreated, tokenResponse)
}

// Helper method to send JSON responses
func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	httputil.SendJSON(w, status, data)
}

// Helper method to send error responses
func (h *Handler) sendError(w http.ResponseWriter, status int, message string, isWarning bool) {
	httputil.SendError(w, status, message, isWarning)
}

// Logout handles user logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method not allowed: %s", r.Method), false)
		return
	}

	// Logout the user
	if err := h.service.Logout(w, r); err != nil {
		h.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to logout user: %s", err.Error()), false)
		return
	}

	// Return success
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// Me returns the current user's information
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method not allowed: %s", r.Method), false)
		return
	}

	// Get the current user
	user, err := h.service.GetCurrentUser(r)
	if err != nil {
		h.sendError(w, http.StatusUnauthorized, fmt.Sprintf("Failed to get current user: %s", err.Error()), false)
		return
	}

	// Return the user data
	h.sendJSON(w, http.StatusOK, user)
}

// LoginJWT handles user login with JWT authentication
func (h *Handler) LoginJWT(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method not allowed: %s", r.Method), false)
		return
	}

	// Parse request body
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %s", err.Error()), false)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		h.sendError(w, http.StatusBadRequest, fmt.Sprintf("Missing required fields: email: %s, password: %s", req.Email, req.Password), true)
		return
	}

	// Login the user with JWT
	tokenResponse, err := h.service.LoginWithJWT(req)
	if err != nil {
		h.sendError(w, http.StatusUnauthorized, fmt.Sprintf("Failed to login user: %s", err.Error()), true)
		return
	}

	// Create a session
	if err := h.service.sessionManager.CreateSession(w, tokenResponse.User.ID); err != nil {
		h.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create session: %s", err.Error()), false)
		return
	}

	// Return the token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tokenResponse)
}

// ValidateToken validates a JWT token
func (h *Handler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, fmt.Sprintf("Method not allowed: %s", r.Method), false)
		return
	}

	// First try to extract token from request
	tokenString, err := ExtractTokenFromRequest(r)
	if err != nil {

		// If no token in header, try to get user from session
		userID, err := h.service.sessionManager.GetUserFromSession(r)
		if err != nil {
			h.sendError(w, http.StatusUnauthorized, fmt.Sprintf("Authentication failed: %s", err.Error()), true)
			return
		}

		// If we have a user ID from session, token is valid
		h.sendJSON(w, http.StatusOK, map[string]string{"message": "Token is valid", "userId": userID})
		return
	}

	// Validate token
	_, err = ValidateToken(tokenString, h.service.jwtConfig)
	if err != nil {
		h.sendError(w, http.StatusUnauthorized, fmt.Sprintf("Invalid token: %s", err.Error()), true)
		return
	}

	// Token is valid
	h.sendJSON(w, http.StatusOK, map[string]string{"message": "Token is valid"})
}
