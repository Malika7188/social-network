'use client';

import { formatDistanceToNow } from 'date-fns';
import styles from '@/styles/Messages.module.css';

export default function Message({ message, contactAvatar, contactName }) {
  const { content, timestamp, isSent, isRead } = message;
  
  // Format timestamp for display
  const formatMessageTime = (timestamp) => {
    if (!timestamp) return '';
    return formatDistanceToNow(new Date(timestamp), { addSuffix: true });
  };
  
  return (
    <div className={`${styles.message} ${isSent ? styles.sent : styles.received}`}>
      {!isSent && (
        <img 
          src={contactAvatar} 
          alt={contactName} 
          className={styles.messageAvatar}
        />
      )}
      <div className={styles.messageContent}>
        <div className={styles.messageText}>{content}</div>
        <div className={styles.messageTime}>
          {formatMessageTime(timestamp)}
          {isSent && (
            <span className={styles.readStatus}>
              {isRead ? (
                <i className="fas fa-check-double"></i>
              ) : (
                <i className="fas fa-check"></i>
              )}
            </span>
          )}
        </div>
      </div>
    </div>
  );
} 