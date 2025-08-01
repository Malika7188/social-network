"use client";

import { useState, useEffect, useRef } from "react";
import styles from "@/styles/NotificationDropdown.module.css";
import { formatDistanceToNow } from "date-fns";
import { useNotificationService } from "@/services/notificationService";
import { useGroupService } from "@/services/groupService";
import { showToast } from "@/components/ui/ToastContainer";

export default function NotificationDropdown() {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef(null);
  const {
    notifications,
    isLoadingNotifications,
    fetchNotifications,
    markNotificationAsRead,
    markAllNotificationsAsRead,
    clearAllNotifications,
    handleFriendRequest,
    DeleteNotification,
  } = useNotificationService();
  const {
    acceptInvitation,
    rejectInvitation,
    acceptJoinRequest,
    rejectJoinRequest,
    respondToEvent
  } = useGroupService();

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Handle friend request or group event response (accept/decline)
  const handleNotificationResponse = async (id, notificationId, action, type) => {
    await handleFriendRequest(id, notificationId, action, type);
    fetchNotifications();
  };

  const handleMarkAllAsRead = async () => {
    await markAllNotificationsAsRead();
    fetchNotifications();
  };

  const handleClearAll = async () => {
    await clearAllNotifications();
    fetchNotifications();
  };

  const handleMarkSingleAsRead = async (notificationId) => {
    await markNotificationAsRead(notificationId);
    fetchNotifications();
  };

  const handleGroupNotificationResponse = async (groupId, notificationId, action) => {
    try {
      // Get user data from localStorage
      const userData = JSON.parse(localStorage.getItem("userData"));
      if (!userData) {
        showToast("User data not found", "error");
        return;
      }

      let success;
      if (action === "accepted") {
        success = await acceptInvitation(groupId, userData.id);
      } else if (action === "rejected") {
        success = await rejectInvitation(groupId, userData.id);
      }

      if (!success) {
        throw new Error(`Failed to ${action} invitation`);
      }

      // Delete the notification after handling
      await DeleteNotification(notificationId);
      fetchNotifications();
      showToast(`Successfully ${action}ed request`, "success");
    } catch (error) {
      console.error(`Error ${action}ing invitation:`, error);
      showToast(`Failed to ${action} invitation`, "error");
    }
  };

  // Add new handler for join requests
  const handleJoinRequestResponse = async (groupId, userId, notificationId, action) => {
    try {

      const userData = JSON.parse(localStorage.getItem("userData"));
      if (!userData) {
        showToast("User data not found", "error");
        return;
      }
      let success = false;

      if (action === "accepted") {
        success = await acceptJoinRequest(groupId, userId);
      } else if (action === "rejected") {
        success = await rejectJoinRequest(groupId, userId);
      }

      if (!success) {
        throw new Error(`Failed to ${action} join request`);
      }

      // Delete the notification after handling
      await DeleteNotification(notificationId);
      fetchNotifications();
      showToast(`Successfully ${action}ed join request`, "success");
    } catch (error) {
      console.error(`Error ${action}ing join request:`, error);
      showToast(`Failed to ${action} join request`, "error");
    }
  };

  // Add new handler for event responses
  const handleEventResponse = async (eventId, notificationId, response) => {
    try {
        const success = await respondToEvent(eventId, response);
        
        if (!success) {
            throw new Error(`Failed to respond to event`);
        }

        // Delete the notification after successful response
        await DeleteNotification(notificationId);
        fetchNotifications();
        
        showToast(`Successfully responded to event`, "success");
    } catch (error) {
        console.error('Error responding to event:', error);
        showToast(`Failed to respond to event`, "error");
    }
  };

  const renderNotificationContent = (notification) => {
    switch (notification.type) {
      case "message":
        return (
          <div className={styles.notification}>
            <div className={styles.avatarContainer}>
              <img
                src={notification.avatar}
                alt={notification.sender}
                className={styles.avatar}
              />
            </div>
            <span className={styles.text}>
              You have a new message from <strong>{notification.sender}</strong>
            </span>
          </div>
        );
      case "reaction":
        return (
          <div className={styles.notification}>
            <div className={styles.avatarContainer}>
              <img
                src={notification.avatar}
                alt={notification.sender}
                className={styles.avatar}
              />
            </div>
            <span className={styles.text}>
              <strong>{notification.sender}</strong> {notification.action} your{" "}
              {notification.contentType}
            </span>
          </div>
        );
      case "comment":
        return (
          <div className={styles.notification}>
            <div className={styles.avatarContainer}>
              <img
                src={notification.avatar}
                alt={notification.sender}
                className={styles.avatar}
              />
            </div>
            <span className={styles.text}>
              <strong>{notification.sender}</strong> commented on your{" "}
              {notification.contentType}
            </span>
          </div>
        );
      case "friendRequest":
        return (
          <div className={styles.notification}>
            <div className={styles.avatarContainer}>
              <img
                src={notification.avatar}
                alt={notification.sender}
                className={styles.avatar}
              />
            </div>
            <div className={styles.textBox}>
              <span className={styles.text}>
                Friend request from <strong>{notification.sender}</strong>
              </span>
              <div className={styles.actions}>
                <button
                  onClick={() =>
                    handleNotificationResponse(notification.senderId, notification.id, "accept", "friendRequest")
                  }
                  className={styles.acceptButton}
                >
                  Accept
                </button>
                <button
                  onClick={() =>
                    handleNotificationResponse(notification.senderId, notification.id, "decline", "friendRequest")
                  }
                  className={styles.declineButton}
                >
                  Decline
                </button>
              </div>
            </div>
          </div>
        );
      case "invitation":
        return (
          <div className={styles.notification}>
            <div className={styles.avatarContainer}>
              <img
                src={notification.avatar}
                alt={notification.sender}
                className={styles.avatar}
              />
            </div>
            <div className={styles.textBox}>
              <span className={styles.text}>
                You have been invited to <strong>{notification.contentType}</strong> by{" "}
                <strong>{notification.sender}</strong>
              </span>
              <div className={styles.actions}>
                <button
                  onClick={() => handleGroupNotificationResponse(
                    notification.target,
                    notification.id,
                    "accepted"
                  )}
                  className={styles.acceptButton}
                >
                  Accept
                </button>
                <button
                  onClick={() => handleGroupNotificationResponse(
                    notification.target,
                    notification.id,
                    "rejected"
                  )}
                  className={styles.declineButton}
                >
                  Decline
                </button>
              </div>
            </div>
          </div>
        );
      case "groupEvent":
        return (
          <div className={styles.notification}>
            <div className={styles.avatarContainer}>
              <img
                src={notification.avatar}
                alt={notification.sender}
                className={styles.avatar}
              />
            </div>
            <div className={styles.textBox}>
              <span className={styles.text}>
                <strong>{notification.sender}</strong> invited you to event{" "}
                <strong>{notification.contentType}</strong>
              </span>
              <div className={styles.actions}>
                <button
                  onClick={() => handleEventResponse(
                    notification.eventId, // eventId
                    notification.id,    // notificationId
                    'going'
                  )}
                  className={styles.acceptButton}
                >
                  <i className="fas fa-check"></i> Going
                </button>
                <button
                  onClick={() => handleEventResponse(
                    notification.eventId,
                    notification.id,
                    'not_going'
                  )}
                  className={styles.declineButton}
                >
                  <i className="fas fa-times"></i> Not Going
                </button>
              </div>
            </div>
          </div>
        );
      case "joinRequest":
        console.log("Join Request Notification:", notification);
        return (
          <div className={styles.notification}>
            <div className={styles.avatarContainer}>
              <img
                src={notification.avatar}
                alt={notification.sender}
                className={styles.avatar}
              />
            </div>
            <div className={styles.textBox}>
              <span className={styles.text}>
                <strong>{notification.sender}</strong> wants to join your group
              </span>
              <div className={styles.actions}>
                <button
                  onClick={() => handleJoinRequestResponse(
                    notification.target, // groupId
                    notification.senderId, // userId
                    notification.id, // notificationId
                    "accepted"
                  )}
                  className={styles.acceptButton}
                >
                  Accept
                </button>
                <button
                  onClick={() => handleJoinRequestResponse(
                    notification.target,
                    notification.senderId,
                    notification.id,
                    "rejected"
                  )}
                  className={styles.declineButton}
                >
                  Decline
                </button>
              </div>
            </div>
          </div>
        );
      default:
        return null;
    }
  };

  return (
    <div className={styles.container} ref={dropdownRef}>
      <button
        className={`${styles.iconButton} ${isOpen ? styles.active : ""}`}
        onClick={() => setIsOpen(!isOpen)}
      >
        <i className="fas fa-bell"></i>
        {notifications.length > 0 && (
          <span className={styles.badge}>{notifications.length}</span>
        )}
      </button>

      {isOpen && (
        <div className={styles.dropdown}>
          <div className={styles.headerActions}>
            <h3 className={styles.headerTitle}>Notifications</h3>
            <div className={styles.actionButtons}>
              <button
                className={styles.actionButton}
                onClick={handleMarkAllAsRead}
              >
                Mark all as read
              </button>
              <button className={styles.actionButton} onClick={handleClearAll}>
                Clear all
              </button>
            </div>
          </div>
          <div className={styles.notificationList}>
            {isLoadingNotifications ? (
              <div className={styles.loadingState}>
                Loading notifications...
              </div>
            ) : notifications.length > 0 ? (
              notifications.map((notification) => (
                <div
                  key={notification.id}
                  className={`${styles.notificationItem} ${!notification.read ? styles.unread : ""
                    }`}
                >
                  <div className={styles.notificationHeader}>
                    {renderNotificationContent(notification)}
                    {!notification.read && (
                      <button
                        className={styles.markReadButton}
                        onClick={() => handleMarkSingleAsRead(notification.id)}
                        title="Mark as read"
                      >
                        <i className="fas fa-check"></i>
                        <span className={styles.tooltip}>Mark as read</span>
                      </button>
                    )}
                  </div>
                  <span className={styles.time}>
                    {formatDistanceToNow(new Date(notification.timestamp), {
                      addSuffix: true,
                    })}
                  </span>
                </div>
              ))
            ) : (
              <div className={styles.emptyState}>No new notifications</div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}