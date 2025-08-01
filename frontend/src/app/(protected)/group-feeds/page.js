'use client';

import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import Header from '@/components/header/Header';
import CreatePost from '@/components/posts/CreatePost';
import LeftSidebar from '@/components/sidebar/LeftSidebar';
import RightSidebar from '@/components/sidebar/RightSidebar';
import styles from '@/styles/page.module.css';
import groupFeeds from "@/styles/GroupFeeds.module.css";
import GroupPost from '@/components/groups/Group-Posts';
import { useGroupService } from '@/services/groupService';
import CreateGroupModal from "@/components/groups/CreateGroupModal";
import { showToast } from "@/components/ui/ToastContainer";

const API_URL = process.env.API_URL || "http://localhost:8080/api";
const BASE_URL = API_URL.replace("/api", "");

export default function GroupFeeds() {
  const router = useRouter();
  const { getallgroups, deleteGroup, joinGroup, leaveGroup } = useGroupService();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [userData, setUserData] = useState(null);

  // Handle user data
  useEffect(() => {
    try {
      const raw = localStorage.getItem("userData");
      if (raw) {
        const parsed = JSON.parse(raw);
        setUserData(parsed);
      }
    } catch (e) {
      console.error("Invalid userData in localStorage:", e);
    }
  }, []);

  const fetchGroups = async () => {
    try {
      const fetchedGroups = await getallgroups();
      setGroups(fetchedGroups);
    } catch (error) {
      console.error('Error fetching groups:', error);
      showToast("Failed to fetch groups", "error");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchGroups();
  }, []);

  const handleGroupAction = async (group, action) => {
    try {
      let success = false;

      switch (action) {
        case 'delete':
          success = await deleteGroup(group.ID);
          if (success) {
            showToast("Group deleted successfully", "success");
          }
          break;
        case 'leave':
          success = await leaveGroup(group.ID);
          if (success) {
            showToast("Left group successfully", "success");
          }
          break;
        case 'join':
          success = await joinGroup(group.ID);
          if (success) {
            showToast("Join Request sent successfully", "success");
          }
          break;
      }

      if (success) {
        // Refresh groups list
        fetchGroups();
      }
    } catch (error) {
      console.error(`Error during ${action} action:`, error);
    }
  };

  const handleGroupClick = (groupId) => {
    router.push(`/groups/${groupId}`);
  };

  return (
    <>
      <Header />
      <div className={styles.container}>
        <aside>
          <LeftSidebar />
        </aside>
        <main className={styles.mainContent}>
          <div className={styles.groupsHeader}>
            <h1>Group Feeds</h1>
            <button
              className={styles.createGroupBtn}
              onClick={() => setIsModalOpen(true)}
            >
              <i className="fas fa-plus"></i> Create New Group
            </button>
          </div>
          {groups.map(group => (
            <div key={group.ID} className={groupFeeds.groupSection}>
              <div className={groupFeeds.groupHeader}>
                <div className={groupFeeds.groupInfo}>
                  <img
                    src={group.ProfilePicPath?.String ? `${BASE_URL}/uploads/${group.ProfilePicPath.String}` : "/avatar.jpg"}
                    alt={group.Name}
                    className={groupFeeds.groupBanner}
                  />
                  <div className={groupFeeds.groupDetails}>
                    <h2
                      onClick={() => handleGroupClick(group.ID)}
                      style={{ cursor: 'pointer' }}
                    >
                      {group.Name}
                    </h2>
                    <p>{group.Description}</p>
                    <div className={groupFeeds.groupMeta}>
                      <span>{group.MemberCount.toLocaleString()} members</span>
                    </div>
                  </div>
                </div>
                <div className={groupFeeds.groupActions}>
                  {group.IsMember ? (
                    userData.id === group.Creator.id ? (
                      <button className={groupFeeds.Join} onClick={() => handleGroupAction(group, 'delete')}>
                        Delete Group
                      </button>
                    ) : (
                      <button className={groupFeeds.Join} onClick={() => handleGroupAction(group, 'leave')}>
                        Leave Group
                      </button>
                    )
                  ) : (
                    <button className={groupFeeds.Join} onClick={() => handleGroupAction(group, 'join')}>
                      Request to Join Group
                    </button>
                  )}
                  {/* <button className={groupFeeds.moreButton}>
                    •••
                  </button> */}
                </div>
              </div>

              {group.posts.slice(0, 2).map(post => (
                <div key={post.ID}>
                  <GroupPost
                    post={post}
                    onPostUpdated={fetchGroups}
                  />
                </div>
              ))}
            </div>
          ))}
        </main>
        <aside>
          <RightSidebar />
        </aside>
      </div>

      <CreateGroupModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onGroupCreated={fetchGroups}
      />
    </>
  );
}
