"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { useAuth } from "@/context/authcontext";
import { BASE_URL } from "@/utils/constants";
import { showToast } from "@/components/ui/ToastContainer";

// Event types that match backend definitions
export const EVENT_TYPES = {
  POST_CREATED: "post_created",
  POST_LIKED: "post_liked",
  USER_STATS_UPDATED: "user_stats_updated",
  USER_STATUS_UPDATE: "user_status_update",
  // Chat events
  PRIVATE_MESSAGE: "private_message",
  MESSAGES_READ: "messages_read",
  USER_TYPING: "user_typing",
  NOTIFICATION_UPDATE: "notification_Update",
  // Add more event types as needed
  // COMMENT_ADDED: 'comment_added',
  // MESSAGE_RECEIVED: 'message_received',
};

// Create a singleton WebSocket instance
export let globalSocket = null;
let globalListeners = {};
let reconnectTimeout = null;
let pingInterval = null;
let reconnectAttempts = 0;
const MAX_RECONNECT_ATTEMPTS = 5;
const BASE_RECONNECT_DELAY = 3000;
let tabId = null;
const STORAGE_KEY = "ws_connection_info";

// Add debounce variables for status messages
let statusMessageTimeout = null;
const STATUS_DEBOUNCE_DELAY = 500; // 500ms debounce
let isOnline = navigator.onLine;
let lastPongTime = Date.now();

export const useWebSocket = () => {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState(null);
  const { isAuthenticated, token, loading } = useAuth();
  const eventListenersRef = useRef({});
  const lastStatusRef = useRef(null); // Track last sent status

  // Add a ref to track if we're in browser environment
  const isBrowser = useRef(typeof window !== "undefined");

  useEffect(() => {
    // Only run in browser environment
    if (!isBrowser.current) return;

    // Generate a unique ID for this browser tab if not already set
    if (!tabId) {
      tabId = `tab_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }

    // Only connect if authenticated and not already connected
    // Only connect if authenticated and not already connected
if (isAuthenticated && token && !loading && !globalSocket) {
  const connectWebSocket = () => {
    const baseUrl = BASE_URL.endsWith("/") ? BASE_URL.slice(0, -1) : BASE_URL;
    const wsUrl = `${baseUrl}/ws?token=${token}&tabId=${tabId}`;
    globalSocket = new WebSocket(wsUrl);

    globalSocket.onopen = () => {
      console.log("WebSocket connected");
      setIsConnected(true);
      window.__websocketConnected = true;
      showToast("WebSocket connection established", "success");
      reconnectAttempts = 0;

      sessionStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({
          tabId,
          connectedAt: Date.now(),
        })
      );

      if (pingInterval) clearInterval(pingInterval);
      pingInterval = setInterval(() => {
        if (globalSocket && globalSocket.readyState === WebSocket.OPEN) {
          globalSocket.send(JSON.stringify({ type: "ping" }));
        }
      }, 30000);
    };

    globalSocket.onclose = (event) => {
      console.log("WebSocket disconnected", event.code);
      setIsConnected(false);
      window.__websocketConnected = false;
      globalSocket = null;

      console.log("Event <code>:", event.code);

      // Prevent reconnect if unauthorized
      if (event.code === 1006) {
        console.warn("WebSocket closed due to unauthorized access. No reconnect.");
        showToast("Session expired or unauthorized. Please log in again.", "error");
        return;
      }

      // Reconnect unless normal closure (1000) or exceeded attempts
      if (event.code !== 1000) {
        reconnectAttempts++;

        if (reconnectAttempts <= MAX_RECONNECT_ATTEMPTS) {
          const delay = BASE_RECONNECT_DELAY * Math.pow(1.5, reconnectAttempts - 1);
          console.log(
            `Attempting to reconnect (${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS}) in ${delay / 1000} seconds`
          );

          reconnectTimeout = setTimeout(() => {
            connectWebSocket();
          }, delay);
        } else {
          console.log("Maximum reconnection attempts reached");
          showToast("Connection lost. Please reload the page to reconnect.", "error");
        }
      }
    };

    globalSocket.onerror = (err) => {
      console.error("WebSocket error:", err);

      // Optionally close the socket to trigger onclose
      if (globalSocket && globalSocket.readyState !== WebSocket.CLOSED) {
        globalSocket.close();
      }
    };

    globalSocket.onmessage = (event) => {
      try {
        const messages = event.data.split(/\n|\r\n/);

        for (const message of messages) {
          if (!message.trim()) continue;

          try {
            const data = JSON.parse(message);

            if (data.type === "ping" || data.type === "pong") {
              if (globalSocket && globalSocket.readyState === WebSocket.OPEN) {
                globalSocket.send(JSON.stringify({ type: "pong" }));
              }
              continue;
            }

            setLastMessage(data);

            const eventType = data.type;
            const payload = data.payload;

            if (globalListeners[eventType]) {
              globalListeners[eventType].forEach((callback) => {
                callback(payload);
              });
            }
          } catch (innerError) {
            console.error("Error parsing WebSocket message:", innerError, "Raw:", message);
          }
        }
      } catch (error) {
        console.error("Error processing WebSocket message:", error, "Raw:", event.data);
      }
    };
  };

  connectWebSocket();
}

    // Handle network changes
    const handleNetworkChange = () => {
      const wasOnline = isOnline;
      isOnline = navigator.onLine;

      console.log(`Network status changed: ${wasOnline ? 'online' : 'offline'} -> ${isOnline ? 'online' : 'offline'}`);

      if (!wasOnline && isOnline) {
        // Just came back online, attempt to reconnect
        console.log('Network restored, attempting to reconnect...');
        reconnectAttempts = 0; // Reset attempts on network recovery
        if (isAuthenticated && token && !loading && (!globalSocket || globalSocket.readyState !== WebSocket.OPEN)) {
          connectWebSocket();
        }
      } else if (wasOnline && !isOnline) {
        // Just went offline
        console.log('Network lost');
        setIsConnected(false);
      }
    };

    // Add network event listeners
    window.addEventListener('online', handleNetworkChange);
    window.addEventListener('offline', handleNetworkChange);

    // Function to send status message with debouncing
    const sendStatusMessage = (statusType) => {
      // Clear any pending status message
      if (statusMessageTimeout) {
        clearTimeout(statusMessageTimeout);
      }

      // Don't send if it's the same as the last status
      if (lastStatusRef.current === statusType) {
        return;
      }

      statusMessageTimeout = setTimeout(() => {
        if (globalSocket && globalSocket.readyState === WebSocket.OPEN) {
          console.log(`Sending ${statusType} message`);
          globalSocket.send(JSON.stringify({ type: statusType }));
          lastStatusRef.current = statusType;
        }
        statusMessageTimeout = null;
      }, STATUS_DEBOUNCE_DELAY);
    };

    // Handle tab visibility changes
    const handleVisibilityChange = () => {
      if (document.visibilityState === "hidden" && globalSocket) {
        // Notify server that user is away from this tab
        if (globalSocket.readyState === WebSocket.OPEN) {
          sendStatusMessage("user_away");
        }
      } else if (document.visibilityState === "visible") {
        // Reconnect or notify server that user is back
        if (!globalSocket || globalSocket.readyState !== WebSocket.OPEN) {
          // Close existing socket if it exists but isn't open
          if (globalSocket) {
            globalSocket.close();
            globalSocket = null;
          }

          // Only attempt to reconnect if authenticated
          if (isAuthenticated && token && !loading) {
            connectWebSocket();
          }
        } else if (globalSocket.readyState === WebSocket.OPEN) {
          // If socket is already open, notify server user is back
          sendStatusMessage("user_active");
        }
      }
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      window.removeEventListener('online', handleNetworkChange);
      window.removeEventListener('offline', handleNetworkChange);
      // Clear any pending status message timeout on unmount
      if (statusMessageTimeout) {
        clearTimeout(statusMessageTimeout);
      }
    };
  }, [isAuthenticated, token, loading]);

  // Subscribe to events
  const subscribe = useCallback((eventType, callback) => {
    if (!eventListenersRef.current[eventType]) {
      eventListenersRef.current[eventType] = [];
    }

    // Add to component-specific listeners
    eventListenersRef.current[eventType].push(callback);

    // Add to global listeners
    if (!globalListeners[eventType]) {
      globalListeners[eventType] = [];
    }
    globalListeners[eventType].push(callback);

    return () => {
      // Remove from global listeners
      if (globalListeners[eventType]) {
        globalListeners[eventType] = globalListeners[eventType].filter(
          (cb) => cb !== callback
        );
      }

      // Remove from component-specific listeners
      if (eventListenersRef.current[eventType]) {
        eventListenersRef.current[eventType] = eventListenersRef.current[
          eventType
        ].filter((cb) => cb !== callback);
      }
    };
  }, []);

  // Add this to ensure we're returning the most up-to-date connection status
  const getConnectionStatus = useCallback(() => {
    if (!isBrowser.current) return false;
    return window.__websocketConnected === true;
  }, []);

  return {
    isConnected: isBrowser.current ? getConnectionStatus() : false,
    lastMessage,
    subscribe,
    getConnectionStatus,
  };
};

export const closeWebSocketConnection = () => {
  // Check if we're in a browser environment
  if (typeof window === "undefined") return;

  if (globalSocket) {
    console.log("Closing WebSocket connection");
    globalSocket.close();
    globalSocket = null;
  }

  if (reconnectTimeout) {
    clearTimeout(reconnectTimeout);
    reconnectTimeout = null;
  }

  if (pingInterval) {
    clearInterval(pingInterval);
    pingInterval = null;
  }

  if (statusMessageTimeout) {
    clearTimeout(statusMessageTimeout);
    statusMessageTimeout = null;
  }

  globalListeners = {};
  reconnectAttempts = 0;
};
