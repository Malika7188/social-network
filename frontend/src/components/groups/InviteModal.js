import { useState, useEffect } from 'react';
import { useFriendService } from '@/services/friendService';
import { useGroupService } from '@/services/groupService';
import styles from '@/styles/InviteModal.module.css';
import { showToast } from '@/components/ui/ToastContainer';

const InviteModal = ({ group, isOpen, onClose }) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [friends, setFriends] = useState([]);
    const [loading, setLoading] = useState(true);
    const { fetchUserFollowers } = useFriendService();
    const { inviteToGroup } = useGroupService();

    useEffect(() => {
        const loadFriends = async () => {
            try {
                const contacts = await fetchUserFollowers();
                setFriends(contacts);
            } catch (error) {
                console.error('Error fetching friends:', error);
                showToast('Failed to load friends', 'error');
            } finally {
                setLoading(false);
            }
        };

        if (isOpen) {
            loadFriends();
        }
    }, [isOpen]);

    const handleInvite = async (friendId) => {
        try {
            await inviteToGroup(group.ID, friendId);
            showToast('Invitation sent successfully', 'success');
        } catch (error) {
            console.error('Error inviting friend:', error);
            showToast('Failed to send invitation', 'error');
        }
    };

    const filteredFriends = friends.filter(friend =>
        friend.name.toLowerCase().includes(searchTerm.toLowerCase())
    );

    if (!isOpen) return null;

    return (
        <div className={styles.modalOverlay}>
            <div className={styles.modal}>
                <div className={styles.modalHeader}>
                    <h2>Invite Friends to {group.Name}</h2>
                    <button className={styles.closeButton} onClick={onClose}>
                        <i className="fas fa-times"></i>
                    </button>
                </div>

                <div className={styles.searchBar}>
                    <i className="fas fa-search"></i>
                    <input
                        type="text"
                        placeholder="Search friends..."
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>

                <div className={styles.friendsList}>
                    {loading ? (
                        <div className={styles.loading}>Loading friends...</div>
                    ) : filteredFriends.length === 0 ? (
                        <div className={styles.noFriends}>
                            {searchTerm ? 'No friends match your search' : 'No friends to invite'}
                        </div>
                    ) : (
                        filteredFriends.map(friend => (
                            <div key={friend.id} className={styles.friendItem}>
                                <div className={styles.friendInfo}>
                                    <img src={friend.image} alt={friend.name} />
                                    <div className={styles.friendDetails}>
                                        <h3>{friend.name}</h3>
                                        <span className={friend.isOnline ? styles.online : styles.offline}>
                                            {friend.isOnline ? 'Online' : 'Offline'}
                                        </span>
                                    </div>
                                </div>
                                <button
                                    className={styles.inviteButton}
                                    onClick={() => handleInvite(friend.contactId)}
                                >
                                    Invite
                                </button>
                            </div>
                        ))
                    )}
                </div>
            </div>
        </div>
    );
};

export default InviteModal;