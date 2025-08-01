import React from 'react';
import groupStyles from '@/styles/Groups.module.css';
import ContactsList from '@/components/contacts/ContactsList';
import { useGroupService } from '@/services/groupService';
import { useRouter } from 'next/navigation';
import { useState, useEffect } from 'react';
import pageStyles from "@/styles/page.module.css";
import styles from "@/styles/Groups.module.css";
import { showToast } from '../ui/ToastContainer';
const API_URL = process.env.API_URL || "http://localhost:8080/api";
const BASE_URL = API_URL.replace("/api", ""); // Remove '/api' to get the base URL


const ProfileGroups = ({ userData }) => {
  const router = useRouter();
  const [userGroups, setUserGroups] = useState([]);
  const { getusergroups, deleteGroup, leaveGroup, joinGroup } = useGroupService();

  let viewerdata = null;
  try {
    const raw = localStorage.getItem("viewerData");
    if (raw) viewerdata = JSON.parse(raw);
  } catch (e) {
    console.error("Invalid viewerData in localStorage:", e);
  }
  const fetchGroups = async () => {
    try {
      const result = await getusergroups(userData.id)
      setUserGroups(result)
    } catch (error) {
      console.error("Error fetching groups:", error);
    }

  }

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
            showToast("Joined group successfully", "success");
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
    <div className={styles.groupsContainer}>
      <div className={styles.mainContent}>
        <div className={groupStyles.groupsHeader}>
          <h2>My Groups</h2>
        </div>

        <div className={styles.groupsGrid}>
          {userGroups === null && (
            <div className={styles.noGroups}>
              <h2>No groups found</h2>
              <p>Start by creating a new group!</p>
            </div>
          )}

          {userGroups != null && userGroups.map(group => (
            <div
              key={group.ID}
              className={styles.groupCard}
              onClick={() => handleGroupClick(group.ID)}
            >
              <div className={styles.groupBanner}>
                <img
                  src={group.BannerPath?.String ?
                    `${BASE_URL}/uploads/${group.BannerPath.String}` :
                    "/banner.png"}
                  alt=""
                  className={styles.bannerImg}
                />
              </div>
              <div className={styles.groupInfo}>
                <img src={group.ProfilePicPath?.String ? `${BASE_URL}/uploads/${group.ProfilePicPath.String}` : "/avatar.jpg"} alt="" className={styles.profilePic} />
                <h3 className={styles.groupName}>{group.Name}</h3>
                <span className={styles.groupPrivacy}>
                  <i className={`fas ${group.IsPublic ? 'fa-globe' : 'fa-lock'}`}></i>
                  {group.IsPublic ? 'Public Group' : 'Private Group'}
                </span>

                <div className={styles.memberInfo}>
                  <div className={styles.memberAvatars}>
                    {group.Members.slice(0, 3).map((member, index) => (
                      <img
                        key={member.ID || member.id || `member-${index}`} // Handle both ID and id cases with fallback
                        src={member.Avatar ? `${BASE_URL}/uploads/${member.Avatar}` : "/avatar.jpg"}
                        alt={`Member ${index + 1}`}
                        className={styles.memberAvatar}
                        style={{ zIndex: 3 - index }}
                      />
                    ))}
                  </div>
                  <span className={styles.memberCount}>
                    {group.MemberCount.toLocaleString()} members
                  </span>
                </div>

                <hr className={styles.divider} />

                <div className={styles.groupActions} onClick={e => e.stopPropagation()}>
                  {group.IsMember && (
                    <button className={styles.inviteBtn}>
                      <i className="fas fa-user-plus"></i> Invite
                    </button>
                  )}

                  {group.IsMember ? (
                    viewerdata?.id === group.Creator?.id ? (
                      <button
                        className={styles.leaveBtn}
                        onClick={() => handleGroupAction(group, 'delete')}
                      >
                        Delete Group
                      </button>
                    ) : (
                      <button
                        className={styles.leaveBtn}
                        onClick={() => handleGroupAction(group, 'leave')}
                      >
                        Leave Group
                      </button>
                    )
                  ) : (
                    <button
                      className={styles.inviteBtn}
                      onClick={() => handleGroupAction(group, 'join')}
                    >
                      Join Group
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default ProfileGroups; 