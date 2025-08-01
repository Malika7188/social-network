"use client";

import React, { createContext, useContext } from "react";
import { useUserStatus } from "@/services/userStatusService";

const UserStatusContext = createContext(null);

export const UserStatusProvider = ({ children }) => {
  const userStatusService = useUserStatus();

  return (
    <UserStatusContext.Provider value={userStatusService}>
      {children}
    </UserStatusContext.Provider>
  );
};

export const useUserStatusContext = () => {
  const context = useContext(UserStatusContext);
  if (!context) {
    throw new Error("useUserStatusContext must be used within a UserStatusProvider");
  }
  return context;
};
