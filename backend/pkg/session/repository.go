package session

import (
	"database/sql"
	"errors"
	"time"

	models "github.com/Athooh/social-network/pkg/models/dbTables"
	"github.com/google/uuid"
)

// SQLiteRepository implements Repository for SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// CreateSession adds a new session to the database
func (r *SQLiteRepository) CreateSession(userID string, expiresAt time.Time) (string, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO sessions (id, user_id, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.Exec(query, sessionID, userID, expiresAt, now)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

// GetSession retrieves a session by ID
func (r *SQLiteRepository) GetSession(sessionID string) (string, time.Time, error) {
	query := `
		SELECT user_id, expires_at
		FROM sessions
		WHERE id = ?
	`

	var userID string
	var expiresAt time.Time

	err := r.db.QueryRow(query, sessionID).Scan(&userID, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", time.Time{}, errors.New("session not found")
		}
		return "", time.Time{}, err
	}

	return userID, expiresAt, nil
}

// DeleteSession removes a session from the database
func (r *SQLiteRepository) DeleteSession(sessionID string) error {
	query := `DELETE FROM sessions WHERE id = ?`
	_, err := r.db.Exec(query, sessionID)
	return err
}

// DeleteUserSessions removes all sessions for a user
func (r *SQLiteRepository) DeleteUserSessions(userID string) error {
	query := `DELETE FROM sessions WHERE user_id = ?`
	_, err := r.db.Exec(query, userID)
	return err
}

// CleanExpired removes all expired sessions
func (r *SQLiteRepository) CleanExpired() error {
	query := `DELETE FROM sessions WHERE expires_at < ?`
	_, err := r.db.Exec(query, time.Now())
	return err
}

// GetUserSessions retrieves all sessions for a user
func (r *SQLiteRepository) GetUserSessions(userID string) ([]models.Session, error) {
	query := `
		SELECT id, user_id, expires_at
		FROM sessions
		WHERE user_id = ?
		ORDER BY expires_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(&session.ID, &session.UserID, &session.ExpiresAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

// HasValidSession checks if a user has any valid (non-expired) sessions
func (r *SQLiteRepository) HasValidSession(userID string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM sessions 
		WHERE user_id = ? AND expires_at > CURRENT_TIMESTAMP
	`

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
