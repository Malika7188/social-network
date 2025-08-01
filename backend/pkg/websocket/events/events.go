package events

// EventType defines the type of WebSocket event
type EventType string

const (
	PostCreated           EventType = "post_created"
	PostLiked             EventType = "post_liked"
	PostCommented         EventType = "post_commented"
	UserStatsUpdated      EventType = "user_stats_updated"
	FollowUpdate          EventType = "follow_update"
	FollowRequest         EventType = "follow_request"
	FollowRequestAccepted EventType = "follow_request_accepted"
	CommentCountUpdate    EventType = "comment_count_update"
	UserStatusUpdate      EventType = "user_status_update"
	GroupEventCreated     EventType = "group_event_created"
	GroupEventUpdated     EventType = "group_event_updated"
	GroupEventDeleted     EventType = "group_event_deleted"
	EventResponseUpdated  EventType = "event_response_updated"

	// Chat events
	PrivateMessage EventType = "private_message"
	MessagesRead   EventType = "messages_read"
	UserTyping     EventType = "user_typing"

	// group events
	GroupMessage EventType = "group_message"

	// header notifications
	HeaderNotificationUpdate EventType = "notification_Update"
)

// Event represents a WebSocket event
type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

// PostCreatedPayload represents the payload for a post_created event
type PostCreatedPayload struct {
	Post     interface{} `json:"post"`
	UserID   string      `json:"userId"`
	UserName string      `json:"userName"`
}

// PostLikedPayload represents the data sent when a post is liked/unliked
type PostLikedPayload struct {
	PostID     int64  `json:"postId"`
	UserID     string `json:"userId"`
	UserName   string `json:"userName"`
	IsLiked    bool   `json:"isLiked"`
	LikesCount int    `json:"likesCount"`
}

type UserStatsUpdatedPayload struct {
	UserID    string `json:"userId"`
	StatsType string `json:"statsType"`
	Count     int    `json:"count"`
}

// UserStatusUpdatePayload represents the payload for a user_status_update event
type UserStatusUpdatePayload struct {
	UserID    string `json:"userId"`
	IsOnline  bool   `json:"isOnline"`
	Timestamp int64  `json:"timestamp"`
}

// PrivateMessagePayload represents the payload for a private_message event
type PrivateMessagePayload struct {
	MessageID    int64  `json:"messageId"`
	SenderID     string `json:"senderId"`
	ReceiverID   string `json:"receiverId"`
	Content      string `json:"content"`
	CreatedAt    string `json:"createdAt"`
	IsRead       bool   `json:"isRead"`
	SenderName   string `json:"senderName"`
	SenderAvatar string `json:"senderAvatar"`
}

// MessagesReadPayload represents the payload for a messages_read event
type MessagesReadPayload struct {
	SenderID   string `json:"senderId"`
	ReceiverID string `json:"receiverId"`
	ReadAt     string `json:"readAt"`
}

// UserTypingPayload represents the payload for a user_typing event
type UserTypingPayload struct {
	SenderID   string `json:"senderId"`
	ReceiverID string `json:"receiverId"`
	Timestamp  string `json:"timestamp"`
}
