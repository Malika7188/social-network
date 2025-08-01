package models

import "time"

// FollowRequest represents a follow request between users
type FollowRequest struct {
	ID          int64     `db:"id,pk,autoincrement"`
	FollowerID  string    `db:"follower_id,notnull" index:"idx_follow_requests_follower_id" references:"users(id) ON DELETE CASCADE"`
	FollowingID string    `db:"following_id,notnull" index:"idx_follow_requests_following_id" references:"users(id) ON DELETE CASCADE"`
	Status      string    `db:"status,notnull" index:"idx_follow_requests_status"`
	CreatedAt   time.Time `db:"created_at,default=CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `db:"updated_at,default=CURRENT_TIMESTAMP"`

	// Add unique constraint for follower_id and following_id
	_ struct{} `db:"unique:follower_id,following_id"`
}

// Follower represents a follower relationship
type Follower struct {
	ID          int64     `db:"id,pk,autoincrement"`
	FollowerID  string    `db:"follower_id,notnull" index:"idx_followers_follower_id" references:"users(id) ON DELETE CASCADE"`
	FollowingID string    `db:"following_id,notnull" index:"idx_followers_following_id" references:"users(id) ON DELETE CASCADE"`
	CreatedAt   time.Time `db:"created_at,default=CURRENT_TIMESTAMP"`

	// Add unique constraint for follower_id and following_id
	_ struct{} `db:"unique:follower_id,following_id"`
}
