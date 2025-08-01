"use client";

import { useState } from 'react';
import styles from '@/styles/CreateGroupModal.module.css';
import { useGroupService } from '@/services/groupService';

const CreateGroupModal = ({ isOpen, onClose, onGroupCreated }) => {
  const [groupData, setGroupData] = useState({
    name: '',
    description: '',
    privacy: 'public',
    banner: null,
    profilePic: null
  });
  
  const { createGroup } = useGroupService();
  const [bannerPreview, setBannerPreview] = useState(null);
  const [profilePreview, setProfilePreview] = useState(null);

  const handleImageChange = (e, type) => {
    const file = e.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.onloadend = () => {
        if (type === 'banner') {
          setBannerPreview(reader.result);
          setGroupData(prev => ({ ...prev, banner: file }));
        } else {
          setProfilePreview(reader.result);
          setGroupData(prev => ({ ...prev, profilePic: file }));
        }
      };
      reader.readAsDataURL(file);
    }
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    const formData = new FormData();
    formData.append('name', groupData.name);
    formData.append('description', groupData.description);
    formData.append('privacy', 'private');
    if (groupData.banner) formData.append('banner', groupData.banner);
    if (groupData.profilePic) formData.append('profilePic', groupData.profilePic);
    createGroup(formData)
    onGroupCreated();
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className={styles.modalOverlay}>
      <div className={styles.modal}>
        <div className={styles.modalHeader}>
          <h2>Create New Group</h2>
          <button className={styles.closeButton} onClick={onClose}>×</button>
        </div>

        <form onSubmit={handleSubmit} className={styles.form}>
          {/* Banner Upload */}
          <div className={styles.imageUpload}>
            <div 
              className={`${styles.bannerUpload} ${bannerPreview ? styles.hasImage : ''}`}
              style={bannerPreview ? { backgroundImage: `url(${bannerPreview})` } : {}}
            >
              <input
                type="file"
                accept="image/*"
                onChange={(e) => handleImageChange(e, 'banner')}
                id="banner-upload"
              />
              <label htmlFor="banner-upload">
                <i className="fas fa-camera"></i>
                <span>Add Cover Photo</span>
              </label>
            </div>

            {/* Profile Picture Upload */}
            <div 
              className={`${styles.profileUpload} ${profilePreview ? styles.hasImage : ''}`}
              style={profilePreview ? { backgroundImage: `url(${profilePreview})` } : {}}
            >
              <input
                type="file"
                accept="image/*"
                onChange={(e) => handleImageChange(e, 'profile')}
                id="profile-upload"
              />
              <label htmlFor="profile-upload">
                <i className="fas fa-camera"></i>
              </label>
            </div>
          </div>

          <div className={styles.formFields}>
            <div className={styles.inputGroup}>
              <label htmlFor="group-name">Group Name</label>
              <input
                type="text"
                id="group-name"
                value={groupData.name}
                onChange={(e) => setGroupData(prev => ({ ...prev, name: e.target.value }))}
                placeholder="Enter group name"
                required
              />
            </div>

            <div className={styles.inputGroup}>
              <label htmlFor="group-description">Description</label>
              <textarea
                id="group-description"
                value={groupData.description}
                onChange={(e) => setGroupData(prev => ({ ...prev, description: e.target.value }))}
                placeholder="What's your group about?"
                rows="3"
              />
            </div>
          </div>

          <div className={styles.modalFooter}>
            <button type="button" className={styles.cancelBtn} onClick={onClose}>
              Cancel
            </button>
            <button type="submit" className={styles.createBtn}>
              Create Group
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateGroupModal; 