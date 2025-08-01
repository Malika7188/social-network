package follow

import (
	"fmt"
	"time"

	notifications "github.com/Athooh/social-network/internal/notifcations"
	"github.com/Athooh/social-network/pkg/logger"
	"github.com/Athooh/social-network/pkg/user"
	"github.com/Athooh/social-network/pkg/websocket"
	"github.com/Athooh/social-network/pkg/websocket/events"
)

// NotificationService handles sending notifications related to follow operations
type NotificationService struct {
	hub              *websocket.Hub
	userRepo         user.Repository
	log              *logger.Logger
	notificationRepo notifications.Service
}

// NewNotificationService creates a new follow notification service
func NewNotificationService(hub *websocket.Hub, userRepo user.Repository, notifcationRepo notifications.Service, log *logger.Logger) *NotificationService {
	return &NotificationService{
		hub:              hub,
		userRepo:         userRepo,
		log:              log,
		notificationRepo: notifcationRepo,
	}
}

// SendFollowRequestNotification sends a WebSocket notification for a follow request
func (s *NotificationService) SendFollowRequestNotification(followerID, followingID string, notification *notifications.NewNotification, follower *user.User) {
	if s.hub == nil {
		s.log.Warn("WebSocket hub is nil, cannot send follow request notification")
		return
	}

	// Create notification in database
	if err := s.notificationRepo.CreateNotification(notification); err != nil {
		s.log.Error("Failed to create follow request notification: %v", err)
		return
	}

	// Retrieve the newly created notification to get its ID and CreatedAt
	notifications, err := s.notificationRepo.GetNotifications(followingID, 1, 0)
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
			"type":         notification.NotficationType,
			"senderId":     followerID,
			"senderName":   follower.FirstName + " " + follower.LastName,
			"senderAvatar": follower.Avatar,
			"message":      notification.Message,
			"createdAt":    dbNotification.CreatedAt.Format(time.RFC3339),
			"isRead":       dbNotification.IsRead,
		},
	}

	// Send to the user receiving the follow request
	s.hub.BroadcastToUser(followingID, event)
}

// SendFollowRequestAcceptedNotification sends a notification when a follow request is accepted
func (s *NotificationService) SendFollowRequestAcceptedNotification(followerID, followingID string) {
	if s.hub == nil {
		return
	}

	// Get following user info
	following, err := s.userRepo.GetByID(followingID)
	if err != nil {
		s.log.Warn("Failed to get following user info for notification: %v", err)
		return
	}

	followingName := fmt.Sprintf("%s %s", following.FirstName, following.LastName)

	// Create notification event
	event := events.Event{
		Type: events.FollowRequestAccepted,
		Payload: map[string]interface{}{
			"followingID":   followingID,
			"followingName": followingName,
			"avatar":        following.Avatar,
			"timestamp":     fmt.Sprintf("%d", time.Now().Unix()),
		},
	}

	// Send to the user whose follow request was accepted
	s.hub.BroadcastToUser(followerID, event)
}

// SendFollowNotification sends a notification when a user follows/unfollows another user
func (s *NotificationService) SendFollowNotification(followerID, followingID string, isFollow bool) {
	if s.hub == nil {
		return
	}

	// Get follower info
	follower, err := s.userRepo.GetByID(followerID)
	if err != nil {
		s.log.Warn("Failed to get follower info for notification: %v", err)
		return
	}

	// Create notification event
	action := "followed"
	if !isFollow {
		action = "unfollowed"
	}

	followerName := fmt.Sprintf("%s %s", follower.FirstName, follower.LastName)

	event := events.Event{
		Type: events.FollowUpdate,
		Payload: map[string]interface{}{
			"followerID":   followerID,
			"followerName": followerName,
			"avatar":       follower.Avatar,
			"action":       action,
			"timestamp":    fmt.Sprintf("%d", time.Now().Unix()),
		},
	}

	// Send to the user being followed/unfollowed
	s.hub.BroadcastToUser(followingID, event)
}

// UpdateFollowerCounts updates the follower and following counts for both users
func (s *NotificationService) UpdateFollowerCounts(followerID, followingID string, repo Repository) {
	// Update follower count for the user being followed
	followersCount, err := repo.GetFollowersCount(followingID)
	if err == nil {
		// Send notification about updated stats
		if s.hub != nil {
			s.hub.BroadcastToUser(followingID, events.Event{
				Type: events.UserStatsUpdated,
				Payload: map[string]interface{}{
					"userId":    followingID,
					"statsType": "followers_count",
					"count":     followersCount,
				},
			})
		}
	}

	// Update following count for the follower
	followingCount, err := repo.GetFollowingCount(followerID)
	if err == nil {
		// Send notification about updated stats
		if s.hub != nil {
			s.hub.BroadcastToUser(followerID, events.Event{
				Type: events.UserStatsUpdated,
				Payload: map[string]interface{}{
					"userId":    followerID,
					"statsType": "following_count",
					"count":     followingCount,
				},
			})
		}
	}
}
