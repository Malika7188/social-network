import React, { useEffect } from "react";
import Image from "next/image";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCircle } from "@fortawesome/free-solid-svg-icons";
import styles from "@/styles/Sidebar.module.css";
import { useUserStatus } from "@/services/userStatusService";
import LoadingSpinner from "@/components/ui/LoadingSpinner";

const ContactsSection = ({ contacts, isLoading, isProfilePage = false }) => {
  const { isUserOnline, initializeStatuses } = useUserStatus();

  // Initialize online statuses from API data
  useEffect(() => {
    if (contacts && contacts.length > 0) {
      initializeStatuses(contacts);
    }
  }, [contacts, initializeStatuses]);

  return (
    <section
      className={`${styles.contacts} ${
        isProfilePage ? styles.profileContacts : ""
      }`}
    >
      <h2>Contacts</h2>
      {isLoading ? (
        <div className={styles.loadingContainer}>
          <LoadingSpinner size="small" color="primary" />
        </div>
      ) : !contacts || contacts.length === 0 ? (
        <p className={styles.emptyState}>No contacts to display</p>
      ) : (
        contacts
          .filter((contact) => contact !== null && contact !== undefined)
          .map((contact) => {
            // Use the API-provided status as default, then override with WebSocket updates
            const isOnline = isUserOnline(contact.contactId, contact.isOnline);
            return (
              <div key={contact.id} className={styles.contactItem}>
                <div className={styles.contactProfile}>
                  <div className={styles.contactImageWrapper}>
                    <img src={contact.image} alt={contact.name} />
                    <span
                      className={`${styles.onlineStatus} ${
                        isOnline ? styles.online : styles.offline
                      }`}
                    />
                  </div>
                  <span className={styles.contactName}>{contact.name}</span>
                </div>
              </div>
            );
          })
      )}
    </section>
  );
};

export default ContactsSection;
