import styles from "@/styles/Posts.module.css";

export default function ConfirmationModal({
  isOpen,
  onClose,
  onConfirm,
  title = "Confirmation",
  message = "Are you sure you want to proceed? This action cannot be undone.",
}) {
  if (!isOpen) return null;

  return (
    <div className={styles.modalOverlay}>
      <div className={styles.modal}>
        <div className={styles.modalHeader}>
          <h2>{title}</h2>
          <button className={styles.closeButton} onClick={onClose}>
            <i className="fas fa-times"></i>
          </button>
        </div>
        <div className={styles.modalContent}>
          <p className={styles.deleteConfirmText}>{message}</p>
          <div className={styles.modalFooter}>
            <button className={styles.cancelBtn} onClick={onClose}>
              Cancel
            </button>
            <button className={styles.deleteBtn} onClick={onConfirm}>
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
