package models

import (
	"database/sql"
	"time"
)

// Group represents a group in the system
type Group struct {
	ID             string         `db:"id,pk"`
	Name           string         `db:"name,notnull"`
	Description    string         `db:"description"`
	CreatorID      string         `db:"creator_id,notnull" index:"idx_groups_creator_id"`
	BannerPath     sql.NullString `db:"banner_path"`
	ProfilePicPath sql.NullString `db:"profile_pic_path"`
	IsPublic       bool           `db:"is_public,default=TRUE"`
	CreatedAt      time.Time      `db:"created_at,default=CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time      `db:"updated_at,default=CURRENT_TIMESTAMP"`

	// Non-DB fields
	MemberCount   int            `db:"-"`
	Creator       *UserBasic     `db:"-"`
	IsMember      bool           `db:"-"`
	MemberStatus  string         `db:"-"`
	Members 	[]*GroupMember   `db:"-"`
}

// GroupMember represents a member of a group
type GroupMember struct {
	ID        string    `db:"id,pk"`
	GroupID   string    `db:"group_id,notnull" index:"idx_group_members_group_id"`
	UserID    string    `db:"user_id,notnull" index:"idx_group_members_user_id"`
	Role      string    `db:"role,notnull"` // admin, moderator, member
	Status    string    `db:"status,notnull" index:"idx_group_members_status"` // pending, accepted, rejected
	InvitedBy string    `db:"invited_by"`
	CreatedAt time.Time `db:"created_at,default=CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `db:"updated_at,default=CURRENT_TIMESTAMP"`
	Avatar    string    `db:"avatar"`

	// Non-DB fields
	User      *UserBasic `db:"-"`
	Inviter   *UserBasic `db:"-"`
}

// GroupPost represents a post in a group
type GroupPost struct {
	ID            int64          `db:"id,pk"`
	GroupID       string         `db:"group_id,notnull" index:"idx_group_posts_group_id"`
	UserID        string         `db:"user_id,notnull" index:"idx_group_posts_user_id"`
	Content       string         `db:"content"`
	ImagePath     sql.NullString `db:"image_path"`
	VideoPath     sql.NullString `db:"video_path"`
	LikesCount    int64          `db:"likes_count,default=0"`
	CommentsCount int64          `db:"comments_count,default=0"`
	CreatedAt     time.Time      `db:"created_at,default=CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time      `db:"updated_at,default=CURRENT_TIMESTAMP"`

	// Non-DB fields
	User  *PostUserData `db:"-"`
	Group *GroupBasic   `db:"-"`
	Isliked bool          `db:"-"`
}

// GroupEvent represents an event in a group
type GroupEvent struct {
	ID          string         `db:"id,pk"`
	GroupID     string         `db:"group_id,notnull" index:"idx_group_events_group_id"`
	CreatorID   string         `db:"creator_id,notnull"`
	Title       string         `db:"title,notnull"`
	Description string         `db:"description"`
	EventDate   time.Time      `db:"event_date,notnull"`
	BannerPath  sql.NullString `db:"banner_path"`
	CreatedAt   time.Time      `db:"created_at,default=CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `db:"updated_at,default=CURRENT_TIMESTAMP"`


	// Non-DB fields
	Creator       *UserBasic  `db:"-"`
	Group         *GroupBasic `db:"-"`
	GoingCount    int         `db:"-"`
	NotGoingCount int         `db:"-"`
	UserResponse  string      `db:"-"` // The current user's response
}

// EventResponse represents a user's response to an event
type EventResponse struct {
	ID        string    `db:"id,pk"`
	EventID   string    `db:"event_id,notnull" index:"idx_event_responses_event_id"`
	UserID    string    `db:"user_id,notnull" index:"idx_event_responses_user_id"`
	Response  string    `db:"response,notnull"` // going, not_going, interested, maybe
	CreatedAt time.Time `db:"created_at,default=CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `db:"updated_at,default=CURRENT_TIMESTAMP"`


	// Non-DB fields
	User *UserBasic `db:"-"`
}

// GroupChatMessage represents a message in a group chat
type GroupChatMessage struct {
	ID        int64     `db:"id,pk,autoincrement"`
	GroupID   string    `db:"group_id,notnull" index:"idx_group_chat_messages_group_id"`
	UserID    string    `db:"user_id,notnull"`
	Content   string    `db:"content,notnull"`
	CreatedAt time.Time `db:"created_at,default=CURRENT_TIMESTAMP" index:"idx_group_chat_messages_created_at"`


	// Non-DB fields
	User *UserBasic `db:"-"`
}

// UserBasic contains basic user information for display
type UserBasic struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Avatar    string `json:"avatar"`
}

// GroupBasic contains basic group information for display
type GroupBasic struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	ProfilePicPath string `json:"profilePicPath"`
}
