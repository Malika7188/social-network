package models

import (
	"database/sql"
	"time"
)

type Notification struct {
	ID            int64          `db:"id,pk,autoincrement"`
	UserID        string         `db:"user_id,notnull" index:"idx_notification_user_id" references:"users(id) ON DELETE CASCADE"` // Recipient
	SenderID      sql.NullString `db:"sender_id,notnull" references:"users(id) ON DELETE CASCADE"`                                // Optional sender
	Type          string         `db:"type,notnull"`                                                                              // e.g., follow_request, group_invite
	Message       string         `db:"message,notnull"`                                                                           // Notification message
	IsRead        bool           `db:"is_read,notnull,default=false"`                                                             // Read status
	CreatedAt     time.Time      `db:"created_at,default=CURRENT_TIMESTAMP"`                                                      // Creation time
	TargetGroupID sql.NullString `db:"target_group_id" references:"groups(id) ON DELETE SET NULL"`                                // Nullable group FK
	TargetEventID sql.NullString `db:"target_event_id" references:"group_events(id) ON DELETE SET NULL"`                          // Nullable event FK
}
