package follow

import (
	"database/sql"
	"errors"
	"fmt"

	notifications "github.com/Athooh/social-network/internal/notifcations"
	"github.com/Athooh/social-network/pkg/logger"
	"github.com/Athooh/social-network/pkg/user"
	"github.com/Athooh/social-network/pkg/websocket"
)

// Service defines the follow service interface
type Service interface {
	// Follow/unfollow operations
	FollowUser(followerID, followingID string) (bool, error)
	UnfollowUser(followerID, followingID string) error

	// Follow request operations
	AcceptFollowRequest(followerID, followingID string) error
	DeclineFollowRequest(followerID, followingID string) error
	GetPendingFollowRequests(userID string) ([]*FollowRequestWithUser, error)

	// Status checks
	IsFollowing(followerID, followingID string) (bool, error)

	// Retrieval operations
	GetFollowers(userID string) ([]*FollowerWithUser, error)
	GetFollowing(userID string) ([]*FollowerWithUser, error)

	GetSuggestedFriends(userID string) ([]*SuggestedFriend, error)
}

// FollowRequestWithUser extends FollowRequest with user information
type FollowRequestWithUser struct {
	FollowRequest
	FollowerName   string
	FollowerAvatar string
	MutualFriends  int
}

// FollowerWithUser extends Follower with user information
type FollowerWithUser struct {
	Follower
	UserName   string
	UserAvatar string
	IsOnline   bool
}

// FollowService implements the Service interface
type FollowService struct {
	repo            Repository
	userRepo        user.Repository
	statusRepo      user.StatusRepository
	log             *logger.Logger
	notificationSvc *NotificationService
}

// NewService creates a new follow service
func NewService(repo Repository, userRepo user.Repository, statusRepo user.StatusRepository, notificationRepo notifications.Service, log *logger.Logger, wsHub *websocket.Hub) Service {
	notificationSvc := NewNotificationService(wsHub, userRepo, notificationRepo, log)

	return &FollowService{
		repo:            repo,
		userRepo:        userRepo,
		statusRepo:      statusRepo,
		log:             log,
		notificationSvc: notificationSvc,
	}
}

// FollowUser handles the logic for a user to follow another user
func (s *FollowService) FollowUser(followerID, followingID string) (bool, error) {
	// Check if users are the same
	if followerID == followingID {
		return false, errors.New("you cannot follow yourself")
	}

	// Check אם already following
	isFollowing, err := s.repo.IsFollowing(followerID, followingID)
	if err != nil {
		return false, err
	}

	if isFollowing {
		return false, errors.New("already following this user")
	}

	// Check if the target user's profile is public
	isPublic, err := s.repo.IsUserProfilePublic(followingID)
	if err != nil {
		return false, err
	}

	// For public profiles, create follower relationship directly
	if isPublic {
		if err := s.repo.CreateFollower(followerID, followingID); err != nil {
			return false, err
		}

		// Update follower counts
		s.notificationSvc.UpdateFollowerCounts(followerID, followingID, s.repo)

		// Send notification to the user being followed
		s.notificationSvc.SendFollowNotification(followerID, followingID, true)

		return true, nil
	}

	// For private profiles, create a follow request
	// Check if a request already exists
	existingRequest, err := s.repo.GetFollowRequest(followerID, followingID)
	if err != nil {
		return false, err
	}

	// Get follower info for notification
	follower, err := s.userRepo.GetByID(followerID)
	if err != nil {
		s.log.Warn("Failed to get follower info for notification: %v", err)
		return false, err
	}
	followerName := fmt.Sprintf("%s %s", follower.FirstName, follower.LastName)

	// Construct notification
	newNotification := &notifications.NewNotification{
		UserId:          followingID,
		SenderId:        sql.NullString{String: followerID, Valid: true},
		NotficationType: "friendRequest",
		Message:         fmt.Sprintf("Friend request from %s", followerName),
	}

	if existingRequest != nil {
		if existingRequest.Status == string(StatusPending) {
			return false, errors.New("follow request already pending")
		}

		// Update existing request back to pending
		if err := s.repo.UpdateFollowRequestStatus(followerID, followingID, string(StatusPending)); err != nil {
			return false, err
		}

		// Send follow request notification with constructed notification
		s.notificationSvc.SendFollowRequestNotification(followerID, followingID, newNotification, follower)
	} else {
		// Create new follow request
		if err := s.repo.CreateFollowRequest(followerID, followingID); err != nil {
			return false, err
		}

		// Send follow request notification with constructed notification
		s.notificationSvc.SendFollowRequestNotification(followerID, followingID, newNotification, follower)
	}

	return false, nil
}

// UnfollowUser handles the logic for a user unfollowing another user
func (s *FollowService) UnfollowUser(followerID, followingID string) error {
	// Check if users are the same
	if followerID == followingID {
		return errors.New("you cannot unfollow yourself")
	}

	// Check if actually following
	isFollowing, err := s.repo.IsFollowing(followerID, followingID)
	if err != nil {
		return err
	}

	if !isFollowing {
		return errors.New("not following this user")
	}

	// Remove follower relationship
	if err := s.repo.DeleteFollower(followerID, followingID); err != nil {
		return err
	}

	// Update follower counts
	s.notificationSvc.UpdateFollowerCounts(followerID, followingID, s.repo)

	// Send notification
	// s.notificationSvc.SendFollowNotification(followerID, followingID, false)

	return nil
}

// AcceptFollowRequest accepts a pending follow request
func (s *FollowService) AcceptFollowRequest(followerID, followingID string) error {
	if followerID == "" || followingID == "" {
		return errors.New("follower or following id is required")
	}
	// Verify the request exists and is pending
	request, err := s.repo.GetFollowRequest(followerID, followingID)
	if err != nil {
		return err
	}

	if request == nil {
		return errors.New("follow request not found")
	}

	if request.Status != string(StatusPending) {
		return errors.New("follow request is not pending")
	}

	// Update request status
	if err := s.repo.UpdateFollowRequestStatus(followerID, followingID, string(StatusAccepted)); err != nil {
		return err
	}

	// Create follower relationship
	if err := s.repo.CreateFollower(followerID, followingID); err != nil {
		return err
	}

	// Update follower counts
	s.notificationSvc.UpdateFollowerCounts(followerID, followingID, s.repo)

	// Send notification to the follower that their request was accepted
	// s.notificationSvc.SendFollowRequestAcceptedNotification(followerID, followingID)

	return nil
}

// DeclineFollowRequest declines a pending follow request
func (s *FollowService) DeclineFollowRequest(followerID, followingID string) error {
	// Verify the request exists and is pending
	request, err := s.repo.GetFollowRequest(followerID, followingID)
	if err != nil {
		return err
	}

	if request == nil {
		return errors.New("follow request not found")
	}

	if request.Status != string(StatusPending) {
		return errors.New("follow request is not pending")
	}

	// Update request status
	return s.repo.UpdateFollowRequestStatus(followerID, followingID, string(StatusDeclined))
}

// GetPendingFollowRequests retrieves all pending follow requests for a user with mutual friends count
func (s *FollowService) GetPendingFollowRequests(userID string) ([]*FollowRequestWithUser, error) {
	requests, err := s.repo.GetPendingFollowRequests(userID)
	if err != nil {
		return nil, err
	}

	var requestsWithUser []*FollowRequestWithUser
	for _, request := range requests {
		// Get follower user info
		follower, err := s.userRepo.GetByID(request.FollowerID)
		if err != nil {
			s.log.Warn("Failed to get follower info: %v", err)
			continue
		}

		// Get mutual followers count
		mutualCount, err := s.repo.GetMutualFollowersCount(userID, request.FollowerID)
		if err != nil {
			s.log.Warn("Failed to get mutual followers count: %v", err)
			mutualCount = 0
		}

		requestWithUser := &FollowRequestWithUser{
			FollowRequest:  *request,
			FollowerName:   fmt.Sprintf("%s %s", follower.FirstName, follower.LastName),
			FollowerAvatar: follower.Avatar,
			MutualFriends:  mutualCount,
		}

		requestsWithUser = append(requestsWithUser, requestWithUser)
	}

	return requestsWithUser, nil
}

// IsFollowing checks if a user is following another user
func (s *FollowService) IsFollowing(followerID, followingID string) (bool, error) {
	return s.repo.IsFollowing(followerID, followingID)
}

// GetFollowers retrieves all followers of a user with user information
func (s *FollowService) GetFollowers(userID string) ([]*FollowerWithUser, error) {
	followers, err := s.repo.GetFollowers(userID)
	if err != nil {
		return nil, err
	}

	var followersWithUser []*FollowerWithUser
	for _, follower := range followers {
		// Get follower user info
		user, err := s.userRepo.GetByID(follower.FollowerID)
		if err != nil {
			s.log.Warn("Failed to get follower info: %v", err)
			continue
		}

		// Check if the user is online
		isOnline, err := s.statusRepo.GetUserStatus(follower.FollowerID)
		if err != nil {
			s.log.Warn("Failed to get online status for user %s: %v", follower.FollowerID, err)
			// Continue with isOnline = false as default
		}

		followerWithUser := &FollowerWithUser{
			Follower:   *follower,
			UserName:   fmt.Sprintf("%s %s", user.FirstName, user.LastName),
			UserAvatar: user.Avatar,
			IsOnline:   isOnline,
		}

		followersWithUser = append(followersWithUser, followerWithUser)
	}

	return followersWithUser, nil
}

// GetFollowing retrieves all users a user is following with user information
func (s *FollowService) GetFollowing(userID string) ([]*FollowerWithUser, error) {
	following, err := s.repo.GetFollowing(userID)
	if err != nil {
		return nil, err
	}

	var followingWithUser []*FollowerWithUser
	for _, follow := range following {
		// Get following user info
		user, err := s.userRepo.GetByID(follow.FollowingID)
		if err != nil {
			s.log.Warn("Failed to get following user info: %v", err)
			continue
		}

		// Check if the user is online
		isOnline, err := s.statusRepo.GetUserStatus(follow.FollowingID)
		if err != nil {
			s.log.Warn("Failed to get online status for user %s: %v", follow.FollowingID, err)
			// Continue with isOnline = false as default
		}

		followWithUser := &FollowerWithUser{
			Follower:   *follow,
			UserName:   fmt.Sprintf("%s %s", user.FirstName, user.LastName),
			UserAvatar: user.Avatar,
			IsOnline:   isOnline,
		}

		followingWithUser = append(followingWithUser, followWithUser)
	}

	return followingWithUser, nil
}

func (s *FollowService) GetSuggestedFriends(userID string) ([]*SuggestedFriend, error) {
	// Get users not followed by current user
	suggestions, err := s.repo.GetUsersNotFollowed(userID)
	if err != nil {
		return nil, err
	}

	var suggestedFriends []*SuggestedFriend
	for _, suggestion := range suggestions {
		// Get user info
		user, err := s.userRepo.GetByID(suggestion.ID)
		if err != nil {
			s.log.Warn("Failed to get user info for suggestion %s: %v", suggestion.ID, err)
			continue
		}

		// Get mutual friends count
		mutualCount, err := s.repo.GetMutualFollowersCount(userID, suggestion.ID)
		if err != nil {
			s.log.Warn("Failed to get mutual followers count for user %s: %v", suggestion.ID, err)
			mutualCount = 0
		}

		// Check if user is online
		isOnline, err := s.statusRepo.GetUserStatus(suggestion.ID)
		if err != nil {
			s.log.Warn("Failed to get online status for user %s: %v", suggestion.ID, err)
			// Continue with isOnline = false as default
		}
		suggestedFriend := &SuggestedFriend{
			ID:            user.ID,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			Nickname:      user.Nickname,
			Avatar:        user.Avatar,
			IsPublic:      user.IsPublic,
			MutualFriends: mutualCount,
			IsOnline:      isOnline,
		}

		suggestedFriends = append(suggestedFriends, suggestedFriend)
	}
	return suggestedFriends, nil
}
