package group

import (
	"time"

	notifications "github.com/Athooh/social-network/internal/notifcations"
	"github.com/Athooh/social-network/pkg/logger"
	models "github.com/Athooh/social-network/pkg/models/dbTables"
	"github.com/Athooh/social-network/pkg/websocket"
	"github.com/Athooh/social-network/pkg/websocket/events"
)

// Notifications handles all websocket notifications for group operations
type Notifications struct {
	repo             Repository
	wsHub            *websocket.Hub
	log              *logger.Logger
	notificationRepo notifications.Service
}

// NewNotifications creates a new notifications handler
func NewNotifications(repo Repository, wsHub *websocket.Hub, log *logger.Logger, notifications notifications.Service) *Notifications {
	return &Notifications{
		repo:             repo,
		wsHub:            wsHub,
		log:              log,
		notificationRepo: notifications,
	}
}

// NotifyGroupCreated notifies about group creation
func (n *Notifications) NotifyGroupCreated(group *models.Group, userID string) {
	event := events.Event{
		Type: "group_created",
		Payload: map[string]interface{}{
			"group":  group,
			"userID": userID,
		},
	}

	n.wsHub.BroadcastToUser(userID, event)
}

// NotifyGroupUpdated notifies about group update
func (n *Notifications) NotifyGroupUpdated(group *models.Group, userID string) {
	event := events.Event{
		Type: "group_updated",
		Payload: map[string]interface{}{
			"group":  group,
			"userID": userID,
		},
	}

	// Notify all members
	members, _ := n.repo.GetGroupMembers(group.ID, "accepted")
	for _, member := range members {
		n.wsHub.BroadcastToUser(member.UserID, event)
	}
}

// NotifyGroupDeleted notifies about group deletion
func (n *Notifications) NotifyGroupDeleted(group *models.Group, userID string) {
	event := events.Event{
		Type: "group_deleted",
		Payload: map[string]interface{}{
			"group":  group,
			"userID": userID,
		},
	}

	// Notify all members
	members, _ := n.repo.GetGroupMembers(group.ID, "accepted")
	for _, member := range members {
		n.wsHub.BroadcastToUser(member.UserID, event)
	}
}

// NotifyGroupInvitation notifies about group invitation
func (n *Notifications) NotifyGroupInvitation(group *models.Group, inviterID, inviteeID string, newNote *notifications.NewNotification, inviterInfo *models.UserBasic) {
	if n.wsHub == nil {
		n.log.Warn("WebSocket hub is nil, cannot send follow request notification")
		return
	}

	// Create notification in database
	if err := n.notificationRepo.CreateNotification(newNote); err != nil {
		n.log.Error("Failed to create follow request notification: %v", err)
		return
	}

	// Retrieve the newly created notification to get its ID and CreatedAt
	notifications, err := n.notificationRepo.GetNotifications(inviteeID, 1, 0)
	if err != nil || len(notifications) == 0 {
		n.log.Error("Failed to retrieve newly created notification: %v", err)
		return
	}
	dbNotification := notifications[0]

	// Create WebSocket event
	event := events.Event{
		Type: events.HeaderNotificationUpdate,
		Payload: map[string]interface{}{
			"id":            dbNotification.ID,
			"type":          newNote.NotficationType,
			"senderId":      inviterID,
			"targetGroupId": group.ID,
			"senderName":    inviterInfo.FirstName + " " + inviterInfo.LastName,
			"senderAvatar":  inviterInfo.Avatar,
			"message":       newNote.Message,
			"createdAt":     dbNotification.CreatedAt.Format(time.RFC3339),
			"isRead":        dbNotification.IsRead,
		},
	}

	// Send to the user receiving the follow request
	n.wsHub.BroadcastToUser(inviteeID, event)
}

// NotifyGroupInvitationAccepted notifies about group invitation acceptance
func (n *Notifications) NotifyGroupInvitationAccepted(group *models.Group, userID string) {
	user, _ := n.repo.GetUserBasicByID(userID)

	event := events.Event{
		Type: "group_invitation_accepted",
		Payload: map[string]interface{}{
			"group": group,
			"user":  user,
		},
	}

	// Notify all members
	members, _ := n.repo.GetGroupMembers(group.ID, "accepted")
	for _, member := range members {
		n.wsHub.BroadcastToUser(member.UserID, event)
	}
}

// NotifyGroupJoinRequest notifies about group join request
func (n *Notifications) NotifyGroupJoinRequest(group *models.Group, inviterID, inviteeID string, newNote *notifications.NewNotification, inviterInfo *models.UserBasic) {
	if n.wsHub == nil {
		n.log.Warn("WebSocket hub is nil, cannot send follow request notification")
		return
	}

	// Create notification in database
	if err := n.notificationRepo.CreateNotification(newNote); err != nil {
		n.log.Error("Failed to create follow request notification: %v", err)
		return
	}

	// Retrieve the newly created notification to get its ID and CreatedAt
	notifications, err := n.notificationRepo.GetNotifications(inviteeID, 1, 0)
	if err != nil || len(notifications) == 0 {
		n.log.Error("Failed to retrieve newly created notification: %v", err)
		return
	}
	dbNotification := notifications[0]

	// Create WebSocket event
	event := events.Event{
		Type: events.HeaderNotificationUpdate,
		Payload: map[string]interface{}{
			"id":            dbNotification.ID,
			"type":          newNote.NotficationType,
			"senderId":      inviterID,
			"targetGroupId": group.ID,
			"senderName":    inviterInfo.FirstName + " " + inviterInfo.LastName,
			"senderAvatar":  inviterInfo.Avatar,
			"message":       newNote.Message,
			"createdAt":     dbNotification.CreatedAt.Format(time.RFC3339),
			"isRead":        dbNotification.IsRead,
		},
	}

	// Send to the user receiving the follow request
	n.wsHub.BroadcastToUser(inviteeID, event)
}

// NotifyGroupJoinRequestAccepted notifies about group join request acceptance
func (n *Notifications) NotifyGroupJoinRequestAccepted(group *models.Group, userID, adminID string) {
	user, _ := n.repo.GetUserBasicByID(userID)
	admin, _ := n.repo.GetUserBasicByID(adminID)

	// Notify user
	userEvent := events.Event{
		Type: "group_join_request_accepted",
		Payload: map[string]interface{}{
			"group": group,
			"admin": admin,
		},
	}
	n.wsHub.BroadcastToUser(userID, userEvent)

	// Notify all members
	memberEvent := events.Event{
		Type: "new_group_member",
		Payload: map[string]interface{}{
			"group": group,
			"user":  user,
		},
	}

	members, _ := n.repo.GetGroupMembers(group.ID, "accepted")
	for _, member := range members {
		if member.UserID != userID {
			n.wsHub.BroadcastToUser(member.UserID, memberEvent)
		}
	}
}

// NotifyGroupMemberLeft notifies about group member leaving
func (n *Notifications) NotifyGroupMemberLeft(group *models.Group, userID string) {
	user, _ := n.repo.GetUserBasicByID(userID)

	event := events.Event{
		Type: "group_member_left",
		Payload: map[string]interface{}{
			"group": group,
			"user":  user,
		},
	}

	// Notify all members
	members, _ := n.repo.GetGroupMembers(group.ID, "accepted")
	for _, member := range members {
		n.wsHub.BroadcastToUser(member.UserID, event)
	}
}

// NotifyGroupMemberRoleUpdated notifies about group member role update
func (n *Notifications) NotifyGroupMemberRoleUpdated(group *models.Group, userID, role, adminID string) {
	user, _ := n.repo.GetUserBasicByID(userID)
	admin, _ := n.repo.GetUserBasicByID(adminID)

	event := events.Event{
		Type: "group_member_role_updated",
		Payload: map[string]interface{}{
			"group": group,
			"user":  user,
			"role":  role,
			"admin": admin,
		},
	}

	// Notify all members
	members, _ := n.repo.GetGroupMembers(group.ID, "accepted")
	for _, member := range members {
		n.wsHub.BroadcastToUser(member.UserID, event)
	}
}

// NotifyGroupMemberRemoved notifies about group member removal
func (n *Notifications) NotifyGroupMemberRemoved(group *models.Group, userID, adminID string) {
	user, _ := n.repo.GetUserBasicByID(userID)
	admin, _ := n.repo.GetUserBasicByID(adminID)

	// Notify removed user
	userEvent := events.Event{
		Type: "removed_from_group",
		Payload: map[string]interface{}{
			"group": group,
			"admin": admin,
		},
	}
	n.wsHub.BroadcastToUser(userID, userEvent)

	// Notify all members
	memberEvent := events.Event{
		Type: "group_member_removed",
		Payload: map[string]interface{}{
			"group": group,
			"user":  user,
			"admin": admin,
		},
	}

	members, _ := n.repo.GetGroupMembers(group.ID, "accepted")
	for _, member := range members {
		n.wsHub.BroadcastToUser(member.UserID, memberEvent)
	}
}

// NotifyGroupPostCreated notifies about group post creation
func (n *Notifications) NotifyGroupPostCreated(post *models.GroupPost) {
	event := events.Event{
		Type: "group_post_created",
		Payload: map[string]interface{}{
			"post": post,
		},
	}

	// Notify all members
	members, _ := n.repo.GetGroupMembers(post.GroupID, "accepted")
	for _, member := range members {
		n.wsHub.BroadcastToUser(member.UserID, event)
	}
}

// NotifyGroupChatMessage notifies about a new group chat message
func (n *Notifications) NotifyGroupChatMessage(message *models.GroupChatMessage) {
	event := events.Event{
		Type: events.GroupMessage,
		Payload: map[string]interface{}{
			"id":      message.ID,
			"Content": message.Content,
			"User": map[string]interface{}{
				"id":        message.UserID,
				"firstName": message.User.FirstName,
				"avatar":    message.User.Avatar,
			},
			"CreatedAt": message.CreatedAt,
			"GroupID":   message.GroupID,
		},
	}

	// Notify all members
	members, _ := n.repo.GetGroupMembers(message.GroupID, "accepted")
	for _, member := range members {
		if member.UserID != message.User.ID {
			n.wsHub.BroadcastToUser(member.UserID, event)
		}
	}
}

// NotifyGroupJoinRequestRejected notifies about group join request rejection
func (n *Notifications) NotifyGroupJoinRequestRejected(group *models.Group, userID, adminID string) {
	admin, _ := n.repo.GetUserBasicByID(adminID)

	event := events.Event{
		Type: "group_join_request_rejected",
		Payload: map[string]interface{}{
			"group": group,
			"admin": admin,
		},
	}

	// Notify the rejected user
	n.wsHub.BroadcastToUser(userID, event)
}

// NotifyGroupInvitationRejected notifies about group invitation rejection
func (n *Notifications) NotifyGroupInvitationRejected(group *models.Group, userID, inviterID string) {
	user, _ := n.repo.GetUserBasicByID(userID)

	event := events.Event{
		Type: "group_invitation_rejected",
		Payload: map[string]interface{}{
			"group": group,
			"user":  user,
		},
	}

	// Notify the inviter
	n.wsHub.BroadcastToUser(inviterID, event)
}
