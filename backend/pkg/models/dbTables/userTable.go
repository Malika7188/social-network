package models

import "time"

// User represents a user in the system
type User struct {
	ID          string    `db:"id,pk"`
	Email       string    `db:"email,notnull,unique" index:"unique"`
	Password    string    `db:"password,notnull"`
	FirstName   string    `db:"first_name,notnull"`
	LastName    string    `db:"last_name,notnull"`
	DateOfBirth string    `db:"date_of_birth,notnull"`
	Avatar      string    `db:"avatar"`
	Nickname    string    `db:"nickname"`
	AboutMe     string    `db:"about_me"`
	IsPublic    bool      `db:"is_public,default=TRUE"`
	CreatedAt   time.Time `db:"created_at,default=CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `db:"updated_at,default=CURRENT_TIMESTAMP"`
}

// UserStats represents additional statistics and metrics for a user
type UserStat struct {
	UserID         string    `db:"user_id,pk" index:"unique"`
	PostsCount     int       `db:"posts_count,default=0"`
	GroupsJoined   int       `db:"groups_joined,default=0"`
	FollowersCount int       `db:"followers_count,default=0"`
	FollowingCount int       `db:"following_count,default=0"`
	CreatedAt      time.Time `db:"created_at,default=CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `db:"updated_at,default=CURRENT_TIMESTAMP"`
}

// UserStatus represents a user's online status
type UserStatus struct {
	UserID       string    `db:"user_id,pk" index:"unique" references:"users(id) ON DELETE CASCADE"`
	IsOnline     bool      `db:"is_online,default=FALSE"`
	LastActivity time.Time `db:"last_activity,default=CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `db:"updated_at,default=CURRENT_TIMESTAMP"`
}

// UserProfile represents the extended profile information for a user
type UserProfile struct {
	ID           string    `db:"id,pk"`
	UserID       string    `db:"user_id,notnull" index:"unique"`
	BannerImage  string    `db:"banner_image"`
	ProfileImage string    `db:"profile_image"`
	Username     string    `db:"username"`
	FullName     string    `db:"full_name"`
	Bio          string    `db:"bio"`
	Work         string    `db:"work"`
	Education    string    `db:"education"`
	Email        string    `db:"email"`
	Phone        string    `db:"phone"`
	Website      string    `db:"website"`
	Location     string    `db:"location"`
	TechSkills   string    `db:"tech_skills"` // Comma-separated list
	SoftSkills   string    `db:"soft_skills"` // Comma-separated list
	Interests    string    `db:"interests"`   // Comma-separated list
	IsPrivate    bool      `db:"is_private,default=FALSE"`
	CreatedAt    time.Time `db:"created_at,default=CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `db:"updated_at,default=CURRENT_TIMESTAMP"`
}
