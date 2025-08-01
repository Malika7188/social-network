'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { showToast } from "@/components/ui/ToastContainer";
// Add new imports for components
import GroupAbout from '@/components/groups/GroupAbout';
import GroupPhotos from '@/components/groups/GroupPhotos';
import GroupMembers from '@/components/groups/GroupMembers';
import GroupEvents from '@/components/groups/GroupEvents';
import { useParams } from 'next/navigation';
import Header from '@/components/header/Header';
import LeftSidebar from '@/components/sidebar/LeftSidebar';
import RightSidebar from '@/components/sidebar/RightSidebar';
import styles from '@/styles/page.module.css';
import groupStyles from '@/styles/GroupPage.module.css';
import GroupCreatePost from '@/components/groups/GroupCreatePost';
import GroupPost from '@/components/groups/Group-Posts';
// Add GroupChat to imports
import GroupChat from '@/components/groups/GroupChat';
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import { useGroupService } from "@/services/groupService";

const API_URL = process.env.API_URL || "http://localhost:8080/api";
const BASE_URL = API_URL.replace("/api", ""); // Remove '/api' to get the base URL

export default function GroupPostPage() {
  const params = useParams();
  const { groupId } = params;
  const router = useRouter();
  const { getgroup, getgroupposts, joinGroup } = useGroupService();
  const [group, setGroup] = useState(null);
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activeSection, setActiveSection] = useState('GroupPost');
  const [unreadMessages, setUnreadMessages] = useState(0);
  const [userData, setUserData] = useState(null);

   // Add handleNavClick function
  const handleNavClick = (e, section) => {
    e.preventDefault();
    setActiveSection(section);
  };
  
  // Move localStorage check into useEffect
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
      const result = await getgroup(groupId);
      const postresults = await getgroupposts(groupId);
      setGroup(result);
      setPosts(postresults);
      setLoading(false);
    } catch (error) {
      console.error("Error fetching groups:", error);
    }
  };

  useEffect(() => {
    fetchGroups();
  }, []);

  // Update the renderContent function to use userData instead of userdata
  const renderContent = () => {
    switch (activeSection) {
      case 'GroupPost':
        return (
          <>
            <GroupCreatePost groupId={groupId} groupName={group.Name} oncreatePost={fetchGroups} />
            {/* <GroupPost post={post} onPostUpdated={() => { }} /> */}
            {posts !== null && posts.map(post => (
              <div key={post.ID}>
                <GroupPost
                  currentuser={userData}
                  post={post}
                  onPostUpdated={fetchGroups}
                />
              </div>
            ))}
          </>
        );
      case 'AboutGroup':
        return <GroupAbout group={group} />;
      case 'photos':
        return <GroupPhotos posts={posts} />;
      case 'GroupMembers':
        let role = 'member'
        if (userData?.id === group.CreatorID) {
          role = 'admin';
        }
        return <GroupMembers group={group} currentUserRole={role} onMemberUpdate={fetchGroups} />;
      case 'GroupEvents':
        return <GroupEvents groupId={groupId} />;
      case 'GroupChat':
        return <GroupChat groupId={groupId} groupName={group.Name} />;
      default:
        return null;
    }
  };

  if (loading || !group) {
    return <LoadingSpinner size="large" fullPage={true} />;
  }

  const handleJoinRequest = async () => {
    try {
      await joinGroup(groupId);
      showToast("Join request sent successfully", "success");
      fetchGroups(); // Refresh group data
    } catch (error) {
      console.error("Error sending join request:", error);
      showToast("Failed to send join request", "error");
    }
  };

  if (!group || group.length === 0) {
    return (
      <>
        <Header />
        <div className={styles.container}>
          <aside className={styles.leftSidebar}>
            <LeftSidebar />
          </aside>
          <main className={`${styles.mainContent} ${groupStyles.emptyGroupContainer}`}>
            <div className={groupStyles.emptyGroup}>
              <i className="fas fa-users-slash"></i>
              <h2>Group Not Found</h2>
              <p>This group might be private or no longer exists.</p>
              <div className={groupStyles.emptyGroupActions}>
                <button 
                  className={groupStyles.joinButton}
                  onClick={handleJoinRequest}
                >
                  Request to Join
                </button>
                <button 
                  className={groupStyles.backButton}
                  onClick={() => router.push('/groups')}
                >
                  Back to Groups
                </button>
              </div>
            </div>
          </main>
          <aside className={styles.rightSidebar}>
            <RightSidebar />
          </aside>
        </div>
      </>
    );
  }

  return (
    <>
      <Header />
      <div className={styles.container}>
        <aside className={styles.leftSidebar}>
          <LeftSidebar />
        </aside>
        <main className={styles.mainContent}>
          <div className={groupStyles.groupHeader}>
            <img
              src={group.BannerPath?.String ?
                `${BASE_URL}/uploads/${group.BannerPath.String}` :
                "/banner.jpg"}
              alt={group.Name}
              className={groupStyles.groupBanner}
            />
            <div className={groupStyles.groupInfo}>
              <h1>{group.Name}</h1>
              <p>{group.Description}</p>
              <div className={groupStyles.groupMeta}>
                <span>{group.MemberCount.toLocaleString()} members</span>
              </div>
            </div>
            <div className={groupStyles.groupNav}>
              <nav>
                <a
                  href="#"
                  className={activeSection === 'GroupPost' ? styles.active : ''}
                  onClick={(e) => handleNavClick(e, 'GroupPost')}
                >
                  Posts
                </a>
                <a
                  href="#"
                  className={activeSection === 'AboutGroup' ? styles.active : ''}
                  onClick={(e) => handleNavClick(e, 'AboutGroup')}
                >
                  About
                </a>
                <a
                  href="#"
                  className={activeSection === 'photos' ? styles.active : ''}
                  onClick={(e) => handleNavClick(e, 'photos')}
                >
                  Photos
                </a>
                <a
                  href="#"
                  className={activeSection === 'GroupMembers' ? styles.active : ''}
                  onClick={(e) => handleNavClick(e, 'GroupMembers')}
                >
                  Members
                </a>
                <a
                  href="#"
                  className={activeSection === 'GroupEvents' ? styles.active : ''}
                  onClick={(e) => handleNavClick(e, 'GroupEvents')}
                >
                  Events
                </a>
                <a
                  href="#"
                  className={activeSection === 'GroupChat' ? styles.active : ''}
                  onClick={(e) => handleNavClick(e, 'GroupChat')}
                >
                  Chat
                  {unreadMessages > 0 && (
                    <span className={styles.unreadBadge}>{unreadMessages}</span>
                  )}
                </a>
              </nav>
            </div>
          </div>

          <div className={groupStyles.content}>
            {renderContent()}
          </div>
        </main>
        <aside className={styles.rightSidebar}>
          <RightSidebar />
        </aside>
      </div>
    </>
  );
}