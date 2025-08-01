package user

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

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

// Create adds a new user to the database and creates a basic profile
func (r *SQLiteRepository) Create(user *User) error {
	// Start a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Generate ID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
        INSERT INTO users (
            id, email, password, first_name, last_name, date_of_birth,
            avatar, nickname, about_me, is_public, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	_, err = tx.Exec(
		query,
		user.ID, user.Email, user.Password, user.FirstName, user.LastName, user.DateOfBirth,
		user.Avatar, user.Nickname, user.AboutMe, user.IsPublic, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Create basic profile with the username (nickname)
	profileQuery := `
        INSERT INTO user_profiles (
            id, user_id, username, full_name, email, is_private, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `

	fullName := fmt.Sprintf("%s %s", user.FirstName, user.LastName)

	_, err = tx.Exec(
		profileQuery,
		uuid.New().String(), user.ID, user.Nickname, fullName,
		user.Email, !user.IsPublic, now, now,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetByID retrieves a user by ID
func (r *SQLiteRepository) GetByID(id string) (*User, error) {
	
	query := `
		SELECT 
			u.id, u.email, u.password, u.first_name, u.last_name, u.date_of_birth,
			u.avatar, u.nickname, u.about_me, u.is_public, u.created_at, u.updated_at,
			COALESCE(us.posts_count, 0) AS posts_count,
			COALESCE(us.groups_joined, 0) AS groups_joined,
			COALESCE(us.followers_count, 0) AS followers_count,
			COALESCE(us.following_count, 0) AS following_count,
			
			-- Profile fields
			p.username, p.full_name, p.bio, p.work, p.education, 
			p.email AS contact_email, p.phone, p.website, p.location,
			p.tech_skills, p.soft_skills, p.interests,
			p.banner_image, p.profile_image, p.is_private
		FROM users u
		LEFT JOIN user_stats us ON u.id = us.user_id
		LEFT JOIN user_profiles p ON u.id = p.user_id
		WHERE u.id = ?
	`

	var user User
	var username, fullName, bio, work, education, contactEmail, phone sql.NullString
	var website, location, techSkills, softSkills, interests sql.NullString
	var bannerImage, profileImage sql.NullString
	var isPrivate sql.NullBool

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.DateOfBirth,
		&user.Avatar, &user.Nickname, &user.AboutMe, &user.IsPublic, &user.CreatedAt, &user.UpdatedAt,
		&user.PostsCount, &user.GroupsJoined, &user.FollowersCount, &user.FollowingCount,

		// Profile fields with null handling
		&username, &fullName, &bio, &work, &education,
		&contactEmail, &phone, &website, &location,
		&techSkills, &softSkills, &interests,
		&bannerImage, &profileImage, &isPrivate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Add profile fields to user struct
	user.Username = nullStringToString(username)
	user.FullName = nullStringToString(fullName)
	user.Bio = nullStringToString(bio)
	user.Work = nullStringToString(work)
	user.Education = nullStringToString(education)
	user.ContactEmail = nullStringToString(contactEmail)
	user.Phone = nullStringToString(phone)
	user.Website = nullStringToString(website)
	user.Location = nullStringToString(location)
	user.TechSkills = nullStringToString(techSkills)
	user.SoftSkills = nullStringToString(softSkills)
	user.Interests = nullStringToString(interests)
	user.BannerImage = nullStringToString(bannerImage)
	user.ProfileImage = nullStringToString(profileImage)

	if isPrivate.Valid {
		user.IsPrivate = isPrivate.Bool
	}

	return &user, nil
}

// Helper function to handle sql.NullString values
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// GetByEmail retrieves a user by email
func (r *SQLiteRepository) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, email, password, first_name, last_name, date_of_birth, 
		       avatar, nickname, about_me, is_public, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	var user User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.DateOfBirth,
		&user.Avatar, &user.Nickname, &user.AboutMe, &user.IsPublic, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// Delete removes a user from the database
func (r *SQLiteRepository) Delete(id string) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}
