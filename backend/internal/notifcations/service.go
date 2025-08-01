package notifications

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Athooh/social-network/pkg/logger"
	models "github.com/Athooh/social-network/pkg/models/dbTables"
	"github.com/Athooh/social-network/pkg/user"
	"github.com/Athooh/social-network/pkg/websocket"
)

// Service defines the notification service interface
type Service interface {
	CreateNotification(notification *NewNotification) error
	GetNotifications(userID string, limit, offset int) ([]*NotificationWithUser, error)
	MarkNotificationAsRead(notificationID int64) error
	MarkAllNotificationsAsRead(userID string) error
	ClearAllNotifications(userID string) error
	DeleteNotification(notificationID int64) error
}

// NotificationWithUser extends Notification with user information
type NotificationWithUser struct {
	models.Notification
	SenderName   string
	SenderAvatar string
}

// NotificationService implements the Service interface
type NotificationService struct {
	repo     Repository
	userRepo user.Repository
	log      *logger.Logger
	wsHub    *websocket.Hub
}

type NewNotification struct {
	UserId          string
	SenderId        sql.NullString
	NotficationType string
	Message         string
	TargetGroupID   sql.NullString
	TargetEventID   sql.NullString
}

// NewService creates a new notification service
func NewService(repo Repository, userRepo user.Repository, log *logger.Logger, wsHub *websocket.Hub) Service {
	return &NotificationService{
		repo:     repo,
		userRepo: userRepo,
		log:      log,
		wsHub:    wsHub,
	}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(notification *NewNotification) error {
	if notification.UserId == "" {
		return errors.New("user ID cannot be empty")
	}

	newNotification := &models.Notification{
		UserID:  notification.UserId,
		Type:    notification.NotficationType,
		Message: notification.Message,
		IsRead:  false,
	}

	// Handle nullable SenderID
	if notification.SenderId.Valid {
		newNotification.SenderID = notification.SenderId
	}

	// Handle nullable TargetGroupID
	if notification.TargetGroupID.Valid {
		newNotification.TargetGroupID = notification.TargetGroupID
	}

	// Handle nullable TargetEventID
	if notification.TargetEventID.Valid {
		newNotification.TargetEventID = notification.TargetEventID
	}

	if err := s.repo.CreateNotification(newNotification); err != nil {
		s.log.Error("Failed to create notification: %v", err)
		return err
	}

	return nil
}

// GetNotifications retrieves notifications for a user with sender information
func (s *NotificationService) GetNotifications(userID string, limit, offset int) ([]*NotificationWithUser, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	notifications, err := s.repo.GetNotifications(userID, limit, offset)
	if err != nil {
		s.log.Error("Failed to get notifications: %v", err)
		return nil, err
	}

	var notificationsWithUser []*NotificationWithUser
	for _, notification := range notifications {
		var senderName, senderAvatar string

		// Get sender information if senderID exists
		if notification.SenderID.Valid && *&notification.SenderID.String != "" {
			sender, err := s.userRepo.GetByID(*&notification.SenderID.String)
			if err != nil {
				s.log.Warn("Failed to get sender info for notification %d: %v", notification.ID, err)
				continue
			}
			senderName = fmt.Sprintf("%s %s", sender.FirstName, sender.LastName)
			senderAvatar = sender.Avatar
		}

		notificationWithUser := &NotificationWithUser{
			Notification: *notification,
			SenderName:   senderName,
			SenderAvatar: senderAvatar,
		}

		notificationsWithUser = append(notificationsWithUser, notificationWithUser)
	}

	return notificationsWithUser, nil
}

// MarkNotificationAsRead marks a single notification as read
func (s *NotificationService) MarkNotificationAsRead(notificationID int64) error {
	if notificationID <= 0 {
		return errors.New("invalid notification ID")
	}

	if err := s.repo.MarkNotificationAsRead(notificationID); err != nil {
		s.log.Error("Failed to mark notification as read: %v", err)
		return err
	}

	return nil
}

// MarkAllNotificationsAsRead marks all notifications for a user as read
func (s *NotificationService) MarkAllNotificationsAsRead(userID string) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}

	if err := s.repo.MarkAllNotificationsAsRead(userID); err != nil {
		s.log.Error("Failed to mark all notifications as read: %v", err)
		return err
	}

	return nil
}

// ClearAllNotifications removes all notifications for a user
func (s *NotificationService) ClearAllNotifications(userID string) error {
	if userID == "" {
		return errors.New("user ID cannot be empty")
	}

	if err := s.repo.ClearAllNotificationsDB(userID); err != nil {
		s.log.Error("Failed to clear all notifications: %v", err)
		return err
	}

	return nil
}

func (s *NotificationService) DeleteNotification(notificationID int64) error {
	if notificationID <= 0 {
		return errors.New("invalid notification ID")
	}

	if err := s.repo.DeleteNotificationDb(notificationID); err != nil {
		s.log.Error("Failed to delete notification: %v", err)
		return err
	}
	return nil
}
