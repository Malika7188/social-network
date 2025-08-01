"use client";

import React, { createContext, useContext, useEffect } from "react";
import { useWebSocket } from "@/services/websocketService";

const WebSocketContext = createContext(null);

export const WebSocketProvider = ({ children }) => {
  const websocketService = useWebSocket();


  return (
    <WebSocketContext.Provider value={websocketService}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocketContext = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error(
      "useWebSocketContext must be used within a WebSocketProvider"
    );
  }
  return context;
};