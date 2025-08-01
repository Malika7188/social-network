import { useAuth } from "@/context/authcontext";
import { showToast } from "@/components/ui/ToastContainer";
import { handleApiError } from "@/utils/errorHandler";
import { useWebSocket, EVENT_TYPES } from "./websocketService";
import { useState, useEffect, useCallback } from "react";
import { BASE_URL } from "@/utils/constants";

export const usePostService = () => {
  const { authenticatedFetch } = useAuth();
  const { subscribe } = useWebSocket();
  const [newPosts, setNewPosts] = useState([]);
  const [allPosts, setAllPosts] = useState([]);
  const { currentUserId } = useAuth();

  // Subscribe to post_created events
  useEffect(() => {
    const unsubscribePostCreated = subscribe(
      EVENT_TYPES.POST_CREATED,
      (payload) => {
        if (payload && payload.post) {
          // Destructure user data with defaults
          const {
            UserData = {
              firstName: "Unknown",
              lastName: "User",
              avatar: "/avatar.png",
            },
          } = payload.post;

          // Ensure avatar has absolute URL if it's a relative path
          const avatar = UserData.avatar?.startsWith("http")
            ? UserData.avatar
            : `${BASE_URL}/uploads/${UserData.avatar}`;

          // Ensure the post has all required fields
          const formattedPost = {
            ...payload.post,
            id: payload.post.ID,
            userData: {
              ...UserData,
              avatar, // Use the processed avatar URL
            },
            likesCount: payload.post.LikesCount || 0,
            comments: payload.post.comments || [],
            createdAt: payload.post.createdAt || new Date().toISOString(),
            imageUrl: payload.post.ImagePath?.String
              ? `/uploads/${payload.post.ImagePath.String}`
              : null,
            videoUrl: payload.post.VideoPath?.String
              ? `/uploads/${payload.post.VideoPath.String}`
              : null,
            content: payload.post.Content,
            privacy: payload.post.Privacy,
          };

          setNewPosts((prev) => [formattedPost, ...prev]);
        }
      }
    );
  }, [subscribe]);

  const createPost = async (formData) => {
    try {
      const response = await authenticatedFetch("posts", {
        method: "POST",
        body: formData,
      });

      // Check if response is not ok
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || errorData.error || "Failed to create post"
        );
      }

      const data = await response.json();
      showToast("Post created successfully!", "success");
      return data;
    } catch (error) {
      console.error("Error creating post:", error);
      showToast(error.message || "Error creating post", "error");
      throw error;
    }
  };

  const getFeedPosts = async (page = 1, pageSize = 10) => {
    try {
      const response = await authenticatedFetch(
        `posts?page=${page}&pageSize=${pageSize}`,
        {
          method: "GET",
        }
      );

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || errorData.error || "Failed to fetch posts"
        );
      }

      const data = await response.json();

      // Update allPosts state when fetching posts
      if (page === 1) {
        setAllPosts(data);
      } else {
        setAllPosts((prev) => [...prev, ...data]);
      }

      return data;
    } catch (error) {
      console.error("Error fetching posts:", error);
      showToast(error.message || "Error fetching posts", "error");
      throw error;
    }
  };

  const getUserPosts = async (id) => {
    try {
      const response = await authenticatedFetch(
        `posts/user/${id}`,
        {
          method: "GET",
        }
      );

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || errorData.error || "Failed to fetch posts"
        );
      }
      const data = await response.json();
      (data)
      return data;
    } catch (error) {
      console.error("Error fetching posts:", error);
      showToast(error.message || "Error fetching posts", "error");
      throw error;
    }
  };

  const getUserPhotos = async (id) => {
    try {
      const response = await authenticatedFetch(
        `posts/photos/${id}`,
        {
          method: "GET",
        }
      );

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || errorData.error || "Failed to fetch photos"
        );
      }
      const data = await response.json();
      return data;
    } catch (error) {
      console.error("Error fetching photos:", error);
      showToast(error.message || "Error fetching posts", "error");
      throw error;
    }
  };

  const likePost = async (postId) => {
    try {
      const response = await authenticatedFetch(`posts/like/${postId}`, {
        method: "POST",
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || errorData.error || "Failed to like post"
        );
      }

      const data = await response.json();

      // Use the utility function to update posts
      updatePostLikes(postId, data.likesCount, data.isLiked);

      return data;
    } catch (error) {
      console.error("Error liking post:", error);
      showToast(error.message || "Error liking post", "error");
      throw error;
    }
  };

  const addComment = async (postId, content, image) => {
    try {
      const formData = new FormData();
      formData.append("content", content);
      if (image) {
        formData.append("image", image);
      }

      const response = await authenticatedFetch(`posts/comments/${postId}`, {
        method: "POST",
        body: formData,
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || errorData.error || "Failed to add comment"
        );
      }

      return await response.json();
    } catch (error) {
      console.error("Error adding comment:", error);
      showToast(error.message || "Error adding comment", "error");
      throw error;
    }
  };

  const getPostComments = async (postId) => {
    try {
      const response = await authenticatedFetch(`posts/comments/${postId}`, {
        method: "GET",
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || errorData.error || "Failed to fetch comments"
        );
      }

      return await response.json();
    } catch (error) {
      console.error("Error fetching comments:", error);
      showToast(error.message || "Error fetching comments", "error");
      throw error;
    }
  };

  const deletePost = async (postId) => {
    try {
      const response = await authenticatedFetch(`posts`, {
        method: "DELETE",
        body: JSON.stringify({ postId }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || errorData.error || "Failed to delete post"
        );
      }

      showToast("Post deleted successfully!", "success");
      return true;
    } catch (error) {
      console.error("Error deleting post:", error);
      showToast(error.message || "Error deleting post", "error");
      throw error;
    }
  };

  const deleteComment = async (commentId, postid) => {
    try {
      const response = await authenticatedFetch(`posts/comments/${commentId}?postid=${postid}`, {
        method: "DELETE",
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || errorData.error || "Failed to delete comment"
        );
      }

      showToast("Comment deleted successfully!", "success");
      return true;
    } catch (error) {
      console.error("Error deleting comment:", error);
      showToast(error.message || "Error deleting comment", "error");
      throw error;
    }
  };

  // Clear new posts and return them
  const getAndClearNewPosts = useCallback(() => {
    const posts = [...newPosts];
    setNewPosts([]);
    return posts;
  }, [newPosts]);

  const updatePostLikes = useCallback(
    (postId, likesCount, isLiked) => {
      // Convert postId to number to ensure consistent comparison
      const numericPostId = Number(postId);

      // Debug: Check if posts exist
      const postExistsInNew = newPosts.some(
        (post) => Number(post.id) === numericPostId
      );
      const postExistsInAll = allPosts.some(
        (post) => Number(post.id) === numericPostId
      );

      // Update the newPosts state
      setNewPosts((prevPosts) =>
        prevPosts.map((post) =>
          Number(post.id) === numericPostId
            ? { ...post, likesCount, isLiked }
            : post
        )
      );

      // Update allPosts state as well
      setAllPosts((prevPosts) =>
        prevPosts.map((post) =>
          Number(post.id) === numericPostId
            ? { ...post, likesCount, isLiked }
            : post
        )
      );

      // Return the updated data so components can use it
      return { postId: numericPostId, likesCount, isLiked };
    },
    [newPosts, allPosts]
  );

  const followUser = async (userId, userName) => {
    try {
      const response = await authenticatedFetch("follow/follow", {
        method: "POST",
        body: JSON.stringify({ userId }),
      });
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || errorData.error || "Failed to follow User"
        );
      }

      const data = await response.json();

      if (data.success) {
        if (data.autoFollowed) {
          showToast(`You are now following ${userName}`, "success");
        } else {
          showToast(`Follow request Sent`, "success");
        }
      }
      return true;
    } catch (error) {
      showToast(error.message || "Error Following user", "error");
    }
  };
  const unfollowUser = async (userId, userName) => {
    try {
      const response = await authenticatedFetch("follow/unfollow", {
        method: "POST",
        body: JSON.stringify({ userId }),
      });
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || errorData.error || "Failed to follow User"
        );
      }

      const data = await response.json();

      if (data.success) {
        showToast(`You unfollowed ${userName}`, "success");
      }
      return true;
    } catch (error) {
      showToast(error.message || "Error UnFollowing user", "error");
      throw err;
    }
  };

  return {
    createPost,
    getFeedPosts,
    getUserPosts,
    getUserPhotos,
    likePost,
    addComment,
    getPostComments,
    deletePost,
    deleteComment,
    newPosts,
    getAndClearNewPosts,
    updatePostLikes,
    allPosts,
    setAllPosts,
    followUser,
    unfollowUser,
  };
};
