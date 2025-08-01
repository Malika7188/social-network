import { useAuth } from "@/context/authcontext";
import { showToast } from "@/components/ui/ToastContainer";
import { useState, useEffect, useCallback } from "react";
import { useWebSocket, EVENT_TYPES } from "./websocketService";
import { BASE_URL } from "@/utils/constants";

export const useFriendService = () => {
  const { authenticatedFetch } = useAuth();
  const [friendRequests, setFriendRequests] = useState([]);
  const [SuggestedUsers, setSuggestedUsers] = useState([]);
  const [contacts, setContacts] = useState([]);
  const [isLoadingRequests, setIsLoadingRequests] = useState(false);
  const [isLoadingContacts, setIsLoadingContacts] = useState(false);
  const { subscribe } = useWebSocket();

  // Fetch pending friend requests
  const fetchFriendRequests = async () => {
    setIsLoadingRequests(true);
    try {
      const response = await authenticatedFetch("follow/pending-requests", {
        method: "GET",
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message ||
          errorData.error ||
          "Failed to fetch friend requests"
        );
      }

      const data = await response.json();

      if (data) {
        // Transform the data to match our component's expected format
        const formattedRequests = data.map((request) => ({
          id: request.ID,
          name: request.FollowerName,
          image: request.FollowerAvatar
            ? `${BASE_URL}/uploads/${request.FollowerAvatar}`
            : "/avatar.png",
          mutualFriends: request.MutualFriends || 0,
          followerId: request.FollowerID,
        }));
        setFriendRequests(formattedRequests);

        return formattedRequests;
      }
      return [];
    } catch (error) {
      console.error("Error fetching friend requests:", error);
      return [];
    } finally {
      setIsLoadingRequests(false);
    }
  };

  const fetchSuggestedUsers = async () => {
    setIsLoadingRequests(true);
    try {
      const response = await authenticatedFetch("follow/suggested-friends", {
        method: "GET",
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message ||
          errorData.error ||
          "Failed to fetch friend requests"
        );
      }

      const data = await response.json();

      if (data) {
        // Transform the data to match our component's expected format
        const formattedSuggested = (data.suggestions || []).map((Suggested) => ({
          SuggestedID: Suggested.id,
          name: `${Suggested.firstName} ${Suggested.lastName} (${Suggested.nickname})`,
          image: Suggested.avatar
            ? `${BASE_URL}/uploads/${Suggested.avatar}`
            : "/avatar.png",
          mutualFriends: Suggested.mutualFriends || 0,
          isPublic: Suggested.isPublic,
          isOnline: Suggested.isOnline,
        }));
        setSuggestedUsers(formattedSuggested);

        return formattedSuggested;
      }
      return [];
    } catch (error) {
      console.error("Error fetching friend requests:", error);
      return [];
    } finally {
      setIsLoadingRequests(false);
    }
  };

  // Fetch contacts (followers/following)
  const fetchContacts = useCallback(async () => {
    setIsLoadingContacts(true);
    try {
      const response = await authenticatedFetch("follow/following", {
        method: "GET",
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || errorData.error || "Failed to fetch contacts"
        );
      }

      const data = await response.json();

      if (!data) {
        return [];
      }
      // Transform the data to match our component's expected format
      const formattedContacts = data.map((contact) => ({
        id: contact.ID,
        name: contact.UserName,
        image: contact.UserAvatar
          ? `${BASE_URL}/uploads/${contact.UserAvatar}`
          : "/avatar.png",
        isOnline: contact.IsOnline || false,
        contactId: contact.FollowingID,
      }));

      setContacts(formattedContacts);
      return formattedContacts;
    } catch (error) {
      console.error("Error fetching contacts:", error);
      return [];
    } finally {
      setIsLoadingContacts(false);
    }
  }, [authenticatedFetch]);

  // Accept a friend request
  const acceptFriendRequest = async (followerId) => {

    try {
      const response = await authenticatedFetch("follow/accept", {
        method: "POST",
        body: JSON.stringify({ followerId }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message ||
          errorData.error ||
          "Failed to accept friend request"
        );
      }


      const data = await response.json();


      if (data.success) {
        showToast("Friend request accepted", "success");
        // Remove the request from the list
        setFriendRequests((prev) =>
          prev.filter((request) => request.followerId !== followerId)
        );
        // Refresh contacts list
        fetchContacts();
      }

      return true;
    } catch (error) {
      showToast(error.message || "Error accepting friend request", "error");
      return false;
    }
  };

  // Decline a friend request
  const declineFriendRequest = async (followerId) => {
    try {
      const response = await authenticatedFetch("follow/decline", {
        method: "POST",
        body: JSON.stringify({ followerId }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message ||
          errorData.error ||
          "Failed to decline friend request"
        );
      }

      const data = await response.json();



      if (data.success) {
        showToast("Friend request declined", "success");
        // Remove the request from the list
        setFriendRequests((prev) =>
          prev.filter((request) => request.followerId !== followerId)
        );
      }

      return true;
    } catch (error) {
      showToast(error.message || "Error declining friend request", "error");
      return false;
    }
  };

  // Subscribe to friend request events
  useEffect(() => {
    // Listen for new friend requests
    const unsubscribeFollowRequest = subscribe("follow_request", (payload) => {
      if (payload) {
        const newRequest = {
          id: Date.now(), // Temporary ID
          name: payload.followerName,
          image: payload.avatar || "/avatar.png",
          mutualFriends: 0, // We might not have this info from the event
          followerId: payload.followerID,
        };

        setFriendRequests((prev) => [newRequest, ...prev]);
      }
    });

    // Listen for accepted friend requests
    const unsubscribeFollowAccepted = subscribe(
      "follow_request_accepted",
      (payload) => {
        if (payload) {
          // Refresh contacts list when a request is accepted
          fetchContacts();
        }
      }
    );

    return () => {
      if (unsubscribeFollowRequest) unsubscribeFollowRequest();
      if (unsubscribeFollowAccepted) unsubscribeFollowAccepted();
    };
  }, [subscribe]);

  // Initial data fetch
  useEffect(() => {
    fetchFriendRequests();
    fetchContacts();
    fetchSuggestedUsers();
  }, []);

  // Add this function to the useFriendService hook
  const fetchUserFollowers = async () => {
    try {
      const response = await authenticatedFetch("follow/followers", {
        method: "GET",
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(
          errorData.message || errorData.error || "Failed to fetch followers"
        );
      }

      const data = await response.json();

      if (!data) {
        return [];
      }

      console.log("Fetched followers:", data);

      // Transform the data to match our component's expected format
      const formattedFollowers = data.map((follower) => ({
        id: follower.FollowerID,
        name: follower.UserName,
        image: follower.UserAvatar
          ? `${BASE_URL}uploads/${follower.UserAvatar}`
          : "/avatar.png",
        isOnline: follower.IsOnline || false,
        contactId: follower.FollowerID,
      }));

      return formattedFollowers;
    } catch (error) {
      console.error("Error fetching followers:", error);
      return [];
    }
  };

  return {
    friendRequests,
    SuggestedUsers,
    contacts,
    acceptFriendRequest,
    fetchSuggestedUsers,
    declineFriendRequest,
    fetchFriendRequests,
    fetchContacts,
    fetchUserFollowers,
    isLoadingRequests,
    isLoadingContacts,
  };
};
