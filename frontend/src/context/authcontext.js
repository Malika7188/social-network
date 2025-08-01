"use client";
import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
} from "react";
import { useRouter } from "next/navigation";
import { showToast } from "@/components/ui/ToastContainer";
import { handleApiError } from "@/utils/errorHandler";
import { closeWebSocketConnection } from "@/services/websocketService";
import { globalSocket } from "@/services/websocketService";
import { API_URL } from "@/utils/constants";

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [token, setToken] = useState(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  // Memoize handleLogout to prevent unnecessary re-renders
  const handleLogout = useCallback(
    async (sendRequest = true) => {
      // First, send a user_away message to the WebSocket if connected
      if (typeof window !== "undefined" && window.__websocketConnected) {
        try {
          const socket = globalSocket; // Access the global socket
          if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ type: "user_away" }));
            // Small delay to ensure the message is sent before closing
            await new Promise((resolve) => setTimeout(resolve, 100));
          }
        } catch (error) {
          console.error("Error sending offline status before logout:", error);
        }
      }

      // Close any active WebSocket connections
      closeWebSocketConnection();

      if (sendRequest && token) {
        try {
          await fetch(`${API_URL}/auth/logout`, {
            method: "POST",
            credentials: "include",
            headers: {
              Authorization: `Bearer ${token}`,
              "Content-Type": "application/json",
            },
          });
        } catch (error) {
          await handleApiError(error, "Logout failed");
        }
      }

      localStorage.removeItem("userData");
      localStorage.removeItem("token");

      // Clear the token cookie
      document.cookie = "token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT";

      setCurrentUser(null);
      setToken(null);
      setLoading(false);

      if (sendRequest) {
        router.push("/");
      }
    },
    [token, router]
  ); // Dependencies: token and router

  const validateToken = useCallback(
    async (currentToken) => {
      try {
        const response = await fetch(`${API_URL}/auth/validate_token`, {
          method: "GET",
          credentials: "include",
          headers: { Authorization: `Bearer ${currentToken}` },
        });

        if (!response.ok) {
          handleLogout(false);
          return;
        }

        setLoading(false);
      } catch (error) {
        console.error("Token validation error", error);
        handleLogout(false);
      }
    },
    [handleLogout]
  ); // Depend on memoized handleLogout
  useEffect(() => {
    const storedUser = localStorage.getItem("userData");
    const storedToken = localStorage.getItem("token");

    if (storedUser && storedToken) {
      setCurrentUser(JSON.parse(storedUser));
      setToken(storedToken);
      validateToken(storedToken);

      // router.push("/home");
    } else {
      setLoading(false);
      handleLogout(false);
    }
  }, [validateToken]);

  const login = async (formData) => {
    try {
      setLoading(true);

      const response = await fetch(`${API_URL}/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          email: formData.email,
          password: formData.password,
        }),
      });

      const data = await response.json();

      if (data && data.user && data.token) {
        localStorage.setItem("userData", JSON.stringify(data.user));
        localStorage.setItem("token", data.token);
        setCurrentUser(data.user);
        setToken(data.token);

        document.cookie = `token=${data.token}; path=/; max-age=86400; samesite=strict`;
        showToast("Logged in successfully!", "success");
        return true;
      } else {
        const errorMessage = data.message || data.error || "Login failed";
        await handleApiError({ message: errorMessage }, errorMessage);
        setLoading(false);
        return false;
      }
    } catch (error) {
      setLoading(false);
      await handleApiError(error, "Login failed");
      return false;
    }
  };

  const getAuthHeader = () => {
    return token ? { Authorization: `Bearer ${token}` } : {};
  };

  const signUp = async (signUpData) => {
    try {
      setLoading(true);

      const formData = new FormData();
      formData.append("email", signUpData.email);
      formData.append("password", signUpData.password);
      formData.append("firstName", signUpData.firstName);
      formData.append("lastName", signUpData.lastName);
      formData.append("dateOfBirth", signUpData.dateOfBirth);
      formData.append("nickname", signUpData.nickname || "");
      formData.append("aboutMe", signUpData.aboutMe || "");

      if (signUpData.avatar) {
        formData.append("avatar", signUpData.avatar);
      }

      const response = await fetch(`${API_URL}/auth/register`, {
        credentials: "include",
        method: "POST",
        body: formData,
      });

      const data = await response.json();

      if (data && data.user && data.token) {
        localStorage.setItem("userData", JSON.stringify(data.user));
        localStorage.setItem("token", data.token);

        document.cookie = `token=${data.token}; path=/; max-age=${data.expires_in}; samesite=strict`;

        setCurrentUser(data.user);
        setToken(data.token);
        showToast("Signed up successfully!", "success");
        return true;
      }
      const errorMessage = data.message || data.error || "Signup failed";
      await handleApiError({ message: errorMessage }, errorMessage);
      setLoading(false);
      return false;
    } catch (error) {
      setLoading(false);
      await handleApiError(error, "Signup failed");
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const fetchUserProfile = async (token) => {
    try {
      const res = await fetch(`${API_URL}/users/me`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) {
        const data = await res.json();

        localStorage.setItem("userData", JSON.stringify(data.user));
        localStorage.setItem("token", data.token);

        setCurrentUser(data.user);
        setToken(data.token);
      } else {
        handleLogout(true);
      }
    } catch (error) {
      await handleApiError(error, "Failed to fetch user");
      handleLogout(true);
    } finally {
      setLoading(false);
    }
  };

  const authenticatedFetch = async (url, options = {}) => {
    try {
      const storedToken = localStorage.getItem("token");
      if (!storedToken) {
        handleLogout(true);
        throw new Error("No token found");
      }

      const authHeader = { Authorization: `Bearer ${storedToken}` };

      const response = await fetch(`${API_URL}/${url}`, {
        ...options,
        headers: {
          ...authHeader,
          ...options.headers,
        },
        credentials: "include",
      });

      if (response.status === 401) {
        handleLogout(true);
        const errorData = await response.json();
        await handleApiError(errorData, "Unauthorized - Please log in again.");
        throw new Error("Unauthorized - Please log in again.");
      }

      return response;
    } catch (error) {
      // await handleApiError(error, "Authenticated fetch error");
      throw error;
    }
  };

  const value = {
    login,
    logout: () => handleLogout(true),
    signUp,
    getAuthHeader,
    authenticatedFetch,
    currentUser,
    token,
    loading,
    isAuthenticated: !!token,
    fetchUserProfile,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  return useContext(AuthContext);
};
