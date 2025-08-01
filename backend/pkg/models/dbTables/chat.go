package models

import "time"

// PrivateMessage represents a message between two users
type PrivateMessage struct {
	ID         int64     `json:"id" db:"id,pk,autoincrement"`
	SenderID   string    `json:"senderId" db:"sender_id,notnull" index:"idx_private_messages_sender_id" references:"users(id) ON DELETE CASCADE"`
	ReceiverID string    `json:"receiverId" db:"receiver_id,notnull" index:"idx_private_messages_receiver_id" references:"users(id) ON DELETE CASCADE"`
	Content    string    `json:"content" db:"content,notnull"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at,default=CURRENT_TIMESTAMP"`
	ReadAt     time.Time `json:"readAt,omitempty" db:"read_at"`
	IsRead     bool      `json:"isRead" db:"is_read,default=FALSE"`

	// Populated fields (not stored in DB)
	Sender   *UserBasic `json:"sender,omitempty" db:"-"`
	Receiver *UserBasic `json:"receiver,omitempty" db:"-"`
}

// ChatContact represents a user that the current user can chat with
type ChatContact struct {
	ID          int64     `json:"id" db:"id,pk,autoincrement"`
	UserID      string    `json:"userId" db:"user_id,notnull" references:"users(id) ON DELETE CASCADE"`
	FirstName   string    `json:"firstName" db:"first_name,notnull"`
	LastName    string    `json:"lastName" db:"last_name,notnull"`
	Avatar      string    `json:"avatar" db:"avatar"`
	IsOnline    bool      `json:"isOnline" db:"is_online,default=FALSE"`
	LastMessage string    `json:"lastMessage,omitempty" db:"last_message"`
	LastSent    time.Time `json:"lastSent,omitempty" db:"last_sent"`
	UnreadCount int       `json:"unreadCount" db:"unread_count,default=0"`

	// Populated fields (not stored in DB)
	LastMessageSenderID string `json:"lastMessageSenderId,omitempty" db:"-"`
}
