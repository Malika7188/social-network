"use client";

import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
} from "react";
import { useChatService } from "@/services/chatService";
import { useAuth } from "@/context/authcontext";

const ChatContext = createContext(null);

export const ChatProvider = ({ children }) => {
  const { currentUser } = useAuth();
  const {
    loadContacts: serviceLoadContacts,
    loadMessages: serviceLoadMessages,
    sendMessage: serviceSendMessage,
    markMessagesAsRead: serviceMarkMessagesAsRead,
    sendTypingIndicator: serviceSendTypingIndicator,
    messages: serviceMessages,
    typingUsers: serviceTypingUsers,
    unreadCounts: serviceUnreadCounts,
    isInitialized,
  } = useChatService();

  const [contacts, setContacts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [isDataLoaded, setIsDataLoaded] = useState(false);

  // Initialize chat service when the component mounts
  useEffect(() => {
    if (currentUser?.id) {
      serviceLoadContacts();
    }
  }, [currentUser, serviceLoadContacts]);

  // Load all chat data (contacts and their last messages)
  const loadChatData = useCallback(async () => {
    if (!currentUser?.id) return;

    setLoading(true);
    try {
      const contactsData = await serviceLoadContacts();
      setContacts(contactsData);
      setIsDataLoaded(true);
      return contactsData;
    } catch (error) {
      console.error("Error loading chat data:", error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [currentUser, serviceLoadContacts]);

  // Update contacts when unread counts change
  useEffect(() => {
    if (contacts.length > 0) {
      // Create a new array to trigger re-render
      const updatedContacts = contacts.map((contact) => {
        if (contact && contact.userId) {
          return {
            ...contact,
            unreadCount: serviceUnreadCounts[contact.userId] || 0,
          };
        }
        return contact;
      });

      setContacts(updatedContacts);
    }
  }, [serviceUnreadCounts]);

  // Update contacts when a new message is received
  useEffect(() => {
    if (!currentUser?.id || !isInitialized || !contacts) return;

    // This will update the contacts list when messages change
    // to reflect the latest message for each contact
    const updateContactsWithLatestMessages = () => {
      const updatedContacts = [...contacts];
      let hasChanges = false;

      // For each contact, check if there are messages and update the last message
      updatedContacts.forEach((contact, index) => {
        if (
          contact &&
          contact.userId &&
          serviceMessages[contact.userId]?.length > 0
        ) {
          const contactMessages = serviceMessages[contact.userId];
          const latestMessage = contactMessages[contactMessages.length - 1];

          if (latestMessage) {
            // Check if this would actually change anything
            if (
              contact.lastMessage !== latestMessage.content ||
              contact.lastMessageSenderId !== latestMessage.senderId ||
              contact.lastSent !== latestMessage.createdAt ||
              contact.unreadCount !== (serviceUnreadCounts[contact.userId] || 0)
            ) {
              updatedContacts[index] = {
                ...contact,
                lastMessage: latestMessage.content,
                lastMessageSenderId: latestMessage.senderId,
                lastSent: latestMessage.createdAt,
                unreadCount: serviceUnreadCounts[contact.userId] || 0,
              };
              hasChanges = true;
            }
          }
        }
      });

      // Only update if something actually changed
      if (hasChanges) {
        setContacts(updatedContacts);
      }
    };

    updateContactsWithLatestMessages();
  }, [
    serviceMessages,
    currentUser,
    isInitialized,
    contacts,
    serviceUnreadCounts,
  ]);

  // Wrap the service functions to update our local state
  const loadMessages = useCallback(
    async (userId) => {
      return await serviceLoadMessages(userId);
    },
    [serviceLoadMessages]
  );

  const sendMessage = useCallback(
    async (receiverId, content) => {
      return await serviceSendMessage(receiverId, content);
    },
    [serviceSendMessage]
  );

  const markMessagesAsRead = useCallback(
    async (senderId) => {
      const result = await serviceMarkMessagesAsRead(senderId);

      // Update the local contacts state to reflect read status
      setContacts((prevContacts) =>
        prevContacts.map((contact) => {
          if (contact.userId === senderId) {
            return {
              ...contact,
              unreadCount: 0,
            };
          }
          return contact;
        })
      );

      return result;
    },
    [serviceMarkMessagesAsRead]
  );

  // Provide all chat functionality to components
  const value = {
    contacts,
    loading,
    isDataLoaded,
    loadChatData,
    loadContacts: serviceLoadContacts,
    loadMessages,
    sendMessage,
    markMessagesAsRead,
    sendTypingIndicator: serviceSendTypingIndicator,
    messages: serviceMessages,
    typingUsers: serviceTypingUsers,
    unreadCounts: serviceUnreadCounts,
  };

  return <ChatContext.Provider value={value}>{children}</ChatContext.Provider>;
};

export const useChatContext = () => {
  const context = useContext(ChatContext);
  if (!context) {
    throw new Error("useChatContext must be used within a ChatProvider");
  }
  return context;
};
