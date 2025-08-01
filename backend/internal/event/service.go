package event

import (
	"database/sql"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	notifications "github.com/Athooh/social-network/internal/notifcations"
	"github.com/Athooh/social-network/pkg/filestore"
	"github.com/Athooh/social-network/pkg/logger"
	models "github.com/Athooh/social-network/pkg/models/dbTables"
	"github.com/Athooh/social-network/pkg/websocket"
	"github.com/google/uuid"
)

// Service defines the interface for event business logic
type Service interface {
	// Event operations
	CreateEvent(groupID, userID, title, description string, eventDate time.Time, banner *multipart.FileHeader, response string) (*models.GroupEvent, error)
	GetEvent(eventID, userID string) (*models.GroupEvent, error)
	GetGroupEvents(groupID, userID string) ([]*models.GroupEvent, error)
	UpdateEvent(eventID, userID, title, description string, eventDate time.Time, banner *multipart.FileHeader) (*models.GroupEvent, error)
	DeleteEvent(eventID, userID string) error

	// Event responses operations
	RespondToEvent(eventID, userID, response string) error
	GetEventResponses(eventID, userID string, responseType string) ([]*models.EventResponse, error)
}

// EventService implements the Service interface
type EventService struct {
	repo                Repository
	fileStore           *filestore.FileStore
	log                 *logger.Logger
	wsHub               *websocket.Hub
	notificationService *NotificationService
}

// NewService creates a new event service
func NewService(repo Repository, fileStore *filestore.FileStore, log *logger.Logger, NotificationRepo notifications.Service, wsHub *websocket.Hub) *EventService {
	notificationSvc := NewNotificationService(wsHub, repo, NotificationRepo, log)

	return &EventService{
		repo:                repo,
		fileStore:           fileStore,
		log:                 log,
		wsHub:               wsHub,
		notificationService: notificationSvc,
	}
}

// CreateEvent creates a new event in a group
func (s *EventService) CreateEvent(groupID, userID, title, description string, eventDate time.Time, banner *multipart.FileHeader, response string) (*models.GroupEvent, error) {
	// Check if user is a member of the group
	isMember, err := s.repo.IsGroupMember(groupID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, errors.New("you must be a member of the group to create an event")
	}

	// Create event
	event := &models.GroupEvent{
		ID:          uuid.New().String(),
		GroupID:     groupID,
		CreatorID:   userID,
		Title:       title,
		Description: description,
		EventDate:   eventDate,
	}

	// Handle banner upload if provided
	if banner != nil {
		bannerPath, err := s.fileStore.SaveFile(banner, "event_banners")
		if err != nil {
			return nil, fmt.Errorf("failed to save banner: %w", err)
		}
		event.BannerPath.String = bannerPath
		event.BannerPath.Valid = true
	}

	// Save event
	if err := s.repo.CreateEvent(event); err != nil {
		return nil, err
	}

	// Get full event with creator info
	fullEvent, err := s.repo.GetEventByID(event.ID)
	if err != nil {
		return nil, err
	}

	//get group members for broadcasting
	groupMembers, err := s.repo.GetGroupMembers(groupID, "accepted")
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	eventcreator, err := s.repo.GetUserBasicByID(userID)
	if err != nil {
		return nil, err
	}

	
	// Broadcast event creation to group members
	for _, member := range groupMembers {
		if member.UserID != userID {
			newNotification := &notifications.NewNotification{
				UserId:          member.UserID,
				SenderId:        sql.NullString{String: userID, Valid: true},
				NotficationType: "groupEvent",
				Message:         fmt.Sprintf(event.Title),
				TargetGroupID:   sql.NullString{String: groupID, Valid: true},
				TargetEventID:   sql.NullString{String: event.ID, Valid: true},
			}
			// Notify group members about new event
			s.notificationService.SendEventCreatedNotification(userID, member.UserID, newNotification, eventcreator)
		}
	}
	
	err = s.RespondToEvent(event.ID, userID, response)
	if err != nil {
		return nil, fmt.Errorf("failed to respond to event: %w", err)
	}

	return fullEvent, nil
}

// GetEvent gets an event by ID
func (s *EventService) GetEvent(eventID, userID string) (*models.GroupEvent, error) {
	// Get event
	event, err := s.repo.GetEventByID(eventID)
	if err != nil {
		return nil, err
	}

	// Check if user is a member of the group
	isMember, err := s.repo.IsGroupMember(event.GroupID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, errors.New("you must be a member of the group to view this event")
	}

	// Get user's response to this event
	userResponse, err := s.repo.GetUserEventResponse(eventID, userID)
	if err != nil {
		return nil, err
	}

	if userResponse != nil {
		event.UserResponse = userResponse.Response
	}

	return event, nil
}

// GetGroupEvents gets all events in a group
func (s *EventService) GetGroupEvents(groupID, userID string) ([]*models.GroupEvent, error) {
	// Check if user is a member of the group
	isMember, err := s.repo.IsGroupMember(groupID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, errors.New("you must be a member of the group to view events")
	}

	// Get events
	events, err := s.repo.GetGroupEvents(groupID)
	if err != nil {
		return nil, err
	}

	// Get user's response to each event
	for _, event := range events {
		userResponse, err := s.repo.GetUserEventResponse(event.ID, userID)
		if err != nil {
			return nil, err
		}

		if userResponse != nil {
			event.UserResponse = userResponse.Response
		}
	}

	return events, nil
}

// UpdateEvent updates an event
func (s *EventService) UpdateEvent(eventID, userID, title, description string, eventDate time.Time, banner *multipart.FileHeader) (*models.GroupEvent, error) {
	// Get event
	event, err := s.repo.GetEventByID(eventID)
	if err != nil {
		return nil, err
	}

	// Check if user is the creator or an admin
	if event.CreatorID != userID {
		// Check if user is an admin
		role, err := s.repo.GetMemberRole(event.GroupID, userID)
		if err != nil {
			return nil, err
		}

		if role != "admin" {
			return nil, errors.New("only the event creator or group admins can update this event")
		}
	}

	// Update fields
	event.Title = title
	event.Description = description
	event.EventDate = eventDate

	// Handle banner upload if provided
	if banner != nil {
		bannerPath, err := s.fileStore.SaveFile(banner, "event_banners")
		if err != nil {
			return nil, fmt.Errorf("failed to save banner: %w", err)
		}
		event.BannerPath.String = bannerPath
		event.BannerPath.Valid = true
	}

	// Update event
	if err := s.repo.UpdateEvent(event); err != nil {
		return nil, err
	}

	// Get updated event
	updatedEvent, err := s.repo.GetEventByID(eventID)
	if err != nil {
		return nil, err
	}

	return updatedEvent, nil
}

// DeleteEvent deletes an event
func (s *EventService) DeleteEvent(eventID, userID string) error {
	// Get event
	event, err := s.repo.GetEventByID(eventID)
	if err != nil {
		return err
	}

	// Check if user is the creator or an admin
	if event.CreatorID != userID {
		// Check if user is an admin
		role, err := s.repo.GetMemberRole(event.GroupID, userID)
		if err != nil {
			return err
		}

		if role != "admin" {
			return errors.New("only the event creator or group admins can delete this event")
		}
	}

	// Delete event
	if err := s.repo.DeleteEvent(eventID); err != nil {
		return err
	}

	return nil
}

// RespondToEvent handles a user's response to an event
func (s *EventService) RespondToEvent(eventID, userID, responseType string) error {
	// Check if response type is valid
	if responseType != "going" && responseType != "not_going" {
		return errors.New("invalid response type")
	}

	// Get event
	event, err := s.repo.GetEventByID(eventID)
	if err != nil {
		return err
	}

	// Check if user is a member of the group
	isMember, err := s.repo.IsGroupMember(event.GroupID, userID)
	if err != nil {
		return err
	}

	if !isMember {
		return errors.New("you must be a member of the group to respond to this event")
	}

	// Create response
	response := &models.EventResponse{
		EventID:  eventID,
		UserID:   userID,
		Response: responseType,
	}

	// Save response
	if err := s.repo.AddEventResponse(response); err != nil {
		return err
	}

	return nil
}

// GetEventResponses gets all responses to an event
func (s *EventService) GetEventResponses(eventID, userID, responseType string) ([]*models.EventResponse, error) {
	// Get event
	event, err := s.repo.GetEventByID(eventID)
	if err != nil {
		return nil, err
	}

	// Check if user is a member of the group
	isMember, err := s.repo.IsGroupMember(event.GroupID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, errors.New("you must be a member of the group to view event responses")
	}

	// Get responses
	responses, err := s.repo.GetEventResponses(eventID, responseType)
	if err != nil {
		return nil, err
	}

	return responses, nil
}
