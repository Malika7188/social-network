import { useAuth } from "@/context/authcontext";
import { showToast } from "@/components/ui/ToastContainer";
import { useState, useEffect, useCallback } from "react";
import { EVENT_TYPES, useWebSocket } from "./websocketService";
import { useFriendService } from "@/services/friendService";
import { BASE_URL } from "@/utils/constants";
import { useGroupService } from "./groupService";

export const useNotificationService = () => {
  const { authenticatedFetch } = useAuth();
  const [notifications, setNotifications] = useState([]);
  const [isLoadingNotifications, setIsLoadingNotifications] = useState(false);
  const { respondToEvent } = useGroupService();
  const {
    acceptFriendRequest,
    declineFriendRequest,
  } = useFriendService();
  const { subscribe } = useWebSocket();

  // Fetch notifications
  const fetchNotifications = useCallback(async (limit = 10, offset = 0) => {
    setIsLoadingNotifications(true);
    try {
        const response = await authenticatedFetch(`notification?limit=${limit}&offset=${offset}/`);
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message ||
            errorData.error ||
            "Failed to fetch notifications"
        );
      }

      const data = await response.json();

      if (data) {
        // Transform the data to match the component's expected format
        const formattedNotifications = data.map((notification) => ({
          id: notification.id,
          type: notification.type,
          senderId: notification.senderId,
          sender: notification.senderName || "Unknown",
          avatar: notification.senderAvatar
            ? `${BASE_URL}/uploads/${notification.senderAvatar}`
            : "/avatar.png",
          timestamp: notification.createdAt,
          read: notification.isRead,
          target: notification.targetGroupId,
          action: notification.type === "reaction" ? notification.message.split(" ")[0] : undefined,
          contentType:
            notification.type === "reaction" || notification.type === "comment"
              ? notification.message.split(" ").pop()
              : notification.type === "invitation"
              ? notification.message.split("to ")[1]
              : notification.type === "groupEvent"
              ? notification.message.split("to ")[1]
              : undefined,
          eventId: notification.type === "groupEvent" ? notification.targetEventId : undefined,
        }));
        setNotifications(formattedNotifications);
        return formattedNotifications;
      }
      return [];
    } catch (error) {
      console.error("Error fetching notifications:", error);
      showToast(error.message || "Error fetching notifications", "error");
      return [];
    } finally {
      setIsLoadingNotifications(false);
    }
  }, [authenticatedFetch]);

  // Mark a single notification as read
  const markNotificationAsRead = async (notificationId) => {
    try {
      const response = await authenticatedFetch(`notification?notificationId=${notificationId}`, {
        method: "PUT",
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message ||
            errorData.error ||
            "Failed to mark notification as read"
        );
      }

      setNotifications((prev) =>
        prev.map((notification) =>
          notification.id === notificationId
            ? { ...notification, read: true }
            : notification
        )
      );
      showToast("Notification marked as read", "success");
      return true;
    } catch (error) {
      console.error("Error marking notification as read:", error);
      showToast(error.message || "Error marking notification as read", "error");
      return false;
    }
  };

  // Mark all notifications as read
  const markAllNotificationsAsRead = async () => {
    try {
      const response = await authenticatedFetch("notification/read", {
        method: "PUT",
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message ||
            errorData.error ||
            "Failed to mark all notifications as read"
        );
      }

      setNotifications((prev) =>
        prev.map((notification) => ({ ...notification, read: true }))
      );
      showToast("All notifications marked as read", "success");
      return true;
    } catch (error) {
      console.error("Error marking all notifications as read:", error);
      showToast(error.message || "Error marking all notifications as read", "error");
      return false;
    }
  };

  // Clear all notifications
  const clearAllNotifications = async () => {
    try {
      const response = await authenticatedFetch("notification", {
        method: "DELETE",
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message ||
            errorData.error ||
            "Failed to clear notifications"
        );
      }

      setNotifications([]);
      showToast("All notifications cleared", "success");
      return true;
    } catch (error) {
      console.error("Error clearing notifications:", error);
      showToast(error.message || "Error clearing notifications", "error");
      return false;
    }
  };

  const DeleteNotification = async (notificationId) => {
    try {
      const response = await authenticatedFetch(`notification/delete?notificationId=${notificationId}`, {
        method: "DELETE",
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message ||
            errorData.error ||
            "Failed to delete notifications"
        );
      }
      return
    } catch (error) {
      console.error(`Error deleting notification:`, error);
      showToast(error.message || `Error deleting notification`, "error");
      return;
    }
  };

  // Handle friend request or group event response (accept/decline)
  const handleFriendRequest = async (id, notificationId, action, type = "friendRequest") => {
    try {
      if (type === "friendRequest") {
        if (action === "accept") {
          const response = await acceptFriendRequest(id);
          if (!response) {
            return;
          }
        } else if (action === "decline") {
          const response = await declineFriendRequest(id);
          if (!response) {
            return;
          }
        }
      } else if (type === "groupEvent") {
        const response = await respondToEvent(id, action);
        if (!response) {
          return;
        }
      }

      setNotifications((prev) =>
        prev.filter((notification) => notification.id !== notificationId)
      );
      DeleteNotification(notificationId);
      return true;
    } catch (error) {
      console.error(`Error ${action}ing ${type}:`, error);
      showToast(error.message || `Error ${action}ing ${type}`, "error");
      return false;
    }
  };

  // Subscribe to notification events
  useEffect(() => {
    // Listen for new notifications
    const unsubscribeNotification = subscribe(EVENT_TYPES.NOTIFICATION_UPDATE, (payload) => {
      if (payload) {
        const newNotification = {
          id: payload.id,
          type: payload.type,
          senderId: payload.senderId,
          sender: payload.senderName || "Unknown",
          avatar: payload.senderAvatar
            ? `${BASE_URL}/uploads/${payload.senderAvatar}`
            : "/avatar.png",
          timestamp: payload.createdAt,
          target: payload.targetGroupId,
          read: payload.isRead,
          action: payload.type === "reaction" ? payload.message.split(" ")[0] : undefined,
          contentType:
            payload.type === "reaction" || payload.type === "comment"
              ? payload.message.split(" ").pop()
              : payload.type === "invitation" || payload.type === "groupEvent"
              ? payload.message.split("to ")[1]
              : undefined,
          eventId: payload.type === "groupEvent" ? payload.eventId : undefined,
        };
        setNotifications((prev) => [newNotification, ...prev]);
        showToast("New notification received", "info");
      }
    });

    return () => {
      if (unsubscribeNotification) unsubscribeNotification();
    };
  }, [subscribe]);

  // Initial data fetch
  useEffect(() => {
    fetchNotifications();
  }, [fetchNotifications]);

  return {
    notifications,
    isLoadingNotifications,
    fetchNotifications,
    markNotificationAsRead,
    markAllNotificationsAsRead,
    clearAllNotifications,
    handleFriendRequest,
    DeleteNotification,
  };
};