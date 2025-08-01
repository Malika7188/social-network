package profile

import (
	"mime/multipart"
	"net/http"

	"github.com/Athooh/social-network/internal/auth"
	httputil "github.com/Athooh/social-network/pkg/httputil"
	"github.com/Athooh/social-network/pkg/logger"
)

// Handler handles HTTP requests for posts
type Handler struct {
	service   Service
	log       *logger.Logger
	uploadDir string
}

// ProfileUpdateRequest represents the request to update a profile
type ProfileUpdateRequest struct {
	Username   string `json:"username"`
	FullName   string `json:"fullName"`
	Bio        string `json:"bio"`
	Work       string `json:"work"`
	Education  string `json:"education"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Website    string `json:"website"`
	Location   string `json:"location"`
	TechSkills string `json:"techSkills"`
	SoftSkills string `json:"softSkills"`
	Interests  string `json:"interests"`
	IsPrivate  bool   `json:"isPrivate"`
}

// NewHandler creates a new chat handler
func NewHandler(service Service, log *logger.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

func (h *Handler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	h.log.Info("Received profile view request")
	if r.Method != http.MethodGet {
		httputil.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", false)
		return
	}
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		httputil.SendError(w, http.StatusUnauthorized, "Unauthorized", true)
		return
	}
	profileID := r.URL.Query().Get("userId")
	if profileID == "" {
		httputil.SendError(w, http.StatusBadRequest, "User ID is required", false)
		return
	}
	shouldView, err := h.service.ValidateProfileViewRequest(userID, profileID)
	if err != nil {
		h.log.Error("Failed to validate view request" + err.Error())
		httputil.SendError(w, http.StatusInternalServerError, "Server error", true)
		return
	}
	if !shouldView {
		httputil.SendError(w, http.StatusForbidden, "Forbidden", true)
		return
	}
	targetProfile, err := h.service.GetProfileByUserID(profileID)
	if err != nil {
		h.log.Error("Failed to fetch target profile" + err.Error())
		httputil.SendError(w, http.StatusInternalServerError, "Server error", true)
		return
	}
	// Return success response with updated profile data
	httputil.SendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Profile fetch successful",
		"profile": targetProfile,
	})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	h.log.Info("Received profile update request")

	if r.Method != http.MethodPut {
		httputil.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", false)
		return
	}
	// Get user ID from context (set by auth middleware)
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID <= "" {
		httputil.SendError(w, http.StatusUnauthorized, "Unauthorized", true)
		return
	}
	err := r.ParseMultipartForm(20 << 20) // 20MB max
	if err != nil {
		h.log.Error("Failed to parse multipart form" + err.Error())
		httputil.SendError(w, http.StatusBadRequest, "Failed to parse form data", false)
		return
	}

	// Extract and validate text fields
	profileUpdate := ProfileUpdateRequest{
		Username:   r.FormValue("username"),
		FullName:   r.FormValue("fullName"),
		Bio:        r.FormValue("bio"),
		Work:       r.FormValue("work"),
		Education:  r.FormValue("education"),
		Email:      r.FormValue("email"),
		Phone:      r.FormValue("phone"),
		Website:    r.FormValue("website"),
		Location:   r.FormValue("location"),
		TechSkills: r.FormValue("techSkills"),
		SoftSkills: r.FormValue("softSkills"),
		Interests:  r.FormValue("interests"),
		IsPrivate:  r.FormValue("isPrivate") == "true",
	}

	var profileImageHeader *multipart.FileHeader
	var profileImagePath string
	if file, header, err := r.FormFile("profileImage"); err == nil {
		defer file.Close()
		profileImageHeader = header
	}

	var bannerImageHeader *multipart.FileHeader
	var bannerImagePath string
	if file, header, err := r.FormFile("bannerImage"); err == nil {
		defer file.Close()
		bannerImageHeader = header
	}

	// Convert the profileUpdate struct to a map for the service
	profileData := map[string]interface{}{
		"username":   profileUpdate.Username,
		"fullName":   profileUpdate.FullName,
		"bio":        profileUpdate.Bio,
		"work":       profileUpdate.Work,
		"education":  profileUpdate.Education,
		"email":      profileUpdate.Email,
		"phone":      profileUpdate.Phone,
		"website":    profileUpdate.Website,
		"location":   profileUpdate.Location,
		"techSkills": profileUpdate.TechSkills,
		"softSkills": profileUpdate.SoftSkills,
		"interests":  profileUpdate.Interests,
		"isPrivate":  profileUpdate.IsPrivate,
	}

	// Handle profile image if provided
	if profileImageHeader != nil {
		profileImagePath, err = h.service.SaveProfileImage(userID, profileImageHeader)
		if err != nil {
			h.log.Error("Failed to save profile image: " + err.Error())
			httputil.SendError(w, http.StatusInternalServerError, "Failed to save profile image", false)
			return
		}

		if profileImagePath != "" {
			profileData["profileImage"] = profileImagePath
		}
	}

	// Handle banner image if provided
	if bannerImageHeader != nil {
		bannerImagePath, err = h.service.SaveBannerImage(userID, bannerImageHeader)
		if err != nil {
			h.log.Error("Failed to save banner image: " + err.Error())
			httputil.SendError(w, http.StatusInternalServerError, "Failed to save banner image", false)
			return
		}

		if bannerImagePath != "" {
			profileData["bannerImage"] = bannerImagePath
		}
	}

	err = h.service.UpdateProfile(userID, profileData)
	if err != nil {
		h.log.Error("Failed to update profile: " + err.Error())
		httputil.SendError(w, http.StatusInternalServerError, "Failed to update profile", false)
		return
	}

	// Fetch the complete updated profile data
	updatedProfile, err := h.service.GetProfileByUserID(userID)
	if err != nil {
		h.log.Error("Failed to get updated profile: " + err.Error())
		// Even if this fails, we still return success since the update worked
		httputil.SendJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Profile updated successfully",
		})
		return
	}

	// Return success response with updated profile data
	httputil.SendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Profile updated successfully",
		"profile": updatedProfile,
	})
}
