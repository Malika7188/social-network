package event

import (
	"database/sql"
	"fmt"
	"time"

	notifications "github.com/Athooh/social-network/internal/notifcations"
	"github.com/Athooh/social-network/pkg/logger"
	models "github.com/Athooh/social-network/pkg/models/dbTables"
	"github.com/Athooh/social-network/pkg/websocket"
	"github.com/Athooh/social-network/pkg/websocket/events"
)

// NotificationService handles sending notifications related to event operations
type NotificationService struct {
	hub              *websocket.Hub
	repo             Repository
	notificationRepo notifications.Service
	log              *logger.Logger
}

// NewNotificationService creates a new event notification service
func NewNotificationService(hub *websocket.Hub, repo Repository, notificationRepo notifications.Service, log *logger.Logger) *NotificationService {
	return &NotificationService{
		hub:              hub,
		repo:             repo,
		notificationRepo: notificationRepo,
		log:              log,
	}
}

// SendEventCreatedNotification sends a notification when a new event is created
func (s *NotificationService) SendEventCreatedNotification(inviterID, inviteeID string, newNote *notifications.NewNotification, inviterInfo *models.UserBasic) {
	if s.hub == nil {
		s.log.Warn("WebSocket hub is nil, cannot send follow request notification")
		return
	}

	// Create notification in database
	if err := s.notificationRepo.CreateNotification(newNote); err != nil {
		s.log.Error("Failed to create follow request notification: %v", err)
		return
	}

	// Retrieve the newly created notification to get its ID and CreatedAt
	notifications, err := s.notificationRepo.GetNotifications(inviteeID, 1, 0)
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
			"type":         newNote.NotficationType,
			"senderId":     inviterID,
			"senderName":   inviterInfo.FirstName + " " + inviterInfo.LastName,
			"senderAvatar": inviterInfo.Avatar,
			"message":      newNote.Message,
			"createdAt":    dbNotification.CreatedAt.Format(time.RFC3339),
			"isRead":       dbNotification.IsRead,
			"eventId":      newNote.TargetEventID.String,
		},
	}

	// Send to the user receiving the follow request
	s.hub.BroadcastToUser(inviteeID, event)
}

// SendEventUpdatedNotification sends a notification when an event is updated
func (s *NotificationService) SendEventUpdatedNotification(event *models.GroupEvent) {
	if s.hub == nil {
		s.log.Warn("WebSocket hub is nil, cannot send event updated notification")
		return
	}

	// Get creator info
	creator, err := s.repo.GetUserBasicByID(event.CreatorID)
	if err != nil {
		s.log.Warn("Failed to get creator info for notification: %v", err)
		return
	}

	creatorName := fmt.Sprintf("%s %s", creator.FirstName, creator.LastName)

	// Get group members to notify
	members, err := s.repo.GetGroupMembers(event.GroupID, "accepted")
	if err != nil {
		s.log.Warn("Failed to get group members for notification: %v", err)
		return
	}

	// Send to all group members
	for _, member := range members {
		// Create notification in database
		notification := &notifications.NewNotification{
			UserId:          member.UserID,
			NotficationType: "groupEventUpdated",
			SenderId:        sql.NullString{String: event.CreatorID, Valid: true},
			Message:         fmt.Sprintf("updated the event %s", event.Title),
		}

		if err := s.notificationRepo.CreateNotification(notification); err != nil {
			s.log.Error("Failed to create event updated notification: %v", err)
			continue
		}

		// Retrieve the newly created notification
		notifications, err := s.notificationRepo.GetNotifications(member.UserID, 1, 0)
		if err != nil || len(notifications) == 0 {
			s.log.Error("Failed to retrieve newly created notification: %v", err)
			continue
		}
		dbNotification := notifications[0]

		// Create WebSocket event
		notificationEvent := events.Event{
			Type: "group_event_updated",
			Payload: map[string]interface{}{
				"id":           dbNotification.ID,
				"eventId":      event.ID,
				"groupId":      event.GroupID,
				"type":         "groupEventUpdated",
				"senderId":     event.CreatorID,
				"senderAvatar": creator.Avatar,
				"senderName":   creatorName,
				"message":      fmt.Sprintf("updated the event %s", event.Title),
				"eventDate":    event.EventDate.Format(time.RFC3339),
				"isRead":       false,
				"createdAt":    dbNotification.CreatedAt.Format(time.RFC3339),
			},
		}

		s.hub.BroadcastToUser(member.UserID, notificationEvent)
	}
}

// SendEventResponseNotification sends a notification when a user responds to an event
func (s *NotificationService) SendEventResponseNotification(eventID, userID, response string) {
	if s.hub == nil {
		s.log.Warn("WebSocket hub is nil, cannot send event response notification")
		return
	}

	// Get event info
	event, err := s.repo.GetEventByID(eventID)
	if err != nil {
		s.log.Warn("Failed to get event info for notification: %v", err)
		return
	}

	// Get user info
	user, err := s.repo.GetUserBasicByID(userID)
	if err != nil {
		s.log.Warn("Failed to get user info for notification: %v", err)
		return
	}

	userName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)

	// Create notification in database
	notification := &notifications.NewNotification{
		UserId:          event.CreatorID,
		NotficationType: "eventResponse",
		SenderId:        sql.NullString{String: event.CreatorID, Valid: true},
		Message:         fmt.Sprintf("%s responded %s to the event %s", userName, response, event.Title),
	}

	if err := s.notificationRepo.CreateNotification(notification); err != nil {
		s.log.Error("Failed to create event response notification: %v", err)
		return
	}

	// Retrieve the newly created notification
	notifications, err := s.notificationRepo.GetNotifications(event.CreatorID, 1, 0)
	if err != nil || len(notifications) == 0 {
		s.log.Error("Failed to retrieve newly created notification: %v", err)
		return
	}
	dbNotification := notifications[0]

	// Create WebSocket event
	notificationEvent := events.Event{
		Type: "event_response_updated",
		Payload: map[string]interface{}{
			"id":           dbNotification.ID,
			"eventId":      eventID,
			"groupId":      event.GroupID,
			"type":         "eventResponse",
			"senderId":     userID,
			"senderAvatar": user.Avatar,
			"senderName":   userName,
			"message":      fmt.Sprintf("%s responded %s to the event %s", userName, response, event.Title),
			"response":     response,
			"isRead":       false,
			"createdAt":    dbNotification.CreatedAt.Format(time.RFC3339),
		},
	}

	// Send to event creator
	s.hub.BroadcastToUser(event.CreatorID, notificationEvent)
}
