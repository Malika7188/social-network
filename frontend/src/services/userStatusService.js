import { useState, useEffect, useCallback } from "react";
import { useWebSocket, EVENT_TYPES } from "./websocketService";

export const useUserStatus = () => {
  const [onlineUsers, setOnlineUsers] = useState({});
  const { subscribe } = useWebSocket();

  // Function to directly set a user's status
  const setUserStatus = useCallback((userId, isOnline) => {
    if (!userId) return;

    setOnlineUsers((prev) => ({
      ...prev,
      [userId]: isOnline,
    }));
  }, []);

  useEffect(() => {
    // Subscribe to user status updates
    const unsubscribe = subscribe(EVENT_TYPES.USER_STATUS_UPDATE, (payload) => {
      if (payload && payload.userId) {
        setUserStatus(payload.userId, payload.isOnline);
      }
    });

    return () => {
      if (unsubscribe) unsubscribe();
    };
  }, [subscribe, setUserStatus]);

  // Function to check if a user is online
  const isUserOnline = useCallback(
    (userId, defaultStatus = false) => {
      // If we have a status update from WebSocket, use that
      if (onlineUsers[userId] !== undefined) {
        return onlineUsers[userId];
      }
      // Otherwise use the provided default status (from API)
      return defaultStatus;
    },
    [onlineUsers]
  );

  // Function to initialize status from API data
  const initializeStatuses = useCallback((contacts) => {
    if (!contacts || !Array.isArray(contacts)) return;

    const newStatuses = {};
    contacts.forEach((contact) => {
      if (contact.contactId && contact.isOnline !== undefined) {
        newStatuses[contact.contactId] = contact.isOnline;
      }
    });

    // Update all statuses at once
    setOnlineUsers((prev) => ({
      ...prev,
      ...newStatuses,
    }));
  }, []);

  // Function to clear a user's status (for logout)
  const clearUserStatus = useCallback(
    (userId) => {
      if (!userId) return;

      setUserStatus(userId, false);
    },
    [setUserStatus]
  );

  return {
    onlineUsers,
    isUserOnline,
    initializeStatuses,
    setUserStatus,
    clearUserStatus,
  };
};
