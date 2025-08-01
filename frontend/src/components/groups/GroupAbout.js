import { useState } from 'react';
import styles from '@/styles/GroupAbout.module.css';
import { formatRelativeTime } from '@/utils/dateUtils';
import { useGroupService } from '@/services/groupService';
import { showToast } from "@/components/ui/ToastContainer";
import InviteModal from '@/components/groups/InviteModal';



const GroupAbout = ({ group }) => {
  const [showConfirmLeave, setShowConfirmLeave] = useState(false);
  const [showConfirmDelete, setShowConfirmDelete] = useState(false);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const { leaveGroup, deleteGroup, joinGroup } = useGroupService();


  let userdata = null;
try {
  const raw = localStorage.getItem("userData");
  if (raw) userdata = JSON.parse(raw);
} catch (e) {
  console.error("Invalid userData in localStorage:", e);
}
  const handleLeaveGroup = async () => {
    try {
      success = await leaveGroup(group.ID);
      if (success) {
        showToast("Left group successfully", "success");
      }
    } catch (error) {
      console.error("Error leaving group:", error);
      showToast("Failed to leave group", "error");
    }
    setShowConfirmLeave(false);
  };

  const handleDeleteGroup = async () => {
    try {
      success = await deleteGroup(group.ID);
      if (success) {
        showToast("Group deleted successfully", "success");
      }
    } catch (error) {
      console.error("Error leaving group:", error);
      showToast("Failed to leave group", "error");
    }
    setShowConfirmDelete(false);
  };

  const handlejoiningroup = async () => {
    try {
      success = await joinGroup(group.ID);
      if (success) {
        showToast("Joined group successfully", "success");
      }
    } catch (error) {
      console.error("Error leaving group:", error);
      showToast("Failed to leave group", "error");
    }
  };
  
  if (group === null || group === undefined || group.length === 0) {
    return (
      <div className={styles.loadingContainer}>
        <p>Can not show group info</p>
      </div>
    )
  }

  return (
    <div className={styles.aboutContainer}>
      <div className={styles.mainInfo}>
        <h2>About This Group</h2>
        <div className={styles.description}>
          <p>{group.Content}</p>
        </div>

        <div className={styles.stats}>
          <div className={styles.statItem}>
            <i className="fas fa-users"></i>
            <div>
              <h3>{group.MemberCount.toLocaleString()}</h3>
              <p>Members</p>
            </div>
          </div>
          <div className={styles.statItem}>
            <i className="fas fa-calendar"></i>
            <div>
              <h3>{formatRelativeTime(group.CreatedAt)}</h3>
              <p>Created</p>
            </div>
          </div>
          <div className={styles.statItem}>
            <i className={`fas ${group.Ispublic === true ? 'fa-lock' : 'fa-globe'}`}></i>
            <div>
              <h3>{group.Ispublic === true ? 'Private' : 'Public'}</h3>
              <p>Privacy</p>
            </div>
          </div>
        </div>

        <div className={styles.actions}>
          {group.IsMember && (
            <button
              className={styles.inviteButton}
              onClick={() => setShowInviteModal(true)}
            >
              <i className="fas fa-user-plus"></i>
              Invite Members
            </button>
          )}
          {group.IsMember ? (
            userdata.id === group.Creator.id ? (
              <button
                className={styles.leaveButton}
                onClick={() => setShowConfirmDelete(true)}
              >
                <i className="fas fa-sign-out-alt"></i>
                Delete Group
              </button>
            ) : (
              <button
                className={styles.leaveButton}
                onClick={() => setShowConfirmLeave(true)}
              >
                <i className="fas fa-sign-out-alt"></i>
                Leave Group
              </button>
            )
          ) : (
            <button
              className={styles.inviteButton}
              onClick={handlejoiningroup}
            >
              Join Group
            </button>
          )}
        </div>
      </div>

      {showConfirmLeave && (
        <div className={styles.confirmDialog}>
          <div className={styles.dialogContent}>
            <h3>Leave Group?</h3>
            <p>Are you sure you want to leave this group?</p>
            <div className={styles.dialogActions}>
              <button
                className={styles.cancelButton}
                onClick={() => setShowConfirmLeave(false)}
              >
                Cancel
              </button>
              <button
                className={styles.confirmButton}
                onClick={handleLeaveGroup}
              >
                Leave
              </button>
            </div>
          </div>
        </div>
      )}

      {showConfirmDelete && (
        <div className={styles.confirmDialog}>
          <div className={styles.dialogContent}>
            <h3>Delete Group?</h3>
            <p>Are you sure you want to Delete this group?</p>
            <div className={styles.dialogActions}>
              <button
                className={styles.cancelButton}
                onClick={() => setShowConfirmDelete(false)}
              >
                Cancel
              </button>
              <button
                className={styles.confirmButton}
                onClick={handleDeleteGroup}
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      )}

      <InviteModal
        group={group}
        isOpen={showInviteModal}
        onClose={() => setShowInviteModal(false)}
      />
    </div>
  );
};

export default GroupAbout;