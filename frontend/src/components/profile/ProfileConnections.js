import React, { useState, useEffect } from "react";
import styles from "@/styles/ProfileConnections.module.css";
import ContactsList from "@/components/contacts/ContactsList";
import { useAuth } from "@/context/authcontext";
import { BASE_URL } from "@/utils/constants";

const ProfileConnections = ({ userData }) => {
  const [searchTerm, setSearchTerm] = useState("");
  const [showMore, setShowMore] = useState(false);
  const [activeDropdown, setActiveDropdown] = useState(null);
  const [activeTab, setActiveTab] = useState("following");

  // Initialize arrays with empty arrays instead of null
  const [following, setFollowing] = useState([]);
  const [followers, setFollowers] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  // Get auth context with authenticatedFetch
  const { authenticatedFetch, isAuthenticated } = useAuth();

  // Get follower/following counts - safely access length
  const followingCount = following?.length || 0;
  const followersCount = followers?.length || 0;

  // Fetch data from API
  useEffect(() => {
    const fetchData = async () => {
      if (!isAuthenticated) {
        setError("Authentication required");
        setIsLoading(false);
        return;
      }

      setIsLoading(true);
      try {
        // Use authenticatedFetch from AuthContext
        const followingResponse = await authenticatedFetch(`follow/following?userId=${userData.id}`);
        const followersResponse = await authenticatedFetch(`follow/followers?userId=${userData.id}`);

        if (!followingResponse.ok || !followersResponse.ok) {
          throw new Error("Failed to fetch data");
        }

        const followingData = await followingResponse.json();
        const followersData = await followersResponse.json();

        setFollowing(followingData);
        setFollowers(followersData);
        setError(null);
      } catch (err) {
        console.error("Error fetching connections data:", err);
        setError("Failed to load connections. Please try again later.");
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [authenticatedFetch, isAuthenticated]);

  const toggleDropdown = (contactId) => {
    setActiveDropdown(activeDropdown === contactId ? null : contactId);
  };

  // Get the appropriate data based on active tab
  const contacts = activeTab === "following" ? following : followers;

  // Apply pagination (when showMore is false, limit to 14 items)
  // Make sure contacts is an array before calling slice
  const displayedContacts = showMore
    ? contacts
    : contacts?.slice?.(0, 14) || [];

  // Filter based on search term
  const filteredContacts = Array.isArray(displayedContacts)
    ? displayedContacts.filter((contact) =>
      contact?.UserName?.toLowerCase().includes(searchTerm.toLowerCase())
    )
    : [];

  return (
    <div className={styles.connectionsContainer}>
      <div className={styles.mainContent}>
        <div className={styles.connectionsHeader}>
          <h2>My Connections</h2>
        </div>
        <div className={styles.connectionsStatsSearch}>
          <div className={styles.stats}>
            <span>{followingCount} Following</span>
            <span>•</span>
            <span>{followersCount} Followers</span>
          </div>
          <div className={styles.searchBar}>
            <i className="fas fa-search"></i>
            <input
              type="text"
              placeholder="Search connections..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>

        <div className={styles.tabsContainer}>
          <button
            className={`${styles.tabButton} ${activeTab === "following" ? styles.active : ""}`}
            onClick={() => setActiveTab("following")}
          >
            Following
          </button>
          <button
            className={`${styles.tabButton} ${activeTab === "followers" ? styles.active : ""}`}
            onClick={() => setActiveTab("followers")}
          >
            Followers
          </button>
        </div>

        {isLoading ? (
          <div className={styles.loadingState}>
            <p>Loading connections...</p>
          </div>
        ) : error ? (
          <div className={styles.errorState}>
            <p>{error}</p>
          </div>
        ) : (
          <>
            <div className={styles.contactsGrid}>
              {filteredContacts.length > 0 ? (
                filteredContacts.map((contact) => (
                  <div
                    key={`conn-${contact.ID ?? `${contact.FollowerID}-${contact.FollowingID}`}`}
                    className={styles.contactCard}
                  >
                    <div className={styles.contactInfo}>
                      <img
                        src={
                          contact.UserAvatar
                            ? `${BASE_URL}/uploads/${contact.UserAvatar}`
                            : "/avatar.png"
                        }
                        alt="avatar"
                        className={styles.avatar}
                      />

                      <div className={styles.details}>
                        <h3>{contact.UserName}</h3>
                        <span className={contact.IsOnline ? styles.online : ""}>
                          {contact.IsOnline ? "Online" : "Offline"}
                        </span>
                      </div>
                    </div>
                  </div>
                ))
              ) : (
                <div className={styles.noResults}>
                  <p>
                    No {activeTab} found
                    {searchTerm ? ' matching "' + searchTerm + '"' : ""}.
                  </p>
                </div>
              )}
            </div>

            {!showMore && (contacts?.length || 0) > 14 && (
              <button
                className={styles.loadMoreButton}
                onClick={() => setShowMore(true)}
              >
                Load More
              </button>
            )}
          </>
        )}
      </div>
      <div className={styles.sidebar}>
        <ContactsList />
      </div>
    </div>
  );
};

export default ProfileConnections;
