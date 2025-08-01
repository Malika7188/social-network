package event

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	models "github.com/Athooh/social-network/pkg/models/dbTables"
	"github.com/google/uuid"
)

// Repository defines the interface for event data access
type Repository interface {
	// Event operations
	CreateEvent(event *models.GroupEvent) error
	GetEventByID(id string) (*models.GroupEvent, error)
	GetGroupEvents(groupID string) ([]*models.GroupEvent, error)
	UpdateEvent(event *models.GroupEvent) error
	DeleteEvent(id string) error

	// Event responses operations
	AddEventResponse(response *models.EventResponse) error
	GetEventResponses(eventID string, responseType string) ([]*models.EventResponse, error)
	GetUserEventResponse(eventID, userID string) (*models.EventResponse, error)
	UpdateEventResponse(eventID, userID, response string) error
	DeleteEventResponse(eventID, userID string) error
	GetEventResponseCounts(eventID string) (going int, notGoing int, err error)

	// User and group operations
	GetUserBasicByID(userID string) (*models.UserBasic, error)
	GetGroupByID(id string) (*models.Group, error)
	IsGroupMember(groupID, userID string) (bool, error)
	GetMemberRole(groupID, userID string) (string, error)
	GetGroupMembers(groupID string, status string) ([]*models.GroupMember, error)
}

// SQLiteRepository implements Repository interface for SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// CreateEvent creates a new event in a group
func (r *SQLiteRepository) CreateEvent(event *models.GroupEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	now := time.Now()
	event.CreatedAt = now
	event.UpdatedAt = now

	query := `
		INSERT INTO group_events (
			id, group_id, creator_id, title, description, event_date, banner_path, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		event.ID,
		event.GroupID,
		event.CreatorID,
		event.Title,
		event.Description,
		event.EventDate,
		event.BannerPath,
		event.CreatedAt,
		event.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// GetEventByID gets an event by ID
func (r *SQLiteRepository) GetEventByID(id string) (*models.GroupEvent, error) {
	query := `
		SELECT id, group_id, creator_id, title, description, event_date, banner_path, created_at, updated_at
		FROM group_events
		WHERE id = ?
	`

	var event models.GroupEvent
	var bannerPath sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&event.ID,
		&event.GroupID,
		&event.CreatorID,
		&event.Title,
		&event.Description,
		&event.EventDate,
		&bannerPath,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	event.BannerPath = bannerPath

	// Get creator info
	creator, err := r.GetUserBasicByID(event.CreatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator info: %w", err)
	}
	event.Creator = creator

	// Get response counts
	going, notGoing, err := r.GetEventResponseCounts(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get response counts: %w", err)
	}

	event.GoingCount = going
	event.NotGoingCount = notGoing

	return &event, nil
}

// GetGroupEvents gets all events in a group
func (r *SQLiteRepository) GetGroupEvents(groupID string) ([]*models.GroupEvent, error) {
	query := `
		SELECT id, group_id, creator_id, title, description, event_date, banner_path, created_at, updated_at
		FROM group_events
		WHERE group_id = ?
		ORDER BY event_date ASC
	`

	rows, err := r.db.Query(query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group events: %w", err)
	}
	defer rows.Close()

	var events []*models.GroupEvent

	for rows.Next() {
		var event models.GroupEvent
		var bannerPath sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.GroupID,
			&event.CreatorID,
			&event.Title,
			&event.Description,
			&event.EventDate,
			&bannerPath,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		event.BannerPath = bannerPath

		// Get creator info
		creator, err := r.GetUserBasicByID(event.CreatorID)
		if err != nil {
			return nil, fmt.Errorf("failed to get creator info: %w", err)
		}
		event.Creator = creator

		// Get response counts
		going, notGoing, err := r.GetEventResponseCounts(event.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get response counts: %w", err)
		}

		event.GoingCount = going
		event.NotGoingCount = notGoing

		events = append(events, &event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %w", err)
	}

	return events, nil
}

// UpdateEvent updates an event
func (r *SQLiteRepository) UpdateEvent(event *models.GroupEvent) error {
	event.UpdatedAt = time.Now()

	query := `
		UPDATE group_events
		SET title = ?, description = ?, event_date = ?, banner_path = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(
		query,
		event.Title,
		event.Description,
		event.EventDate,
		event.BannerPath,
		event.UpdatedAt,
		event.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	return nil
}

// DeleteEvent deletes an event
func (r *SQLiteRepository) DeleteEvent(id string) error {
	_, err := r.db.Exec("DELETE FROM group_events WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	return nil
}

// AddEventResponse adds a user's response to an event
func (r *SQLiteRepository) AddEventResponse(response *models.EventResponse) error {
	if response.ID == "" {
		response.ID = uuid.New().String()
	}

	now := time.Now()
	response.CreatedAt = now
	response.UpdatedAt = now

	// Check if response already exists
	existingResponse, err := r.GetUserEventResponse(response.EventID, response.UserID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check existing response: %w", err)
	}

	// If response exists, update it
	if existingResponse != nil {
		return r.UpdateEventResponse(response.EventID, response.UserID, response.Response)
	}

	// Otherwise, insert new response
	query := `
		INSERT INTO event_responses (
			id, event_id, user_id, response, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(
		query,
		response.ID,
		response.EventID,
		response.UserID,
		response.Response,
		response.CreatedAt,
		response.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add event response: %w", err)
	}

	return nil
}

// GetEventResponses gets all responses to an event with optional response type filter
func (r *SQLiteRepository) GetEventResponses(eventID string, responseType string) ([]*models.EventResponse, error) {
	var query string
	var args []interface{}

	if responseType != "" {
		query = `
			SELECT id, event_id, user_id, response, created_at, updated_at
			FROM event_responses
			WHERE event_id = ? AND response = ?
			ORDER BY created_at DESC
		`
		args = []interface{}{eventID, responseType}
	} else {
		query = `
			SELECT id, event_id, user_id, response, created_at, updated_at
			FROM event_responses
			WHERE event_id = ?
			ORDER BY created_at DESC
		`
		args = []interface{}{eventID}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get event responses: %w", err)
	}
	defer rows.Close()

	var responses []*models.EventResponse

	for rows.Next() {
		var response models.EventResponse

		err := rows.Scan(
			&response.ID,
			&response.EventID,
			&response.UserID,
			&response.Response,
			&response.CreatedAt,
			&response.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan response row: %w", err)
		}

		// Get user info
		user, err := r.GetUserBasicByID(response.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user info: %w", err)
		}
		response.User = user

		responses = append(responses, &response)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating response rows: %w", err)
	}

	return responses, nil
}

// GetUserEventResponse gets a user's response to an event
func (r *SQLiteRepository) GetUserEventResponse(eventID, userID string) (*models.EventResponse, error) {
	query := `
		SELECT id, event_id, user_id, response, created_at, updated_at
		FROM event_responses
		WHERE event_id = ? AND user_id = ?
	`

	var response models.EventResponse

	err := r.db.QueryRow(query, eventID, userID).Scan(
		&response.ID,
		&response.EventID,
		&response.UserID,
		&response.Response,
		&response.CreatedAt,
		&response.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Response not found
		}
		return nil, fmt.Errorf("failed to get user event response: %w", err)
	}

	// Get user info
	user, err := r.GetUserBasicByID(response.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	response.User = user

	return &response, nil
}

// UpdateEventResponse updates a user's response to an event
func (r *SQLiteRepository) UpdateEventResponse(eventID, userID, response string) error {
	query := `
		UPDATE event_responses
		SET response = ?, updated_at = ?
		WHERE event_id = ? AND user_id = ?
	`

	_, err := r.db.Exec(
		query,
		response,
		time.Now(),
		eventID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update event response: %w", err)
	}

	return nil
}

// DeleteEventResponse deletes a user's response to an event
func (r *SQLiteRepository) DeleteEventResponse(eventID, userID string) error {
	_, err := r.db.Exec(
		"DELETE FROM event_responses WHERE event_id = ? AND user_id = ?",
		eventID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete event response: %w", err)
	}

	return nil
}

// GetEventResponseCounts gets the counts of going and not going responses
func (r *SQLiteRepository) GetEventResponseCounts(eventID string) (going int, notGoing int, err error) {
	// Get going count
	err = r.db.QueryRow(
		"SELECT COUNT(*) FROM event_responses WHERE event_id = ? AND response = 'going'",
		eventID,
	).Scan(&going)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get going count: %w", err)
	}

	// Get not going count
	err = r.db.QueryRow(
		"SELECT COUNT(*) FROM event_responses WHERE event_id = ? AND response = 'not_going'",
		eventID,
	).Scan(&notGoing)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get not going count: %w", err)
	}

	return going, notGoing, nil
}

// GetUserBasicByID gets basic user information by ID
func (r *SQLiteRepository) GetUserBasicByID(userID string) (*models.UserBasic, error) {
	query := `
		SELECT id, first_name, last_name, avatar
		FROM users
		WHERE id = ?
	`

	var user models.UserBasic
	var avatar sql.NullString

	err := r.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&avatar,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Avatar = avatar.String

	return &user, nil
}

// GetGroupByID gets a group by ID
func (r *SQLiteRepository) GetGroupByID(id string) (*models.Group, error) {
	query := `
		SELECT id, name, description, creator_id, banner_path, profile_pic_path, is_public, created_at, updated_at
		FROM groups
		WHERE id = ?
	`

	var group models.Group
	var bannerPath, profilePicPath sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&group.CreatorID,
		&bannerPath,
		&profilePicPath,
		&group.IsPublic,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("group not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	group.BannerPath = bannerPath
	group.ProfilePicPath = profilePicPath

	return &group, nil
}

// IsGroupMember checks if a user is a member of a group
func (r *SQLiteRepository) IsGroupMember(groupID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ? AND status = 'accepted'",
		groupID, userID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check group membership: %w", err)
	}

	return count > 0, nil
}

// GetMemberRole gets a member's role in a group
func (r *SQLiteRepository) GetMemberRole(groupID, userID string) (string, error) {
	var role string
	err := r.db.QueryRow(
		"SELECT role FROM group_members WHERE group_id = ? AND user_id = ? AND status = 'accepted'",
		groupID, userID,
	).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user is not a member of this group")
		}
		return "", fmt.Errorf("failed to get member role: %w", err)
	}

	return role, nil
}

// GetGroupMembers gets all members of a group with optional status filter
func (r *SQLiteRepository) GetGroupMembers(groupID string, status string) ([]*models.GroupMember, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `
			SELECT id, group_id, user_id, role, status, invited_by, created_at, updated_at
			FROM group_members
			WHERE group_id = ? AND status = ?
			ORDER BY created_at ASC
		`
		args = []interface{}{groupID, status}
	} else {
		query = `
			SELECT id, group_id, user_id, role, status, invited_by, created_at, updated_at
			FROM group_members
			WHERE group_id = ?
			ORDER BY created_at ASC
		`
		args = []interface{}{groupID}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	defer rows.Close()

	var members []*models.GroupMember

	for rows.Next() {
		var member models.GroupMember
		var invitedBy sql.NullString

		err := rows.Scan(
			&member.ID,
			&member.GroupID,
			&member.UserID,
			&member.Role,
			&member.Status,
			&invitedBy,
			&member.CreatedAt,
			&member.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member row: %w", err)
		}

		member.InvitedBy = invitedBy.String

		// Get user data
		userData, err := r.GetUserBasicByID(member.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user data: %w", err)
		}

		member.User = userData
		members = append(members, &member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating member rows: %w", err)
	}

	return members, nil
}
