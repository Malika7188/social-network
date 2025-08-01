package follow

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Repository defines the interface for follow data access
type Repository interface {
	// Follow requests
	CreateFollowRequest(followerID, followingID string) error
	GetFollowRequest(followerID, followingID string) (*FollowRequest, error)
	UpdateFollowRequestStatus(followerID, followingID, status string) error
	GetPendingFollowRequests(userID string) ([]*FollowRequest, error)

	// Followers
	CreateFollower(followerID, followingID string) error
	DeleteFollower(followerID, followingID string) error
	IsFollowing(followerID, followingID string) (bool, error)
	GetFollowers(userID string) ([]*Follower, error)
	GetFollowing(userID string) ([]*Follower, error)
	GetFollowersCount(userID string) (int, error)
	GetFollowingCount(userID string) (int, error)

	// User profile check
	IsUserProfilePublic(userID string) (bool, error)

	// Mutual friends
	GetMutualFollowers(userID1, userID2 string) ([]*Follower, error)
	GetMutualFollowersCount(userID1, userID2 string) (int, error)

	GetUsersNotFollowed(userID string) ([]*BasicUser, error)
}

// SQLiteRepository implements Repository interface for SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// CreateFollowRequest creates a new follow request
func (r *SQLiteRepository) CreateFollowRequest(followerID, followingID string) error {
	now := time.Now()

	query := `
		INSERT INTO follow_requests (follower_id, following_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query, followerID, followingID, string(StatusPending), now, now)
	return err
}

// GetFollowRequest retrieves a follow request
func (r *SQLiteRepository) GetFollowRequest(followerID, followingID string) (*FollowRequest, error) {
	query := `
		SELECT id, follower_id, following_id, status, created_at, updated_at
		FROM follow_requests
		WHERE follower_id = ? AND following_id = ?
	`

	var request FollowRequest
	err := r.db.QueryRow(query, followerID, followingID).Scan(
		&request.ID,
		&request.FollowerID,
		&request.FollowingID,
		&request.Status,
		&request.CreatedAt,
		&request.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &request, nil
}

// UpdateFollowRequestStatus updates the status of a follow request
func (r *SQLiteRepository) UpdateFollowRequestStatus(followerID, followingID, status string) error {
	now := time.Now()

	query := `
		UPDATE follow_requests
		SET status = ?, updated_at = ?
		WHERE follower_id = ? AND following_id = ?
	`

	_, err := r.db.Exec(query, status, now, followerID, followingID)
	return err
}

// GetPendingFollowRequests retrieves all pending follow requests for a user
func (r *SQLiteRepository) GetPendingFollowRequests(userID string) ([]*FollowRequest, error) {
	query := `
		SELECT id, follower_id, following_id, status, created_at, updated_at
		FROM follow_requests
		WHERE following_id = ? AND status = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID, string(StatusPending))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*FollowRequest
	for rows.Next() {
		var request FollowRequest
		err := rows.Scan(
			&request.ID,
			&request.FollowerID,
			&request.FollowingID,
			&request.Status,
			&request.CreatedAt,
			&request.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &request)
	}

	return requests, rows.Err()
}

func (r *SQLiteRepository) CreateFollower(followerID, followingID string) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	// Create follower relationship
	query := `
		INSERT INTO followers (follower_id, following_id, created_at)
		VALUES (?, ?, ?)
	`
	_, err = tx.Exec(query, followerID, followingID, now)
	if err != nil {
		return err
	}

	// Update user_stats for the person being followed (increase followers_count)
	err = r.updateUserStats(tx, followingID, "followers_count", true)
	if err != nil {
		return err
	}

	// Update user_stats for the follower (increase following_count)
	err = r.updateUserStats(tx, followerID, "following_count", true)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

func (r *SQLiteRepository) DeleteFollower(followerID, followingID string) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Remove follower relationship
	query := `
		DELETE FROM followers
		WHERE follower_id = ? AND following_id = ?
	`
	_, err = tx.Exec(query, followerID, followingID)
	if err != nil {
		return err
	}

	// Update user_stats for the person being unfollowed (decrease followers_count)
	err = r.updateUserStats(tx, followingID, "followers_count", false)
	if err != nil {
		return err
	}

	// Update user_stats for the follower (decrease following_count)
	err = r.updateUserStats(tx, followerID, "following_count", false)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

func (r *SQLiteRepository) updateUserStats(tx *sql.Tx, userID string, statsType string, increment bool) error {
	now := time.Now()

	// Check if user stats record exists
	var exists bool
	err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM user_stats WHERE user_id = ?)", userID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Update existing record
		var updateQuery string
		if increment {
			updateQuery = fmt.Sprintf("UPDATE user_stats SET %s = %s + 1, updated_at = ? WHERE user_id = ?", statsType, statsType)
		} else {
			updateQuery = fmt.Sprintf("UPDATE user_stats SET %s = CASE WHEN %s > 0 THEN %s - 1 ELSE 0 END, updated_at = ? WHERE user_id = ?", statsType, statsType, statsType)
		}
		_, err = tx.Exec(updateQuery, now, userID)
	} else {
		// Create new record with default values
		var value int
		if increment {
			value = 1
		} else {
			value = 0
		}
		insertQuery := fmt.Sprintf("INSERT INTO user_stats (user_id, %s, created_at, updated_at) VALUES (?, ?, ?, ?)", statsType)
		_, err = tx.Exec(insertQuery, userID, value, now, now)
	}

	return err
}

// IsFollowing checks if a user is following another user
func (r *SQLiteRepository) IsFollowing(followerID, followingID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM followers
		WHERE follower_id = ? AND following_id = ?
	`

	var count int
	err := r.db.QueryRow(query, followerID, followingID).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetFollowers retrieves all followers of a user
func (r *SQLiteRepository) GetFollowers(userID string) ([]*Follower, error) {
	query := `
		SELECT id, follower_id, following_id, created_at
		FROM followers
		WHERE following_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var followers []*Follower
	for rows.Next() {
		var follower Follower
		err := rows.Scan(
			&follower.ID,
			&follower.FollowerID,
			&follower.FollowingID,
			&follower.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		followers = append(followers, &follower)
	}

	return followers, rows.Err()
}

// GetFollowing retrieves all users a user is following
func (r *SQLiteRepository) GetFollowing(userID string) ([]*Follower, error) {
	query := `
		SELECT id, follower_id, following_id, created_at
		FROM followers
		WHERE follower_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var following []*Follower
	for rows.Next() {
		var follower Follower
		err := rows.Scan(
			&follower.ID,
			&follower.FollowerID,
			&follower.FollowingID,
			&follower.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		following = append(following, &follower)
	}

	return following, rows.Err()
}

// GetFollowersCount retrieves the number of followers for a user
func (r *SQLiteRepository) GetFollowersCount(userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM followers
		WHERE following_id = ?
	`

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetFollowingCount retrieves the number of users a user is following
func (r *SQLiteRepository) GetFollowingCount(userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM followers
		WHERE follower_id = ?
	`

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// IsUserProfilePublic checks if a user's profile is public
func (r *SQLiteRepository) IsUserProfilePublic(userID string) (bool, error) {
	query := `
		SELECT is_public
		FROM users
		WHERE id = ?
	`

	var isPublic bool
	err := r.db.QueryRow(query, userID).Scan(&isPublic)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("user not found")
		}
		return false, err
	}

	return isPublic, nil
}

// GetMutualFollowers retrieves mutual followers between two users
func (r *SQLiteRepository) GetMutualFollowers(userID1, userID2 string) ([]*Follower, error) {
	query := `
		SELECT f.id, f.follower_id, f.following_id, f.created_at
		FROM followers f
		JOIN followers f2 ON f.follower_id = f2.following_id AND f2.follower_id = f.following_id
		WHERE f.follower_id = ? AND f2.follower_id = ?
	`

	rows, err := r.db.Query(query, userID1, userID2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mutualFollowers []*Follower
	for rows.Next() {
		var follower Follower
		err := rows.Scan(
			&follower.ID,
			&follower.FollowerID,
			&follower.FollowingID,
			&follower.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		mutualFollowers = append(mutualFollowers, &follower)
	}

	return mutualFollowers, rows.Err()
}

// GetMutualFollowersCount retrieves the count of mutual followers between two users
func (r *SQLiteRepository) GetMutualFollowersCount(userID1, userID2 string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM followers f1
		JOIN followers f2 ON f1.follower_id = f2.follower_id
		WHERE f1.following_id = ? AND f2.following_id = ?
	`

	var count int
	err := r.db.QueryRow(query, userID1, userID2).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *SQLiteRepository) GetUsersNotFollowed(userID string) ([]*BasicUser, error) {
	query := `
		SELECT u.id
		FROM users u
		WHERE u.id != ? 
		AND u.id NOT IN (
			SELECT f.following_id 
			FROM followers f 
			WHERE f.follower_id = ?
		)
		AND u.id NOT IN (
			SELECT fr.following_id 
			FROM follow_requests fr 
			WHERE fr.follower_id = ? AND fr.status = 'pending'
		)
		ORDER BY u.created_at DESC
	`
	rows, err := r.db.Query(query, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*BasicUser
	for rows.Next() {
		var user BasicUser
		err := rows.Scan(&user.ID)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}
