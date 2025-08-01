package chat

import (
	"errors"
	"time"

	"github.com/Athooh/social-network/pkg/logger"
	models "github.com/Athooh/social-network/pkg/models/dbTables"
	"github.com/Athooh/social-network/pkg/websocket"
)

// Service defines the chat service interface
type Service interface {
	// Message operations
	SendMessage(senderID, receiverID, content string) (*models.PrivateMessage, error)
	GetMessages(userID1, userID2 string, limit, offset int) ([]*models.PrivateMessage, error)
	MarkAsRead(senderID, receiverID string) error

	// Contact operations
	GetContacts(userID string) ([]*models.ChatContact, error)

	// Typing indicator
	SendTypingIndicator(senderID, receiverID string) error

	// Message search
	SearchMessages(userID, query, otherUserID string, limit int) ([]*models.PrivateMessage, error)
}

// ChatService implements the Service interface
type ChatService struct {
	repo            Repository
	log             *logger.Logger
	notificationSvc *NotificationService
}

// NewService creates a new chat service
func NewService(repo Repository, log *logger.Logger, wsHub *websocket.Hub) Service {
	notificationSvc := NewNotificationService(wsHub)

	return &ChatService{
		repo:            repo,
		log:             log,
		notificationSvc: notificationSvc,
	}
}

// SendMessage sends a private message from one user to another
func (s *ChatService) SendMessage(senderID, receiverID, content string) (*models.PrivateMessage, error) {
	// Check if users can message each other
	canSend, err := s.repo.CanSendMessage(senderID, receiverID)
	if err != nil {
		return nil, err
	}

	if !canSend {
		return nil, errors.New("you cannot send messages to this user")
	}

	// Create the message
	message := &models.PrivateMessage{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		CreatedAt:  time.Now(),
		IsRead:     false,
	}

	// Save to database
	if err := s.repo.SaveMessage(message); err != nil {
		return nil, err
	}

	// Get sender and receiver info for the response
	sender, err := s.repo.GetUserBasicByID(senderID)
	if err == nil {
		message.Sender = sender
	}

	receiver, err := s.repo.GetUserBasicByID(receiverID)
	if err == nil {
		message.Receiver = receiver
	}

	// Send notification via WebSocket
	go s.notificationSvc.NotifyNewMessage(message)

	return message, nil
}

// GetMessages gets messages between two users with pagination
func (s *ChatService) GetMessages(userID1, userID2 string, limit, offset int) ([]*models.PrivateMessage, error) {
	// Check if users can view messages
	canSend, err := s.repo.CanSendMessage(userID1, userID2)
	if err != nil {
		return nil, err
	}

	if !canSend {
		return nil, errors.New("you cannot view messages with this user")
	}

	// Get messages
	messages, err := s.repo.GetMessagesBetweenUsers(userID1, userID2, limit, offset)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// MarkAsRead marks messages from a sender to a receiver as read
func (s *ChatService) MarkAsRead(senderID, receiverID string) error {
	// Check if users can view messages
	canSend, err := s.repo.CanSendMessage(receiverID, senderID)
	if err != nil {
		return err
	}

	if !canSend {
		return errors.New("you cannot mark messages as read from this user")
	}

	// Mark messages as read in the database
	if err := s.repo.MarkMessagesAsRead(senderID, receiverID); err != nil {
		return err
	}

	// Send notification via WebSocket
	go s.notificationSvc.NotifyMessageRead(senderID, receiverID, time.Now())

	return nil
}

// GetContacts gets all users that the current user can chat with
func (s *ChatService) GetContacts(userID string) ([]*models.ChatContact, error) {
	return s.repo.GetChatContacts(userID)
}

// SendTypingIndicator sends a typing indicator to a receiver
func (s *ChatService) SendTypingIndicator(senderID, receiverID string) error {
	// Check if users can message each other
	canSend, err := s.repo.CanSendMessage(senderID, receiverID)
	if err != nil {
		return err
	}

	if !canSend {
		return errors.New("you cannot send typing indicators to this user")
	}

	// Send notification via WebSocket
	go s.notificationSvc.NotifyTyping(senderID, receiverID)

	return nil
}

// SearchMessages searches for messages containing the query string
func (s *ChatService) SearchMessages(userID, query, otherUserID string, limit int) ([]*models.PrivateMessage, error) {
	return s.repo.SearchMessages(userID, query, otherUserID, limit)
}
