import { useState, useEffect } from "react";
import styles from "@/styles/Messages.module.css";
import { useAuth } from "@/context/authcontext";
import { useFriendService } from "@/services/friendService";
import { showToast } from "@/components/ui/ToastContainer";

export default function NewMessageModal({ onClose, onSelectContact }) {
  const { currentUser } = useAuth();
  const { fetchContacts } = useFriendService();
  const [contacts, setContacts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");

  useEffect(() => {
    const loadContacts = async () => {
      try {
        const contactsList = await fetchContacts();
        setContacts(contactsList);
      } catch (error) {
        console.error("Error loading contacts:", error);
        showToast("Failed to load contacts", "error");
      } finally {
        setLoading(false);
      }
    };

    loadContacts();
  }, [fetchContacts]);

  const filteredContacts = contacts.filter((contact) =>
    contact.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleSelectContact = (contact) => {
    onSelectContact(contact);
    onClose();
  };

  return (
    <div className={styles.modalOverlay}>
      <div className={styles.modalContent}>
        <div className={styles.modalHeader}>
          <h3>New Message</h3>
          <button className={styles.closeButton} onClick={onClose}>
            <i className="fas fa-times"></i>
          </button>
        </div>
        <div className={styles.modalBody}>
          <input
            type="text"
            placeholder="Search people..."
            className={styles.searchPeople}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            autoFocus
          />
          <div className={styles.suggestedPeople}>
            {loading ? (
              <div className={styles.loadingContacts}>
                <i className="fas fa-spinner fa-spin"></i>
                <p>Loading contacts...</p>
              </div>
            ) : filteredContacts.length > 0 ? (
              filteredContacts.map((contact) => (
                <div
                  key={contact.contactId}
                  className={styles.contactItem}
                  onClick={() => handleSelectContact(contact)}
                >
                  <div className={styles.contactAvatar}>
                    <img
                      src={contact.image || "/avatar.png"}
                      alt={contact.name}
                      className={styles.avatar}
                    />
                  </div>
                  <div className={styles.contactInfo}>
                    <h4>{contact.name}</h4>
                  </div>
                </div>
              ))
            ) : (
              <div className={styles.noResults}>
                <p>No contacts found</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
