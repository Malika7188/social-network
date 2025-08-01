package user

import "time"

// Repository defines the user repository interface
type Repository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	Delete(id string) error
}

// StatusRepository defines the interface for user status operations
type StatusRepository interface {
	SetUserOnline(userID string) error
	SetUserOffline(userID string) error
	GetUserStatus(userID string) (bool, error)
	GetFollowersForStatusUpdate(userID string) ([]string, error)
	GetAllOnlineUsers() ([]string, error)
}

// User represents a user in the system
type User struct {
	ID             string
	Email          string
	Password       string
	FirstName      string
	LastName       string
	DateOfBirth    string
	Avatar         string
	Nickname       string
	AboutMe        string
	IsPublic       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	PostsCount     int
	GroupsJoined   int
	FollowersCount int
	FollowingCount int
	
	// Profile fields
	Username     string
	FullName     string
	Bio          string
	Work         string
	Education    string
	ContactEmail string
	Phone        string
	Website      string
	Location     string
	TechSkills   string
	SoftSkills   string
	Interests    string
	BannerImage  string
	ProfileImage string
	IsPrivate    bool
}
