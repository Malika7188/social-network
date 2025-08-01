'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import Header from '@/components/header/Header'
import ProtectedRoute from '@/components/auth/ProtectedRoute'
import LeftSidebar from '@/components/sidebar/LeftSidebar'
import styles from '@/styles/Friends.module.css'
import { useFriendService } from '@/services/friendService'
import { usePostService } from '@/services/postService'
import { showToast } from '@/components/ui/ToastContainer'


export default function FriendsPage() {
  const { followUser } = usePostService();

  const { friendRequests, SuggestedUsers, acceptFriendRequest, declineFriendRequest } = useFriendService()
  const [activeTab, setActiveTab] = useState('requests');
  const router = useRouter();

  const handleCardClick = (userId) => {
    router.push(`/profile/${userId}`);
  };

  const handleConfirm = async (friend) => {
    const success = await acceptFriendRequest(friend.followerId);
    if (success) {
      console.log(`Confirmed friend request from ${friend.name} (followerId: ${friend.followerId})`);
    }
  };

  const handleDecline = async (friend) => {
    const success = await declineFriendRequest(friend.followerId);
    if (success) {
      console.log(`Declined friend request from ${friend.name} (followerId: ${friend.followerId})`);
    }
  };

  const handleSendRequestOrFollow = async (user) => {
    await followUser(user.SuggestedID, user.name)
    router.push("/friends")
  };

  return (
    <ProtectedRoute>
      <Header />
      <div className={styles.container}>
        <aside className={styles.sidebar}>
          <LeftSidebar />
        </aside>
        <main className={styles.mainContent}>
          <div className={styles.friendsHeader}>
            {/* <h1>Friends</h1> */}
            <div className={styles.tabsContainer}>
              <button
                className={`${styles.tabButton} ${activeTab === 'requests' ? styles.activeTab : ''}`}
                onClick={() => setActiveTab('requests')}
              >
                Friend Requests <span className={styles.requestCount}>{friendRequests.length}</span>
              </button>
              <button
                className={`${styles.tabButton} ${activeTab === 'suggestions' ? styles.activeTab : ''}`}
                onClick={() => setActiveTab('suggestions')}
              >
                People You May Know <span className={styles.requestCount}>{SuggestedUsers.length}</span>
              </button>
            </div>
          </div>

          {activeTab === 'requests' ? (
            <>
              <h2 className={styles.sectionTitle}>Friend Requests</h2>
              <div className={styles.friendsGrid}>
                {friendRequests.map(friend => (
                  <div
                    key={friend.id}
                    className={styles.friendCard}
                    onClick={() => handleCardClick(friend.followerId)}
                    style={{ cursor: 'pointer' }}
                  >
                    <img
                      src={friend.image}
                      alt={friend.name}
                      className={styles.profileImage}
                    />
                    <h3 className={styles.friendName}>{friend.name}</h3>
                    <div className={styles.mutualFriends}>

                      <span>{friend.mutualFriends} mutual friends</span>
                    </div>
                    <div className={styles.actions}>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleConfirm(friend);
                        }}
                        className={styles.confirmButton}
                      >
                        Confirm
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleDecline(friend);
                        }}
                        className={styles.declineButton}
                      >
                        Decline
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </>
          ) : (
            <>
              <h2 className={styles.sectionTitle}>People You May Know</h2>
              <div className={styles.friendsGrid}>
                {SuggestedUsers.map(user => (
                  <div
                    key={user.SuggestedID}
                    className={styles.friendCard}
                    onClick={() => handleCardClick(user.SuggestedID)}
                    style={{ cursor: 'pointer' }}
                  >
                    <img
                      src={user.image}
                      alt={user.name}
                      className={styles.profileImage}
                    />
                    {user.isOnline && <div className={styles.onlineIndicator}></div>}
                    <h3 className={styles.friendName}>{user.name}</h3>
                    <div className={styles.mutualFriends}>
                      <span>{user.mutualFriends} mutual friends</span>
                    </div>
                    <div className={styles.actions}>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleSendRequestOrFollow(user);
                        }}
                        className={styles.confirmButton}
                      >
                        {user.isPublic ? 'Follow' : 'Send Friend Request'}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </>
          )}
        </main>
      </div>
    </ProtectedRoute>
  )
}