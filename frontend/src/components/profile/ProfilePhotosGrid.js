import React from "react";
import Image from "next/image";
import styles from "@/styles/ProfilePhotoGrid.module.css";

const ProfilePhotosGrid = ({
  photos = [],
  totalPhotos = 0,
  isLoading = false,
  error = null,
  BASE_URL = "",
}) => {
  // Check for valid photos array
  const hasPhotos = Array.isArray(photos) && photos.length > 0;

  // Calculate total if not provided
  const actualTotal = totalPhotos || (hasPhotos ? photos.length : 0);

  // Show only first 6 photos or all if less than 6
  const displayPhotos = hasPhotos ? photos.slice(0, 6) : [];

  // Calculate remaining photos (if any)
  // Only show "+X more" if we have more than 6 photos total
  const remainingPhotos = actualTotal > 6 ? actualTotal - 6 : 0;

  // Render loading state
  if (isLoading) {
    return (
      <div className={styles.photosSection}>
        <div className={styles.sectionHeader}>
          <h3>Photos</h3>
        </div>
        <div className={styles.loadingContainer}>
          <p>Loading photos...</p>
        </div>
      </div>
    );
  }

  // Render error state
  if (error) {
    return (
      <div className={styles.photosSection}>
        <div className={styles.sectionHeader}>
          <h3>Photos</h3>
        </div>
        <div className={styles.errorContainer}>
          <p>Unable to load photos</p>
        </div>
      </div>
    );
  }

  // Render empty state
  if (!hasPhotos) {
    return (
      <div className={styles.photosSection}>
        <div className={styles.sectionHeader}>
          <h3>Photos</h3>
        </div>
        <div className={styles.emptyContainer}>
          <p>No photos available</p>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.photosSection}>
      <div className={styles.sectionHeader}>
        <h3>Photos</h3>
        {hasPhotos && (
          <a href="#" className={styles.seeAll}>
            See all
          </a>
        )}
      </div>
      <div className={styles.photoGrid}>
        {displayPhotos.map((photo, index) => {
          // Create the image URL with BASE_URL prefix - fix string concatenation
          const imageUrl = photo.imageUrl
            ? `${BASE_URL}${photo.imageUrl}`
            : "/default-photo.jpg";

          return (
            <div
              key={photo.id || `photo-${index}`}
              className={styles.photoItem}
            >
              {/* Only show overlay on the last item when there are more photos */}
              {index === displayPhotos.length - 1 && remainingPhotos > 0 ? (
                <>
                  <Image
                    src={imageUrl}
                    alt={`Photo ${index + 1}`}
                    fill
                    sizes="(max-width: 768px) 100vw, 33vw"
                    className={styles.photo}
                  />
                  <div className={styles.overlay}>
                    <span>+{remainingPhotos}</span>
                  </div>
                </>
              ) : (
                <Image
                  src={imageUrl}
                  alt={`Photo ${index + 1}`}
                  fill
                  sizes="(max-width: 768px) 100vw, 33vw"
                  className={styles.photo}
                />
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default ProfilePhotosGrid;
