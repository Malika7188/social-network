package post

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	models "github.com/Athooh/social-network/pkg/models/dbTables"
)

// Repository defines the interface for post data access
type Repository interface {
	CreatePost(post *models.Post) error
	GetPostByID(id int64) (*models.Post, error)
	GetPostsByUserID(userID string) ([]*models.Post, error)
	GetPublicPosts(limit, offset int) ([]*models.Post, error)
	UpdatePost(post *models.Post) error
	DeletePost(id int64) error

	// Privacy related methods
	AddPostViewer(postID int64, userID string) error
	RemovePostViewer(postID int64, userID string) error
	GetPostViewers(postID int64) ([]string, error)
	CanViewPost(postID int64, userID string) (bool, error)

	// Comment methods
	CreateComment(comment *models.Comment) error
	UpdatePostCommentCount(postId int64, increase bool) (int, error)
	GetCommentsByPostID(postID int64) ([]*models.Comment, error)
	DeleteComment(id int64) error

	// Like-related methods
	LikePost(postID int64, userID string) error
	UnlikePost(postID int64, userID string) error
	HasLiked(postID int64, userID string) (bool, error)
	GetLikesCount(postID int64) (int, error)
	GetFeedPosts(userID string, limit, offset int) ([]*models.Post, error)

	// User data method
	GetUserDataByID(userID string) (*models.PostUserData, error)

	// New method
	GetUserFollowers(userID string) ([]string, error)
	UpdateUserStats(userID string, statsType string, increment bool) (int, error)
	getNextAvailableID() (int64, error)
}

// SQLiteRepository implements Repository interface for SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// CreatePost creates a new post
func (r *SQLiteRepository) CreatePost(post *models.Post) error {
	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now
    newid, err := r.getNextAvailableID()
	if err != nil {
		return err
	}
	post.ID = newid
	query := `
		INSERT INTO posts (id, user_id, content, image_path, video_path, privacy, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(
		query,
		post.ID,
		post.UserID,
		post.Content,
		post.ImagePath.String,
		post.VideoPath.String,
		post.Privacy,
		post.CreatedAt,
		post.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetPostByID retrieves a post by ID
func (r *SQLiteRepository) GetPostByID(id int64) (*models.Post, error) {
	query := `
		SELECT id, user_id, content, image_path, video_path, privacy, likes_count, created_at, updated_at
		FROM (
			SELECT id, user_id, content, image_path, video_path, privacy, likes_count, created_at, updated_at
			FROM posts
			WHERE id = ?
			UNION ALL
			SELECT id, user_id, content, image_path, video_path, 'public' as privacy, likes_count, created_at, updated_at
			FROM group_posts
			WHERE id = ?
		)
		LIMIT 1
	`

	post := &models.Post{}
	var imagePath, videoPath sql.NullString

	err := r.db.QueryRow(query, id, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Content,
		&imagePath,
		&videoPath,
		&post.Privacy,
		&post.LikesCount,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if imagePath.Valid {
		post.ImagePath = imagePath
	}
	if videoPath.Valid {
		post.VideoPath = videoPath
	}

	return post, nil
}

// GetPostsByUserID retrieves all posts by a user
func (r *SQLiteRepository) GetPostsByUserID(userID string) ([]*models.Post, error) {
	query := `
		SELECT id, user_id, content, image_path, video_path, privacy, likes_count, created_at, updated_at
		FROM posts
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	var imagePath, videoPath sql.NullString
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&imagePath,
			&videoPath,
			&post.Privacy,
			&post.LikesCount,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if imagePath.Valid {
			post.ImagePath = imagePath
		}
		if videoPath.Valid {
			post.VideoPath = videoPath
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// GetPublicPosts retrieves public posts with pagination
func (r *SQLiteRepository) GetPublicPosts(limit, offset int) ([]*models.Post, error) {
	query := `
		SELECT id, user_id, content, image_path, video_path, privacy, likes_count, created_at, updated_at
		FROM posts
		WHERE privacy = 'public'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	var imagePath, videoPath sql.NullString
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&imagePath,
			&videoPath,
			&post.Privacy,
			&post.LikesCount,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// UpdatePost updates an existing post
func (r *SQLiteRepository) UpdatePost(post *models.Post) error {
	post.UpdatedAt = time.Now()

	query := `
		UPDATE posts
		SET content = ?, image_path = ?, privacy = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(
		query,
		post.Content,
		post.ImagePath,
		post.Privacy,
		post.UpdatedAt,
		post.ID,
	)

	return err
}

// DeletePost deletes a post
func (r *SQLiteRepository) DeletePost(id int64) error {
	query := "DELETE FROM posts WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// AddPostViewer adds a user who can view a private post
func (r *SQLiteRepository) AddPostViewer(postID int64, userID string) error {
	query := `
		INSERT INTO post_viewers (post_id, user_id)
		VALUES (?, ?)
	`
	_, err := r.db.Exec(query, postID, userID)
	return err
}

// RemovePostViewer removes a user from viewing a private post
func (r *SQLiteRepository) RemovePostViewer(postID int64, userID string) error {
	query := "DELETE FROM post_viewers WHERE post_id = ? AND user_id = ?"
	_, err := r.db.Exec(query, postID, userID)
	return err
}

// GetPostViewers gets all users who can view a private post
func (r *SQLiteRepository) GetPostViewers(postID int64) ([]string, error) {
	query := "SELECT user_id FROM post_viewers WHERE post_id = ?"
	rows, err := r.db.Query(query, postID)
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}

// CanViewPost checks if a user can view a post based on privacy settings
func (r *SQLiteRepository) CanViewPost(postID int64, userID string) (bool, error) {
	// Get the post to check its privacy setting
	post, err := r.GetPostByID(postID)
	if err != nil {
		return false, err
	}
	if post == nil {
		return false, errors.New("post not found")
	}

	// If the user is the post creator, they can always view it
	if post.UserID == userID {
		return true, nil
	}

	// Check based on privacy setting
	switch post.Privacy {
	case models.PrivacyPublic:
		return true, nil
	case models.PrivacyAlmostPrivate:
		// Check if the user is a follower of the post creator
		query := "SELECT COUNT(*) FROM followers WHERE following_id = ? AND follower_id = ?"
		var count int
		err := r.db.QueryRow(query, post.UserID, userID).Scan(&count)
		if err != nil {
			return false, err
		}
		return count > 0, nil
	case models.PrivacyPrivate:
		// Check if the user is in the post_viewers table
		query := "SELECT COUNT(*) FROM post_viewers WHERE post_id = ? AND user_id = ?"
		var count int
		err := r.db.QueryRow(query, postID, userID).Scan(&count)
		if err != nil {
			return false, err
		}
		return count > 0, nil
	default:
		return false, errors.New("invalid privacy setting")
	}
}

// CreateComment creates a new comment
func (r *SQLiteRepository) CreateComment(comment *models.Comment) error {
	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now

	query := `
		INSERT INTO comments (post_id, user_id, content, image_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		comment.PostID,
		comment.UserID,
		comment.Content,
		comment.ImagePath.String,
		comment.CreatedAt,
		comment.UpdatedAt,
	).Scan(&comment.ID)

	return err
}

// GetCommentsByPostID retrieves all comments for a post
func (r *SQLiteRepository) GetCommentsByPostID(postID int64) ([]*models.Comment, error) {
	query := `
		SELECT id, post_id, user_id, content, image_path, created_at, updated_at
		FROM comments
		WHERE post_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.Comment
	var imagePath sql.NullString
	for rows.Next() {
		comment := &models.Comment{}
		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&imagePath,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if imagePath.Valid {
			comment.ImagePath.String = imagePath.String
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// DeleteComment deletes a comment
func (r *SQLiteRepository) DeleteComment(id int64) error {
	query := "DELETE FROM comments WHERE id = ?"
	_, err := r.db.Exec(query, id)
	return err
}

// LikePost adds a like to a post and increments the likes count
func (r *SQLiteRepository) LikePost(postID int64, userID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert like
	_, err = tx.Exec(`
		INSERT INTO post_likes (post_id, user_id)
		VALUES (?, ?)
	`, postID, userID)
	if err != nil {
		return err
	}

	// Increment likes count in both tables
	_, err = tx.Exec(`
		UPDATE posts 
		SET likes_count = likes_count + 1 
		WHERE id = ?;

		UPDATE group_posts 
		SET likes_count = likes_count + 1 
		WHERE id = ?;
	`, postID, postID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// UnlikePost removes a like from a post and decrements the likes count
func (r *SQLiteRepository) UnlikePost(postID int64, userID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Remove like
	result, err := tx.Exec(`
		DELETE FROM post_likes 
		WHERE post_id = ? AND user_id = ?
	`, postID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected > 0 {
		// Decrement likes count in both tables if a like was actually removed
		_, err = tx.Exec(`
			UPDATE posts 
			SET likes_count = likes_count - 1 
			WHERE id = ? AND likes_count > 0;

			UPDATE group_posts 
			SET likes_count = likes_count - 1 
			WHERE id = ? AND likes_count > 0;
		`, postID, postID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// HasLiked checks if a user has liked a post
func (r *SQLiteRepository) HasLiked(postID int64, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM post_likes 
			WHERE post_id = ? AND user_id = ?
		)
	`, postID, userID).Scan(&exists)
	return exists, err
}

// GetLikesCount gets the number of likes for a post
func (r *SQLiteRepository) GetLikesCount(postID int64) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COALESCE(
			(SELECT likes_count FROM posts WHERE id = ?),
			(SELECT likes_count FROM group_posts WHERE id = ?),
			0
		)
	`, postID, postID).Scan(&count)
	return count, err
}

// GetFeedPosts gets posts visible to the user
func (r *SQLiteRepository) GetFeedPosts(userID string, limit, offset int) ([]*models.Post, error) {
	query := `
		SELECT DISTINCT p.* 
		FROM posts p
		LEFT JOIN followers f ON p.user_id = f.following_id
		LEFT JOIN post_viewers pv ON p.id = pv.post_id
		WHERE 
			p.privacy = 'public'
			OR p.user_id = ?
			OR (p.privacy = 'almost_private' AND f.follower_id = ?)
			OR (p.privacy = 'private' AND pv.user_id = ?)
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, userID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		var imagePath, videoPath sql.NullString

		// Make sure the order matches exactly what's returned from the database
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&imagePath,
			&videoPath,
			&post.Privacy,
			&post.LikesCount,
			&post.CommentsCount,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if imagePath.Valid {
			post.ImagePath = imagePath
		}

		if videoPath.Valid {
			post.VideoPath = videoPath
		}

		user, err := r.GetUserDataByID(post.UserID)
		if err != nil {
			return nil, err
		}
		post.UserData = user

		posts = append(posts, post)
	}

	return posts, rows.Err()
}

// GetUserDataByID retrieves user data by ID
func (r *SQLiteRepository) GetUserDataByID(userID string) (*models.PostUserData, error) {
	query := `
		SELECT id, first_name, last_name, avatar
		FROM users
		WHERE id = ?
	`

	user := &models.PostUserData{}
	err := r.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Avatar,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *SQLiteRepository) UpdatePostCommentCount(postId int64, increase bool) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	// First, check which table the postId exists in
	var isPost bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)", postId).Scan(&isPost)
	if err != nil {
		return 0, err
	}

	var isGroupPost bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM group_posts WHERE id = ?)", postId).Scan(&isGroupPost)
	if err != nil {
		return 0, err
	}

	if !isPost && !isGroupPost {
		return 0, errors.New("post not found in either posts or group_posts")
	}

	// Choose target table
	var tableName string
	if isPost {
		tableName = "posts"
	} else {
		tableName = "group_posts"
	}

	// Initialize comment count if nil
	initQuery := fmt.Sprintf(`
		UPDATE %s 
		SET comments_count = COALESCE(comments_count, (SELECT COUNT(*) FROM comments WHERE post_id = ?)), 
		    updated_at = ? 
		WHERE id = ?`, tableName)
	_, err = tx.Exec(initQuery, postId, now, postId)
	if err != nil {
		return 0, err
	}

	// Increment the comment count
	updateQuery := ""
	if increase {
		updateQuery = fmt.Sprintf("UPDATE %s SET comments_count = comments_count + 1, updated_at = ? WHERE id = ?", tableName)
	} else {
		updateQuery = fmt.Sprintf("UPDATE %s SET comments_count = comments_count - 1, updated_at = ? WHERE id = ?", tableName)
	}
	result, err := tx.Exec(updateQuery, now, postId)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		return 0, errors.New("failed to update comment count")
	}

	// Retrieve the new comment count
	var newCount int
	selectQuery := fmt.Sprintf("SELECT comments_count FROM %s WHERE id = ?", tableName)
	err = tx.QueryRow(selectQuery, postId).Scan(&newCount)
	if err != nil {
		return 0, err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return newCount, nil
}

// GetUserFollowers returns the IDs of users who follow the specified user
func (r *SQLiteRepository) GetUserFollowers(userID string) ([]string, error) {
	var followerIDs []string

	query := `SELECT follower_id FROM followers WHERE following_id = $1`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		followerIDs = append(followerIDs, id)
	}

	return followerIDs, nil
}

// UpdateUserStats updates or inserts a user's statistics
func (r *SQLiteRepository) UpdateUserStats(userID string, statsType string, increment bool) (int, error) {
	// If statsType is posts_count, get the actual count from posts table
	if statsType == "posts_count" {
		var postsCount int
		countQuery := "SELECT COUNT(*) FROM posts WHERE user_id = ?"
		err := r.db.QueryRow(countQuery, userID).Scan(&postsCount)
		if err != nil {
			return 0, err
		}

		// Now update the user_stats table with the actual count
		return r.updateUserStatsWithValue(userID, statsType, postsCount)
	}

	// For other stats types, proceed with the original logic
	return r.updateUserStatsNormal(userID, statsType, increment)
}

// Helper method for the original increment/decrement logic
func (r *SQLiteRepository) updateUserStatsNormal(userID string, statsType string, increment bool) (int, error) {
	// Check if user stats record exists
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM user_stats WHERE user_id = ?)"
	err := r.db.QueryRow(checkQuery, userID).Scan(&exists)
	if err != nil {
		return 0, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()
	var newCount int

	if exists {
		// Update existing record
		var updateQuery string
		var selectQuery string
		if increment {
			updateQuery = fmt.Sprintf("UPDATE user_stats SET %s = %s + 1, updated_at = ? WHERE user_id = ?", statsType, statsType)
		} else {
			updateQuery = fmt.Sprintf("UPDATE user_stats SET %s = CASE WHEN %s > 0 THEN %s - 1 ELSE 0 END, updated_at = ? WHERE user_id = ?", statsType, statsType, statsType)
		}
		_, err = tx.Exec(updateQuery, now, userID)
		if err != nil {
			return 0, err
		}

		// Get the new count
		selectQuery = fmt.Sprintf("SELECT %s FROM user_stats WHERE user_id = ?", statsType)
		err = tx.QueryRow(selectQuery, userID).Scan(&newCount)
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
		if err != nil {
			return 0, err
		}

		newCount = value
	}

	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return newCount, nil
}

// Helper method to update stats with a specific value
func (r *SQLiteRepository) updateUserStatsWithValue(userID string, statsType string, value int) (int, error) {
	// Check if user stats record exists
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT 1 FROM user_stats WHERE user_id = ?)"
	err := r.db.QueryRow(checkQuery, userID).Scan(&exists)
	if err != nil {
		return 0, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	if exists {
		// Update existing record with the provided value
		updateQuery := fmt.Sprintf("UPDATE user_stats SET %s = ?, updated_at = ? WHERE user_id = ?", statsType)
		_, err = tx.Exec(updateQuery, value, now, userID)
	} else {
		// Create new record with the provided value
		insertQuery := fmt.Sprintf("INSERT INTO user_stats (user_id, %s, created_at, updated_at) VALUES (?, ?, ?, ?)", statsType)
		_, err = tx.Exec(insertQuery, userID, value, now, now)
	}

	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return value, nil
}

// getNextAvailableID checks 'posts' and 'group_posts' for max(id) and returns next available int64 ID.
func (r *SQLiteRepository) getNextAvailableID() (int64, error) {
    var maxPostID, maxGroupPostID sql.NullInt64

    // Query max id from 'posts'
    err := r.db.QueryRow("SELECT MAX(id) FROM posts").Scan(&maxPostID)
    if err != nil {
        return 0, fmt.Errorf("error querying posts: %v", err)
    }

    // Query max id from 'group_posts'
    err = r.db.QueryRow("SELECT MAX(id) FROM group_posts").Scan(&maxGroupPostID)
    if err != nil {
        return 0, fmt.Errorf("error querying group_posts: %v", err)
    }

    // Calculate the next ID
    maxID := int64(0)
    if maxPostID.Valid && maxPostID.Int64 > maxID {
        maxID = maxPostID.Int64
    }
    if maxGroupPostID.Valid && maxGroupPostID.Int64 > maxID {
        maxID = maxGroupPostID.Int64
    }

    return maxID + 1, nil
}