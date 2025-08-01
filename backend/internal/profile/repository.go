package profile

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// SQLiteRepository implements Repository interface for SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// Repository interface for profile data persistence
type Repository interface {
	UpdateUserProfile(userID string, profileData map[string]interface{}) error
	GetUserProfileByID(userID string) (*UserProfileData, error)
	IsUserProfilePublic(userID string) (bool, error)
	IsUserFollowing(followerID string, followingID string) (bool, error)
}

// NewSQLiteRepository creates a new SQLite repository
func NewSQLiteRepository(db *sql.DB) Repository {
	return &SQLiteRepository{
		db: db,
	}
}

func (r *SQLiteRepository) IsUserProfilePublic(userID string) (bool, error) {
	var isPublic bool
	query := `SELECT is_public FROM users WHERE id = ?`
	err := r.db.QueryRow(query, userID).Scan(&isPublic)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, fmt.Errorf("user not found")
		}
		return false, fmt.Errorf("failed to query user visibility: %w", err)
	}
	return isPublic, nil
}

func (r *SQLiteRepository) IsUserFollowing(followerID string, followingID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM followers WHERE follower_id = ? AND following_id = ?`
	err := r.db.QueryRow(query, followerID, followingID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check following status: %w", err)
	}
	return count > 0, nil
}

func (r *SQLiteRepository) UpdateUserProfile(userID string, profileData map[string]interface{}) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := time.Now()

	if len(profileData) > 0 {
		// Build the user table update query dynamically based on what data was provided
		userFields := []string{"updated_at = ?"}
		userValues := []interface{}{now}

		// Fields from User that can be updated from profile edit
		if nickname, ok := profileData["username"].(string); ok && nickname != "" {
			userFields = append(userFields, "nickname = ?")
			userValues = append(userValues, nickname)
		}

		if aboutMe, ok := profileData["bio"].(string); ok && aboutMe != "" {
			userFields = append(userFields, "about_me = ?")
			userValues = append(userValues, aboutMe)
		}

		if avatar, ok := profileData["profileImage"].(string); ok && avatar != "" {
			userFields = append(userFields, "avatar = ?")
			userValues = append(userValues, avatar)
		}

		if isPrivate, ok := profileData["isPrivate"].(bool); ok {
			userFields = append(userFields, "is_public = ?")
			userValues = append(userValues, !isPrivate) // Note: isPrivate is the inverse of isPublic
		}

		// Add the WHERE clause values
		userValues = append(userValues, userID)

		if len(userFields) > 1 {
			userQuery := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(userFields, ", "))
			_, err = tx.Exec(userQuery, userValues...)
			if err != nil {
				return err
			}
		}
	}

	// Update the user_profiles table
	profileFields := []string{"updated_at = ?"}
	profileValues := []interface{}{now}
	// Map each field from the form to the database field
	fieldMappings := map[string]string{
		"username":     "username",
		"fullName":     "full_name",
		"bio":          "bio",
		"work":         "work",
		"education":    "education",
		"email":        "email", // This is contact email, not authentication email
		"phone":        "phone",
		"website":      "website",
		"location":     "location",
		"techSkills":   "tech_skills",
		"softSkills":   "soft_skills",
		"interests":    "interests",
		"bannerImage":  "banner_image",
		"profileImage": "profile_image",
		"isPrivate":    "is_private",
	}

	// Add each field that exists in the profileData
	for formField, dbField := range fieldMappings {
		if value, exists := profileData[formField]; exists {
			// Special handling for boolean values
			if dbField == "is_private" {
				if boolVal, ok := value.(bool); ok {
					profileFields = append(profileFields, dbField+" = ?")
					profileValues = append(profileValues, boolVal)
				}
				continue
			}

			// Only update string fields when they're non-empty
			if strVal, ok := value.(string); ok && strVal != "" {
				profileFields = append(profileFields, dbField+" = ?")
				profileValues = append(profileValues, strVal)
			}
		}
	}

	// Add WHERE clause value
	profileValues = append(profileValues, userID)

	// Execute the profile update if there are fields to update
	if len(profileFields) > 1 { // More than just updated_at
		profileQuery := fmt.Sprintf("UPDATE user_profiles SET %s WHERE user_id = ?", strings.Join(profileFields, ", "))
		_, err = tx.Exec(profileQuery, profileValues...)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SQLiteRepository implementation of GetUserProfileByID
// SQLiteRepository implementation of GetUserProfileByID
func (r *SQLiteRepository) GetUserProfileByID(userID string) (*UserProfileData, error) {
	// We need to query both the users and user_profiles tables and combine the results
	query := `
	SELECT 
		u.id, u.email, u.first_name, u.last_name, 
		COALESCE(u.nickname, '') as nickname, 
		COALESCE(u.about_me, '') as about_me, 
		COALESCE(u.avatar, '') as avatar, 
		u.is_public, u.created_at, u.updated_at,
		COALESCE(p.username, '') as username, 
		COALESCE(p.full_name, '') as full_name, 
		COALESCE(p.bio, '') as bio, 
		COALESCE(p.work, '') as work, 
		COALESCE(p.education, '') as education, 
		COALESCE(p.email, '') as contact_email, 
		COALESCE(p.phone, '') as phone, 
		COALESCE(p.website, '') as website, 
		COALESCE(p.location, '') as location, 
		COALESCE(p.tech_skills, '') as tech_skills, 
		COALESCE(p.soft_skills, '') as soft_skills, 
		COALESCE(p.interests, '') as interests, 
		COALESCE(p.banner_image, '') as banner_image, 
		COALESCE(p.profile_image, '') as profile_image, 
		COALESCE(p.is_private, 0) as is_private, 
		COALESCE(p.created_at, u.created_at) as profile_created_at, 
		COALESCE(p.updated_at, u.updated_at) as profile_updated_at,
		COALESCE(us.followers_count, 0) AS followers_count,
		COALESCE(us.following_count, 0) AS following_count
	FROM users u
	LEFT JOIN user_stats us ON u.id = us.user_id
	LEFT JOIN user_profiles p ON u.id = p.user_id
	WHERE u.id = ?`

	row := r.db.QueryRow(query, userID)

	var profileData UserProfileData
	var createdAt, updatedAt, profileCreatedAt, profileUpdatedAt string // SQLite returns dates as strings

	err := row.Scan(
		// User fields
		&profileData.ID, &profileData.Email, &profileData.FirstName, &profileData.LastName,
		&profileData.Nickname, &profileData.AboutMe, &profileData.Avatar, &profileData.IsPublic,
		&createdAt, &updatedAt,

		// Profile fields
		&profileData.Username, &profileData.FullName, &profileData.Bio, &profileData.Work,
		&profileData.Education, &profileData.ContactEmail, &profileData.Phone, &profileData.Website,
		&profileData.Location, &profileData.TechSkills, &profileData.SoftSkills, &profileData.Interests,
		&profileData.BannerImage, &profileData.ProfileImage, &profileData.IsPrivate,
		&profileCreatedAt, &profileUpdatedAt, &profileData.FollowersCount, &profileData.FollowingCount, 
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user profile not found")
		}
		return nil, fmt.Errorf("error scanning user profile: %w", err)
	}

	// Parse the timestamps
	parsedCreatedAt, err := time.Parse(time.RFC3339, createdAt)
	if err == nil {
		profileData.CreatedAt = parsedCreatedAt
	}

	parsedUpdatedAt, err := time.Parse(time.RFC3339, updatedAt)
	if err == nil {
		profileData.UpdatedAt = parsedUpdatedAt
	}

	parsedProfileCreatedAt, err := time.Parse(time.RFC3339, profileCreatedAt)
	if err == nil {
		profileData.ProfileCreatedAt = parsedProfileCreatedAt
	}

	parsedProfileUpdatedAt, err := time.Parse(time.RFC3339, profileUpdatedAt)
	if err == nil {
		profileData.ProfileUpdatedAt = parsedProfileUpdatedAt
	}

	return &profileData, nil
}
