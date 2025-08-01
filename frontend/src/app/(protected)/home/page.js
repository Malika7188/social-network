"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import Header from "@/components/header/Header";
import LeftSidebar from "@/components/sidebar/LeftSidebar";
import RightSidebar from "@/components/sidebar/RightSidebar";
import CreatePost from "@/components/posts/CreatePost";
import Post from "@/components/posts/Post";
import styles from "@/styles/page.module.css";
import postStyles from "@/styles/Posts.module.css";
import ProtectedRoute from "@/components/auth/ProtectedRoute";
// import ChatSidebarFloat from "@/components/chat/ChatSidebarFloat";
import { usePostService } from "@/services/postService";
import { showToast } from "@/components/ui/ToastContainer";

const contacts = [
  {
    id: 1,
    name: "Jane Smith",
    avatar: "/avatar1.png",
    online: true,
    unreadCount: 2,
    messages: [
      { id: 1, content: "Hey there!", isSent: false },
      { id: 2, content: "Hi! How are you?", isSent: true },
    ],
  },
  {
    id: 2,
    name: "John Doe",
    avatar: "/avatar4.png",
    online: false,
    unreadCount: 0,
    messages: [],
  },
  // Add more contacts as needed
];

const pageSize = 10;

export default function Home() {
  const { getFeedPosts, newPosts, getAndClearNewPosts } =
    usePostService() || {};
  const [posts, setPosts] = useState([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const observer = useRef();

  // Function to load posts
  const loadPosts = useCallback(
    async (pageNum = 1, replace = false) => {
      if (loading) return;
      if (!getFeedPosts) {
        console.error("getFeedPosts function is not available");
        return;
      }

      setLoading(true);
      try {
        let data = await getFeedPosts(pageNum, pageSize);

        // Ensure data is an array
        if (!data || !Array.isArray(data)) {
          data = [];
        }

        if (data.length < pageSize) {
          setHasMore(false);
        }

        if (replace) {
          setPosts(data);
        } else {
          setPosts((prev) => [...(prev || []), ...data]);
        }

        setPage(pageNum);
      } catch (error) {
        console.error("Error loading posts:", error);
      } finally {
        setLoading(false);
      }
    },
    [getFeedPosts, loading]
  );

  // Initial load - adding a ref to prevent multiple calls
  const initialLoadRef = useRef(false);
  useEffect(() => {
    if (!initialLoadRef.current && getFeedPosts) {
      loadPosts(1, true);
      initialLoadRef.current = true;
    }
  }, [loadPosts, getFeedPosts]);

  // Replace the refreshTrigger effect with a more controlled approach
  const refreshFeed = useCallback(() => {
    loadPosts(1, true);
  }, [loadPosts]);

  // Handle new posts from WebSocket
  useEffect(() => {
    if (Array.isArray(newPosts) && newPosts.length > 0 && getAndClearNewPosts) {
      const incomingPosts = getAndClearNewPosts();
      if (Array.isArray(incomingPosts) && incomingPosts.length > 0) {
        setPosts((prev) => [...incomingPosts, ...(prev || [])]);
      }
    }
  }, [newPosts, getAndClearNewPosts]);

  // Infinite scroll setup
  const lastPostElementRef = useCallback(
    (node) => {
      if (loading) return;
      if (observer.current) observer.current.disconnect();

      observer.current = new IntersectionObserver((entries) => {
        if (entries[0]?.isIntersecting && hasMore) {
          loadPosts(page + 1);
        }
      });

      if (node) observer.current.observe(node);
    },
    [loading, hasMore, page, loadPosts]
  );

  return (
    <ProtectedRoute>
      <Header />
      <div className={styles.container}>
        <aside className="styles.leftSidebar">
          <LeftSidebar />
        </aside>
        <main className={styles.mainContent}>
          <section className={styles.createPostSection}>
            <CreatePost />
          </section>
          <section className={styles.feedSection}>
            {Array.isArray(posts) &&
              posts.map((post, index) => {
                if (!post) return null;
                // Create a unique key that combines post.id with the index
                const uniqueKey = `post-${post.id || index}-${index}`;

                if ((posts?.length || 0) === index + 1) {
                  return (
                    <div
                      ref={lastPostElementRef}
                      key={`${uniqueKey}-container`}
                    >
                      <Post
                        key={uniqueKey}
                        post={post}
                        onPostUpdated={refreshFeed}
                      />
                    </div>
                  );
                } else {
                  return (
                    <Post
                      key={uniqueKey}
                      post={post}
                      onPostUpdated={refreshFeed}
                    />
                  );
                }
              })}

            {loading && (
              <div className={postStyles.loadingState}>
                <div className={postStyles.spinner}></div>
                <p>Loading posts...</p>
              </div>
            )}

            {!loading &&
              !hasMore &&
              Array.isArray(posts) &&
              (posts?.length || 0) > 0 && (
                <div className={postStyles.endOfFeed}>
                  <p>You&apos;ve reached the end of your feed!</p>
                </div>
              )}
          </section>
        </main>
        <aside className={styles.rightSidebar}>
          <RightSidebar />
        </aside>
      </div>
      {/* <ChatSidebarFloat contacts={contacts} /> */}
    </ProtectedRoute>
  );
}
