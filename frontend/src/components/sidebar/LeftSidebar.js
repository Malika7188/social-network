"use client";

import Link from "next/link";
import styles from "@/styles/Sidebar.module.css";
import { usePathname } from "next/navigation";
import { useWebSocket, EVENT_TYPES } from "@/services/websocketService";
import { useEffect, useState } from "react";
import { useAuth } from "@/context/authcontext";

// Define base URL for media assets
const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";
const BASE_URL = API_URL.replace("/api", ""); // Remove '/api' to get the base URL

export default function LeftSidebar() {
  const pathname = usePathname();
  const [person, setPerson] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const { subscribe, isConnected } = useWebSocket();

  // Get authenticatedFetch from auth context
  const { authenticatedFetch, isAuthenticated } = useAuth();

  useEffect(() => {
    const fetchUserData = async () => {
      if (!isAuthenticated) {
        setError("Authentication required");
        setIsLoading(false);
        return;
      }

      try {
        const response = await authenticatedFetch("users/me");

        if (!response.ok) {
          throw new Error("Failed to fetch user data");
        }

        const userData = await response.json();
        setPerson(userData);
        setError(null);
      } catch (err) {
        console.error("Error fetching user data in sidebar:", err);
        setError("Failed to load user data");

        // Fallback to localStorage if API fails
        const storedUser = JSON.parse(localStorage.getItem("userData") || "{}");
        if (storedUser && Object.keys(storedUser).length > 0) {
          setPerson(storedUser);
        }
      } finally {
        setIsLoading(false);
      }
    };

    fetchUserData();
  }, [authenticatedFetch, isAuthenticated]);

  useEffect(() => {
    // Only set up WebSocket subscription if connected and user data is available
    if (!isConnected || !person) {
      return;
    }

    // Subscribe to user stats updates
    const handleUserStatsUpdate = (payload) => {
      if (person && payload.userId === person.id) {
        setPerson((prevPerson) => {
          const updatedPerson = { ...prevPerson };

          // Update the specific stat that changed
          if (payload.statsType === "Posts") {
            updatedPerson.numPosts = payload.count;
          } else if (payload.statsType === "Followers") {
            updatedPerson.followersCount = payload.count;
          } else if (payload.statsType === "Following") {
            updatedPerson.followingCount = payload.count;
          } else if (payload.statsType === "Groups") {
            updatedPerson.groupsJoined = payload.count;
          }

          return updatedPerson;
        });
      }
    };

    const unsubscribe = subscribe(
      EVENT_TYPES.USER_STATS_UPDATED,
      handleUserStatsUpdate
    );

    return () => {
      if (unsubscribe) unsubscribe();
    };
  }, [subscribe, isConnected, person]);

  if (isLoading) {
    return (
      <div className={styles.sidebar}>
        <div className={styles.loadingIndicator}>Loading...</div>
      </div>
    );
  }

  if (error && !person) {
    return (
      <div className={styles.sidebar}>
        <div className={styles.errorMessage}>{error}</div>
      </div>
    );
  }

  if (!person) return null;

  const stats = [
    { label: "Posts", count: person.numPosts },
    { label: "Groups", count: person.groupsJoined },
    { label: "Followers", count: person.followersCount },
    { label: "Following", count: person.followingCount },
  ];

  const navLinks = [
    { icon: "fas fa-home", label: "Home", path: "/home" },
    { icon: "fas fa-users", label: "Groups", path: "/groups" },
    // { icon: "fas fa-calendar-alt", label: "Events", path: "/events" },
    { icon: "fas fa-user-friends", label: "Friends", path: "/friends" },
    { icon: "fas fa-envelope", label: "Messages", path: "/messages" },
    { icon: "fas fa-user", label: "Profile", path: "/profile" },
    // { icon: "fa-solid fa-gear", label: "Settings", path: "/settings" },
  ];

  return (
    <div className={styles.sidebar}>
      <div className={styles.profileSection}>
        <img
          src={
            person.avatar
              ? `${BASE_URL}/uploads/${person.avatar}`
              : "/avatar4.png"
          }
          alt="Profile"
          className={styles.profilePic}
        />
        <h2 className={styles.userName}>{person.fullName}</h2>
        <p className={styles.userProfession}>{person.aboutMe}</p>

        <div className={styles.statsGrid}>
          {stats.map((stat) => (
            <div key={stat.label} className={styles.statItem}>
              <span className={styles.statCount}>{stat.count || 0}</span>
              <span className={styles.statLabel}>{stat.label}</span>
            </div>
          ))}
        </div>
      </div>

      <nav className={styles.sidebarNav}>
        {navLinks.map((link) => (
          <Link
            key={link.path}
            href={link.path}
            className={`${styles.navLink} ${
              pathname === link.path ? styles.active : ""
            }`}
          >
            <i className={link.icon}></i>
            <span>{link.label}</span>
          </Link>
        ))}
      </nav>
    </div>
  );
}
