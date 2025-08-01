"use client";

import React, { useState, useEffect, useCallback, useRef } from "react";
import Post from "./Post";
import styles from "@/styles/PostList.module.css";
import postStyles from "@/styles/Posts.module.css";
import { usePostService } from "@/services/postService";
import { BASE_URL } from "@/utils/constants";
import { showToast } from "@/components/ui/ToastContainer";

const PostList = ({ userData }) => {
  const { getUserPosts } = usePostService();
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const isMounted = useRef(true);
  const initialLoadRef = useRef(false);

  // Function to load user posts
  const loadUserPosts = useCallback(async () => {
    if (!userData || !userData.id) {
      setLoading(false);
      return;
    }

    try {
      setLoading(true);
      const userPosts = await getUserPosts(userData.id);

      // Only update state if component is still mounted
      if (isMounted.current) {
        // Ensure data is an array
        if (!userPosts || !Array.isArray(userPosts)) {
          setPosts([]);
        } else {
          setPosts(userPosts);
        }
      }
    } catch (err) {
      console.error("Error fetching user posts:", err);
      if (isMounted.current) {
        setError(err.message || "Failed to load posts");
        showToast("Failed to load posts", "error");
      }
    } finally {
      if (isMounted.current) {
        setLoading(false);
      }
    }
  }, [userData?.id, getUserPosts]);

  // Initial load of posts - using a ref to prevent multiple calls
  useEffect(() => {
    // Set up the isMounted ref
    isMounted.current = true;

    // Only load posts once on component mount
    if (!initialLoadRef.current && getUserPosts && userData?.id) {
      initialLoadRef.current = true;
      loadUserPosts();
    }

    // Cleanup function
    return () => {
      isMounted.current = false;
    };
  }, [loadUserPosts, getUserPosts, userData?.id]);

  // Function to refresh posts (can be used after updates)
  const refreshPosts = useCallback(() => {
    if (initialLoadRef.current) {
      loadUserPosts();
    }
  }, [loadUserPosts]);

  return (
    <section className={styles.postListContainer}>
      {Array.isArray(posts) && posts.length > 0 ? (
        posts.map((post, index) => {
          if (!post) return null;
          // Create a unique key that combines post.id with the index
          const uniqueKey = `post-${post.id || index}-${index}`;

          return (
            <Post key={uniqueKey} post={post} onPostUpdated={refreshPosts} />
          );
        })
      ) : loading ? (
        <div className={postStyles.loadingState}>
          <div className={postStyles.spinner}></div>
          <p>Loading posts...</p>
        </div>
      ) : error ? (
        <div className={postStyles.error}>
          <p>Error: {error}</p>
        </div>
      ) : (
        <div className={postStyles.endOfFeed}>
          <p>No posts found</p>
        </div>
      )}
    </section>
  );
};

export default PostList;
