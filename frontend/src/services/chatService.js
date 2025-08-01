import { useState, useEffect, useCallback } from "react";
import { useWebSocket, EVENT_TYPES } from "./websocketService";
import { useAuth } from "@/context/authcontext";

// The EVENT_TYPES in websocketService.js already includes these events

export const useChatService = () => {
  const { currentUser, authenticatedFetch } = useAuth();
  const { subscribe, getConnectionStatus } = useWebSocket();
  const [messages, setMessages] = useState({});
  const [typingUsers, setTypingUsers] = useState({});
  const [unreadCounts, setUnreadCounts] = useState({});
  const [isInitialized, setIsInitialized] = useState(false);

  // Load initial contacts and messages
  const loadContacts = useCallback(async () => {
    try {
      const response = await authenticatedFetch("chat/contacts");
      if (!response.ok) throw new Error("Failed to load contacts");
      const data = await response.json();
      return data;
    } catch (error) {
      console.error("Error loading contacts:", error);
      return [];
    }
  }, [authenticatedFetch]);

  // Load messages for a specific contact
  const loadMessages = useCallback(
    async (contactId, limit = 100, offset = 0) => {
      try {
        const response = await authenticatedFetch(
          `chat/messages?userId=${contactId}&limit=${limit}&offset=${offset}`
        );
        if (!response.ok) throw new Error("Failed to load messages");
        const data = await response.json();

        // Update messages state
        setMessages((prev) => ({
          ...prev,
          [contactId]: data,
        }));

        return data;
      } catch (error) {
        console.error("Error loading messages:", error);
        return [];
      }
    },
    [authenticatedFetch]
  );

  // Send a message to a contact
  const sendMessage = useCallback(
    async (receiverId, content) => {
      try {
        const response = await authenticatedFetch("chat/send", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            receiverId,
            content,
          }),
        });

        if (!response.ok) throw new Error("Failed to send message");
        const message = await response.json();

        return message;
      } catch (error) {
        console.error("Error sending message:", error);
        throw error;
      }
    },
    [authenticatedFetch]
  );

  // Mark messages from a sender as read
  const markMessagesAsRead = useCallback(
    async (senderId) => {
      try {
        const response = await authenticatedFetch("chat/mark-read", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            senderId,
          }),
        });

        if (!response.ok) throw new Error("Failed to mark messages as read");

        // Update unread counts locally
        setUnreadCounts((prev) => ({
          ...prev,
          [senderId]: 0,
        }));

        // Update message read status locally
        setMessages((prev) => {
          const senderMessages = prev[senderId] || [];
          const updatedMessages = senderMessages.map((msg) => {
            if (msg.senderId === senderId && !msg.isRead) {
              return { ...msg, isRead: true, readAt: new Date().toISOString() };
            }
            return msg;
          });

          return {
            ...prev,
            [senderId]: updatedMessages,
          };
        });

        return true;
      } catch (error) {
        console.error("Error marking messages as read:", error);
        return false;
      }
    },
    [authenticatedFetch]
  );

  // Send typing indicator
  const sendTypingIndicator = useCallback(
    async (receiverId) => {
      try {
        await authenticatedFetch("chat/typing", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            receiverId,
          }),
        });
        return true;
      } catch (error) {
        console.error("Error sending typing indicator:", error);
        return false;
      }
    },
    [authenticatedFetch]
  );

  // Initialize WebSocket subscriptions
  const initializeWebSocketSubscriptions = useCallback(() => {
    if (!currentUser?.id || isInitialized) return;

    // Handle new messages
    const messageUnsubscribe = subscribe(
      EVENT_TYPES.PRIVATE_MESSAGE,
      (payload) => {
        if (!payload) return;

        const {
          senderId,
          receiverId,
          content,
          createdAt,
          messageId,
          isRead,
          senderName,
          senderAvatar,
        } = payload;

        // Determine which contact this message belongs to
        const contactId = currentUser.id === senderId ? receiverId : senderId;

        // Create message object
        const message = {
          id: messageId,
          senderId,
          receiverId,
          content,
          createdAt,
          isRead,
          sender: {
            id: senderId,
            firstName: senderName?.split(" ")[0] || "",
            lastName: senderName?.split(" ")[1] || "",
            avatar: senderAvatar,
          },
        };

        // Update messages state
        setMessages((prev) => {
          const contactMessages = prev[contactId] || [];

          // Check if message already exists to avoid duplicates
          // This checks both server-generated IDs and content/timestamp for temp messages
          const messageExists = contactMessages.some(
            (msg) =>
              msg.id === messageId ||
              (msg.isTemp &&
                msg.senderId === senderId &&
                msg.content === content)
          );

          if (messageExists) {
            // If it exists but might be a temp message, replace it with the server version
            return {
              ...prev,
              [contactId]: contactMessages.map((msg) =>
                msg.isTemp &&
                msg.senderId === senderId &&
                msg.content === content
                  ? message
                  : msg
              ),
            };
          }

          return {
            ...prev,
            [contactId]: [...contactMessages, message],
          };
        });

        // Update unread counts if we're the receiver
        if (receiverId === currentUser.id && !isRead) {
          setUnreadCounts((prev) => ({
            ...prev,
            [senderId]: (prev[senderId] || 0) + 1,
          }));
        }
      }
    );

    // Handle messages read
    const readUnsubscribe = subscribe(EVENT_TYPES.MESSAGES_READ, (payload) => {
      if (!payload) return;

      const { senderId, receiverId, readAt } = payload;

      // If we're the sender, update our messages to the receiver as read
      if (senderId === currentUser.id) {
        setMessages((prev) => {
          const contactMessages = prev[receiverId] || [];
          const updatedMessages = contactMessages.map((msg) => {
            if (msg.senderId === currentUser.id && !msg.isRead) {
              return { ...msg, isRead: true, readAt };
            }
            return msg;
          });

          return {
            ...prev,
            [receiverId]: updatedMessages,
          };
        });
      }

      // If we're the receiver, update our unread count
      if (receiverId === currentUser.id) {
        setUnreadCounts((prev) => ({
          ...prev,
          [senderId]: 0,
        }));
      }
    });

    // Handle typing indicators
    const typingUnsubscribe = subscribe(EVENT_TYPES.USER_TYPING, (payload) => {
      if (!payload) return;

      const { senderId, timestamp } = payload;

      // Set typing status
      setTypingUsers((prev) => ({
        ...prev,
        [senderId]: parseInt(timestamp),
      }));

      // Clear typing status after 3 seconds
      setTimeout(() => {
        setTypingUsers((prev) => {
          const current = { ...prev };
          if (current[senderId] === parseInt(timestamp)) {
            delete current[senderId];
          }
          return current;
        });
      }, 3000);
    });

    setIsInitialized(true);

    return () => {
      messageUnsubscribe();
      readUnsubscribe();
      typingUnsubscribe();
    };
  }, [currentUser, subscribe, isInitialized]);

  // Add a polling mechanism to check connection status
  useEffect(() => {
    let attempts = 0;
    const MAX_ATTEMPTS = 5;

    const checkConnection = () => {
      const isConnected = getConnectionStatus();

      // If connected and not initialized, initialize
      if (currentUser?.id && isConnected && !isInitialized) {
        initializeWebSocketSubscriptions();
      }

      attempts++;
    };

    // Check immediately
    checkConnection();

    // Then check every second until initialized or max attempts reached
    const interval = setInterval(() => {
      if (isInitialized || attempts >= MAX_ATTEMPTS) {
        clearInterval(interval);
        return;
      }
      checkConnection();
    }, 1000);

    return () => clearInterval(interval);
  }, [
    currentUser,
    getConnectionStatus,
    isInitialized,
    initializeWebSocketSubscriptions,
  ]);

  // Keep your existing useEffect as a backup
  useEffect(() => {
    // Only initialize if user exists, WebSocket is connected, and not already initialized
    if (currentUser?.id && getConnectionStatus() && !isInitialized) {
      const cleanup = initializeWebSocketSubscriptions();
      return cleanup;
    } else if (!currentUser?.id) {
      // Reset initialization when user is not available
      setIsInitialized(false);
    }
  }, [
    currentUser,
    initializeWebSocketSubscriptions,
    isInitialized,
    getConnectionStatus,
  ]);

  return {
    loadContacts,
    loadMessages,
    sendMessage,
    markMessagesAsRead,
    sendTypingIndicator,
    messages,
    typingUsers,
    unreadCounts,
    isInitialized,
  };
};
