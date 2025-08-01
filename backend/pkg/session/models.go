package session

import (
	"time"

	models "github.com/Athooh/social-network/pkg/models/dbTables"
)

// SessionStore defines the interface for session storage
type SessionStore interface {
	CreateSession(userID string, expiresAt time.Time) (string, error)
	GetSession(sessionID string) (string, time.Time, error)
	DeleteSession(sessionID string) error
	DeleteUserSessions(userID string) error
	CleanExpired() error
	GetUserSessions(userID string) ([]models.Session, error)
	HasValidSession(userID string) (bool, error)
}
