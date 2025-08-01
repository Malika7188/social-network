import { useState } from 'react';
import styles from '@/styles/GroupMembers.module.css';
import { useGroupService } from '@/services/groupService';
import { showToast } from '@/components/ui/ToastContainer';

const API_URL = process.env.API_URL || "http://localhost:8080/api";
const BASE_URL = API_URL.replace("/api", "");// Replace with your actual base URL

const GroupMembers = ({ group, currentUserRole, onMemberUpdate }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [activeTab, setActiveTab] = useState('all');
  const [currentPage, setCurrentPage] = useState(1);
  const membersPerPage = 14;
  const { acceptJoinRequest, rejectJoinRequest } = useGroupService();

  // Convert raw member data into usable format
  const members = (group?.Members || []).filter(member => member.Status === 'accepted').map((member) => ({
    id: member.User?.id,
    name: `${member.User?.firstName || ''} ${member.User?.lastName || ''}`,
    avatar: member.Avatar
      ? `${BASE_URL}/uploads/${member.Avatar}`
      : '/avatar.jpg',
    role: member.Role?.toLowerCase() === 'admin' ? 'Admin' : 'Member',
    joinDate: member.CreatedAt,
    isOnline: false,
    status: member.Status
  }));

  // Get pending requests
  const pendingRequests = (group?.Members || [])
    .filter(member => member.Status === 'pending')
    .map((member) => ({
      id: member.User?.id,
      name: `${member.User?.firstName || ''} ${member.User?.lastName || ''}`,
      avatar: member.Avatar
        ? `${BASE_URL}/uploads/${member.Avatar}`
        : '/avatar.jpg',
      joinDate: member.CreatedAt
    }));

  const handleRequestAction = async (userId, action) => {
    try {
      if (action === 'accept') {
        await acceptJoinRequest(group.ID, userId);
        showToast('Request accepted successfully', 'success');
      } else {
        await rejectJoinRequest(group.ID, userId);
        showToast('Request rejected successfully', 'success');
      }
      // You might want to refresh the group data here
      if (onMemberUpdate) onMemberUpdate();
    } catch (error) {
      console.error(`Error ${action}ing request:`, error);
      showToast(`Failed to ${action} request`, 'error');
    }
  };

  // Filter by search
  let filteredMembers = members.filter((member) =>
    member.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  // Filter by tab
  if (activeTab === 'admins') {
    filteredMembers = filteredMembers.filter((member) => member.role === 'Admin');
  }

  // Pagination
  const totalPages = Math.ceil(
    (activeTab === 'requests' ? pendingRequests : filteredMembers).length / membersPerPage
  );
  const indexOfLast = currentPage * membersPerPage;
  const indexOfFirst = indexOfLast - membersPerPage;
  const currentItems = (activeTab === 'requests' ? pendingRequests : filteredMembers)
    .slice(indexOfFirst, indexOfLast);

  return (
    <div className={styles.membersContainer}>
      <div className={styles.header}>
        <h2>Group Members</h2>
        <div className={styles.searchBar}>
          <i className="fas fa-search"></i>
          <input
            type="text"
            placeholder="Search members..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>

      <div className={styles.tabsContainer}>
        <button
          className={`${styles.tabButton} ${activeTab === 'all' ? styles.active : ''}`}
          onClick={() => setActiveTab('all')}
        >
          All Members
        </button>
        <button
          className={`${styles.tabButton} ${activeTab === 'admins' ? styles.active : ''}`}
          onClick={() => setActiveTab('admins')}
        >
          Admins
        </button>
        {currentUserRole === 'admin' && (
          <button
            className={`${styles.tabButton} ${activeTab === 'requests' ? styles.active : ''}`}
            onClick={() => setActiveTab('requests')}
          >
            Requests {pendingRequests.length > 0 && (
              <span className={styles.requestBadge}>{pendingRequests.length}</span>
            )}
          </button>
        )}
      </div>

      <div className={styles.membersGrid}>
        {activeTab === 'requests' ? (
          pendingRequests.length === 0 ? (
            <div className={styles.noRequests}>No pending requests</div>
          ) : (
            currentItems.map((request) => (
              <div key={request.id} className={styles.requestCard}>
                <div className={styles.memberInfo}>
                  <img src={request.avatar} alt={request.name} className={styles.avatar} />
                  <div className={styles.details}>
                    <h3>{request.name}</h3>
                    <span className={styles.requestDate}>
                      Requested {new Date(request.joinDate).toLocaleDateString()}
                    </span>
                  </div>
                </div>
                <div className={styles.requestActions}>
                  <button
                    className={styles.acceptButton}
                    onClick={() => handleRequestAction(request.id, 'accept')}
                  >
                    Accept
                  </button>
                  <button
                    className={styles.rejectButton}
                    onClick={() => handleRequestAction(request.id, 'reject')}
                  >
                    Reject
                  </button>
                </div>
              </div>
            ))
          )
        ) : (
          currentItems.map((member) => (
            <div key={member.id} className={styles.memberCard}>
              <div className={styles.memberInfo}>
                <img src={member.avatar} alt={member.name} className={styles.avatar} />
                <div className={styles.details}>
                  <h3>{member.name}</h3>
                  <span className={member.isOnline ? styles.online : styles.offline}>
                    {member.isOnline ? 'Online' : 'Offline'}
                  </span>
                  {member.role === 'Admin' && (
                    <span className={styles.adminBadge}>Admin</span>
                  )}
                </div>
              </div>
              {/* <div className={styles.actions}>
              <button className={styles.followButton}>Follow</button>
            </div> */}
            </div>
          ))
        )}
      </div>

      {totalPages > 1 && (
        <div className={styles.pagination}>
          {Array.from({ length: totalPages }, (_, i) => (
            <button
              key={i + 1}
              onClick={() => setCurrentPage(i + 1)}
              className={currentPage === i + 1 ? styles.active : ''}
            >
              {i + 1}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default GroupMembers;
