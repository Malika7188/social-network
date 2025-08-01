"use client";

import { useState, useEffect } from "react";
import { usePostService } from "@/services/postService";
import styles from "@/styles/Posts.module.css";
import { showToast } from "@/components/ui/ToastContainer";
import EmojiPicker from "@/components/ui/EmojiPicker";
import Image from "next/image";
import { useFriendService } from "@/services/friendService";
import LoadingSpinner from "@/components/ui/LoadingSpinner";
import { BASE_URL } from "@/utils/constants";

export default function CreatePost() {
  const { createPost } = usePostService();
  const { fetchUserFollowers } = useFriendService();
  const [postText, setPostText] = useState("");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedFiles, setSelectedFiles] = useState([]);
  const [previewUrls, setPreviewUrls] = useState([]);
  const [privacy, setPrivacy] = useState("public");
  const [showPrivacyDropdown, setShowPrivacyDropdown] = useState(false);
  const [showEmojiPicker, setShowEmojiPicker] = useState(false);
  const [showSettingsModal, setShowSettingsModal] = useState(false);
  const [selectedViewers, setSelectedViewers] = useState([]);
  const [followers, setFollowers] = useState([]);
  const [searchTerm, setSearchTerm] = useState("");
  const [isLoadingFollowers, setIsLoadingFollowers] = useState(false);
  const [followersError, setFollowersError] = useState(null);
  const [userData, setUserData] = useState(null);

  // Load user data from localStorage when component mounts
  useEffect(() => {
    try {
      const storedUserData = localStorage.getItem("userData");
      if (storedUserData) {
        setUserData(JSON.parse(storedUserData));
      }
    } catch (error) {
      console.error("Error loading user data from localStorage:", error);
    }
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();

    const formData = new FormData();
    formData.append("content", postText);
    formData.append("privacy", privacy);

    if (privacy === "private" && selectedViewers.length === 0) {
      showToast("Please select at least one follower", "error");
      return;
    }

    // Add selected viewers if privacy is private
    if (privacy === "private" && selectedViewers.length > 0) {
      selectedViewers.forEach((viewerId) => {
        formData.append(`viewers`, viewerId);
      });
    }

    if (selectedFiles.length > 1) {
      showToast(
        "Please select only one image or video the backend does not support multiple files currently😁",
        "error"
      );
      return;
    }
    // Separate files by type
    selectedFiles.forEach((file) => {
      if (file.type.startsWith("image/")) {
        formData.append("image", file);
      } else if (file.type.startsWith("video/")) {
        formData.append("video", file);
      }
    });

    if (!postText.trim() && !selectedFiles.length) {
      showToast("Please enter a post content or add media", "error");
      return;
    }

    try {
      await createPost(formData);

      setPostText("");
      setSelectedFiles([]);
      setPreviewUrls([]);
      setIsModalOpen(false);
      setSelectedViewers([]);
    } catch (error) {
      console.error("Error submitting post:", error);
    }
  };

  const handleFileSelect = (e, fileType) => {
    const files = Array.from(e.target.files);

    const filteredFiles = files.filter((file) => {
      if (fileType === "image") {
        // Only accept png, jpg/jpeg, and gif
        return (
          file.type === "image/png" ||
          file.type === "image/jpeg" ||
          file.type === "image/jpg" ||
          file.type === "image/gif"
        );
      }
      if (fileType === "video") {
        return file.type.startsWith("video/");
      }
      return false; // Reject if not matching criteria
    });

    if (filteredFiles.length === 0) {
      if (fileType === "image") {
        showToast("Please select PNG, JPG, or GIF images only", "error");
      } else {
        showToast("Please select valid video files", "error");
      }
      return;
    }

    setSelectedFiles((prev) => [...prev, ...filteredFiles]);

    // Create preview URLs
    const urls = filteredFiles.map((file) => URL.createObjectURL(file));
    setPreviewUrls((prev) => [...prev, ...urls]);
  };

  const removeFile = (index) => {
    setSelectedFiles((prev) => prev.filter((_, i) => i !== index));
    setPreviewUrls((prev) => prev.filter((_, i) => i !== index));
  };

  const getPrivacyIcon = () => {
    switch (privacy) {
      case "public":
        return "fas fa-globe";
      case "private":
        return "fas fa-lock";
      case "almost_private":
        return "fas fa-user-friends";
      default:
        return "fas fa-globe";
    }
  };

  const getPrivacyText = () => {
    switch (privacy) {
      case "public":
        return "Public";
      case "private":
        return "Private";
      case "almost_private":
        return "Almost Private";
      default:
        return "Public";
    }
  };

  const addEmoji = (emoji) => {
    setPostText((prev) => prev + emoji);
    setShowEmojiPicker(false);
  };

  // Get user avatar with fallback
  const getUserAvatar = () => {
    if (userData?.avatar) {
      return `${BASE_URL}uploads/${userData.avatar}`;
    }
    return "/avatar.png"; // Default avatar path
  };

  // Get user's full name or nickname
  const getUserDisplayName = () => {
    if (userData) {
      return (
        `${userData.firstName || ""} ${userData.lastName || ""}`.trim() ||
        "User"
      );
    }
    return "User";
  };

  const fetchFollowers = async () => {
    setIsLoadingFollowers(true);
    setFollowersError(null);

    try {
      const followersList = await fetchUserFollowers();

      if (followersList && followersList.length > 0) {
        // Transform the data to match our component's expected format
        const formattedFollowers = followersList.map((follower) => ({
          id: follower.id,
          name: follower.name,
          avatar: follower.image,
        }));
        setFollowers(formattedFollowers);
      } else {
        setFollowers([]);
        setFollowersError(
          "No followers found. Add some friends to share private posts with them."
        );
      }
    } catch (error) {
      console.error("Error fetching followers:", error);
      setFollowers([]);
      setFollowersError("Failed to load followers. Please try again later.");
    } finally {
      setIsLoadingFollowers(false);
    }
  };

  const toggleViewer = (followerId) => {
    setSelectedViewers((prev) => {
      if (prev.includes(followerId)) {
        return prev.filter((id) => id !== followerId);
      } else {
        return [...prev, followerId];
      }
    });
  };

  const filteredFollowers = () => {
    if (!searchTerm.trim()) return followers;

    return followers.filter((follower) =>
      follower.name.toLowerCase().includes(searchTerm.toLowerCase())
    );
  };

  const openSettingsModal = () => {
    fetchFollowers(); // Fetch followers when opening settings
    setIsModalOpen(false); // Close the create post modal
    setShowSettingsModal(true);
  };

  const closeSettingsModal = () => {
    setShowSettingsModal(false);
    setIsModalOpen(true); // Reopen the create post modal
  };

  return (
    <>
      <div className={styles.createPostCard}>
        <div className={styles.createPostHeader}>
          <img
            src={getUserAvatar()}
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
              placeholder={`What's on your mind, ${userData?.firstName || "there"}?`}
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
                handleFileSelect(e, "image");
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
                handleFileSelect(e, "video");
                setIsModalOpen(true);
              }}
              hidden
            />
            <i className="fas fa-video"></i>
            <span>Video</span>
          </label>
          <button className={styles.mediaButton}>
            <i className="fas fa-face-smile"></i>
            <span>Feeling</span>
          </button>
        </div>
      </div>

      {isModalOpen && (
        <div className={styles.modalOverlay}>
          <div className={styles.modal}>
            <div className={styles.modalHeader}>
              <h2>Create Post</h2>
              <button
                className={styles.closeButton}
                onClick={() => setIsModalOpen(false)}
              >
                <i className="fas fa-times"></i>
              </button>
            </div>

            <div className={styles.modalContent}>
              <div className={styles.userInfo}>
                <img
                  src={getUserAvatar()}
                  alt="Profile"
                  width={40}
                  height={40}
                  className={styles.profilePic}
                />
                <div>
                  <h3>{getUserDisplayName()}</h3>
                  <div className={styles.privacySelector}>
                    <button
                      className={styles.privacyButton}
                      onClick={openSettingsModal}
                    >
                      <i className={getPrivacyIcon()}></i>
                      {getPrivacyText()}
                      <i className="fas fa-caret-down"></i>
                    </button>
                  </div>
                </div>
              </div>

              <form onSubmit={handleSubmit}>
                <textarea
                  value={postText}
                  onChange={(e) => setPostText(e.target.value)}
                  placeholder={`What's on your mind, ${userData?.firstName || "there"}?`}
                  autoFocus
                />

                {previewUrls.length > 0 && (
                  <div className={styles.previewGrid}>
                    {previewUrls.map((url, index) => (
                      <div key={index} className={styles.previewItem}>
                        {selectedFiles[index].type.startsWith("image/") ? (
                          <Image src={url} fill alt="Preview" />
                        ) : (
                          <video
                            src={url}
                            controls
                            preload="metadata"
                            className={styles.videoPreview}
                            onError={(e) =>
                              console.error("Video preview error:", e)
                            }
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
                        onChange={(e) => handleFileSelect(e, "image")}
                        hidden
                      />
                      <i
                        className="fas fa-image"
                        style={{ color: "#45bd62" }}
                        title="Upload Images (PNG, JPG, GIF)"
                      ></i>
                    </label>
                    <label className={styles.mediaLabel}>
                      <input
                        type="file"
                        multiple
                        accept="video/*"
                        onChange={(e) => handleFileSelect(e, "video")}
                        hidden
                      />
                      <i
                        className="fas fa-video"
                        style={{ color: "#e42645" }}
                        title="Upload Videos"
                      ></i>
                    </label>
                    <button type="button">
                      <i
                        className="fas fa-user-tag"
                        style={{ color: "#1877f2" }}
                      ></i>
                    </button>
                    <button
                      type="button"
                      onClick={() => setShowEmojiPicker(!showEmojiPicker)}
                    >
                      <i
                        className="fas fa-face-smile"
                        style={{ color: "#f7b928" }}
                      ></i>
                    </button>
                    <button type="button">
                      <i
                        className="fas fa-map-marker-alt"
                        style={{ color: "#f5533d" }}
                      ></i>
                    </button>
                    <button
                      type="button"
                      onClick={openSettingsModal}
                      title="Post Settings"
                    >
                      <i
                        className="fas fa-cog"
                        style={{ color: "#1877f2" }}
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

      {showSettingsModal && (
        <div className={styles.modalOverlay}>
          <div className={`${styles.modal} ${styles.settingsModal}`}>
            <div className={styles.modalHeader}>
              <button
                className={styles.backButton}
                onClick={closeSettingsModal}
              >
                <i className="fas fa-arrow-left"></i>
              </button>
              <h2>Post Settings</h2>
              <div></div> {/* Empty div for flex spacing */}
            </div>

            <div className={styles.modalContent}>
              <div className={styles.settingsSection}>
                <h3>Privacy</h3>
                <div className={styles.privacyOptions}>
                  <div className={styles.privacyOption}>
                    <label
                      className={`${styles.radioLabel} ${
                        privacy === "public" ? styles.selectedPrivacy : ""
                      }`}
                    >
                      <div className={styles.radioIcon}>
                        <input
                          type="radio"
                          name="privacy"
                          value="public"
                          checked={privacy === "public"}
                          onChange={() => setPrivacy("public")}
                        />
                        <i className="fas fa-globe"></i>
                      </div>
                      <div className={styles.privacyContent}>
                        <span className={styles.privacyTitle}>Public</span>
                        <p className={styles.privacyDescription}>
                          Anyone can see this post
                        </p>
                      </div>
                    </label>
                  </div>

                  <div className={styles.privacyOption}>
                    <label
                      className={`${styles.radioLabel} ${
                        privacy === "almost_private"
                          ? styles.selectedPrivacy
                          : ""
                      }`}
                    >
                      <div className={styles.radioIcon}>
                        <input
                          type="radio"
                          name="privacy"
                          value="almost_private"
                          checked={privacy === "almost_private"}
                          onChange={() => setPrivacy("almost_private")}
                        />
                        <i className="fas fa-user-friends"></i>
                      </div>
                      <div className={styles.privacyContent}>
                        <span className={styles.privacyTitle}>
                          Almost Private
                        </span>
                        <p className={styles.privacyDescription}>
                          Only your followers can see this post
                        </p>
                      </div>
                    </label>
                  </div>

                  <div className={styles.privacyOption}>
                    <label
                      className={`${styles.radioLabel} ${
                        privacy === "private" ? styles.selectedPrivacy : ""
                      }`}
                    >
                      <div className={styles.radioIcon}>
                        <input
                          type="radio"
                          name="privacy"
                          value="private"
                          checked={privacy === "private"}
                          onChange={() => setPrivacy("private")}
                        />
                        <i className="fas fa-lock"></i>
                      </div>
                      <div className={styles.privacyContent}>
                        <span className={styles.privacyTitle}>Private</span>
                        <p className={styles.privacyDescription}>
                          Only selected followers can see this post
                        </p>
                      </div>
                    </label>
                  </div>
                </div>
              </div>

              {privacy === "private" && (
                <div className={styles.settingsSection}>
                  <h3>Who can see this post?</h3>
                  <div className={styles.searchContainer}>
                    <i className="fas fa-search"></i>
                    <input
                      type="text"
                      placeholder="Search followers..."
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      className={styles.searchInput}
                    />
                  </div>

                  <div className={styles.followersList}>
                    {isLoadingFollowers ? (
                      <div className={styles.loadingContainer}>
                        <LoadingSpinner size="small" color="primary" />
                        <p>Loading followers...</p>
                      </div>
                    ) : followersError ? (
                      <div className={styles.errorContainer}>
                        <i className="fas fa-exclamation-circle"></i>
                        <p>{followersError}</p>
                      </div>
                    ) : filteredFollowers().length > 0 ? (
                      filteredFollowers().map((follower) => (
                        <div
                          key={follower.id}
                          className={`${styles.followerItem} ${
                            selectedViewers.includes(follower.id)
                              ? styles.selected
                              : ""
                          }`}
                          onClick={() => toggleViewer(follower.id)}
                        >
                          <div className={styles.followerInfo}>
                            <img
                              src={follower.avatar}
                              width={40}
                              height={40}
                              alt={follower.name}
                            />
                            <span>{follower.name}</span>
                          </div>
                          <div className={styles.checkboxContainer}>
                            {selectedViewers.includes(follower.id) && (
                              <i className="fas fa-check-circle"></i>
                            )}
                          </div>
                        </div>
                      ))
                    ) : (
                      <div className={styles.noResults}>
                        No followers found matching `{searchTerm}`
                      </div>
                    )}
                  </div>
                </div>
              )}

              <button
                className={styles.saveSettingsButton}
                onClick={closeSettingsModal}
              >
                Save Settings
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
