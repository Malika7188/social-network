"use client";

import { useState, useEffect } from "react";
import styles from "@/styles/Toast.module.css";

const Toast = ({ message, type = "info", duration = 3000, onClose }) => {
  const [isRemoving, setIsRemoving] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => {
      setIsRemoving(true);
      setTimeout(() => {
        onClose();
      }, 300); // Wait for slide-down animation
    }, duration);

    return () => clearTimeout(timer);
  }, [duration, onClose]);

  const getIcon = () => {
    switch (type) {
      case "success":
        return "fas fa-check-circle";
      case "error":
        return "fas fa-exclamation-circle";
      case "warning":
        return "fas fa-exclamation-triangle";
      case "info":
        return "fas fa-info-circle";
      default:
        return "fas fa-info-circle";
    }
  };

  return (
    <div
      className={`${styles.toast} ${styles[type]} ${
        isRemoving ? styles.removing : ""
      }`}
    >
      <div className={styles.icon}>
        <i className={getIcon()}></i>
      </div>
      <div className={styles.message}>{message}</div>
      <button
        className={styles.closeButton}
        onClick={() => setIsRemoving(true)}
      >
        <i className="fas fa-times"></i>
      </button>
    </div>
  );
};

export default Toast;
