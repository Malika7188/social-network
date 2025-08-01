"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import styles from "@/styles/Header.module.css";
import { useAuth } from "@/context/authcontext";
import { showToast } from "@/components/ui/ToastContainer";

const API_URL = process.env.API_URL || "http://localhost:8080/api";
const BASE_URL = API_URL.replace("/api", ""); // Remove '/api' to get the base URL

export default function ProfileDropdown() {
  const [isOpen, setIsOpen] = useState(false);
  const [isDarkMode, setIsDarkMode] = useState(false);
  const router = useRouter();
  const { logout, currentUser } = useAuth();

  const toggleDropdown = () => setIsOpen(!isOpen);
  const toggleDarkMode = () => setIsDarkMode(!isDarkMode);

  const handleLogout = () => {
    logout();
    showToast("Logged out successfully!", "success");
  };

  // Use currentUser data if available, otherwise fallback to sample data
  const userName = currentUser
    ? `${currentUser.firstName} ${currentUser.lastName}`
    : "John Doe";
  const userRole = currentUser?.nickname || "Web Developer";

  // Handle avatar URL construction
  const getAvatarUrl = () => {
    if (!currentUser?.avatar) return "/avatar.png"; // Default avatar

    // If avatar is a full URL, use it directly
    if (currentUser.avatar.startsWith("http")) {
      return currentUser.avatar;
    }

    // If avatar is a relative path, construct the full URL
    if (currentUser.avatar.startsWith("/uploads/")) {
      return `${BASE_URL}${currentUser.avatar}`;
    }

    // If avatar doesn't have the /uploads/ prefix, add it
    return `${BASE_URL}/uploads/${currentUser.avatar}`;
  };

  const userAvatar = getAvatarUrl();

  return (
    <div className={styles.profileDropdownContainer}>
      <img
        src={userAvatar}
        alt="Profile"
        className={styles.profileIcon}
        onClick={toggleDropdown}
      />

      {isOpen && (
        <div className={styles.dropdownMenu}>
          <div className={styles.profileHeader}>
            <img
              src={userAvatar}
              alt="Profile"
              className={styles.dropdownProfilePic}
            />
            <div className={styles.profileInfo}>
              <h3>{userName}</h3>
              <span>@{userRole}</span>
            </div>
          </div>

          <div className={styles.dropdownDivider} />

          <Link href="/profile" className={styles.dropdownItem}>
            <i className="fas fa-user"></i>
            View Profile
          </Link>

          <div className={styles.dropdownDivider} />

          <button onClick={toggleDarkMode} className={styles.dropdownItem}>
            <i className={isDarkMode ? "fas fa-sun" : "fas fa-moon"}></i>
            {isDarkMode ? "Light Mode" : "Dark Mode"}
          </button>

          <div className={styles.dropdownDivider} />

          <button onClick={handleLogout} className={styles.dropdownItem}>
            <i className="fas fa-sign-out-alt"></i>
            Sign Out
          </button>
        </div>
      )}
    </div>
  );
}
