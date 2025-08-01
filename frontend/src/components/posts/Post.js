"use client";

import { useState, useEffect } from "react";
import styles from "@/styles/Posts.module.css";
import { usePostService } from "@/services/postService";
import { showToast } from "@/components/ui/ToastContainer";
import { useAuth } from "@/context/authcontext";
import { formatRelativeTime } from "@/utils/dateUtils";
import { BASE_URL } from "@/utils/constants";
import { useWebSocketContext } from "@/context/websocketContext";
import { EVENT_TYPES } from "@/services/websocketService";
import ConfirmationModal from "@/components/ui/ConfirmationModal";
import { useWebSocket } from "@/services/websocketService";
import Image from "next/image";
import { useRouter } from 'next/navigation';

export default function Post({ post, onPostUpdated }) {
  const {
    likePost,
    addComment,
    getPostComments,
    deletePost,
    deleteComment,
    updatePostLikes,
    followUser,
    unfollowUser,
  } = usePostService();
  const { subscribe } = useWebSocket();
  const { currentUser } = useAuth();
  const [isLiked, setIsLiked] = useState(false);
  const [showComments, setShowComments] = useState(false);
  const [commentText, setCommentText] = useState("");
  const [showOptions, setShowOptions] = useState(false);
  const [loadingComments, setLoadingComments] = useState(false);
  const [commentImage, setCommentImage] = useState(null);
  const [commentImagePreview, setCommentImagePreview] = useState(null);
  const [showConfirmModal, setShowConfirmModal] = useState(false);
  const [confirmModalConfig, setConfirmModalConfig] = useState({
    title: "",
    message: "",
    onConfirm: () => {},
  });
  const router = useRouter();

  // Format the post data to match component expectations
  const formattedPost = {
    id: post.id,
    authorName: post.userData
      ? `${post.userData.firstName} ${post.userData.lastName}`
      : "Unknown User",
    authorImage: post.userData?.avatar
      ? post.userData.avatar.startsWith("http")
        ? post.userData.avatar
        : `${BASE_URL}${post.userData.avatar}`
      : "/avatar.png",
    content: post.content,
    timestamp: formatRelativeTime(post.createdAt),
    privacy: post.privacy,
    likes: post.likesCount || 0,
    commentCount: post.comments?.length || 0,
    shares: post.shares || 0,
    image: post.imageUrl ? `${BASE_URL}${post.imageUrl}` : null,
    video: post.videoUrl ? `${BASE_URL}${post.videoUrl}` : null,
    userId: post.userId,
  };

  // Use formatted data instead of raw post data
  const [likesCount, setLikesCount] = useState(formattedPost.likes);
  const [comments, setComments] = useState(post.comments || []);

  // Add click outside handler
  const handleClickOutside = (e) => {
    if (!e.target.closest(`.${styles.postOptionsContainer}`)) {
      setShowOptions(false);
    }
  };

  // Add event listener when dropdown is open
  useEffect(() => {
    if (showOptions) {
      document.addEventListener("click", handleClickOutside);
    }
    return () => {
      document.removeEventListener("click", handleClickOutside);
    };
  }, [showOptions]);

  // Fetch comments when comment section is opened
  useEffect(() => {
    if (showComments) {
      fetchComments();
    }
  }, [showComments]);

  const fetchComments = async () => {
    if (loadingComments) return;

    setLoadingComments(true);
    try {
      const commentsData = await getPostComments(post.id);
      if (commentsData) {
        const formattedComments = commentsData.map((comment) => ({
          ...comment,
          imageUrl: comment.imageUrl ? `${BASE_URL}${comment.imageUrl}` : null,
          authorName: `${comment.userData.firstName} ${comment.userData.lastName}`,
          authorImage: comment.userData.avatar
            ? comment.userData.avatar.startsWith("http")
              ? comment.userData.avatar
              : `${BASE_URL}/uploads/${comment.userData.avatar}`
            : "/avatar.png",
        }));

        formattedComments.forEach((comment) => {
          comment.timestamp = formatRelativeTime(comment.createdAt);
        });

        setComments(formattedComments);
      } else {
        setComments([]);
      }
    } catch (error) {
      console.error("Error fetching comments:", error);
      setComments([]);
    } finally {
      setLoadingComments(false);
    }
  };

  // Add state for comment deletion
  const [commentToDelete, setCommentToDelete] = useState(null);

  // Function to show confirmation modal with custom config
  const showConfirmation = (config) => {
    setConfirmModalConfig(config);
    setShowConfirmModal(true);
  };
  // Update handleOptionClick to use the new confirmation approach
  const handleOptionClick = async (action) => {
    switch (action) {
      case "edit":
        break;
      case "delete":
        showConfirmation({
          title: "Delete Post",
          message:
            "Are you sure you want to delete this post? This action cannot be undone.",
          onConfirm: confirmDelete,
        });
        break;
      case "follow":
        const userId = post.userId ?? post.UserID;
        await followUser(userId, formattedPost.authorName);
        break;
      case "unfollow":
        await unfollowUser(post.userId, formattedPost.authorName);
        break;
    }
    setShowOptions(false);
  };

  // Function to handle comment deletion
  const handleDeleteComment = (comment, postid) => {
    showConfirmation({
      title: "Delete Comment",
      message:
        "Are you sure you want to delete this comment? This action cannot be undone.",
      onConfirm: () => confirmDeleteComment(comment.id, postid),
    });
  };

  // Function to confirm comment deletion
  const confirmDeleteComment = async (commentId, postid) => {
    try {
      await deleteComment(commentId, postid);
      // Remove the deleted comment from the state
      setComments(comments.filter((comment) => comment.id !== commentId));
      if (onPostUpdated) onPostUpdated();
    } catch (error) {
      console.error("Error deleting comment:", error);
    } finally {
      setShowConfirmModal(false);
    }
  };

  // Update the existing confirmDelete function
  const confirmDelete = async () => {
    try {
      await deletePost(post.id);
      if (onPostUpdated) onPostUpdated();
    } catch (error) {
      console.error("Error deleting post:", error);
    } finally {
      setShowConfirmModal(false);
    }
  };

  const handleLike = async () => {
    if (!post.id || post.id === 0) {
      showToast("Cannot like this post yet. Please refresh the page.", "error");
      return;
    }

    try {
      const response = await likePost(post.id);
      // Update local state
      setIsLiked(response.isLiked);
      setLikesCount(response.likesCount);

      if (onPostUpdated) onPostUpdated();
    } catch (error) {
      console.error("Error liking post:", error);
      showToast("Failed to like post", "error");
    }
  };

  const avatar = currentUser?.avatar
    ? currentUser.avatar.startsWith("http")
      ? currentUser.avatar
      : `${BASE_URL}/uploads/${currentUser.avatar}`
    : "/avatar.png";

  const currentUserFullName =
    currentUser?.firstName + " " + currentUser?.lastName;

  const handleComment = async (e) => {
    e.preventDefault();
    if (!commentText.trim() && !commentImage) {
      showToast("Please enter a comment or add an image", "error");
      return;
    }

    try {
      let newComment;
      if (post.id) {
        newComment = await addComment(post.id, commentText, commentImage);
      } else {
        console.error("Post ID is required to add a comment");
        return;
      }

      const commentImageUrl = newComment.imagePath
        ? `${BASE_URL}/uploads/comments/${newComment.imagePath}`
        : null;
      // Format the new comment to match our component's expected format
      const formattedComment = {
        id: newComment.id,
        authorName: currentUserFullName || "You",
        authorImage: avatar,
        content: commentText,
        timestamp: "just now",
        imageUrl: commentImageUrl,
      };

      setComments((prev) => [...prev, formattedComment]);
      setCommentText("");
      setCommentImage(null);
      setCommentImagePreview(null);

      if (onPostUpdated) onPostUpdated();
    } catch (error) {
      console.error("Error adding comment:", error);
    }
  };

  const handleCommentImageChange = (e) => {
    const file = e.target.files[0];
    if (!file) return;

    if (!file.type.startsWith("image/")) {
      showToast("Please select an image file", "error");
      return;
    }

    setCommentImage(file);
    setCommentImagePreview(URL.createObjectURL(file));
  };

  const removeCommentImage = () => {
    setCommentImage(null);
    if (commentImagePreview) {
      URL.revokeObjectURL(commentImagePreview);
      setCommentImagePreview(null);
    }
  };

  const isCurrentUserPost = currentUser?.id === post.userId;

  // Update the comment rendering to include delete button for owner
  const renderComments = (postid) => {
    return comments.map((comment) => {
      const isCommentOwner = currentUser?.id === comment.userData?.id;

      return (
        <div key={`comment-${comment.id}`} className={styles.comment}>
          <Image
            src={comment.authorImage || "/avatar.png"}
            alt={`${comment.authorName}'s avatar`}
            height={32}
            width={32}
            className={styles.commentAvatar}
            onError={(e) => {
              // Fallback if image fails to load
              e.target.src = "/avatar.png";
            }}
          />
          <div className={styles.commentContent}>
            <div className={styles.commentBubble}>
              <h4>{comment.authorName}</h4>
              <p>{comment.content}</p>
              {comment.imageUrl && (
                <div className={styles.commentImage}>
                  <img src={comment.imageUrl} alt="Comment attachment" />
                </div>
              )}
            </div>
            <div className={styles.commentActions}>
              <button>Like</button>
              <button>Reply</button>
              {isCommentOwner && (
                <button
                  className={styles.deleteCommentBtn}
                  onClick={() => handleDeleteComment(comment, postid)}
                >
                  Delete
                </button>
              )}
              <span>{comment.timestamp}</span>
            </div>
          </div>
        </div>
      );
    });
  };

  // Add this useEffect to handle real-time like updates from WebSocket
  useEffect(() => {
    // Check if the post ID exists
    if (!post.id) return;

    // Subscribe to post like updates from WebSocket
    const unsubscribe = subscribe(EVENT_TYPES.POST_LIKED, (payload) => {
      // Make sure this update is for our post
      if (Number(payload.postId) === Number(post.id)) {
        setIsLiked(payload.isLiked);
        setLikesCount(payload.likesCount);
      }
    });

    // Clean up subscription when component unmounts
    return () => {
      if (typeof unsubscribe === "function") {
        unsubscribe();
      }
    };
  }, [post.id, subscribe]);

  const handleUserClick = (userId) => {
    // Navigate to profile page regardless of whether it's the current user or not
    router.push(`/profile/${userId}`);
  };

  return (
    <article className={styles.post}>
      <div className={styles.postHeader}>
        <div className={styles.postAuthor}>
          <img
            src={formattedPost.authorImage}
            alt={formattedPost.authorName}
            className={styles.authorAvatar}
            onClick={() => handleUserClick(formattedPost.userId)}
            style={{ cursor: 'pointer' }}
          />
          <div className={styles.authorInfo}>
            <h3 
              onClick={() => handleUserClick(formattedPost.userId)}
              style={{ cursor: 'pointer' }}
              className={styles.authorName}
            >
              {formattedPost.authorName}
            </h3>
            <div className={styles.postMeta}>
              <span>{formattedPost.timestamp}</span>
              <span className={styles.dot}>•</span>
              <i
                className={`fas ${
                  formattedPost.privacy === "public"
                    ? "fa-globe-americas"
                    : formattedPost.privacy === "private"
                    ? "fa-lock"
                    : "fa-user-friends"
                }`}
              ></i>
            </div>
          </div>
        </div>
        <div className={styles.postOptionsContainer}>
          <button
            className={styles.postOptions}
            onClick={(e) => {
              e.stopPropagation();
              setShowOptions(!showOptions);
            }}
          >
            <i className="fas fa-ellipsis-h"></i>
          </button>
          {showOptions && (
            <div className={styles.optionsDropdown}>
              {isCurrentUserPost && (
                <>
                  <button onClick={() => handleOptionClick("edit")}>
                    <i className="fas fa-edit"></i>
                    Edit Post
                  </button>
                  <button onClick={() => handleOptionClick("delete")}>
                    <i className="fas fa-trash-alt"></i>
                    Delete Post
                  </button>
                  <div className={styles.dropdownDivider} />
                </>
              )}
              <button onClick={() => handleOptionClick("follow")}>
                <i className="fas fa-user-plus"></i>
                Follow
              </button>
              <button onClick={() => handleOptionClick("unfollow")}>
                <i className="fas fa-user-minus"></i>
                Unfollow
              </button>
            </div>
          )}
        </div>
      </div>

      <div className={styles.postContent}>
        <p>{formattedPost.content}</p>
        {formattedPost.image && (
          <div className={styles.postMedia}>
            <img src={formattedPost.image} alt="Post content" />
          </div>
        )}
        {formattedPost.video && (
          <div className={styles.postMedia}>
            <video
              src={formattedPost.video}
              controls
              className={styles.videoContent}
            />
          </div>
        )}
      </div>

      <div className={styles.postStats}>
        <div className={styles.reactions}>
          <span className={styles.reactionIcons}>
            <i className="fas fa-thumbs-up" style={{ color: "#2078f4" }}></i>
            <i className="fas fa-heart" style={{ color: "#f33e58" }}></i>
          </span>
          <span>{likesCount} likes</span>
        </div>
        <div className={styles.engagement}>
          <span>{formattedPost.commentCount} comments</span>
          <span>{formattedPost.shares} shares</span>
        </div>
      </div>

      <div className={styles.postActions}>
        <button
          className={`${styles.actionButton} ${isLiked ? styles.liked : ""}`}
          onClick={handleLike}
        >
          <i className="fas fa-thumbs-up"></i>
          <span>Like</span>
        </button>
        <button
          className={styles.actionButton}
          onClick={() => setShowComments(!showComments)}
        >
          <i className="fas fa-comment"></i>
          <span>Comment</span>
        </button>
        <button className={styles.actionButton}>
          <i className="fas fa-share"></i>
          <span>Share</span>
        </button>
      </div>

      {showComments && (
        <div className={styles.commentsSection}>
          <form onSubmit={handleComment} className={styles.commentForm}>
            <img
              src={avatar}
              alt="Your avatar"
              className={styles.commentAvatar}
            />
            <div className={styles.commentInput}>
              <input
                type="text"
                placeholder="Write a comment..."
                value={commentText}
                onChange={(e) => setCommentText(e.target.value)}
              />
              <label className={styles.commentImageLabel}>
                <i className="fas fa-camera"></i>
                <input
                  type="file"
                  accept="image/*"
                  onChange={handleCommentImageChange}
                  style={{ display: "none" }}
                />
              </label>
              <button
                type="submit"
                disabled={!commentText.trim() && !commentImage}
              >
                <i className="fas fa-paper-plane"></i>
              </button>
            </div>
          </form>

          {commentImagePreview && (
            <div className={styles.commentImagePreview}>
              <img src={commentImagePreview} alt="Comment attachment" />
              <button
                className={styles.removeCommentImage}
                onClick={removeCommentImage}
              >
                <i className="fas fa-times"></i>
              </button>
            </div>
          )}

          {loadingComments && (
            <div className={styles.commentsLoading}>
              <div className={styles.spinner}></div>
              <p>Loading comments...</p>
            </div>
          )}

          {!loadingComments && comments && comments.length === 0 && (
            <div className={styles.noComments}>
              <p>No comments yet. Be the first to comment!</p>
            </div>
          )}

          {comments && renderComments(formattedPost.id)}
        </div>
      )}

      <ConfirmationModal
        isOpen={showConfirmModal}
        onClose={() => setShowConfirmModal(false)}
        onConfirm={confirmModalConfig.onConfirm}
        title={confirmModalConfig.title}
        message={confirmModalConfig.message}
      />
    </article>
  );
}
