"use client";

import { useState, useRef, useEffect } from 'react';
import styles from '@/styles/ChatWindow.module.css';
import { useUserStatus } from "@/services/userStatusService";

const ChatWindow = ({ contact, onClose, onMinimize, isMinimized }) => {
  const [newMessage, setNewMessage] = useState('');
  const messagesEndRef = useRef(null);
  const { isUserOnline } = useUserStatus();

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [contact.messages]);

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!newMessage.trim()) return;
    
    setNewMessage('');
  };

  return (
    <div className={`${styles.chatWindow} ${isMinimized ? styles.minimized : ''}`}>
      {/* Chat Header */}
      <div className={styles.header}>
        <div className={styles.userInfo}>
          <img src={contact.avatar} alt={contact.name} />
          <div className={styles.nameStatus}>
            <span className={styles.name}>{contact.name}</span>
            <span className={`${styles.status} ${isUserOnline(contact.userId, contact.online) ? styles.online : styles.offline}`}>
              {isUserOnline(contact.userId, contact.online) ? 'Active Now' : 'Offline'}
            </span>
          </div>
        </div>
        <div className={styles.controls}>
          <button onClick={onMinimize} className={styles.minimizeBtn}>
            {isMinimized ? '□' : '−'}
          </button>
          <button onClick={onClose} className={styles.closeBtn}>×</button>
        </div>
      </div>

      {/* Messages Container */}
      <div className={styles.messagesContainer}>
        {contact.messages?.map((message, index) => (
          <div
            key={message.id || index}
            className={`${styles.message} ${message.isSent ? styles.sent : styles.received}`}
          >
            <div className={styles.messageContent}>
              {message.content}
              <span className={styles.messageTime}>
                {message.timestamp || '12:00 PM'}
              </span>
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* Message Input */}
      <form onSubmit={handleSubmit} className={styles.inputContainer}>
        <input
          type="text"
          value={newMessage}
          onChange={(e) => setNewMessage(e.target.value)}
          placeholder="Type a message..."
          className={styles.input}
        />
        <button type="submit" className={styles.sendButton}>
          <i className="fas fa-paper-plane"></i>
        </button>
      </form>
    </div>
  );
};

export default ChatWindow; 