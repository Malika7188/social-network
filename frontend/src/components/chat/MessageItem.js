"use client";

import { useState } from "react";
import styles from "@/styles/Messages.module.css";
import { useAuth } from "@/context/authcontext";
import { formatRelativeTime } from "@/utils/dateUtils";
import { BASE_URL } from "@/utils/constants";

export default function MessageItem({ message, isOwnMessage, avatar }) {
  const { currentUser } = useAuth();

  return (
    <div
      className={`${styles.messageItem} ${
        isOwnMessage ? styles.sent : styles.received
      }`}
    >
      {!isOwnMessage && (
        <img
          src={
                avatar
                ? `${BASE_URL}/uploads/${avatar}`
                : "/avatar.png"
            }
          alt={message.sender?.firstName || "User"}
          className={styles.messageAvatar}
        />
      )}
      <div className={styles.messageContent}>
        <div className={styles.messageText}>{message.content}</div>
        <div className={styles.messageTime}>
          {formatRelativeTime(new Date(message.createdAt))}
          {isOwnMessage && (
            <span className={styles.readStatus}>
              {message.isRead ? (
                <i className="fas fa-check-double" title="Read"></i>
              ) : (
                <i className="fas fa-check" title="Sent"></i>
              )}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}
