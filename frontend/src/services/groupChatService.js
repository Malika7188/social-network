import { useState, useEffect, useCallback } from "react";
import { useWebSocket, EVENT_TYPES } from "./websocketService";
import { useAuth } from "@/context/authcontext";

// Assuming EVENT_TYPES in websocketService.js includes group-specific events
const GROUP_EVENT_TYPES = {
  GROUP_MESSAGE: "group_message",
  GROUP_MESSAGES_READ: "group_messages_read",
  // GROUP_USER_TYPING: "group_user_typing",
  GROUP_USER_JOINED: "group_user_joined",
  GROUP_USER_LEFT: "group_user_left",
};

export const useGroupChatService = (groupId) => {
  const { currentUser, authenticatedFetch } = useAuth();
  const { subscribe, getConnectionStatus } = useWebSocket();
  const [messages, setMessages] = useState([]);
  const [typingUsers, setTypingUsers] = useState({});
  const [unreadCounts, setUnreadCounts] = useState(0);
  const [activeUsers, setActiveUsers] = useState([]);
  const [isInitialized, setIsInitialized] = useState(false);

  // Load group messages
  const loadMessages = useCallback(
    async (limit = 50, offset = 0) => {
      try {
        const response = await authenticatedFetch(
          `groups/get-messages?groupId=${groupId}&limit=${limit}&offset=${offset}`
        );
          if (!response.ok) throw new Error("Failed to load group messages");
          
          const data = await response.json();
          
          if (!data) {
              setMessages([]); 
            return [];  
          }
            setMessages((prev) => {
                const newMessages = data.filter(
                (msg) => !prev.some((m) => m.id === msg.id)
                );
                return [...prev, ...newMessages];
            });
        return data;
      } catch (error) {
        console.error("Error loading group messages:", error);
        return [];
      }
    },
    [authenticatedFetch, groupId]
  );

  // Send a message to the group
  const sendMessage = useCallback(
    async (content) => {
      try {
        const response = await authenticatedFetch("groups/send-message", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            groupId,
            content,
          }),
        });

        if (!response.ok) throw new Error("Failed to send group message");
        const message = await response.json();

        return message;
      } catch (error) {
        console.error("Error sending group message:", error);
        throw error;
      }
    },
    [authenticatedFetch, groupId]
  );




  
  // Initialize WebSocket subscriptions
  const initializeWebSocketSubscriptions = useCallback(() => {
    if (!currentUser?.id || isInitialized || !groupId) return;

    // Handle new group messages
    const messageUnsubscribe = subscribe(GROUP_EVENT_TYPES.GROUP_MESSAGE, (payload) => {
      if (!payload || payload.GroupID !== groupId) return;
    
      const {
        id,
        Content,
        CreatedAt,
        GroupID,
        User
      } = payload;
    
      const message = {
        id,
        Content,
        CreatedAt,
        groupId: GroupID,
        senderId: User.id,
        senderName: User.firstName,
        senderAvatar: User.avatar,
        User,
        isRead: false, // Optional: may be updated elsewhere
      };
    
      setMessages((prev) => {
        const messageExists = prev.some(
          (msg) =>
            msg.id === id ||
            (msg.isTemp && msg.senderId === User.id && msg.content === Content)
        );
    
        if (messageExists) {
          return prev.map((msg) =>
            msg.isTemp && msg.senderId === User.id && msg.content === Content
              ? message
              : msg
          );
        }
    
        return [...prev, message];
      });
    
      if (User.id !== currentUser.id && !message.isRead) {
        setUnreadCounts((prev) => prev + 1);
      }
    });
    
    // Handle group messages read
    const readUnsubscribe = subscribe(GROUP_EVENT_TYPES.GROUP_MESSAGES_READ, (payload) => {
      if (!payload || payload.groupId !== groupId) return;

      const { readAt, userId } = payload;

      if (userId === currentUser.id) {
        setUnreadCounts(0);
        setMessages((prev) =>
          prev.map((msg) => {
            if (!msg.isRead) {
              return { ...msg, isRead: true, readAt };
            }
            return msg;
          })
        );
      }
    });

    // Handle group typing indicators
    const typingUnsubscribe = subscribe(GROUP_EVENT_TYPES.GROUP_USER_TYPING, (payload) => {
      if (!payload || payload.groupId !== groupId) return;

      const { senderId, timestamp, senderName } = payload;

      setTypingUsers((prev) => ({
        ...prev,
        [senderId]: { timestamp: parseInt(timestamp), name: senderName },
      }));

        setTimeout(() => {
          
        setTypingUsers((prev) => {
          const current = { ...prev };
          if (current[senderId]?.timestamp === parseInt(timestamp)) {
            delete current[senderId];
          }
          return current;
        });
      }, 3000);
    });

    // Handle user joined group
    const joinUnsubscribe = subscribe(GROUP_EVENT_TYPES.GROUP_USER_JOINED, (payload) => {
      if (!payload || payload.groupId !== groupId) return;

      const { userId, userName, avatar } = payload;
      setActiveUsers((prev) => {
        if (prev.some((user) => user.id === userId)) return prev;
        return [...prev, { id: userId, name: userName, avatar, isOnline: true }];
      });
    });

    // Handle user left group
    const leaveUnsubscribe = subscribe(GROUP_EVENT_TYPES.GROUP_USER_LEFT, (payload) => {
      if (!payload || payload.groupId !== groupId) return;

      const { userId } = payload;
      setActiveUsers((prev) => prev.filter((user) => user.id !== userId));
    });

    setIsInitialized(true);

    return () => {
      messageUnsubscribe();
      readUnsubscribe();
      typingUnsubscribe();
      joinUnsubscribe();
      leaveUnsubscribe();
    };
  }, [currentUser, subscribe, isInitialized, groupId]);

  // Polling mechanism for WebSocket connection
  useEffect(() => {
    let attempts = 0;
    const MAX_ATTEMPTS = 5;

    const checkConnection = () => {
      const isConnected = getConnectionStatus();

      if (currentUser?.id && isConnected && !isInitialized && groupId) {
        initializeWebSocketSubscriptions();
      }

      attempts++;
    };

    checkConnection();

    const interval = setInterval(() => {
      if (isInitialized || attempts >= MAX_ATTEMPTS) {
        clearInterval(interval);
        return;
      }
      checkConnection();
    }, 1000);

    return () => clearInterval(interval);
  }, [currentUser, getConnectionStatus, isInitialized, initializeWebSocketSubscriptions, groupId]);

  // Backup initialization
  useEffect(() => {
    if (currentUser?.id && getConnectionStatus() && !isInitialized && groupId) {
      const cleanup = initializeWebSocketSubscriptions();
      return cleanup;
    } else if (!currentUser?.id || !groupId) {
      setIsInitialized(false);
    }
  }, [currentUser, initializeWebSocketSubscriptions, isInitialized, getConnectionStatus, groupId]);

  // Initial data load
  useEffect(() => {
    if (currentUser?.id && groupId) {
      loadMessages();
    }
  }, [currentUser, groupId, loadMessages]);

  return {
    loadMessages,
    sendMessage,
    // markMessagesAsRead,
    // loadActiveUsers,
    messages,
    setMessages, 
    typingUsers,
    unreadCounts,
    activeUsers,
    isInitialized,
  };
};