"use client";

import { useRouter } from 'next/navigation';
import Header from "@/components/header/Header";
import LeftSidebar from "@/components/sidebar/LeftSidebar";
import ProtectedRoute from "@/components/auth/ProtectedRoute";
import pageStyles from "@/styles/page.module.css";
import styles from "@/styles/Groups.module.css";
import CreateGroupModal from "@/components/groups/CreateGroupModal";
import { useState, useEffect } from 'react';
import { useGroupService } from "@/services/groupService";
import { showToast } from "@/components/ui/ToastContainer";
import InviteModal from '@/components/groups/InviteModal';

const API_URL = process.env.API_URL || "http://localhost:8080/api";
const BASE_URL = API_URL.replace("/api", "");

export default function Groups() {
    const router = useRouter();
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [selectedGroupForInvite, setSelectedGroupForInvite] = useState(null);
    const [allGroups, setAllGroups] = useState([]);
    const [loading, setLoading] = useState(true);
    const [userData, setUserData] = useState(null);
    const { getgrouponly, deleteGroup, leaveGroup, joinGroup, acceptInvitation, rejectInvitation } = useGroupService();

    // Handle user data
    useEffect(() => {
        const getUserData = () => {
            try {
                const raw = localStorage.getItem("userData");
                if (raw) {
                    const parsed = JSON.parse(raw);
                    setUserData(parsed);
                }
            } catch (e) {
                console.error("Invalid userData in localStorage:", e);
            } finally {
                setLoading(false);
            }
        };

        getUserData();
    }, []);

    // Fetch groups data
    useEffect(() => {
        if (!loading) {
            fetchGroups();
        }
    }, [loading]);

    const fetchGroups = async () => {
        try {
            const result = await getgrouponly();
            setAllGroups(result);
        } catch (error) {
            console.error("Error fetching groups:", error);
            showToast("Failed to fetch groups", "error");
        }
    };

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
                        showToast("Join request sent successfully", "success");
                    }
                    break;
                case 'accept':
                    success = await acceptInvitation(group.ID, userData?.id);
                    if (success) {
                        showToast("Invitation accepted successfully", "success");
                    }
                    break;
                case 'reject':
                    success = await rejectInvitation(group.ID, userData?.id);
                    if (success) {
                        showToast("Invitation rejected successfully", "success");
                    }
                    break;
            }

            if (success) {
                // Refresh groups list
                fetchGroups();
            }
        } catch (error) {
            console.error(`Error during ${action} action:`, error);
            showToast(`Failed to ${action} group`, "error");
        }
    };

    const handleGroupClick = (groupId) => {
        router.push(`/groups/${groupId}`);
    };

    // Add this helper function at the top of your component
    const getMembershipStatus = (group, userId) => {
        if (!group || !userId) return 'not_member';
        
        const member = group.Members.find(m => m.UserID === userId);
        if (!member) {
            return 'not_member';
        }

        // If status is pending, check InvitedBy
        if (member.Status === 'pending') {
            // If InvitedBy is not empty, user was invited
            // If InvitedBy is empty, user requested to join
            return member.InvitedBy ? 'invited' : 'pending';
        }

        // Status is 'accepted'
        return 'accepted';
    };

    return (
        <ProtectedRoute>
            <Header />
            <div className={styles.container}>
                <aside className={pageStyles.leftSidebar}>
                    <LeftSidebar />
                </aside>
                <main className={styles.mainContent}>
                    <div className={styles.groupsHeader}>
                        <h1>Groups</h1>
                        <button
                            className={styles.createGroupBtn}
                            onClick={() => setIsModalOpen(true)}
                        >
                            <i className="fas fa-plus"></i> Create New Group
                        </button>
                    </div>

                    <div className={styles.groupsGrid}>
                        {allGroups === null && (
                            <div className={styles.noGroups}>
                                <h2>No groups found</h2>
                                <p>Start by creating a new group!</p>
                            </div>
                        )}

                        {allGroups != null && allGroups.map(group => (
                            console.log("Group:", group), // Debugging line to check group data
                            <div
                                key={group.ID}
                                className={styles.groupCard}
                                onClick={() => handleGroupClick(group.ID)}
                            >
                                <div className={styles.groupBanner}>
                                    <img
                                        src={group.BannerPath?.String ?
                                            `${BASE_URL}/uploads/${group.BannerPath.String}` :
                                            "/banner5.jpg"}
                                        alt=""
                                        className={styles.bannerImg}
                                    />
                                </div>
                                <div className={styles.groupInfo}>
                                    <img src={group.ProfilePicPath?.String ? `${BASE_URL}/uploads/${group.ProfilePicPath.String}` : "/avatar.jpg"} alt="" className={styles.profilePic} />
                                    <h3 className={styles.groupName}>{group.Name}</h3>

                                    <div className={styles.memberInfo}>
                                        <div className={styles.memberAvatars}>
                                            {group.Members
                                                .filter(member => member.Status === 'accepted')
                                                .slice(0, 3)
                                                .map((member, index) => (
                                                    <img
                                                        key={member.ID || member.id || `member-${index}`}
                                                        src={member.Avatar ? `${BASE_URL}/uploads/${member.Avatar}` : "/avatar.jpg"}
                                                        alt={`${index + 1}`}
                                                        className={styles.memberAvatar}
                                                        style={{ zIndex: 3 - index }}
                                                    />
                                                ))
                                            }
                                        </div>
                                        <span className={styles.memberCount}>
                                            {group.MemberCount.toLocaleString()} members
                                        </span>
                                    </div>

                                    <hr className={styles.divider} />

                                    <div className={styles.groupActions} onClick={e => e.stopPropagation()}>
                                        {(() => {
                                            const membershipStatus = getMembershipStatus(group, userData?.id);
                                            console.log("Member Status:", membershipStatus); // Debugging

                                            switch (membershipStatus) {
                                                case 'accepted':
                                                    return (
                                                        <>
                                                            <button 
                                                                className={styles.inviteBtn} 
                                                                onClick={(e) => {
                                                                    e.stopPropagation();
                                                                    setSelectedGroupForInvite(group);
                                                                }}
                                                            >
                                                                <i className="fas fa-user-plus"></i> Invite
                                                            </button>
                                                            {userData?.id === group.Creator?.id ? (
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
                                                            )}
                                                        </>
                                                    );
                                                    
                                                case 'invited':
                                                    return (
                                                        <div className={styles.invitationActions}>
                                                            <button
                                                                className={styles.acceptBtn}
                                                                onClick={() => handleGroupAction(group, 'accept')}
                                                            >
                                                                Accept Invitation
                                                            </button>
                                                            <button
                                                                className={styles.rejectBtn}
                                                                onClick={() => handleGroupAction(group, 'reject')}
                                                            >
                                                                Reject Invitation
                                                            </button>
                                                        </div>
                                                    );
                                                    
                                                case 'pending':
                                                    return (
                                                        <button className={styles.pendingBtn} disabled>
                                                            Request Pending
                                                        </button>
                                                    );
                                                    
                                                default:
                                                    return (
                                                        <button
                                                            className={styles.joinBtn}
                                                            onClick={() => handleGroupAction(group, 'join')}
                                                        >
                                                            Request to Join
                                                        </button>
                                                    );
                                            }
                                        })()}
                                    </div>
                                </div>
                                {/* <InviteModal
                                    group={group}
                                    isOpen={selectedGroupForInvite === group}
                                    onClose={() => setSelectedGroupForInvite(null)}
                                /> */}
                            </div>
                        ))}
                    </div>
                </main>
            </div>

            {selectedGroupForInvite && (
                <InviteModal
                    group={selectedGroupForInvite}
                    isOpen={!!selectedGroupForInvite}
                    onClose={() => setSelectedGroupForInvite(null)}
                />
            )}

            <CreateGroupModal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                onGroupCreated={fetchGroups} // Refresh the page after creating a group
            />
        </ProtectedRoute>
    );
}