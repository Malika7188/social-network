"use client";

import { useState, useEffect, useRef } from "react";
import Header from "@/components/header/Header";
import styles from "@/styles/Messages.module.css";
import NewMessageModal from "./NewMessageModal";
import ProtectedRoute from "@/components/auth/ProtectedRoute";
import { useChatContext } from "@/context/chatContext";
import { useAuth } from "@/context/authcontext";
import { useUserStatus } from "@/services/userStatusService";
import MessageItem from "@/components/chat/MessageItem";
import { formatRelativeTime } from "@/utils/dateUtils";
import { showToast } from "@/components/ui/ToastContainer";
import { BASE_URL } from "@/utils/constants";
import EmojiPicker from "@/components/ui/EmojiPicker";

export default function Messages() {
  const { currentUser } = useAuth();
  const {
    loadContacts,
    loadMessages,
    sendMessage,
    markMessagesAsRead,
    sendTypingIndicator,
    messages,
    typingUsers,
    unreadCounts,
    loadChatData,
    isDataLoaded,
    contacts,
    loading,
  } = useChatContext();

  const { isUserOnline, initializeStatuses } = useUserStatus();
  const [selectedChat, setSelectedChat] = useState(null);
  const [showNewMessageModal, setShowNewMessageModal] = useState(false);
  const [newMessageText, setNewMessageText] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [localLoading, setLocalLoading] = useState(false);
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const [isMobile, setIsMobile] = useState(false);
  const messagesEndRef = useRef(null);
  const messageInputRef = useRef(null);
  const typingTimeoutRef = useRef(null);

  // Detect mobile view
  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth <= 768);
    };

    window.addEventListener("resize", handleResize);
    handleResize(); // Initial check

    return () => {
      window.removeEventListener("resize", handleResize);
    };
  }, []);

  // Load contacts only when the page is loaded
  useEffect(() => {
    const fetchData = async () => {
      if (!isDataLoaded) {
        try {
          setLocalLoading(true);
          await loadChatData();
        } catch (error) {
          console.error("Error loading contacts:", error);
          showToast("Failed to load contacts", "error");
        } finally {
          setLocalLoading(false);
        }
      }
    };

    fetchData();
  }, [loadChatData, isDataLoaded]);

  // Load messages when a chat is selected
  useEffect(() => {
    if (selectedChat && selectedChat.userId) {
      const fetchMessages = async () => {
        try {
          await loadMessages(selectedChat.userId);
          // Mark messages as read when opening a chat
          await markMessagesAsRead(selectedChat.userId);
          messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
        } catch (error) {
          console.error("Error loading messages:", error);
          showToast("Failed to load messages", "error");
        }
      };

      fetchMessages();
    }
  }, [selectedChat, loadMessages, markMessagesAsRead]);

  // Scroll to bottom when messages change
  useEffect(() => {
    if (selectedChat && selectedChat.userId) {
      messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages, selectedChat]);

  // Focus input when chat is selected
  useEffect(() => {
    if (selectedChat) {
      messageInputRef.current?.focus();
    }
  }, [selectedChat]);

  // Handle typing indicator
  useEffect(() => {
    if (selectedChat) {
      // Send typing indicator
      sendTypingIndicator(selectedChat.userId);
    }
  }, [selectedChat]);

  // Handle typing indication timeout
  const handleTyping = () => {
    typingTimeoutRef.current = setTimeout(() => {
      typingTimeoutRef.current = null;
    }, 3000);
    if (!selectedChat || !selectedChat.userId) return;
  };

  // Clear existing timeout
  if (typingTimeoutRef.current) {
    clearTimeout(typingTimeoutRef.current);
  }

  // Handle sending a message
  const handleSendMessage = async (e) => {
    e.preventDefault();

    if (!newMessageText.trim() || !selectedChat || !selectedChat.userId) return;
    // Send typing indicator
    sendTypingIndicator(selectedChat.userId);

    try {
      await sendMessage(selectedChat.userId, newMessageText);
      setNewMessageText("");
    } catch (error) {
      console.error("Error sending message:", error);
      showToast("Failed to send message", "error");
    }
  };

  // Handle selecting a chat
  const handleSelectChat = async (contact) => {
    if (!contact || !contact.userId) {
      console.error("Invalid contact selected:", contact);
      return;
    }
    try {
      setSelectedChat(contact);
      // Mark messages as read when selecting a chat
      if (unreadCounts[contact.userId] > 0) {
        await markMessagesAsRead(contact.userId);
      }
    } catch (error) {
      console.error("Error marking messages as read:", error);
    }
  };

  // Handle emoji selection
  const handleEmojiSelect = (emoji) => {
    if (emoji) {
      setNewMessageText((prev) => prev + emoji);
    }
    setShowEmojiPicker(false);
  };

  // Close emoji picker
  const handleCloseEmojiPicker = () => {
    setShowEmojiPicker(false);
  };

  // Filter contacts based on search query
  const filteredContacts = contacts?.filter(
    (contact) =>
      contact &&
      contact.userId && // Ensure userId exists
      ((contact.firstName && contact.firstName.toLowerCase().includes(searchQuery.toLowerCase())) ||
        (contact.lastName && contact.lastName.toLowerCase().includes(searchQuery.toLowerCase())) ||
        (contact.name && contact.name.toLowerCase().includes(searchQuery.toLowerCase())))
  ) || [];

  // Sort contacts by most recent message timestamp, then alphabetically
  const sortedContacts = [...filteredContacts].sort((a, b) => {
    const aName = `${a.firstName || ""} ${a.lastName || ""}`.trim();
    const bName = `${b.firstName || ""} ${b.lastName || ""}`.trim();

    // Use the lastSent field from the backend data
    if (a.lastSent && b.lastSent) {
      const aTime = new Date(a.lastSent).getTime();
      const bTime = new Date(b.lastSent).getTime();
      if (aTime !== bTime) return bTime - aTime; // Most recent first
    }
    // If only one has lastSent, prioritize that one
    else if (a.lastSent) return -1;
    else if (b.lastSent) return 1;

    // Fallback to alphabetical order by name
    return aName.localeCompare(bName);
  });

  // Get current chat messages
  const currentChatMessages =
    selectedChat && selectedChat.userId
      ? messages[selectedChat.userId] || []
      : [];

  // Check if the selected user is typing
  const isTyping =
    selectedChat && selectedChat.userId && typingUsers[selectedChat.userId];

  // Initialize online statuses from API data
  useEffect(() => {
    if (contacts && contacts.length > 0) {
      initializeStatuses(contacts);
    }
  }, [contacts, initializeStatuses]);

  return (
    <ProtectedRoute>
      <Header />
      <div className={styles.messagesContainer}>
        <aside className={`${styles.conversationsList} ${selectedChat && isMobile ? styles.hidden : ''}`}>
          <div className={styles.conversationsHeader}>
            <h2>{isMobile ? 'Chats' : 'Messages'}</h2>
          </div>
          <div className={styles.searchContainer}>
            <input
              type="text"
              placeholder="Search messages..."
              className={styles.searchInput}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
          <div className={styles.conversations}>
            {localLoading ? (
              <div className={styles.loadingState}>
                <i className="fas fa-spinner fa-spin"></i>
                <p>Loading conversations...</p>
              </div>
            ) : sortedContacts.length > 0 ? (
              sortedContacts.map((contact) => {
                if (!contact || !contact.userId) return null;

                return (
                  <div
                    key={contact.userId}
                    className={`${styles.conversationItem} ${selectedChat?.userId === contact.userId ? styles.selected : ""} ${contact.unreadCount > 0 ? styles.unread : ""}`}
                    onClick={() => handleSelectChat(contact)}
                  >
                    <div className={styles.avatarContainer}>
                      <img
                        src={
                          contact.avatar
                            ? `${BASE_URL}/uploads/${contact.avatar}`
                            : "/avatar.png"
                        }
                        alt={`${contact.firstName} ${contact.lastName}`}
                        className={styles.avatar}
                      />
                      <span
                        className={`${styles.statusIndicator} ${isUserOnline(contact.userId, contact.isOnline) ? styles.online : styles.offline}`}
                      ></span>
                    </div>
                    <div className={styles.previewContent}>
                      <h4 className={contact.unreadCount > 0 ? styles.boldName : ""}>
                        {contact.firstName} {contact.lastName}
                      </h4>
                      <p className={contact.unreadCount > 0 ? styles.boldPreview : styles.messagePreview}>
                        {contact.lastMessage ? (
                          contact.lastMessageSenderId === currentUser?.id ? (
                            <span>
                              You:{" "}
                              {contact.lastMessage.length > 30
                                ? `${contact.lastMessage.substring(0, 30)}...`
                                : contact.lastMessage}
                            </span>
                          ) : (
                            contact.lastMessage
                          )
                        ) : (
                          "Start a conversation"
                        )}
                      </p>
                    </div>
                    <div className={styles.conversationMeta}>
                      <span className={`${styles.timestamp} ${contact.unreadCount > 0 ? styles.boldTimestamp : ""}`}>
                        {contact.lastSent && contact.lastSent !== "0001-01-01T00:00:00Z"
                          ? formatRelativeTime(new Date(contact.lastSent))
                          : ""}
                      </span>
                      {contact.unreadCount > 0 && (
                        <span className={styles.unreadBadge}>
                          {contact.unreadCount}
                        </span>
                      )}
                    </div>
                  </div>
                );
              })
            ) : (
              <div className={styles.emptyState}>
                <p>No conversations found</p>
              </div>
            )}
          </div>
        </aside>

        <main className={`${styles.chatArea} ${!selectedChat && isMobile ? styles.hidden : ''}`}>
          {selectedChat && selectedChat.userId ? (
            <>
              <div className={styles.chatHeader}>
                {isMobile && (
                  <button
                    className={styles.backButton}
                    onClick={() => setSelectedChat(null)}
                  >
                    <i className="fas fa-arrow-left"></i>
                  </button>
                )}
                <div className={styles.avatarContainer}>
                  <img
                    src={
                      selectedChat.avatar
                        ? `${BASE_URL}/uploads/${selectedChat.avatar}`
                        : "/avatar.png"
                    }
                    alt={`${selectedChat.firstName} ${selectedChat.lastName}`}
                    className={styles.avatar}
                  />
                  <span
                    className={`${styles.statusIndicator} ${isUserOnline(selectedChat.userId, selectedChat.isOnline) ? styles.online : styles.offline}`}
                  ></span>
                </div>
                <div className={styles.userDetails}>
                  <h2>
                    {selectedChat.firstName} {selectedChat.lastName}
                  </h2>
                  <span className={styles.userStatus}>
                    {isUserOnline(selectedChat.userId, selectedChat.isOnline) ? "Online" : "Offline"}
                  </span>
                </div>
                <div className={styles.chatActions}>
                  <button className={styles.actionButton}>
                    <i className="fas fa-phone"></i>
                  </button>
                  <button className={styles.actionButton}>
                    <i className="fas fa-video"></i>
                  </button>
                  <button className={styles.actionButton}>
                    <i className="fas fa-info-circle"></i>
                  </button>
                </div>
              </div>

              <div className={styles.messagesList}>
                {currentChatMessages.length > 0 ? (
                  currentChatMessages.map((message) => (
                    <MessageItem
                      key={`${message.id}-${message.createdAt}`}
                      message={message}
                      isOwnMessage={message.senderId === currentUser?.id}
                      avatar={
                        message.senderId === currentUser?.id
                          ? currentUser?.avatar
                          : selectedChat?.avatar
                      }
                    />
                  ))
                ) : (
                  <div className={styles.emptyMessages}>
                    <p>No messages yet. Start a conversation!</p>
                  </div>
                )}
                <div ref={messagesEndRef} />
              </div>

              {isTyping && (
                <div className={styles.typingIndicatorContainer}>
                  <div className={styles.typingDot}></div>
                  <div className={styles.typingDot}></div>
                  <div className={styles.typingDot}></div>
                </div>
              )}

              <form
                className={styles.messageInputContainer}
                onSubmit={handleSendMessage}
              >
                <button
                  type="button"
                  className={styles.emojiButton}
                  onClick={() => setShowEmojiPicker(!showEmojiPicker)}
                >
                  😊
                </button>
                {showEmojiPicker && (
                  <EmojiPicker
                    onEmojiSelect={handleEmojiSelect}
                    onClose={handleCloseEmojiPicker}
                  />
                )}
                <input
                  type="text"
                  placeholder="Type a message..."
                  className={styles.messageInput}
                  value={newMessageText}
                  onChange={(e) => setNewMessageText(e.target.value)}
                  onKeyDown={handleTyping}
                  ref={messageInputRef}
                />
                <button
                  type="submit"
                  className={styles.sendButton}
                  disabled={!newMessageText.trim()}
                >
                  <i className="fas fa-paper-plane"></i>
                </button>
              </form>
            </>
          ) : (
            <div className={styles.noChat}>
              <div className={styles.noChatContent}>
                <i className="fas fa-comments"></i>
                <h2>Select a conversation to start messaging</h2>
                <p>
                  Choose from your existing conversations or start a new one
                </p>
              </div>
            </div>
          )}
        </main>
      </div>
      {showNewMessageModal && (
        <NewMessageModal
          onClose={() => setShowNewMessageModal(false)}
          onSelectContact={handleSelectChat}
        />
      )}
    </ProtectedRoute>
  );
}
