package post

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	notifications "github.com/Athooh/social-network/internal/notifcations"
	"github.com/Athooh/social-network/pkg/logger"
	models "github.com/Athooh/social-network/pkg/models/dbTables"
	"github.com/Athooh/social-network/pkg/user"
	"github.com/Athooh/social-network/pkg/websocket"
	"github.com/Athooh/social-network/pkg/websocket/events"
)

type NotificationService struct {
	hub              *websocket.Hub
	userRepo         user.Repository
	log              *logger.Logger
	notificationSRVC notifications.Service
}

func NewNotificationService(hub *websocket.Hub, userRepo user.Repository, notifcationSRVC notifications.Service, log *logger.Logger) *NotificationService {
	return &NotificationService{
		hub:              hub,
		userRepo:         userRepo,
		log:              log,
		notificationSRVC: notifcationSRVC,
	}
}

func (s *NotificationService) NotifyPostCreated(post interface{}, userID, userName string) error {
	event := events.Event{
		Type: events.PostCreated,
		Payload: events.PostCreatedPayload{
			Post:     post,
			UserID:   userID,
			UserName: userName,
		},
	}

	// Convert event to JSON and send it to the broadcast channel
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	s.hub.Broadcast <- eventJSON
	return nil
}

// NotifyPostCreatedToSpecificUsers sends notifications about a new post to specific users
func (s *NotificationService) NotifyPostCreatedToSpecificUsers(post *models.Post, userID string, userName string, recipientIDs []string) error {
	// Create event payload
	payload := events.PostCreatedPayload{
		Post:     post,
		UserID:   userID,
		UserName: userName,
	}

	// Create event
	event := events.Event{
		Type:    events.PostCreated,
		Payload: payload,
	}

	// Send to each specific recipient
	for _, recipientID := range recipientIDs {
		// Don't notify the post creator
		if recipientID == userID {
			continue
		}

		s.hub.BroadcastToUser(recipientID, event)
	}

	return nil
}

// NotifyPostLiked sends a notification when a post is liked or unliked
func (s *NotificationService) NotifyPostLiked(post *models.Post, userID string, userName string, isLiked bool) error {
	// Create event payload
	payload := events.PostLikedPayload{
		PostID:     post.ID,
		UserID:     userID,
		UserName:   userName,
		IsLiked:    isLiked,
		LikesCount: int(post.LikesCount),
	}

	// Create event
	event := events.Event{
		Type:    events.PostLiked,
		Payload: payload,
	}

	// Notify the post owner if they're not the one who liked/unliked
	// if post.UserID != userID {
	// 	s.hub.BroadcastToUser(post.UserID, event)
	// }

	// Also broadcast to anyone viewing the post
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	s.hub.Broadcast <- eventJSON
	return nil
}

// NotifyUserStatsUpdated sends a notification when a user's stats are updated
func (s *NotificationService) NotifyUserStatsUpdated(userID string, statsType string, count int) error {
	// Create event payload
	payload := events.UserStatsUpdatedPayload{
		UserID:    userID,
		StatsType: statsType, // e.g., "followers", "following", "posts"
		Count:     count,
	}

	// Create event
	event := events.Event{
		Type:    events.UserStatsUpdated,
		Payload: payload,
	}

	// Broadcast to the specific user
	s.hub.BroadcastToUser(userID, event)

	return nil
}

func (s *NotificationService) NotifyPostsCommentUpdateToSpecifUsers(userID string, statsType string, count int, recipientIDs []string) error {
	// Create event payload
	payload := events.UserStatsUpdatedPayload{
		UserID:    userID,
		StatsType: statsType,
		Count:     count,
	}

	// Create event
	event := events.Event{
		Type:    events.CommentCountUpdate,
		Payload: payload,
	}

	// Send to each specific recipient including the current
	for _, recipientID := range recipientIDs {
		s.hub.BroadcastToUser(recipientID, event)
	}

	return nil
}

func (s *NotificationService) NotifyPostsCommentUpdate(userID string, statsType string, count int) error {
	// Create event payload
	payload := events.UserStatsUpdatedPayload{
		UserID:    userID,
		StatsType: statsType,
		Count:     count,
	}

	// Create event
	event := events.Event{
		Type:    events.CommentCountUpdate,
		Payload: payload,
	}

	// Broadcast to the specific user
	s.hub.BroadcastToUser(userID, event)

	return nil
}

// SendCommentNotification sends a WebSocket notification when a user comments on a post
func (s *NotificationService) SendCommentNotificationToOwner(userID, commenterID string) {
	if s.hub == nil {
		s.log.Warn("WebSocket hub is nil, cannot send comment notification")
		return
	}

	if userID == commenterID {
		return
	}

	// Fetch commenter details
	commenter, err := s.userRepo.GetByID(commenterID)
	if err != nil {
		s.log.Error("Failed to fetch commenter details: %v", err)
		return
	}
	commenterName := commenter.FirstName + " " + commenter.LastName

	// Create notification in database
	notification := &notifications.NewNotification{
		UserId:          userID,
		NotficationType: "comment",
		SenderId:        sql.NullString{String: commenterID, Valid: true},
		Message:         fmt.Sprintf("%s commented on your post.", commenterName),
	}
	if err := s.notificationSRVC.CreateNotification(notification); err != nil {
		s.log.Error("Failed to create comment notification: %v", err)
		return
	}

	// Retrieve the newly created notification to get its ID and CreatedAt
	notifications, err := s.notificationSRVC.GetNotifications(userID, 1, 0)
	if err != nil || len(notifications) == 0 {
		s.log.Error("Failed to retrieve newly created notification: %v", err)
		return
	}
	dbNotification := notifications[0]

	// Create WebSocket event
	event := events.Event{
		Type: events.HeaderNotificationUpdate,
		Payload: map[string]interface{}{
			"id":           dbNotification.ID,
			"type":         "comment",
			"senderId":     commenterID,
			"senderName":   commenterName,
			"senderAvatar": commenter.Avatar,
			"message":      notification.Message,
			"createdAt":    dbNotification.CreatedAt.Format(time.RFC3339),
			"isRead":       dbNotification.IsRead,
		},
	}

	s.hub.BroadcastToUser(userID, event)
}
