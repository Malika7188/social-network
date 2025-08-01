import { useState } from 'react';
import styles from '@/styles/GroupPhotos.module.css';  // Update import to use correct CSS module
import { BASE_URL } from '@/utils/constants';

const GroupPhotos = ({ posts }) => {
  const [selectedPhotoIndex, setSelectedPhotoIndex] = useState(null);
  const [currentPage, setCurrentPage] = useState(1);
  if (posts == null || posts.length === 0) { 
    return <div>No photos available</div>;
  }

  const photos = posts
  .filter(post => post.ImagePath?.Valid)
  .map(post => ({
    id: post.ID,
    url: `${BASE_URL}/uploads/${post.ImagePath.String}`,
  }));


  const photosPerPage = 20;

  const handlePhotoClick = (index) => {
    setSelectedPhotoIndex(index);
  };

  const handleCloseModal = () => {
    setSelectedPhotoIndex(null);
  };

  // Calculate pagination
  const totalPages = Math.ceil(photos.length / photosPerPage);
  const indexOfLastPhoto = currentPage * photosPerPage;
  const indexOfFirstPhoto = indexOfLastPhoto - photosPerPage;
  const currentPhotos = photos.slice(indexOfFirstPhoto, indexOfLastPhoto);

  return (
    <div className={styles.photosContainer}>
      <div className={styles.mainContent}>
        <h2>Group Photos</h2>
        <div className={styles.photoGrid}>
          {currentPhotos.map((photo, index) => (
            <div 
              key={photo.id} 
              className={styles.photoItem}
              onClick={() => handlePhotoClick(index)}
            >
              <img 
                src={photo.url} 
                alt={`Group photo ${index + 1}`}
                loading="lazy" // Add lazy loading for better performance
              />
            </div>
          ))}
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

      {selectedPhotoIndex !== null && (
        <div className={styles.modal} onClick={handleCloseModal}>
          <div className={styles.modalContent}>
            <button className={styles.closeButton} onClick={handleCloseModal}>×</button>
            <img 
              src={photos[selectedPhotoIndex].url} 
              alt={`Selected photo ${selectedPhotoIndex + 1}`} 
            />
          </div>
        </div>
      )}
    </div>
  );
};

export default GroupPhotos;