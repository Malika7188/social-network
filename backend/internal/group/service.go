package group

import (
	"database/sql"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	notifications "github.com/Athooh/social-network/internal/notifcations"
	"github.com/Athooh/social-network/pkg/filestore"
	"github.com/Athooh/social-network/pkg/logger"
	models "github.com/Athooh/social-network/pkg/models/dbTables"
	"github.com/Athooh/social-network/pkg/websocket"
	"github.com/google/uuid"
)

// Service defines the interface for group business logic
type Service interface {
	// Group operations
	CreateGroup(userID, name, description string, isPublic bool, banner, profilePic *multipart.FileHeader) (*models.Group, error)
	GetGroup(id, userID string) (*models.Group, error)
	GetUserGroups(userID, viewerID string) ([]*models.Group, error)
	GetAllGroups(userID string, limit, offset int) ([]*models.Group, error)
	UpdateGroup(id, userID, name, description string, isPublic bool, banner, profilePic *multipart.FileHeader) (*models.Group, error)
	DeleteGroup(id, userID string) error

	// Group membership operations
	InviteToGroup(groupID, inviterID, inviteeID string) error
	JoinGroup(groupID, userID string) error
	LeaveGroup(groupID, userID string) error
	AcceptInvitation(groupID, userID string) error
	RejectInvitation(groupID, userID string) error
	AcceptJoinRequest(groupID, adminID, userID string) error
	RejectJoinRequest(groupID, adminID, userID string) error
	UpdateMemberRole(groupID, adminID, userID, role string) error
	RemoveMember(groupID, adminID, userID string) error
	GetGroupMembers(groupID, userID string, status string) ([]*models.GroupMember, error)

	// Group posts operations
	CreateGroupPost(groupID, userID, content string, image, video *multipart.FileHeader) (*models.GroupPost, error)
	GetGroupPosts(groupID, userID string, limit, offset int) ([]*models.GroupPost, error)
	DeleteGroupPost(postID int64, userID string) error

	// Group chat operations
	SendChatMessage(groupID, userID, content string) (*models.GroupChatMessage, error)
	GetGroupChatMessages(groupID, userID string, limit, offset int) ([]*models.GroupChatMessage, error)
}

// GroupService implements the Service interface
type GroupService struct {
	repo          Repository
	fileStore     *filestore.FileStore
	log           *logger.Logger
	wsHub         *websocket.Hub
	notifications *Notifications
}

// NewService creates a new group service
func NewService(repo Repository, fileStore *filestore.FileStore, log *logger.Logger, wsHub *websocket.Hub, notificationRepo notifications.Service) *GroupService {
	notifications := NewNotifications(repo, wsHub, log, notificationRepo)

	return &GroupService{
		repo:          repo,
		fileStore:     fileStore,
		log:           log,
		wsHub:         wsHub,
		notifications: notifications,
	}
}

// CreateGroup creates a new group
func (s *GroupService) CreateGroup(userID, name, description string, isPublic bool, banner, profilePic *multipart.FileHeader) (*models.Group, error) {
	if name == "" {
		return nil, errors.New("group name is required")
	}

	group := &models.Group{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatorID:   userID,
		IsPublic:    isPublic,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Handle banner upload if provided
	if banner != nil {
		bannerPath, err := s.fileStore.SaveFile(banner, "group_banners")
		if err != nil {
			return nil, fmt.Errorf("failed to save banner: %w", err)
		}
		group.BannerPath.String = bannerPath
		group.BannerPath.Valid = true
	}

	// Handle profile pic upload if provided
	if profilePic != nil {
		profilePicPath, err := s.fileStore.SaveFile(profilePic, "group_profiles")
		if err != nil {
			return nil, fmt.Errorf("failed to save profile picture: %w", err)
		}
		group.ProfilePicPath.String = profilePicPath
		group.ProfilePicPath.Valid = true
	}

	// Create the group
	if err := s.repo.CreateGroup(group); err != nil {
		return nil, err
	}

	// Get the complete group with member count and creator info
	createdGroup, err := s.repo.GetGroupByID(group.ID)
	if err != nil {
		return nil, err
	}

	// Notify about group creation
	s.notifyGroupCreated(createdGroup, userID)

	return createdGroup, nil
}

// GetGroup gets a group by ID
func (s *GroupService) GetGroup(id, userID string) (*models.Group, error) {
	group, err := s.repo.GetGroupByID(id)
	if err != nil {
		return nil, err
	}

	// Check if user is a member
	isMember, err := s.repo.IsGroupMember(id, userID)
	if err != nil {
		return nil, err
	}
	group.IsMember = isMember

	// If not a member and group is not public, deny access
	if !group.IsPublic && !isMember && group.CreatorID != userID {
		return nil, errors.New("access denied: private group")
	}

	// Get member status if not a member
	if !isMember {
		member, err := s.repo.GetMemberByID(id, userID)
		if err != nil {
			return nil, err
		}
		if member != nil {
			group.MemberStatus = member.Status
		}
	}

	return group, nil
}

// GetUserGroups gets all groups a user is a member of
func (s *GroupService) GetUserGroups(userID, viewerID string) ([]*models.Group, error) {
	return s.repo.GetUserGroups(userID, viewerID)
}

// GetAllGroups gets all public groups and private groups the user is a member of
func (s *GroupService) GetAllGroups(userID string, limit, offset int) ([]*models.Group, error) {
	group, err := s.repo.GetAllGroups(userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return group, nil
}

// UpdateGroup updates a group's information
func (s *GroupService) UpdateGroup(id, userID, name, description string, isPublic bool, banner, profilePic *multipart.FileHeader) (*models.Group, error) {
	// Check if user is admin or creator
	role, err := s.repo.GetMemberRole(id, userID)
	if err != nil {
		return nil, err
	}

	if role != "admin" {
		return nil, errors.New("only group admins can update the group")
	}

	// Get current group
	group, err := s.repo.GetGroupByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields
	group.Name = name
	group.Description = description
	group.IsPublic = isPublic
	group.UpdatedAt = time.Now()

	// Handle banner upload if provided
	if banner != nil {
		bannerPath, err := s.fileStore.SaveFile(banner, "group_banners")
		if err != nil {
			return nil, fmt.Errorf("failed to save banner: %w", err)
		}
		group.BannerPath.String = bannerPath
		group.BannerPath.Valid = true
	}

	// Handle profile pic upload if provided
	if profilePic != nil {
		profilePicPath, err := s.fileStore.SaveFile(profilePic, "group_profiles")
		if err != nil {
			return nil, fmt.Errorf("failed to save profile picture: %w", err)
		}
		group.ProfilePicPath.String = profilePicPath
		group.ProfilePicPath.Valid = true
	}

	// Update the group
	if err := s.repo.UpdateGroup(group); err != nil {
		return nil, err
	}

	// Notify about group update
	s.notifyGroupUpdated(group, userID)

	return group, nil
}

// DeleteGroup deletes a group
func (s *GroupService) DeleteGroup(id, userID string) error {
	// Check if user is the creator
	group, err := s.repo.GetGroupByID(id)
	if err != nil {
		return err
	}

	if group.CreatorID != userID {
		return errors.New("only the group creator can delete the group")
	}

	// Delete the group
	if err := s.repo.DeleteGroup(id); err != nil {
		return err
	}

	// Notify about group deletion
	s.notifyGroupDeleted(group, userID)

	return nil
}

// InviteToGroup invites a user to a group
func (s *GroupService) InviteToGroup(groupID, inviterID, inviteeID string) error {
	// Check if inviter is a member
	isMember, err := s.repo.IsGroupMember(groupID, inviterID)
	if err != nil {
		return err
	}

	if !isMember {
		return errors.New("only group members can invite others")
	}

	// Check if invitee is already a member or has a pending invitation
	existingMember, err := s.repo.GetMemberByID(groupID, inviteeID)
	if err != nil {
		return err
	}

	if existingMember != nil {
		if existingMember.Status == "accepted" {
			return errors.New("user is already a member of this group")
		} else if existingMember.Status == "pending" && existingMember.InvitedBy != "" {
			return errors.New("user already has a pending invitation")
		}
	}

	// Create member with pending status
	member := &models.GroupMember{
		GroupID:   groupID,
		UserID:    inviteeID,
		Role:      "member",
		Status:    "pending",
		InvitedBy: inviterID,
	}

	if err := s.repo.AddMember(member); err != nil {
		return err
	}

	// Get group info for notification
	group, err := s.repo.GetGroupByID(groupID)
	if err != nil {
		return err
	}
	newNotification := &notifications.NewNotification{
		UserId:          inviteeID,
		SenderId:        sql.NullString{String: inviterID, Valid: true},
		NotficationType: "invitation",
		Message:         fmt.Sprintf("Invite to group %s", group.Name),
		TargetGroupID:   sql.NullString{String: groupID, Valid: true},
	}
	inviterInfo, err := s.repo.GetUserBasicByID(inviterID)
	if err != nil {
		return err
	}
	// Notify invitee
	s.notifications.NotifyGroupInvitation(group, inviterID, inviteeID, newNotification, inviterInfo)

	return nil
}

// JoinGroup sends a request to join a group
func (s *GroupService) JoinGroup(groupID, userID string) error {
	// Check if user is already a member or has a pending request
	existingMember, err := s.repo.GetMemberByID(groupID, userID)
	if err != nil {
		return err
	}

	if existingMember != nil {
		if existingMember.Status == "accepted" {
			return errors.New("you are already a member of this group")
		} else if existingMember.Status == "pending" && existingMember.InvitedBy == "" {
			return errors.New("you already have a pending join request")
		} else if existingMember.Status == "pending" && existingMember.InvitedBy != "" {
			return errors.New("you have a pending invitation to this group")
		}
	}

	// Get group to check if it's public
	group, err := s.repo.GetGroupByID(groupID)
	if err != nil {
		return err
	}

	// If group is public, automatically accept
	status := "pending"

	// Create member
	member := &models.GroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    "member",
		Status:  status,
	}

	if err := s.repo.AddMember(member); err != nil {
		return err
	}

	newNotification := &notifications.NewNotification{
		UserId:          group.CreatorID,
		SenderId:        sql.NullString{String: userID, Valid: true},
		NotficationType: "joinRequest",
		Message:         fmt.Sprintf("Wants to join %s", group.Name),
		TargetGroupID:   sql.NullString{String: groupID, Valid: true},
	}
	requesterInfo, err := s.repo.GetUserBasicByID(userID)
	if err != nil {
		return err
	}

	s.notifications.NotifyGroupJoinRequest(group, userID, group.CreatorID, newNotification, requesterInfo)

	return nil
}

// LeaveGroup allows a user to leave a group
func (s *GroupService) LeaveGroup(groupID, userID string) error {
	// Check if user is a member
	isMember, err := s.repo.IsGroupMember(groupID, userID)
	if err != nil {
		return err
	}

	if !isMember {
		return errors.New("you are not a member of this group")
	}

	// Check if user is the creator
	group, err := s.repo.GetGroupByID(groupID)
	if err != nil {
		return err
	}

	if group.CreatorID == userID {
		return errors.New("the group creator cannot leave the group, please delete the group or transfer ownership")
	}

	// Remove the member
	if err := s.repo.RemoveMember(groupID, userID); err != nil {
		return err
	}

	// Notify about member leaving
	s.notifyGroupMemberLeft(group, userID)

	return nil
}

// AcceptInvitation accepts a group invitation
func (s *GroupService) AcceptInvitation(groupID, userID string) error {
	// Check if invitation exists
	member, err := s.repo.GetMemberByID(groupID, userID)
	if err != nil {
		return err
	}

	if member == nil {
		return errors.New("invitation not found")
	}

	if member.Status != "pending" || member.InvitedBy == "" {
		return errors.New("no pending invitation found")
	}

	// Update status to accepted
	if err := s.repo.UpdateMemberStatus(groupID, userID, "accepted"); err != nil {
		return err
	}

	// Get group info for notification
	group, err := s.repo.GetGroupByID(groupID)
	if err != nil {
		return err
	}

	// Notify about invitation acceptance
	s.notifyGroupInvitationAccepted(group, userID)

	return nil
}

// RejectInvitation rejects a group invitation
func (s *GroupService) RejectInvitation(groupID, userID string) error {
	// Check if invitation exists
	member, err := s.repo.GetMemberByID(groupID, userID)
	if err != nil {
		return err
	}

	if member == nil {
		return errors.New("invitation not found")
	}

	if member.Status != "pending" || member.InvitedBy == "" {
		return errors.New("no pending invitation found")
	}

	// Remove member record
	if err := s.repo.RemoveMember(groupID, userID); err != nil {
		return err
	}

	// Get group info for notification
	group, err := s.repo.GetGroupByID(groupID)
	if err != nil {
		return err
	}

	// Notify about invitation rejection
	s.notifyGroupInvitationRejected(group, userID, member.InvitedBy)

	return nil
}

// AcceptJoinRequest accepts a request to join a group
func (s *GroupService) AcceptJoinRequest(groupID, adminID, userID string) error {
	// Check if admin is an admin
	role, err := s.repo.GetMemberRole(groupID, adminID)
	if err != nil {
		return err
	}

	if role != "admin" {
		return errors.New("only group admins can accept join requests")
	}

	// Check if join request exists
	member, err := s.repo.GetMemberByID(groupID, userID)
	if err != nil {
		return err
	}

	if member == nil {
		return errors.New("join request not found")
	}

	fmt.Println(member.Status, member.InvitedBy)
	if member.Status != "pending" || member.InvitedBy != "" {
		return errors.New("no pending join request found")
	}

	// Update status to accepted
	if err := s.repo.UpdateMemberStatus(groupID, userID, "accepted"); err != nil {
		return err
	}

	// Get group info for notification
	group, err := s.repo.GetGroupByID(groupID)
	if err != nil {
		return err
	}

	// Notify about join request acceptance
	s.notifyGroupJoinRequestAccepted(group, userID, adminID)

	return nil
}

// RejectJoinRequest rejects a request to join a group
func (s *GroupService) RejectJoinRequest(groupID, adminID, userID string) error {
	// Check if admin is an admin
	role, err := s.repo.GetMemberRole(groupID, adminID)
	if err != nil {
		return err
	}

	if role != "admin" {
		return errors.New("only group admins can reject join requests")
	}

	// Check if join request exists
	member, err := s.repo.GetMemberByID(groupID, userID)
	if err != nil {
		return err
	}

	if member == nil {
		return errors.New("join request not found")
	}

	if member.Status != "pending" || member.InvitedBy != "" {
		return errors.New("no pending join request found")
	}

	// Remove member record
	if err := s.repo.RemoveMember(groupID, userID); err != nil {
		return err
	}

	// Get group info for notification
	group, err := s.repo.GetGroupByID(groupID)
	if err != nil {
		return err
	}

	// Notify about join request rejection
	s.notifyGroupJoinRequestRejected(group, userID, adminID)

	return nil
}

// UpdateMemberRole updates a member's role
func (s *GroupService) UpdateMemberRole(groupID, adminID, userID, role string) error {
	// Check if admin is an admin
	adminRole, err := s.repo.GetMemberRole(groupID, adminID)
	if err != nil {
		return err
	}

	if adminRole != "admin" {
		return errors.New("only group admins can update member roles")
	}

	// Check if user is a member
	isMember, err := s.repo.IsGroupMember(groupID, userID)
	if err != nil {
		return err
	}

	if !isMember {
		return errors.New("user is not a member of this group")
	}

	// Get group to check if user is the creator
	group, err := s.repo.GetGroupByID(groupID)
	if err != nil {
		return err
	}

	if group.CreatorID == userID {
		return errors.New("cannot change the role of the group creator")
	}

	// Validate role
	if role != "admin" && role != "moderator" && role != "member" {
		return errors.New("invalid role")
	}

	// Update role
	if err := s.repo.UpdateMemberRole(groupID, userID, role); err != nil {
		return err
	}

	// Notify about role update
	s.notifyGroupMemberRoleUpdated(group, userID, role, adminID)

	return nil
}

// RemoveMember removes a member from a group
func (s *GroupService) RemoveMember(groupID, adminID, userID string) error {
	// Check if admin is an admin
	adminRole, err := s.repo.GetMemberRole(groupID, adminID)
	if err != nil {
		return err
	}

	if adminRole != "admin" {
		return errors.New("only group admins can remove members")
	}

	// Check if user is a member
	isMember, err := s.repo.IsGroupMember(groupID, userID)
	if err != nil {
		return err
	}

	if !isMember {
		return errors.New("user is not a member of this group")
	}

	// Get group to check if user is the creator
	group, err := s.repo.GetGroupByID(groupID)
	if err != nil {
		return err
	}

	if group.CreatorID == userID {
		return errors.New("cannot remove the group creator")
	}

	// Remove member
	if err := s.repo.RemoveMember(groupID, userID); err != nil {
		return err
	}

	// Notify about member removal
	s.notifyGroupMemberRemoved(group, userID, adminID)

	return nil
}

// GetGroupMembers gets all members of a group
func (s *GroupService) GetGroupMembers(groupID, userID string, status string) ([]*models.GroupMember, error) {
	// Check if user is a member or admin
	isMember, err := s.repo.IsGroupMember(groupID, userID)
	if err != nil {
		return nil, err
	}

	role, err := s.repo.GetMemberRole(groupID, userID)
	if err != nil {
		return nil, err
	}

	// Only admins can see pending members
	if status == "pending" && role != "admin" {
		return nil, errors.New("only group admins can view pending members")
	}

	// Non-members can't view member list
	if !isMember {
		return nil, errors.New("only group members can view the member list")
	}

	// Get members
	members, err := s.repo.GetGroupMembers(groupID, status)
	if err != nil {
		return nil, err
	}

	return members, nil
}

// CreateGroupPost creates a new post in a group
func (s *GroupService) CreateGroupPost(groupID, userID, content string, image, video *multipart.FileHeader) (*models.GroupPost, error) {
	// Check if user is a member
	isMember, err := s.repo.IsGroupMember(groupID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, errors.New("only group members can create posts")
	}

	post := &models.GroupPost{
		GroupID:   groupID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Handle image upload if provided
	if image != nil {
		imagePath, err := s.fileStore.SaveFile(image, "group_post_images")
		if err != nil {
			return nil, fmt.Errorf("failed to save image: %w", err)
		}
		post.ImagePath.String = imagePath
		post.ImagePath.Valid = true
	}

	// Handle video upload if provided
	if video != nil {
		videoPath, err := s.fileStore.SaveFile(video, "group_post_videos")
		if err != nil {
			return nil, fmt.Errorf("failed to save video: %w", err)
		}
		post.VideoPath.String = videoPath
		post.VideoPath.Valid = true
	}

	// Create the post
	if err := s.repo.CreateGroupPost(post); err != nil {
		return nil, err
	}

	// Get user data
	user, err := s.repo.GetUserBasicByID(userID)
	if err != nil {
		return nil, err
	}
	post.User = &models.PostUserData{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Avatar:    user.Avatar,
	}

	// Notify about post creation
	s.notifyGroupPostCreated(post)

	return post, nil
}

// GetGroupPosts gets all posts in a group
func (s *GroupService) GetGroupPosts(groupID, userID string, limit, offset int) ([]*models.GroupPost, error) {
	// Check if user is a member
	isMember, err := s.repo.IsGroupMember(groupID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, errors.New("only group members can view posts")
	}

	// Get posts
	posts, err := s.repo.GetGroupPosts(groupID, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Get user data for each post
	for _, post := range posts {
		user, err := s.repo.GetUserBasicByID(post.UserID)
		if err != nil {
			return nil, err
		}
		post.User = &models.PostUserData{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Avatar:    user.Avatar,
		}
	}

	return posts, nil
}

// DeleteGroupPost deletes a post from a group
func (s *GroupService) DeleteGroupPost(postID int64, userID string) error {
	// Get the post
	post, err := s.repo.GetGroupPostByID(postID)
	if err != nil {
		return err
	}

	if post == nil {
		return errors.New("post not found")
	}

	// Check if user is the post creator or an admin
	if post.UserID != userID {
		role, err := s.repo.GetMemberRole(post.GroupID, userID)
		if err != nil {
			return err
		}

		if role != "admin" && role != "moderator" {
			return errors.New("only the post creator, moderators, or admins can delete posts")
		}
	}

	// Delete the post
	if err := s.repo.DeleteGroupPost(postID); err != nil {
		return err
	}

	return nil
}

// SendChatMessage sends a message to a group chat
func (s *GroupService) SendChatMessage(groupID, userID, content string) (*models.GroupChatMessage, error) {
	// Check if user is a member
	isMember, err := s.repo.IsGroupMember(groupID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, errors.New("only group members can send messages")
	}

	if content == "" {
		return nil, errors.New("message content is required")
	}

	// Create message
	message := &models.GroupChatMessage{
		GroupID:   groupID,
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	// Add message to database
	if err := s.repo.AddChatMessage(message); err != nil {
		return nil, err
	}

	// Get user info
	user, err := s.repo.GetUserBasicByID(userID)
	if err != nil {
		return nil, err
	}
	message.User = user

	// Notify about new message
	s.notifyGroupChatMessage(message)

	return message, nil
}

// GetGroupChatMessages gets messages from a group chat
func (s *GroupService) GetGroupChatMessages(groupID, userID string, limit, offset int) ([]*models.GroupChatMessage, error) {
	// Check if user is a member
	isMember, err := s.repo.IsGroupMember(groupID, userID)
	if err != nil {
		return nil, err
	}

	if !isMember {
		return nil, errors.New("only group members can view messages")
	}

	// Get messages
	messages, err := s.repo.GetGroupChatMessages(groupID, limit, offset)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// notifyGroupCreated notifies about group creation
func (s *GroupService) notifyGroupCreated(group *models.Group, userID string) {
	s.notifications.NotifyGroupCreated(group, userID)
}

// notifyGroupUpdated notifies about group update
func (s *GroupService) notifyGroupUpdated(group *models.Group, userID string) {
	s.notifications.NotifyGroupUpdated(group, userID)
}

// notifyGroupDeleted notifies about group deletion
func (s *GroupService) notifyGroupDeleted(group *models.Group, userID string) {
	s.notifications.NotifyGroupDeleted(group, userID)
}

// notifyGroupInvitationAccepted notifies about group invitation acceptance
func (s *GroupService) notifyGroupInvitationAccepted(group *models.Group, userID string) {
	s.notifications.NotifyGroupInvitationAccepted(group, userID)
}

// notifyGroupJoinRequest notifies about group join request
// func (s *GroupService) notifyGroupJoinRequest(group *models.Group, userID string) {
// 	s.notifications.NotifyGroupJoinRequest(group, userID)
// }

// notifyGroupJoinRequestAccepted notifies about group join request acceptance
func (s *GroupService) notifyGroupJoinRequestAccepted(group *models.Group, userID, adminID string) {
	s.notifications.NotifyGroupJoinRequestAccepted(group, userID, adminID)
}

// notifyGroupMemberLeft notifies about group member leaving
func (s *GroupService) notifyGroupMemberLeft(group *models.Group, userID string) {
	s.notifications.NotifyGroupMemberLeft(group, userID)
}

// notifyGroupMemberRoleUpdated notifies about group member role update
func (s *GroupService) notifyGroupMemberRoleUpdated(group *models.Group, userID, role, adminID string) {
	s.notifications.NotifyGroupMemberRoleUpdated(group, userID, role, adminID)
}

// notifyGroupMemberRemoved notifies about group member removal
func (s *GroupService) notifyGroupMemberRemoved(group *models.Group, userID, adminID string) {
	s.notifications.NotifyGroupMemberRemoved(group, userID, adminID)
}

// notifyGroupPostCreated notifies about group post creation
func (s *GroupService) notifyGroupPostCreated(post *models.GroupPost) {
	s.notifications.NotifyGroupPostCreated(post)
}

// notifyGroupChatMessage notifies about group chat message
func (s *GroupService) notifyGroupChatMessage(message *models.GroupChatMessage) {
	s.notifications.NotifyGroupChatMessage(message)
}

// notifyGroupJoinRequestRejected sends a notification when a join request is rejected
func (s *GroupService) notifyGroupJoinRequestRejected(group *models.Group, userID, adminID string) {
	s.notifications.NotifyGroupJoinRequestRejected(group, userID, adminID)
}

// notifyGroupInvitationRejected sends a notification when a group invitation is rejected
func (s *GroupService) notifyGroupInvitationRejected(group *models.Group, userID, inviterID string) {
	s.notifications.NotifyGroupInvitationRejected(group, userID, inviterID)
}
