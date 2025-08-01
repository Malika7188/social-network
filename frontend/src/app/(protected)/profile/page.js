"use client";

import React, { useState, useEffect, useCallback, useRef } from "react";
import Header from "@/components/header/Header";
import styles from "@/styles/ProfilePage.module.css";
import { use } from "react";
import ProfileBanner from "@/components/profile/ProfileBanner";
import ProfileAboutSideBar from "@/components/profile/ProfileAboutSideBar";
import CreatePost from "@/components/posts/CreatePost";
import PostList from "@/components/posts/PostList";
import ProfilePhotosGrid from "@/components/profile/ProfilePhotosGrid";
import ContactsSection from "@/components/contacts/ContactsList";
import ProfileAbout from "@/components/profile/ProfileAbout";
import ProfilePhotos from "@/components/profile/ProfilePhotos";
import ProfileGroups from "@/components/profile/ProfileGroups";
import ProfileConnections from "@/components/profile/ProfileConnections";
import { useFriendService } from "@/services/friendService";
import { useAuth } from "@/context/authcontext";
import { usePostService } from "@/services/postService";
import { showToast } from "@/components/ui/ToastContainer";

// Define base URL for media assets
const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";
const BASE_URL = API_URL.replace("/api", ""); // Remove '/api' to get the base URL

export default function ProfilePage({ params }) {
  const { getUserPhotos } = usePostService();

  const [photos, setPhotos] = useState([]);
  const [userData, setUserData] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const [activeSection, setActiveSection] = useState("posts");
  const [photosLoading, setPhotosLoading] = useState(true);
  const [photosError, setPhotosError] = useState(null);

  // Create ref for component mounting state
  const isMounted = useRef(true);
  // Flag to track if photos have been loaded
  const photosLoadedRef = useRef(false);

  // Get authenticatedFetch from auth context
  const { authenticatedFetch, isAuthenticated } = useAuth();

  const resolvedParams = use(params);
  const { contacts, isLoadingContacts } = useFriendService();

  // Set isMounted.current to true on initial render and to false on cleanup
  useEffect(() => {
    isMounted.current = true;

    return () => {
      isMounted.current = false;
    };
  }, []);

  // Fetch user data when component mounts
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

        const data = await response.json();

        if (isMounted.current) {
          setUserData(data);
          setError(null);
        }
      } catch (err) {
        console.error("Error fetching user data:", err);

        if (isMounted.current) {
          setError("Failed to load user profile. Please try again later.");

          // Fallback to localStorage if API fails
          const storedUser = JSON.parse(
            localStorage.getItem("userData") || "{}"
          );
          if (storedUser && Object.keys(storedUser).length > 0) {
            setUserData(storedUser);
          }
        }
      } finally {
        if (isMounted.current) {
          setIsLoading(false);
        }
      }
    };

    fetchUserData();
  }, [authenticatedFetch, isAuthenticated]);

  // Function to load user photos - defined with useCallback to prevent recreation
  const loadUserPhotos = useCallback(async () => {

    if (!userData || !userData.id) {
      if (isMounted.current) {
        setPhotosLoading(false);
      }
      return;
    }

    try {
      // Set loading state before fetching
      if (isMounted.current) {
        setPhotosLoading(true);
      }

      // Fetch photos
      const userPhotos = await getUserPhotos(userData.id);

      // Only update state if component is still mounted
      if (isMounted.current) {
        // Ensure data is an array
        if (!userPhotos || !Array.isArray(userPhotos)) {
          setPhotos([]);
        } else {
          setPhotos(userPhotos);
        }
        // Reset error state if successful
        setPhotosError(null);
        // Mark photos as loaded
        photosLoadedRef.current = true;
      }
    } catch (err) {
      console.error("Error fetching user photos:", err);
      if (isMounted.current) {
        setPhotosError(err.message || "Failed to load photos");
        showToast("Failed to load photos", "error");
      }
    } finally {
      // Always set loading to false when done
      if (isMounted.current) {
        setPhotosLoading(false);
      }
    }
  }, [userData?.id, getUserPhotos]);

  // Load photos once when userData becomes available
  useEffect(() => {
    // Load photos only if:
    // 1. We have userData with an ID
    // 2. Photos haven't been loaded yet
    if (userData?.id && !photosLoadedRef.current) {
      loadUserPhotos();
    }
  }, [userData, loadUserPhotos]);

  // Function to refresh photos (can be called after updates)
  const refreshPhotos = useCallback(() => {
    if (userData?.id) {
      // Reset the flag to allow loading again
      photosLoadedRef.current = false;
      loadUserPhotos();
    }
  }, [loadUserPhotos, userData?.id]);

  const renderContent = () => {
    // If still loading data, show loading state
    if (isLoading) {
      return (
        <div className={styles.loadingContainer}>Loading profile data...</div>
      );
    }

    // If error and no fallback data, show error
    if (error && !userData) {
      return <div className={styles.errorContainer}>{error}</div>;
    }

    switch (activeSection) {
      case "about":
        return (
          <ProfileAbout
            photos={photos}
            isLoading={photosLoading}
            error={photosError}
            BASE_URL={BASE_URL}
            userData={userData}
          />
        );
      case "posts":
        return (
          <div className={styles.contentLayout}>
            <div className={styles.leftSidebar}>
              <ProfileAboutSideBar userData={userData} />
            </div>
            <div className={styles.mainContent}>
              <CreatePost />
              <PostList userData={userData} />
            </div>
            <div className={styles.rightSidebar}>
              <ProfilePhotosGrid
                photos={photos}
                totalPhotos={photos.length}
                isLoading={photosLoading}
                error={photosError}
                BASE_URL={BASE_URL}
              />
              <ContactsSection
                contacts={contacts}
                isLoading={isLoadingContacts}
                isProfilePage={true}
              />
            </div>
          </div>
        );
      case "photos":
        return (
          <ProfilePhotos
            photos={photos}
            isLoading={photosLoading}
            error={photosError}
            onRefresh={refreshPhotos}
            BASE_URL={BASE_URL}
          />
        );
      case "groups":
        return <ProfileGroups userData={userData} />;
      case "connections":
        return <ProfileConnections userData={userData} />;
      default:
        return null;
    }
  };

  return (
    <>
      <Header />
      <div className={styles.profileContainer}>
        {userData && (
          <ProfileBanner
            userData={userData}
            onNavClick={setActiveSection}
            activeSection={activeSection}
            isOwnProfile={true}
            BASE_URL={BASE_URL}
          />
        )}
        {renderContent()}
      </div>
    </>
  );
}
