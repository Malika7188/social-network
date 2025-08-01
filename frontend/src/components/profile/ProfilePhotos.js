import React, { useState, useEffect } from 'react';
import styles from '@/styles/ProfilePhotos.module.css';
import ContactsList from '@/components/contacts/ContactsList';

const ProfilePhotos = ({
  photos = [],
  isLoading = false,
  error = null,
  BASE_URL = ""
}) => {
  const [selectedPhotoIndex, setSelectedPhotoIndex] = useState(null);

  const handlePhotoClick = (index) => {
    setSelectedPhotoIndex(index);
  };

  const handleCloseModal = () => {
    setSelectedPhotoIndex(null);
  };

  const handlePrevPhoto = () => {
    setSelectedPhotoIndex((prev) =>
      prev > 0 ? prev - 1 : photos.length - 1
    );
  };

  const handleNextPhoto = () => {
    setSelectedPhotoIndex((prev) =>
      prev < photos.length - 1 ? prev + 1 : 0
    );
  };

  // Handle keyboard navigation
  const handleKeyDown = (e) => {
    if (selectedPhotoIndex === null) return;
    switch(e.key) {
      case 'ArrowLeft':
        handlePrevPhoto();
        break;
      case 'ArrowRight':
        handleNextPhoto();
        break;
      case 'Escape':
        handleCloseModal();
        break;
      default:
        break;
    }
  };

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [selectedPhotoIndex]);

  // Render loading state
  if (isLoading) {
    return (
      <div className={styles.photosContainer}>
        <div className={styles.mainContent}>
          <h2>Photos</h2>
          <div className={styles.loadingContainer}>
            <p>Loading photos...</p>
          </div>
        </div>
        <div className={styles.sidebar}>
          <ContactsList />
        </div>
      </div>
    );
  }

  // Render error state
  if (error) {
    return (
      <div className={styles.photosContainer}>
        <div className={styles.mainContent}>
          <h2>Photos</h2>
          <div className={styles.errorContainer}>
            <p>Error loading photos: {error}</p>
          </div>
        </div>
        <div className={styles.sidebar}>
          <ContactsList />
        </div>
      </div>
    );
  }

  // Render empty state
  if (!photos.length) {
    return (
      <div className={styles.photosContainer}>
        <div className={styles.mainContent}>
          <h2>Photos</h2>
          <div className={styles.emptyContainer}>
            <p>No photos available</p>
          </div>
        </div>
        <div className={styles.sidebar}>
          <ContactsList />
        </div>
      </div>
    );
  }

  return (
    <div className={styles.photosContainer}>
      <div className={styles.mainContent}>
        <h2>Photos</h2>
        <div className={styles.photoGrid}>
          {photos.map((photo, index) => {
            // Create the image URL with BASE_URL prefix
            const imageUrl = photo.imageUrl
              ? `${BASE_URL}${photo.imageUrl}`
              : "/default-photo.jpg";
              
            return (
              <div
                key={photo.id || `photo-${index}`}
                className={styles.photoItem}
                onClick={() => handlePhotoClick(index)}
              >
                <img src={imageUrl} alt={`Photo ${index + 1}`} />
              </div>
            );
          })}
        </div>
      </div>
      <div className={styles.sidebar}>
        <ContactsList />
      </div>
      
      {selectedPhotoIndex !== null && photos.length > 0 && (
        <div className={styles.modal} onClick={handleCloseModal}>
          <div className={styles.modalContent} onClick={e => e.stopPropagation()}>
            <button className={styles.closeButton} onClick={handleCloseModal}>
              ×
            </button>
            <button
              className={`${styles.navButton} ${styles.prevButton}`}
              onClick={handlePrevPhoto}
            >
              ‹
            </button>
            <button
              className={`${styles.navButton} ${styles.nextButton}`}
              onClick={handleNextPhoto}
            >
              ›
            </button>
            {photos[selectedPhotoIndex] && (
              <img 
                src={photos[selectedPhotoIndex].imageUrl 
                  ? `${BASE_URL}${photos[selectedPhotoIndex].imageUrl}` 
                  : "/default-photo.jpg"} 
                alt="Selected photo" 
              />
            )}
            <div className={styles.photoCounter}>
              {selectedPhotoIndex + 1} / {photos.length}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ProfilePhotos;