"use client";

import styles from "@/styles/Sidebar.module.css";
import Link from "next/link";
import { useRouter, usePathname } from "next/navigation";
import { useFriendService } from "@/services/friendService";
import { useGroupService } from "@/services/groupService";
import { useEffect, useState } from "react";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import { useUserStatus } from "@/services/userStatusService";
import ContactsSection from "@/components/contacts/ContactsList";
import { showToast } from "@/components/ui/ToastContainer";

const API_URL = process.env.API_URL || "http://localhost:8080/api";
const BASE_URL = API_URL.replace("/api", "");
// Separate components for better organization
const FriendRequestSection = ({
  friendRequests,
  onAccept,
  onDecline,
  isLoading,
}) => (
  <section className={styles.friendRequests}>
    <h2>Friend Requests</h2>
    {isLoading ? (
      <div className={styles.loadingContainer}>
        <LoadingSpinner size="small" color="primary" />
      </div>
    ) : friendRequests.length === 0 ? (
      <p className={styles.emptyState}>No pending friend requests</p>
    ) : (
      friendRequests.map((request) => (
        <div key={request.id} className={styles.requestCard}>
          <div className={styles.requestProfile}>
            <img src={request.image} alt={request.name} />
            <div className={styles.requestInfo}>
              <h3>{request.name}</h3>
              <span>{request.mutualFriends} mutual friends</span>
            </div>
          </div>
          <div className={styles.requestActions}>
            <button
              className={styles.acceptButton}
              onClick={() => onAccept(request.followerId)}
            >
              Accept
            </button>
            <button
              className={styles.declineButton}
              onClick={() => onDecline(request.followerId)}
            >
              Decline
            </button>
          </div>
        </div>
      ))
    )}
  </section>
);
const GroupsSection = ({ groups, onJoin }) => (
  <section className={styles.groups}>
    <h2>Groups</h2>
    {groups
      .filter(group => !group.IsMember)
      .slice(0, 3)
      .map((group) => (
        <div key={group.ID} className={styles.groupItem}>
          <div className={styles.groupProfile}>
            <img src={group.ProfilePicPath?.String ?
              `${BASE_URL}/uploads/${group.ProfilePicPath.String}` :
              "/avatar.jpg"}
              alt={group.Name} />
            <div className={styles.groupInfo}>
              <h3>{group.Name}</h3>
              <span>{group.MemberCount} members</span>
            </div>
          </div>
          <div className={styles.joinActions}>
            <button
              className={styles.joinButton}
              onClick={() => onJoin(group.ID)}
            >Request to Join</button>
          </div>
        </div>
      ))}
    <Link href="/groups">
      <div className={styles.viewGroup}>
        <button className={styles.viewButton}>
          See All Groups <i className="fas fa-arrow-right"></i>
        </button>
      </div>
    </Link>
  </section>
);

export default function RightSidebar() {
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const {
    friendRequests,
    contacts,
    acceptFriendRequest,
    declineFriendRequest,
    isLoadingRequests,
    isLoadingContacts,
  } = useFriendService();
  const { getgrouponly, joinGroup } = useGroupService();
  const router = useRouter();

  // Limit contacts to display (to avoid cluttering the sidebar)
  const displayedContacts = contacts.slice(0, 10);

  useEffect(() => {
    fetchGroups();
  }, []);

  const fetchGroups = async () => {
    try {
      const fetchedGroups = await getgrouponly();
      setGroups(fetchedGroups);
    } catch (error) {
      console.error("Error fetching groups:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleJoinGroup = async (group) => {
    try {
      const success = await joinGroup(group);
      if (success) {
        showToast("Joined group successfully", "success");
        fetchGroups(); // Refresh the groups list
      }
    } catch (error) {
      console.error("Error joining group:", error);
      showToast("Failed to join group", "error");
    }
  };

  return (
    <div className={styles.rightSidebar}>
      <FriendRequestSection
        friendRequests={friendRequests}
        onAccept={acceptFriendRequest}
        onDecline={declineFriendRequest}
        isLoading={isLoadingRequests}
      />

      <ContactsSection
        contacts={displayedContacts}
        isLoading={isLoadingContacts}
        isProfilePage={false}
      />

      {loading ? (
        <div className={styles.loadingContainer}>
          <LoadingSpinner size="small" color="primary" />
        </div>
      ) : (
        <GroupsSection groups={groups} onJoin={handleJoinGroup} />
      )}
    </div>
  );
}
