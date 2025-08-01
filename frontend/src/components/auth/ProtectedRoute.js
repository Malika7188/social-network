"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/authcontext";
import LoadingSpinner from "@/components/ui/LoadingSpinner";

export default function ProtectedRoute({ children }) {
  const { isAuthenticated, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    // If authentication check is complete and user is not authenticated
    if (!loading && !isAuthenticated) {
      router.push("/");
    }
  }, [isAuthenticated, loading, router]);

  // Show loading spinner while loading
  if (loading) {
    return <LoadingSpinner size="large" fullPage={true} />;
  }

  // If authenticated, show the children components
  return isAuthenticated ? children : null;
}
