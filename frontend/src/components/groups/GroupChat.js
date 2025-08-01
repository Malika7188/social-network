import { useState, useEffect, useRef } from 'react';
import styles from '@/styles/GroupChat.module.css';
import { useAuth } from '@/context/authcontext';
import { useGroupChatService } from '@/services/groupChatService'; 
import { BASE_URL } from "@/utils/constants";
import EmojiPicker from '../ui/EmojiPicker';


const GroupChat = ({ groupId, groupName }) => {
  const [newMessage, setNewMessage] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [selectedEmoji, setSelectedEmoji] = useState(null);
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const messagesEndRef = useRef(null);
  const { currentUser } = useAuth();
  const { messages, setMessages, typingUsers, sendMessage, loadMessages } = useGroupChatService(groupId);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    const fetchMessages = async () => {
      setIsLoading(true);
      try {
        await loadMessages();
      } catch (error) {
        console.error("Failed to fetch group messages:", error);
      } finally {
        setIsLoading(false);
      }
    };

    if (currentUser && groupId) {
      fetchMessages();
    }
  }, [currentUser, groupId, loadMessages]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSendMessage = async (e) => {
    e.preventDefault();
    if (!newMessage.trim() || !currentUser) return;

    const avatar = currentUser?.avatar
        ? currentUser.avatar.startsWith("http")
          ? currentUser.avatar
          : `${BASE_URL}/uploads/${currentUser.avatar}`
        : "/avatar.png";

    const tempMessage = {
      id: Date.now(),
      Content: newMessage,
      User: {
        id: currentUser.id || 'temp-id',
        firstName: currentUser.firstName || 'Anonymous',
        avatar: avatar|| '/default-avatar.png',
      },
      CreatedAt: new Date().toISOString(),
      isTemp: true,
    };

    // Optimistically update UI
    setMessages((prev) => [...prev, tempMessage]);
    setNewMessage('');

    try {
      await sendMessage(newMessage);
    } catch (error) {
      // Remove temp message on failure
      setMessages((prev) => prev.filter((msg) => msg.id !== tempMessage.id));
    }
  };

  const handleReaction = (messageId, emoji) => {
    if (!currentUser) return;

    setMessages((prev) =>
      prev.map((msg) => {
        if (msg.id === messageId) {
          const reactions = [...msg.reactions];
          const existingReaction = reactions.findIndex((r) => r.userId === currentUser.id);

          if (existingReaction >= 0) {
            reactions[existingReaction].emoji = emoji;
          } else {
            reactions.push({ userId: currentUser.id, emoji });
          }

          return { ...msg, reactions };
        }
        return msg;
      })
    );
  };

  // Handle typing indicator
  const handleTyping = (e) => {
    setNewMessage(e.target.value);
   
  };
  const handleEmojiSelect = (emoji) => {
    if (emoji) {
      setNewMessage((prev) => prev + emoji);
    }
    setShowEmojiPicker(false);
  };

  const handleCloseEmojiPicker = () => {
    setShowEmojiPicker(false);
  };


  if (isLoading) {
    return <div className={styles.loading}>Loading chat...</div>;
  }

  return (
    <div className={styles.chatContainer}>
      <div className={styles.chatHeader}>
        <h2>{groupName} Chat</h2>
      </div>

      <div className={styles.chatMessages}>
        {messages.map((message) => (
          <div
            key={message.ID * Math.random(50)}
            className={`${styles.messageWrapper} ${
              currentUser && message.User.id === currentUser.id ? styles.ownMessage : ''
            }`}
          >
            <div className={styles.message}>
              <img
                src={message.User.avatar ? message.User.avatar.startsWith("http") ? message.User.avatar : `${BASE_URL}/uploads/${message.User.avatar}` : '/avatar.png'}
                alt={message.User.firstName}
                className={styles.senderAvatar}
              />
              <div className={styles.messageContent}>
                <div className={styles.messageHeader}>
                  <span className={styles.senderName}>{message.User.firstName}</span>
                  <span className={styles.timestamp}>
                    {new Date(message.CreatedAt).toLocaleTimeString()}
                  </span>
                </div>
                <p>{message.Content}</p>
                
              </div>
             
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      <form onSubmit={handleSendMessage} className={styles.chatInput}>
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
          value={newMessage}
          onChange={handleTyping}
          placeholder="Type a message..."
        />
        <button type="submit" className={styles.sendButton}>
          <i className="fas fa-paper-plane"></i>
        </button>
      </form>
    </div>
  );
};

export default GroupChat;