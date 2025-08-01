package models

import "time"

// Session represents a user session
type Session struct {
	ID        string    `db:"id,pk,"`
	UserID    string    `db:"user_id,notnull" index:"" references:"users(id) ON DELETE CASCADE"`
	ExpiresAt time.Time `db:"expires_at,notnull" index:""`
	CreatedAt time.Time `db:"created_at,default=CURRENT_TIMESTAMP"`
}
