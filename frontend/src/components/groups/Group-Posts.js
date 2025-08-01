'use client';

import { useState, useEffect } from "react";
import styles from '@/styles/Posts.module.css';
import { usePostService } from '@/services/postService';
import { showToast } from '@/components/ui/ToastContainer';
import { useAuth } from '@/context/authcontext';
import { formatRelativeTime } from '@/utils/dateUtils';
import { BASE_URL } from '@/utils/constants';
import { useWebSocket } from '@/services/websocketService';
import ConfirmationModal from '@/components/ui/ConfirmationModal';
import Image from 'next/image';
import { useRouter } from 'next/navigation';



export default function GroupPost({ currentUser, post, onPostUpdated, isDetailView = false }) {
  const {
    likePost,
    addComment,
    getPostComments,
    deletePost,
    deleteComment,
    updatePostLikes,
  } = usePostService();
  const { subscribe } = useWebSocket();
  const [isLiked, setIsLiked] = useState(post.Isliked);
  const [showComments, setShowComments] = useState(isDetailView);
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
  const [comments, setComments] = useState([]);
  const [likesCount, setLikesCount] = useState(post.LikesCount || 0);
  const router = useRouter();
  
  let userdata = null;
  try {
    const raw = localStorage.getItem("userData");
    if (raw) userdata = JSON.parse(raw);
  } catch (e) {
    console.error("Invalid userData in localStorage:", e);
  }
  // Format the post data
  const formattedPost = {
    id: post.ID,
    authorName: post.User
      ? `${post.User.firstName} ${post.User.lastName}`
      : "Unknown User",
    authorImage: post.User?.avatar
      ? `${BASE_URL}/uploads/${post.User.avatar}`
      : "/avatar.jpg",
    content: post.Content,
    timestamp: formatRelativeTime(post.CreatedAt),
    likes: post.LikesCount || 0,
    commentCount: post.CommentsCount || 0,
    image: post.ImagePath?.Valid
      ? `${BASE_URL}/uploads/${post.ImagePath.String}`
      : null,
    video: post.VideoPath?.Valid
      ? `${BASE_URL}/${post.VideoPath.String}`
      : null,
    userId: post.UserID,
    groupName: post.Group?.name || null,
  };
  

  useEffect(() => {
    if (showComments || isDetailView) {
      fetchComments();
    }
  }, [showComments, isDetailView]);

  const fetchComments = async () => {
    if (loadingComments) return;
    setLoadingComments(true);
    try {
      const commentsData = await getPostComments(formattedPost.id);
      if (commentsData) {
        setComments(commentsData.map(comment => ({
          ...comment,
          imageUrl: comment.imageUrl ? `${BASE_URL}${comment.imageUrl}` : null,
          authorName: `${comment.userData.firstName} ${comment.userData.lastName}`,
          authorImage: comment.userData.avatar
            ? comment.userData.avatar.startsWith("http")
              ? comment.userData.avatar
              : `${BASE_URL}/uploads/${comment.userData.avatar}`
            : "/avatar.png",
          timestamp: formatRelativeTime(comment.createdAt)
        })));
      }
    } catch (error) {
      console.error("Error fetching comments:", error);
      setComments([]);
    } finally {
      setLoadingComments(false);
    }
  };

  const handleOptionClick = async (action) => {
    switch (action) {
      case "delete":
        showConfirmation({
          title: "Delete Post",
          message: "Are you sure you want to delete this post? This action cannot be undone.",
          onConfirm: confirmDelete,
        });
        break;
    }
    setShowOptions(false);
  };

  const handleDeleteComment = (comment, postid) => {
    showConfirmation({
      title: "Delete Comment",
      message: "Are you sure you want to delete this comment?",
      onConfirm: () => confirmDeleteComment(comment.id, postid),
    });
  };

  const confirmDeleteComment = async (commentId,postid) => {
    try {
      await deleteComment(commentId,postid);
      setComments(comments.filter((comment) => comment.id !== commentId));
      if (onPostUpdated) onPostUpdated();
    } catch (error) {
      console.error("Error deleting comment:", error);
    } finally {
      setShowConfirmModal(false);
    }
  };

  const confirmDelete = async () => {
    try {
      await deletePost(formattedPost.id);
      if (onPostUpdated) onPostUpdated();
    } catch (error) {
      console.error("Error deleting post:", error);
    } finally {
      setShowConfirmModal(false);
    }
  };

  const handleLike = async () => {
    try {
      const response = await likePost(formattedPost.id);
      setIsLiked(response.isLiked);
      setLikesCount(response.likesCount);
      if (onPostUpdated) onPostUpdated();
    } catch (error) {
      console.error("Error liking post:", error);
      showToast("Failed to like post", "error");
    }
  };

  const handleComment = async (e) => {
    e.preventDefault();
    if (!commentText.trim() && !commentImage) {
      showToast("Please enter a comment or add an image", "error");
      return;
    }

    try {
      const newComment = await addComment(formattedPost.id, commentText, commentImage);
      setComments(prev => [...prev, {
        ...newComment,
        authorName: currentUser ? `${currentUser.firstName} ${currentUser.lastName}` : "You",
        authorImage: currentUser ? `${currentUser.avatar}` : "/avatar.png",
        timestamp: "just now"
      }]);
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

  const showConfirmation = (config) => {
    setConfirmModalConfig(config);
    setShowConfirmModal(true);
  };

  const handleUserClick = (userId) => {
    router.push(`/profile/${userId}`);
  };

  return (
    <article className={styles.post}>
      <div className={styles.postHeader}>
        <div className={styles.postAuthor}>
          <img
            src= {formattedPost.authorImage}
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
              <span>{post.groupName}</span>
            </div>
          </div>
        </div>
        {currentUser?.id === post.userId && (
          <div className={styles.postOptionsContainer}>
            <button className={styles.postOptions} onClick={() => setShowOptions(!showOptions)}>
              <i className="fas fa-ellipsis-h"></i>
            </button>
            {showOptions && (
              <div className={styles.optionsDropdown}>
                <button onClick={() => handleOptionClick("delete")}>
                  <i className="fas fa-trash-alt"></i> Delete Post
                </button>
              </div>
            )}
          </div>
        )}
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
          <span>{formattedPost.likes} likes</span>
        </div>
        <div className={styles.engagement}>
          <span>{formattedPost.commentCount} comments</span>
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

      {(showComments || isDetailView) && (
        <div className={styles.commentsSection}>
          <form onSubmit={handleComment} className={styles.commentForm}>
            <img
              src={`${BASE_URL}uploads/${userdata.avatar}` || "/avatar.png"}
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

          {comments && comments.map((comment) => {
            const isCommentOwner = currentUser?.id === comment.userData?.id;

            return (
              <div key={`comment-${comment.id}`} className={styles.comment}>
                <Image
                  src={comment.authorImage ? `${comment.authorImage}` : "/avatar.png"}
                  alt={`${comment.authorName}'s avatar`}
                  height={32}
                  width={32}
                  className={styles.commentAvatar}
                  onError={(e) => {
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
                        onClick={() => handleDeleteComment(comment, formattedPost.id)}
                      >
                        Delete
                      </button>
                    )}
                    <span>{comment.timestamp}</span>
                  </div>
                </div>
              </div>
            );
          })}
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
