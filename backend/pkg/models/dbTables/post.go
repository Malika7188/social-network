package models

import (
	"database/sql"
	"time"
)

// Privacy level constants
const (
	PrivacyPublic        = "public"
	PrivacyAlmostPrivate = "almost_private"
	PrivacyPrivate       = "private"
)

// Post represents a user post in the database
type Post struct {
	ID            int64          `db:"id,pk"`
	UserID        string         `db:"user_id,notnull" index:"idx_post_user_id"`
	Content       string         `db:"content,notnull"`
	ImagePath     sql.NullString `db:"image_path"`
	VideoPath     sql.NullString `db:"video_path"`
	Privacy       string         `db:"privacy,notnull"`
	LikesCount    int64          `db:"likes_count,default=0"`
	CommentsCount int64          `db:"comments_count,default=0"`
	CreatedAt     time.Time      `db:"created_at,default=CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time      `db:"updated_at,notnull"`
	UserData      *PostUserData  `db:"-"`
}

// PostViewer represents which users can view a private post
type PostViewer struct {
	ID     int64  `db:"id,pk,autoincrement"`
	PostID int64  `db:"post_id,notnull" index:"idx_post_viewer_post_id"`
	UserID string `db:"user_id,notnull" index:"idx_post_viewer_user_id"`
}

// Comment represents a comment on a post
type Comment struct {
	ID        int64          `db:"id,pk,autoincrement"`
	PostID    int64          `db:"post_id,notnull" index:"idx_comment_post_id"`
	UserID    string         `db:"user_id,notnull" index:"idx_comment_user_id"`
	Content   string         `db:"content,notnull"`
	ImagePath sql.NullString `db:"image_path"`
	CreatedAt time.Time      `db:"created_at,default=CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `db:"updated_at,notnull"`
	UserData  *PostUserData  `db:"-"`
}

// PostLike represents a like on a post
type PostLike struct {
	ID        int64     `db:"id,pk,autoincrement"`
	PostID    int64     `db:"post_id,notnull" index:"idx_post_likes_post_id"`
	UserID    string    `db:"user_id,notnull" index:"idx_post_likes_user_id"`
	CreatedAt time.Time `db:"created_at,default=CURRENT_TIMESTAMP"`
}

type PostUserData struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Avatar    string `json:"avatar"`
}
