package models

import (
	"time"
)

// RegisterRequest represents the data needed for user registration
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	DateOfBirth string `json:"dateOfBirth"`
	Avatar      string `json:"avatar"`
	Nickname    string `json:"nickname"`
	AboutMe     string `json:"aboutMe"`
}

// LoginRequest represents the data needed for user login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserResponse represents the user data returned to the client
type UserResponse struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName"`
	DateOfBirth    string    `json:"dateOfBirth"`
	Avatar         string    `json:"avatar"`
	Nickname       string    `json:"nickname"`
	AboutMe        string    `json:"aboutMe"`
	IsPublic       bool      `json:"isPublic"`
	CreatedAt      time.Time `json:"createdAt"`
	NumPosts       int       `json:"numPosts"`
	GroupsJoined   int       `json:"groupsJoined"`
	FollowersCount int       `json:"followersCount"`
	FollowingCount int       `json:"followingCount"`

	// Profile information
	Username     string `json:"username,omitempty"`
	FullName     string `json:"fullName,omitempty"`
	Bio          string `json:"bio,omitempty"`
	Work         string `json:"work,omitempty"`
	Education    string `json:"education,omitempty"`
	ContactEmail string `json:"contactEmail,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Website      string `json:"website,omitempty"`
	Location     string `json:"location,omitempty"`
	TechSkills   string `json:"techSkills,omitempty"`
	SoftSkills   string `json:"softSkills,omitempty"`
	Interests    string `json:"interests,omitempty"`
	BannerImage  string `json:"bannerImage,omitempty"`
	ProfileImage string `json:"profileImage,omitempty"`
	IsPrivate    bool   `json:"isPrivate,omitempty"`
}

// TokenResponse represents the JWT token response
type TokenResponse struct {
	Token     string       `json:"token"`
	ExpiresIn int          `json:"expires_in"`
	User      UserResponse `json:"user"`
}

// Add a new Session struct
type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
}
