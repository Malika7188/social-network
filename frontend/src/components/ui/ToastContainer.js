"use client";

import { useState, useCallback, useEffect } from "react";
import Toast from "./Toast";
import styles from "@/styles/Toast.module.css";

const ToastContainer = () => {
  const [toasts, setToasts] = useState([]);

  const removeToast = useCallback((id) => {
    setToasts((prevToasts) => prevToasts.filter((toast) => toast.id !== id));
  }, []);

  useEffect(() => {
    const handleShowToast = (event) => {
      const { message, type, duration, id } = event.detail;
      setToasts((prevToasts) => [
        ...prevToasts,
        { message, type, duration, id },
      ]);
    };

    // Add event listener
    document.addEventListener("showToast", handleShowToast);

    // Cleanup
    return () => {
      document.removeEventListener("showToast", handleShowToast);
    };
  }, []);

  return (
    <div className={styles.toastContainer}>
      {toasts.map((toast) => (
        <Toast
          key={toast.id}
          message={toast.message}
          type={toast.type}
          duration={toast.duration}
          onClose={() => removeToast(toast.id)}
        />
      ))}
    </div>
  );
};

// Export the component and a function to show toasts
export const showToast = (() => {
  let toastContainer = null;
  let counter = 0; // Add a counter to ensure uniqueness

  return (message, type = "info", duration = 3000) => {
    if (!toastContainer) {
      const div = document.createElement("div");
      document.body.appendChild(div);
      toastContainer = div;
    }

    // Use both timestamp and counter to ensure uniqueness
    const id = `${Date.now()}-${counter++}`;

    const event = new CustomEvent("showToast", {
      detail: {
        message,
        type,
        duration,
        id,
      },
    });
    document.dispatchEvent(event);
  };
})();

export default ToastContainer;
