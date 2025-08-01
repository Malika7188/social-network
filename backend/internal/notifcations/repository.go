package notifications

import (
	"database/sql"
	"time"

	models "github.com/Athooh/social-network/pkg/models/dbTables"
	"github.com/Athooh/social-network/pkg/utils"
)

type Repository interface {
	CreateNotification(notification *models.Notification) error
	GetNotifications(userID string, limit, offset int) ([]*models.Notification, error)
	MarkNotificationAsRead(notificationID int64) error
	MarkAllNotificationsAsRead(userID string) error
	ClearAllNotificationsDB(userId string) error
	DeleteNotificationDb(notificationId int64) error
}

// SQLiteRepository implements Repository interface for SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) CreateNotification(notification *models.Notification) error {
	now := time.Now()
	notification.CreatedAt = now

	query := `
		INSERT INTO notifications (
			user_id, sender_id, type, message, is_read, created_at, target_group_id, target_event_id
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		notification.UserID,
		utils.NullableString(notification.SenderID),
		notification.Type,
		notification.Message,
		notification.IsRead,
		notification.CreatedAt,
		utils.NullableString(notification.TargetGroupID),
		utils.NullableString(notification.TargetEventID),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteRepository) GetNotifications(userID string, limit, offset int) ([]*models.Notification, error) {
	query := `
		SELECT 
			id, user_id, sender_id, type, message, is_read, created_at, target_group_id, target_event_id
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.Notification

	for rows.Next() {
		var n models.Notification
		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.SenderID,
			&n.Type,
			&n.Message,
			&n.IsRead,
			&n.CreatedAt,
			&n.TargetGroupID,
			&n.TargetEventID,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, &n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *SQLiteRepository) MarkNotificationAsRead(notificationID int64) error {
	query := `
		UPDATE notifications
		SET is_read = TRUE
		WHERE id = ?
	`

	_, err := r.db.Exec(query, notificationID)
	return err
}

func (r *SQLiteRepository) MarkAllNotificationsAsRead(userID string) error {
	query := `
		UPDATE notifications
		SET is_read = TRUE
		WHERE user_id = ? AND is_read = FALSE
	`

	_, err := r.db.Exec(query, userID)
	return err
}

func (r *SQLiteRepository) ClearAllNotificationsDB(userId string) error {
	_, err := r.db.Exec("DELETE FROM notifications WHERE user_id =  ?", userId)
	return err
}

func (r *SQLiteRepository) DeleteNotificationDb(notificationId int64) error {
	_, err := r.db.Exec("DELETE FROM notifications WHERE id = ?", notificationId)
	return err
}
