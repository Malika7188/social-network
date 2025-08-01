package profile

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// Service interface defines the operations for profile management
type Service interface {
	UpdateProfile(userID string, profileData map[string]interface{}) error
	SaveProfileImage(userID string, fileHeader *multipart.FileHeader) (string, error)
	SaveBannerImage(userID string, fileHeader *multipart.FileHeader) (string, error)
	GetProfileByUserID(userID string) (*UserProfileData, error)
	ValidateProfileViewRequest(userID string, targetID string) (bool, error)
}

// UserProfileData represents the combined user and profile data
type UserProfileData struct {
	// User data
	ID               string    `json:"id"`
	Email            string    `json:"email,omitempty"` // Authentication email (omitted from JSON)
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	Nickname         string    `json:"nickname"`
	AboutMe          string    `json:"aboutMe"`
	Avatar           string    `json:"avatar"`
	IsPublic         bool      `json:"isPublic"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	Username         string    `json:"username"`
	FullName         string    `json:"fullName"`
	Bio              string    `json:"bio"`
	Work             string    `json:"work"`
	Education        string    `json:"education"`
	ContactEmail     string    `json:"contactEmail"` // Contact email
	Phone            string    `json:"phone"`
	Website          string    `json:"website"`
	Location         string    `json:"location"`
	TechSkills       string    `json:"techSkills"`
	SoftSkills       string    `json:"softSkills"`
	Interests        string    `json:"interests"`
	BannerImage      string    `json:"bannerImage"`
	ProfileImage     string    `json:"profileImage"`
	IsPrivate        bool      `json:"isPrivate"`
	ProfileCreatedAt time.Time `json:"profileCreatedAt"`
	ProfileUpdatedAt time.Time `json:"profileUpdatedAt"`
	FollowersCount   int       `json:"followersCount"`
	FollowingCount   int       `json:"followingCount"`
}

// ProfileService implements the Service interface
type ProfileService struct {
	repo      Repository
	uploadDir string
}

// NewService creates a new profile service
func NewService(repo Repository, uploadDir string) Service {
	// Ensure upload directory exists
	os.MkdirAll(uploadDir, os.ModePerm)
	return &ProfileService{
		repo:      repo,
		uploadDir: uploadDir,
	}
}

// UpdateProfile updates a user's profile
func (s *ProfileService) UpdateProfile(userID string, profileData map[string]interface{}) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	return s.repo.UpdateUserProfile(userID, profileData)
}

// SaveProfileImage saves a profile image and returns the path
func (s *ProfileService) SaveProfileImage(userID string, fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", nil // No file provided, not an error
	}
	return s.saveImage(userID, fileHeader, "avatars")
}

// SaveBannerImage saves a banner image and returns the path
func (s *ProfileService) SaveBannerImage(userID string, fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", nil // No file provided, not an error
	}
	return s.saveImage(userID, fileHeader, "banners")
}

// GetProfileByUserID retrieves a user's complete profile data
func (s *ProfileService) GetProfileByUserID(userID string) (*UserProfileData, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	return s.repo.GetUserProfileByID(userID)
}

// saveImage is a helper function to save images
func (s *ProfileService) saveImage(userID string, fileHeader *multipart.FileHeader, imageType string) (string, error) {
	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a unique filename
	fileExt := filepath.Ext(fileHeader.Filename)
	fileName := fmt.Sprintf("%s_%s_%d%s", userID, imageType, time.Now().Unix(), fileExt)

	// Create user directory if it doesn't exist
	userDir := filepath.Join(s.uploadDir, imageType)
	if err := os.MkdirAll(userDir, os.ModePerm); err != nil {
		return "", err
	}

	// Create the destination file
	filePath := filepath.Join(userDir, fileName)
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	// Return the relative path to be stored in the database
	// This assumes the uploadDir is accessible via a URL path
	return filepath.Join(imageType, fileName), nil
}

func (s *ProfileService) ValidateProfileViewRequest(userID string, targetID string) (bool, error) {
	if userID == targetID {
		return true, nil
	}
	isPublic, err := s.repo.IsUserProfilePublic(targetID)
	if err != nil {
		return false, fmt.Errorf("failed to check profile visibility: %w", err)
	}
	if isPublic {
		return true, nil
	}
	isFollowing, err := s.repo.IsUserFollowing(userID, targetID)
	if err != nil {
		return false, fmt.Errorf("failed to check following status: %w", err)
	}
	return isFollowing, nil
}
