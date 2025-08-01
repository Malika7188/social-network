package user

import (
	"github.com/Athooh/social-network/pkg/logger"
	"github.com/Athooh/social-network/pkg/session"
	"github.com/Athooh/social-network/pkg/user"
	"github.com/Athooh/social-network/pkg/websocket"
	"github.com/Athooh/social-network/pkg/websocket/events"
)

// StatusService handles user online/offline status
type StatusService struct {
	statusRepo  user.StatusRepository
	sessionRepo session.SessionStore
	hub         *websocket.Hub
	log         *logger.Logger
}

// NewStatusService creates a new user status service
func NewStatusService(statusRepo user.StatusRepository, sessionRepo session.SessionStore, hub *websocket.Hub, log *logger.Logger) *StatusService {
	return &StatusService{
		statusRepo:  statusRepo,
		sessionRepo: sessionRepo,
		hub:         hub,
		log:         log,
	}
}

// SetUserOnline marks a user as online and notifies followers
func (s *StatusService) SetUserOnline(userID string) error {
	// Update database
	if err := s.statusRepo.SetUserOnline(userID); err != nil {
		s.log.Error("Failed to set user online status: %v", err)
		return err
	}

	// Get users to notify (both followers and following)
	userIDs, err := s.statusRepo.GetFollowersForStatusUpdate(userID)
	if err != nil {
		s.log.Error("Failed to get users for status update: %v", err)
		return err
	}

	// Create status update event
	event := events.Event{
		Type: "user_status_update",
		Payload: map[string]interface{}{
			"userId":   userID,
			"isOnline": true,
		},
	}

	// Notify users (both followers and following)
	for _, userID := range userIDs {
		s.hub.BroadcastToUser(userID, event)
	}

	return nil
}

// SetUserOffline marks a user as offline and notifies followers
func (s *StatusService) SetUserOffline(userID string) error {
	// Update database
	if err := s.statusRepo.SetUserOffline(userID); err != nil {
		s.log.Error("Failed to set user offline status: %v", err)
		return err
	}

	// Get users to notify (both followers and following)
	userIDs, err := s.statusRepo.GetFollowersForStatusUpdate(userID)
	if err != nil {
		s.log.Error("Failed to get users for status update: %v", err)
		return err
	}

	// Create status update event
	event := events.Event{
		Type: "user_status_update",
		Payload: map[string]interface{}{
			"userId":   userID,
			"isOnline": false,
		},
	}

	// Notify users (both followers and following)
	for _, userID := range userIDs {
		s.hub.BroadcastToUser(userID, event)
	}

	s.log.Info("User %s is offline", userID)

	return nil
}

// GetUserStatus gets a user's online status
func (s *StatusService) GetUserStatus(userID string) (bool, error) {
	return s.statusRepo.GetUserStatus(userID)
}

// CleanupUserStatuses checks all online users and marks them offline if they don't have a valid session
func (s *StatusService) CleanupUserStatuses() {
	s.log.Info("Starting user status cleanup...")

	// Get all users currently marked as online
	onlineUsers, err := s.statusRepo.GetAllOnlineUsers()
	if err != nil {
		s.log.Error("Failed to get online users during cleanup: %v", err)
		return
	}

	s.log.Info("Found %d users marked as online, checking session validity", len(onlineUsers))

	// For each online user, check if they have a valid session
	for _, userID := range onlineUsers {
		// Check if user has a valid session
		hasValidSession, err := s.sessionRepo.HasValidSession(userID)
		if err != nil {
			s.log.Error("Error checking session for user %s: %v", userID, err)
			continue
		}

		// If no valid session, mark user as offline
		if !hasValidSession {
			s.log.Info("User %s has no valid session but is marked online, setting to offline", userID)
			if err := s.SetUserOffline(userID); err != nil {
				s.log.Error("Failed to set user %s offline during cleanup: %v", userID, err)
			}
		}
	}

	s.log.Info("User status cleanup completed")
}
