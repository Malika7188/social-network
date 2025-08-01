package auth

import (
	"errors"
	"net/http"

	"github.com/Athooh/social-network/pkg/logger"
	models "github.com/Athooh/social-network/pkg/models/authModels"
	"github.com/Athooh/social-network/pkg/session"
	"github.com/Athooh/social-network/pkg/user"
)

// Service provides authentication functionality
type Service struct {
	userRepo       user.Repository
	sessionManager *session.SessionManager
	jwtConfig      JWTConfig
	statusRepo     user.StatusRepository
}

// NewService creates a new authentication service
func NewService(userRepo user.Repository, sessionManager *session.SessionManager, jwtConfig JWTConfig, statusRepo user.StatusRepository) *Service {
	return &Service{
		userRepo:       userRepo,
		sessionManager: sessionManager,
		jwtConfig:      jwtConfig,
		statusRepo:     statusRepo,
	}
}

// Register creates a new user account
func (s *Service) Register(req models.RegisterRequest) (*models.TokenResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already registered")
	}

	// Hash the password
	hashedPassword, err := session.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create the user
	newUser := &user.User{
		Email:       req.Email,
		Password:    hashedPassword,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DateOfBirth: req.DateOfBirth,
		Avatar:      req.Avatar,
		Nickname:    req.Nickname,
		AboutMe:     req.AboutMe,
		IsPublic:    true, // Default to public profile
	}

	if err := s.userRepo.Create(newUser); err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := GenerateToken(newUser.ID, s.jwtConfig)
	if err != nil {
		return nil, err
	}

	// Return the token
	return &models.TokenResponse{
		Token:     token,
		ExpiresIn: int(s.jwtConfig.TokenDuration.Seconds()),
		User: models.UserResponse{
			ID:          newUser.ID,
			Email:       newUser.Email,
			FirstName:   newUser.FirstName,
			LastName:    newUser.LastName,
			DateOfBirth: newUser.DateOfBirth,
			Avatar:      newUser.Avatar,
			Nickname:    newUser.Nickname,
			AboutMe:     newUser.AboutMe,
			IsPublic:    newUser.IsPublic,
			CreatedAt:   newUser.CreatedAt,
		},
	}, nil
}

// Logout ends a user's session and marks the user as offline
func (s *Service) Logout(w http.ResponseWriter, r *http.Request) error {
	// Get user ID from session before clearing it
	userID, err := s.sessionManager.GetUserFromSession(r)
	if err == nil && userID != "" {
		// Mark user as offline in the status repository
		if err := s.statusRepo.SetUserOffline(userID); err != nil {
			// Log the error but continue with logout
			logger.Error("Failed to mark user offline during logout: %v", err)
		}
	}

	// Clear the session
	return s.sessionManager.ClearSession(w, r)
}

// GetCurrentUser retrieves the currently authenticated user
func (s *Service) GetCurrentUser(r *http.Request) (*models.UserResponse, error) {
	// Get user ID from session
	userID, err := s.sessionManager.GetUserFromSession(r)
	if err != nil {
		return nil, err
	}

	// Get user data with profile
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Return user data with profile information
	return &models.UserResponse{
		ID:             user.ID,
		Email:          user.Email,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		DateOfBirth:    user.DateOfBirth,
		Avatar:         user.Avatar,
		Nickname:       user.Nickname,
		AboutMe:        user.AboutMe,
		IsPublic:       user.IsPublic,
		CreatedAt:      user.CreatedAt,
		NumPosts:       user.PostsCount,
		FollowersCount: user.FollowersCount,
		FollowingCount: user.FollowingCount,
		GroupsJoined:   user.GroupsJoined,
		
		// Include profile fields
		Username:     user.Username,
		FullName:     user.FullName,
		Bio:          user.Bio,
		Work:         user.Work,
		Education:    user.Education,
		ContactEmail: user.ContactEmail,
		Phone:        user.Phone,
		Website:      user.Website,
		Location:     user.Location,
		TechSkills:   user.TechSkills,
		SoftSkills:   user.SoftSkills,
		Interests:    user.Interests,
		BannerImage:  user.BannerImage,
		ProfileImage: user.ProfileImage,
		IsPrivate:    user.IsPrivate,
	}, nil
}

// LoginWithJWT authenticates a user and generates a JWT token
func (s *Service) LoginWithJWT(req models.LoginRequest) (*models.TokenResponse, error) {
	// Find the user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check the password
	if !session.CheckPassword(user.Password, req.Password) {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := GenerateToken(user.ID, s.jwtConfig)
	if err != nil {
		return nil, err
	}

	// Return the token
	return &models.TokenResponse{
		Token:     token,
		ExpiresIn: int(s.jwtConfig.TokenDuration.Seconds()),
		User: models.UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			DateOfBirth: user.DateOfBirth,
			Avatar:      user.Avatar,
			Nickname:    user.Nickname,
			AboutMe:     user.AboutMe,
			IsPublic:    user.IsPublic,
			CreatedAt:   user.CreatedAt,
		},
	}, nil
}
