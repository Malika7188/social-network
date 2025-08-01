package group

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	models "github.com/Athooh/social-network/pkg/models/dbTables"
	"github.com/google/uuid"
)

// Repository defines the interface for group data access
type Repository interface {
	// Group operations
	CreateGroup(group *models.Group) error
	GetGroupByID(id string) (*models.Group, error)
	GetGroupsByUserID(userID, viewerID string) ([]*models.Group, error)
	GetAllGroups(userid string, limit, offset int) ([]*models.Group, error)
	UpdateGroup(group *models.Group) error
	DeleteGroup(id string) error
	DeleteMembers(groupID string) error
	GetGroupMemberCount(groupID string) (int, error)

	// Group membership operations
	AddMember(member *models.GroupMember) error
	GetMemberByID(groupID, userID string) (*models.GroupMember, error)
	GetGroupMembers(groupID string, status string) ([]*models.GroupMember, error)
	GetUserGroups(userID, viewerID string) ([]*models.Group, error)
	UpdateMemberStatus(groupID, userID, status string) error
	UpdateMemberRole(groupID, userID, role string) error
	RemoveMember(groupID, userID string) error
	IsGroupMember(groupID, userID string) (bool, error)
	GetMemberRole(groupID, userID string) (string, error)

	// Group posts operations
	CreateGroupPost(post *models.GroupPost) error
	GetGroupPosts(groupID string, currentUserID string, limit, offset int) ([]*models.GroupPost, error)
	GetGroupPostByID(id int64) (*models.GroupPost, error)
	DeleteGroupPost(id int64) error

	// Group chat operations
	AddChatMessage(message *models.GroupChatMessage) error
	GetGroupChatMessages(groupID string, limit, offset int) ([]*models.GroupChatMessage, error)

	// User data operations
	GetUserBasicByID(userID string) (*models.UserBasic, error)
	UpdateUserGroupCount(userID string, increment bool) (int, error)
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

// CreateGroup creates a new group
func (r *SQLiteRepository) CreateGroup(group *models.Group) error {
	if group.ID == "" {
		group.ID = uuid.New().String()
	}

	now := time.Now()
	group.CreatedAt = now
	group.UpdatedAt = now

	query := `
		INSERT INTO groups (
			id, name, description, creator_id, banner_path, profile_pic_path, 
			is_public, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		group.ID,
		group.Name,
		group.Description,
		group.CreatorID,
		group.BannerPath,
		group.ProfilePicPath,
		group.IsPublic,
		group.CreatedAt,
		group.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	// Add creator as admin member
	member := &models.GroupMember{
		ID:        uuid.New().String(),
		GroupID:   group.ID,
		UserID:    group.CreatorID,
		Role:      "admin",
		Status:    "accepted",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := r.AddMember(member); err != nil {
		return fmt.Errorf("failed to add creator as member: %w", err)
	}

	// // Update user's group count
	// _, err = r.UpdateUserGroupCount(group.CreatorID, true)
	// if err != nil {
	// 	return fmt.Errorf("failed to update user group count: %w", err)
	// }

	return nil
}

// GetGroupByID retrieves a group by ID
func (r *SQLiteRepository) GetGroupByID(id string) (*models.Group, error) {
	query := `
		SELECT id, name, description, creator_id, banner_path, profile_pic_path, 
		       is_public, created_at, updated_at
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

	// Get member count
	count, err := r.GetGroupMemberCount(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get member count: %w", err)
	}
	group.MemberCount = count

	// Get creator info
	creator, err := r.GetUserBasicByID(group.CreatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator info: %w", err)
	}
	group.Creator = creator

	members, err := r.GetGroupMembers(group.ID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	group.Members = members

	return &group, nil
}

// GetGroupsByUserID retrieves all groups a user is a member of
func (r *SQLiteRepository) GetGroupsByUserID(userID, viewerID string) ([]*models.Group, error) {
	query := `
		SELECT g.id, g.name, g.description, g.creator_id, g.banner_path, g.profile_pic_path, 
		       g.is_public, g.created_at, g.updated_at, gm.role, gm.status
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = ? AND gm.status = 'accepted'
		ORDER BY g.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}
	defer rows.Close()

	var groups []*models.Group

	for rows.Next() {
		var group models.Group
		var bannerPath, profilePicPath sql.NullString
		var role, status string

		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
			&group.CreatorID,
			&bannerPath,
			&profilePicPath,
			&group.IsPublic,
			&group.CreatedAt,
			&group.UpdatedAt,
			&role,
			&status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group row: %w", err)
		}

		group.BannerPath = bannerPath
		group.ProfilePicPath = profilePicPath

		// Get member count
		count, err := r.GetGroupMemberCount(group.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get member count: %w", err)
		}
		group.MemberCount = count

		// Get creator info
		creator, err := r.GetUserBasicByID(group.CreatorID)
		if err != nil {
			return nil, fmt.Errorf("failed to get creator info: %w", err)
		}
		group.Creator = creator

		members, err := r.GetGroupMembers(group.ID, "accepted")
		if err != nil {
			return nil, fmt.Errorf("failed to get group members: %w", err)
		}
		group.Members = members

		member, err := r.GetMemberByID(group.ID, viewerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get member info: %w", err)
		}
		if member != nil {
			group.IsMember = true
		}
		groups = append(groups, &group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group rows: %w", err)
	}

	return groups, nil
}

// GetAllGroups retrieves all groups with pagination
func (r *SQLiteRepository) GetAllGroups(userid string, limit, offset int) ([]*models.Group, error) {
	query := `
		SELECT id, name, description, creator_id, banner_path, profile_pic_path, 
		       is_public, created_at, updated_at
		FROM groups
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}
	defer rows.Close()

	var groups []*models.Group

	for rows.Next() {
		var group models.Group
		var bannerPath, profilePicPath sql.NullString

		err := rows.Scan(
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
			return nil, fmt.Errorf("failed to scan group row: %w", err)
		}

		group.BannerPath = bannerPath
		group.ProfilePicPath = profilePicPath

		// Get member count
		count, err := r.GetGroupMemberCount(group.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get member count: %w", err)
		}
		group.MemberCount = count

		// Get creator info
		creator, err := r.GetUserBasicByID(group.CreatorID)
		if err != nil {
			return nil, fmt.Errorf("failed to get creator info: %w", err)
		}
		group.Creator = creator

		members, err := r.GetGroupMembers(group.ID, "")
		if err != nil {
			return nil, fmt.Errorf("failed to get group members: %w", err)
		}
		group.Members = members

		member, err := r.GetMemberByID(group.ID, userid)
		if err != nil {
			return nil, fmt.Errorf("failed to get member info: %w", err)
		}
		if member != nil && member.Status == "accepted" {
			group.IsMember = true
		}

		groups = append(groups, &group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group rows: %w", err)
	}

	return groups, nil
}

// UpdateGroup updates a group's information
func (r *SQLiteRepository) UpdateGroup(group *models.Group) error {
	group.UpdatedAt = time.Now()

	query := `
		UPDATE groups
		SET name = ?, description = ?, banner_path = ?, profile_pic_path = ?, 
		    is_public = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(
		query,
		group.Name,
		group.Description,
		group.BannerPath,
		group.ProfilePicPath,
		group.IsPublic,
		group.UpdatedAt,
		group.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	return nil
}

// DeleteGroup deletes a group and all its members
func (r *SQLiteRepository) DeleteGroup(id string) error {
	// Get all members to update their group counts
	members, err := r.GetGroupMembers(id, "accepted")
	if err != nil {
		return fmt.Errorf("failed to get group members: %w", err)
	}

	// Start a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if there's an error

	// Delete the group
	_, err = tx.Exec("DELETE FROM groups WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Delete all members from the group
	err = r.DeleteMembers(id)
	if err != nil {
		return fmt.Errorf("failed to delete group members: %w", err)
	}

	// Update group counts for all members
	for _, member := range members {
		_, err = r.UpdateUserGroupCount(member.UserID, false)
		if err != nil {
			return fmt.Errorf("failed to update user group count for user %s: %w", member.UserID, err)
		}
	}

	return nil
}

func (r *SQLiteRepository) DeleteMembers(groupID string) error {
	// Start a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Delete all members from the group
	_, err = tx.Exec("DELETE FROM group_members WHERE group_id = ?", groupID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete group members: %w", err)
	}
	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
// GetGroupMemberCount gets the count of accepted members in a group
func (r *SQLiteRepository) GetGroupMemberCount(groupID string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM group_members 
		WHERE group_id = ? AND status = 'accepted'
	`

	var count int
	err := r.db.QueryRow(query, groupID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get member count: %w", err)
	}

	return count, nil
}

// AddMember adds a member to a group
func (r *SQLiteRepository) AddMember(member *models.GroupMember) error {
	if member.ID == "" {
		member.ID = uuid.New().String()
	}

	now := time.Now()
	member.CreatedAt = now
	member.UpdatedAt = now

	query := `
		INSERT INTO group_members (
			id, group_id, user_id, role, status, invited_by, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		member.ID,
		member.GroupID,
		member.UserID,
		member.Role,
		member.Status,
		member.InvitedBy,
		member.CreatedAt,
		member.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	// If status is accepted, update user's group count
	if member.Status == "accepted" {
		_, err = r.UpdateUserGroupCount(member.UserID, true)
		if err != nil {
			return fmt.Errorf("failed to update user group count: %w", err)
		}
	}

	return nil
}

// GetMemberByID gets a member by group ID and user ID
func (r *SQLiteRepository) GetMemberByID(groupID, userID string) (*models.GroupMember, error) {
	query := `
		SELECT id, group_id, user_id, role, status, invited_by, created_at, updated_at
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`

	var member models.GroupMember

	err := r.db.QueryRow(query, groupID, userID).Scan(
		&member.ID,
		&member.GroupID,
		&member.UserID,
		&member.Role,
		&member.Status,
		&member.InvitedBy,
		&member.CreatedAt,
		&member.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Member not found
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	// Get user info
	user, err := r.GetUserBasicByID(member.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	member.User = user

	// Get inviter info if available
	if member.InvitedBy != "" {
		inviter, err := r.GetUserBasicByID(member.InvitedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to get inviter info: %w", err)
		}
		member.Inviter = inviter
	}

	return &member, nil
}

// GetGroupMembers gets all members of a group with optional status filter
func (r *SQLiteRepository) GetGroupMembers(groupID string, status string) ([]*models.GroupMember, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `
			SELECT 
				gm.id, gm.group_id, gm.user_id, gm.role, gm.status, 
				gm.invited_by, gm.created_at, gm.updated_at,
				u.avatar
			FROM group_members gm
			JOIN users u ON gm.user_id = u.id
			WHERE gm.group_id = ? AND gm.status = ?
			ORDER BY gm.created_at DESC
		`
		args = []interface{}{groupID, status}
	} else {
		query = `
			SELECT 
				gm.id, gm.group_id, gm.user_id, gm.role, gm.status, 
				gm.invited_by, gm.created_at, gm.updated_at,
				u.avatar
			FROM group_members gm
			JOIN users u ON gm.user_id = u.id
	        WHERE gm.group_id = ?
			ORDER BY gm.created_at DESC
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
			&member.Avatar,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group member row: %w", err)
		}

		// Get inviter info if available
		if invitedBy.Valid {
			member.InvitedBy = invitedBy.String
		}

		// Get user info
		user, err := r.GetUserBasicByID(member.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user info: %w", err)
		}
		member.User = user

		// Get inviter info if available
		if member.InvitedBy != "" {
			inviter, err := r.GetUserBasicByID(member.InvitedBy)
			if err != nil {
				return nil, fmt.Errorf("failed to get inviter info: %w", err)
			}
			member.Inviter = inviter
		}

		members = append(members, &member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group member rows: %w", err)
	}

	return members, nil
}

// GetUserGroups gets all groups a user is a member of
func (r *SQLiteRepository) GetUserGroups(userID, viewerID string) ([]*models.Group, error) {
	return r.GetGroupsByUserID(userID, viewerID)
}

// UpdateMemberStatus updates a member's status
func (r *SQLiteRepository) UpdateMemberStatus(groupID, userID, status string) error {
	// Get current member status
	member, err := r.GetMemberByID(groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to get member: %w", err)
	}

	if member == nil {
		return errors.New("member not found")
	}

	// Start a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Update member status
	_, err = tx.Exec(
		"UPDATE group_members SET status = ?, updated_at = ? WHERE group_id = ? AND user_id = ?",
		status,
		time.Now(),
		groupID,
		userID,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update member status: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update user's group count if status changed to/from accepted
	if member.Status != "accepted" && status == "accepted" {
		// Increment group count
		_, err = r.UpdateUserGroupCount(userID, true)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to increment user group count: %w", err)
		}
	} else if member.Status == "accepted" && status != "accepted" {
		// Decrement group count
		_, err = r.UpdateUserGroupCount(userID, false)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to decrement user group count: %w", err)
		}
	}

	return nil
}

// UpdateMemberRole updates a member's role
func (r *SQLiteRepository) UpdateMemberRole(groupID, userID, role string) error {
	query := `
		UPDATE group_members 
		SET role = ?, updated_at = ? 
		WHERE group_id = ? AND user_id = ?
	`

	_, err := r.db.Exec(
		query,
		role,
		time.Now(),
		groupID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	return nil
}

// RemoveMember removes a member from a group
func (r *SQLiteRepository) RemoveMember(groupID, userID string) error {
	// Check if member exists and is accepted
	member, err := r.GetMemberByID(groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to get member: %w", err)
	}

	if member == nil {
		return errors.New("member not found")
	}

	// Start a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Delete the member
	_, err = tx.Exec(
		"DELETE FROM group_members WHERE group_id = ? AND user_id = ?",
		groupID,
		userID,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update user's group count if status was accepted
	if member.Status == "accepted" {
		_, err = r.UpdateUserGroupCount(userID, false)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update user group count: %w", err)
		}
	}
	return nil
}

// IsGroupMember checks if a user is a member of a group
func (r *SQLiteRepository) IsGroupMember(groupID, userID string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM group_members 
		WHERE group_id = ? AND user_id = ? AND status = 'accepted'
	`

	var count int
	err := r.db.QueryRow(query, groupID, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}

	return count > 0, nil
}

// GetMemberRole gets a member's role in a group
func (r *SQLiteRepository) GetMemberRole(groupID, userID string) (string, error) {
	query := `
		SELECT role 
		FROM group_members 
		WHERE group_id = ? AND user_id = ? AND status = 'accepted'
	`

	var role string
	err := r.db.QueryRow(query, groupID, userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Not a member
		}
		return "", fmt.Errorf("failed to get member role: %w", err)
	}

	return role, nil
}

// CreateGroupPost creates a new post in a group
func (r *SQLiteRepository) CreateGroupPost(post *models.GroupPost) error {
	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now
	newid, err := r.getNextAvailableID()
	if err != nil {
		return err
	}
	post.ID = newid

	query := `
		INSERT INTO group_posts (
			id, group_id, user_id, content, image_path, video_path, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(
		query,
		post.ID,
		post.GroupID,
		post.UserID,
		post.Content,
		post.ImagePath,
		post.VideoPath,
		post.CreatedAt,
		post.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create group post: %w", err)
	}

	return nil
}

// GetGroupPosts gets all posts in a group with pagination
func (r *SQLiteRepository) GetGroupPosts(groupID string, currentUserID string, limit, offset int) ([]*models.GroupPost, error) {
	query := `
        SELECT gp.id, gp.group_id, gp.user_id, gp.content, gp.image_path, gp.video_path, 
               gp.likes_count, gp.comments_count, gp.created_at, gp.updated_at,
               CASE WHEN pl.user_id IS NOT NULL THEN 1 ELSE 0 END as is_liked
        FROM group_posts gp
        LEFT JOIN post_likes pl ON pl.post_id = gp.id AND pl.user_id = ?
        WHERE gp.group_id = ?
        ORDER BY gp.created_at DESC
        LIMIT ? OFFSET ?
    `

	rows, err := r.db.Query(query, currentUserID, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get group posts: %w", err)
	}
	defer rows.Close()

	var posts []*models.GroupPost

	for rows.Next() {
		var post models.GroupPost
		var imagePath, videoPath sql.NullString
		var isLiked bool

		err := rows.Scan(
			&post.ID,
			&post.GroupID,
			&post.UserID,
			&post.Content,
			&imagePath,
			&videoPath,
			&post.LikesCount,
			&post.CommentsCount,
			&post.CreatedAt,
			&post.UpdatedAt,
			&isLiked,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}

		post.ImagePath = imagePath
		post.VideoPath = videoPath
		post.Isliked = isLiked

		// Get user data
		userData, err := r.GetUserBasicByID(post.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user data: %w", err)
		}

		post.User = &models.PostUserData{
			ID:        userData.ID,
			FirstName: userData.FirstName,
			LastName:  userData.LastName,
			Avatar:    userData.Avatar,
		}

		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating post rows: %w", err)
	}

	return posts, nil
}

// GetGroupPostByID gets a post by ID
func (r *SQLiteRepository) GetGroupPostByID(id int64) (*models.GroupPost, error) {
	query := `
		SELECT id, group_id, user_id, content, image_path, video_path, likes_count, comments_count, created_at, updated_at
		FROM group_posts
		WHERE id = ?
	`

	var post models.GroupPost
	var imagePath, videoPath sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&post.ID,
		&post.GroupID,
		&post.UserID,
		&post.Content,
		&imagePath,
		&videoPath,
		&post.LikesCount,
		&post.CommentsCount,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("post not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	post.ImagePath = imagePath
	post.VideoPath = videoPath

	// Get user data
	userData, err := r.GetUserBasicByID(post.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user data: %w", err)
	}

	post.User = &models.PostUserData{
		ID:        userData.ID,
		FirstName: userData.FirstName,
		LastName:  userData.LastName,
		Avatar:    userData.Avatar,
	}

	return &post, nil
}

// DeleteGroupPost deletes a post
func (r *SQLiteRepository) DeleteGroupPost(id int64) error {
	_, err := r.db.Exec("DELETE FROM group_posts WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	return nil
}

// AddChatMessage adds a message to a group chat
func (r *SQLiteRepository) AddChatMessage(message *models.GroupChatMessage) error {
	query := `
		INSERT INTO group_chat_messages (
			group_id, user_id, content, created_at
		) VALUES (?, ?, ?, ?)
	`

	result, err := r.db.Exec(
		query,
		message.GroupID,
		message.UserID,
		message.Content,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to add chat message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	message.ID = id
	return nil
}

// GetGroupChatMessages gets messages from a group chat with pagination
func (r *SQLiteRepository) GetGroupChatMessages(groupID string, limit, offset int) ([]*models.GroupChatMessage, error) {
	query := `
		SELECT id, group_id, user_id, content, created_at
		FROM group_chat_messages
		WHERE group_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.GroupChatMessage

	for rows.Next() {
		var message models.GroupChatMessage

		err := rows.Scan(
			&message.ID,
			&message.GroupID,
			&message.UserID,
			&message.Content,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message row: %w", err)
		}

		// Get user info
		user, err := r.GetUserBasicByID(message.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user info: %w", err)
		}
		message.User = user

		messages = append(messages, &message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating message rows: %w", err)
	}

	// Reverse the order to get oldest messages first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
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
		return nil, fmt.Errorf("failed to get user basic info: %w", err)
	}

	if avatar.Valid {
		user.Avatar = avatar.String
	}

	return &user, nil
}

// UpdateUserGroupCount updates a user's group count
func (r *SQLiteRepository) UpdateUserGroupCount(userID string, increment bool) (int, error) {
	// Start a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Get current count
	var currentCount int
	err = tx.QueryRow(
		"SELECT groups_joined FROM user_stats WHERE user_id = ?",
		userID,
	).Scan(&currentCount)

	if err != nil {
		// If user stats don't exist, create them
		if err == sql.ErrNoRows {
			if increment {
				_, err = tx.Exec(
					"INSERT INTO user_stats (user_id, groups_joined) VALUES (?, 1)",
					userID,
				)
			} else {
				_, err = tx.Exec(
					"INSERT INTO user_stats (user_id, groups_joined) VALUES (?, 0)",
					userID,
				)
			}

			if err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("failed to create user stats: %w", err)
			}

			if increment {
				currentCount = 1
			} else {
				currentCount = 0
			}
		} else {
			tx.Rollback()
			return 0, fmt.Errorf("failed to get current group count: %w", err)
		}
	} else {
		// Update count
		if increment {
			currentCount++
		} else if currentCount > 0 {
			currentCount--
		}

		_, err = tx.Exec(
			"UPDATE user_stats SET groups_joined = ?, updated_at = ? WHERE user_id = ?",
			currentCount,
			time.Now(),
			userID,
		)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to update group count: %w", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return currentCount, nil
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
