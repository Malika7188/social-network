'use client';

import { useState } from 'react';
import styles from '@/styles/Posts.module.css';
import { useGroupService } from '@/services/groupService';
import { showToast } from '@/components/ui/ToastContainer';
import Image from 'next/image';
import { BASE_URL } from '@/utils/constants';
import EmojiPicker from '@/components/ui/EmojiPicker';

export default function GroupCreatePost({ groupId, groupName, oncreatePost }) {
  const { createPost } = useGroupService();
  const [postText, setPostText] = useState('');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedFiles, setSelectedFiles] = useState([]);
  const [previewUrls, setPreviewUrls] = useState([]);
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const [privacyNote, setPrivacyNote] = useState('Members');

  const handleSubmit = async (e) => {
    e.preventDefault();

    const formData = new FormData();
    formData.append('content', postText);
    formData.append('groupId', groupId);
    formData.append('isGroupPost', 'true');

    if (selectedFiles.length > 1) {
      showToast('Please select only one image or video', 'error');
      return;
    }

    selectedFiles.forEach(file => {
      if (file.type.startsWith('image/')) {
        formData.append('image', file);
      } else if (file.type.startsWith('video/')) {
        formData.append('video', file);
      }
    });

    if (!postText.trim() && !selectedFiles.length) {
      showToast('Please enter post content or add media', 'error');
      return;
    }

    try {
      await createPost(formData);
      setPostText('');
      setSelectedFiles([]);
      setPreviewUrls([]);
      setIsModalOpen(false);
      oncreatePost();
      showToast('Post created successfully!', 'success');
    } catch (error) {
      console.error('Error submitting post:', error);
      showToast('Failed to create post', 'error');
    }
  };

  const handleFileSelect = (e, fileType) => {
    const files = Array.from(e.target.files);
    
    const filteredFiles = files.filter(file => {
      if (fileType === 'image') {
        return file.type === 'image/png' || 
               file.type === 'image/jpeg' || 
               file.type === 'image/jpg' || 
               file.type === 'image/gif';
      }
      if (fileType === 'video') {
        return file.type.startsWith('video/');
      }
      return false;
    });

    if (filteredFiles.length === 0) {
      if (fileType === 'image') {
        showToast('Please select PNG, JPG, or GIF images only', 'error');
      } else {
        showToast('Please select valid video files', 'error');
      }
      return;
    }

    setSelectedFiles(prev => [...prev, ...filteredFiles]);
    const urls = filteredFiles.map(file => URL.createObjectURL(file));
    setPreviewUrls(prev => [...prev, ...urls]);
  };

  const removeFile = (index) => {
    setSelectedFiles(prev => prev.filter((_, i) => i !== index));
    setPreviewUrls(prev => prev.filter((_, i) => i !== index));
  };

  const addEmoji = (emoji) => {
    setPostText(prev => prev + emoji);
    setShowEmojiPicker(false);
  };

  const UserAvatar = `${BASE_URL}/uploads/${JSON.parse(localStorage.getItem('userData'))?.avatar || '/avatar.png'}`;

  return (
    <>
      <div className={styles.createPostCard}>
        <div className={styles.createPostHeader}>
          <img
            src={UserAvatar}
            alt="Profile"
            width={40}
            height={40}
            className={styles.profilePic}
          />
          <div 
            className={styles.createPostInput}
            onClick={() => setIsModalOpen(true)}
          >
            <input
              type="text"
              placeholder={`Write something to ${groupName}...`}
              readOnly
            />
          </div>
        </div>
        <div className={styles.createPostFooter}>
          <label className={styles.mediaButton}>
            <input
              type="file"
              multiple
              accept=".png,.jpg,.jpeg,.gif,image/png,image/jpeg,image/jpg,image/gif"
              onChange={(e) => {
                handleFileSelect(e, 'image');
                setIsModalOpen(true);
              }}
              hidden
            />
            <i className="fas fa-images"></i>
            <span>Photo</span>
          </label>
          <label className={styles.mediaButton}>
            <input
              type="file"
              multiple
              accept="video/*"
              onChange={(e) => {
                handleFileSelect(e, 'video');
                setIsModalOpen(true);
              }}
              hidden
            />
            <i className="fas fa-video"></i>
            <span>Video</span>
          </label>
        </div>
      </div>

      {isModalOpen && (
        <div className={styles.modalOverlay}>
          <div className={styles.modal}>
            <div className={styles.modalHeader}>
              <h2>Create Group Post</h2>
              <button
                className={styles.closeButton}
                onClick={() => setIsModalOpen(false)}
              >
                <i className="fas fa-times"></i>
              </button>
            </div>

            <div className={styles.modalContent}>
              <div className={styles.userInfo}>
                <Image
                  src={UserAvatar}
                  alt="Profile"
                  width={40}
                  height={40}
                  className={styles.profilePic}
                />
                <div>
                  <h3>Post to {groupName}</h3>
                  <div className={styles.privacySelector}>
                    <i className="fas fa-users"></i>
                    {privacyNote}
                  </div>
                </div>
              </div>

              <form onSubmit={handleSubmit}>
                <textarea
                  value={postText}
                  onChange={(e) => setPostText(e.target.value)}
                  placeholder="What's on your mind?"
                  autoFocus
                />

                {previewUrls.length > 0 && (
                  <div className={styles.previewGrid}>
                    {previewUrls.map((url, index) => (
                      <div key={index} className={styles.previewItem}>
                        {selectedFiles[index].type.startsWith('image/') ? (
                          <Image src={url} fill alt="Preview" />
                        ) : (
                          <video
                            src={url}
                            controls
                            preload="metadata"
                            className={styles.videoPreview}
                          />
                        )}
                        <button
                          type="button"
                          onClick={() => removeFile(index)}
                          className={styles.removePreview}
                        >
                          <i className="fas fa-times"></i>
                        </button>
                      </div>
                    ))}
                  </div>
                )}

                <div className={styles.addToPost}>
                  <h4>Add to your post</h4>
                  <div className={styles.mediaButtons}>
                    <label className={styles.mediaLabel}>
                      <input
                        type="file"
                        multiple
                        accept=".png,.jpg,.jpeg,.gif,image/png,image/jpeg,image/jpg,image/gif"
                        onChange={(e) => handleFileSelect(e, 'image')}
                        hidden
                      />
                      <i
                        className="fas fa-image"
                        style={{ color: '#45bd62' }}
                        title="Upload Images (PNG, JPG, GIF)"
                      ></i>
                    </label>
                    <label className={styles.mediaLabel}>
                      <input
                        type="file"
                        multiple
                        accept="video/*"
                        onChange={(e) => handleFileSelect(e, 'video')}
                        hidden
                      />
                      <i
                        className="fas fa-video"
                        style={{ color: '#e42645' }}
                        title="Upload Videos"
                      ></i>
                    </label>
                    <button type="button">
                      <i
                        className="fas fa-user-tag"
                        style={{ color: '#1877f2' }}
                      ></i>
                    </button>
                    <button
                      type="button"
                      onClick={() => setShowEmojiPicker(!showEmojiPicker)}
                    >
                      <i
                        className="fas fa-face-smile"
                        style={{ color: '#f7b928' }}
                      ></i>
                    </button>
                    <button type="button">
                      <i
                        className="fas fa-map-marker-alt"
                        style={{ color: '#f5533d' }}
                      ></i>
                    </button>
                  </div>
                </div>

                {showEmojiPicker && (
                  <EmojiPicker
                    onEmojiSelect={(emoji) => addEmoji(emoji)}
                    onClose={() => setShowEmojiPicker(false)}
                  />
                )}

                <button
                  type="submit"
                  onClick={handleSubmit}
                  className={styles.postButton}
                  disabled={!postText && !selectedFiles.length}
                >
                  Post
                </button>
              </form>
            </div>
          </div>
        </div>
      )}
    </>
  );
}