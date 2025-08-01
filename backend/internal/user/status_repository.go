package user

import (
	"database/sql"
)

// SQLiteStatusRepository implements StatusRepository for SQLite
type SQLiteStatusRepository struct {
	db *sql.DB
}

// NewSQLiteStatusRepository creates a new SQLite status repository
func NewSQLiteStatusRepository(db *sql.DB) *SQLiteStatusRepository {
	return &SQLiteStatusRepository{db: db}
}

// SetUserOnline marks a user as online
func (r *SQLiteStatusRepository) SetUserOnline(userID string) error {
	query := `
		INSERT INTO user_status (user_id, is_online, last_activity, updated_at)
		VALUES (?, TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id) DO UPDATE SET
		is_online = TRUE,
		last_activity = CURRENT_TIMESTAMP,
		updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.Exec(query, userID)
	return err
}

// SetUserOffline marks a user as offline
func (r *SQLiteStatusRepository) SetUserOffline(userID string) error {
	query := `
		INSERT INTO user_status (user_id, is_online, last_activity, updated_at)
		VALUES (?, FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id) DO UPDATE SET
		is_online = FALSE,
		last_activity = CURRENT_TIMESTAMP,
		updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.Exec(query, userID)
	return err
}

// GetUserStatus gets a user's online status
func (r *SQLiteStatusRepository) GetUserStatus(userID string) (bool, error) {
	query := `
		SELECT is_online FROM user_status
		WHERE user_id = ?
	`
	var isOnline bool
	err := r.db.QueryRow(query, userID).Scan(&isOnline)
	if err == sql.ErrNoRows {
		return false, nil // Default to offline if no record
	}
	return isOnline, err
}

// GetFollowersForStatusUpdate gets the list of users who should be notified of status changes
// This includes both followers and users that the user is following (bidirectional relationship)
func (r *SQLiteStatusRepository) GetFollowersForStatusUpdate(userID string) ([]string, error) {
	query := `
		SELECT DISTINCT user_id FROM (
			-- Users who follow this user (followers)
			SELECT follower_id as user_id FROM followers
			WHERE following_id = ?
			UNION
			-- Users that this user follows (following)
			SELECT following_id as user_id FROM followers
			WHERE follower_id = ?
		)
	`
	rows, err := r.db.Query(query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// GetAllOnlineUsers returns the IDs of all users currently marked as online
func (r *SQLiteStatusRepository) GetAllOnlineUsers() ([]string, error) {
	query := `
		SELECT user_id FROM user_status
		WHERE is_online = TRUE
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}
