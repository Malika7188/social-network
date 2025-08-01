"use client";

import styles from "@/styles/auth.module.css";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useState, useEffect } from "react";
import { useAuth } from "@/context/authcontext";
import PasswordInput from "@/components/inputs/PasswordInput";
import LoadingSpinner from "@/components/ui/LoadingSpinner";

export default function Login() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [formData, setFormData] = useState({
    email: "",
    password: "",
  });
  const [error, setError] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const { login, isAuthenticated, loading } = useAuth();

  // Get the 'from' parameter to redirect after login
  const from = searchParams.get("from") || "/home";

  // Redirect if user is already authenticated
  useEffect(() => {
    if (!loading && isAuthenticated) {
      router.push(from);
    }
  }, [isAuthenticated, loading, router]);

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");

    try {
      const success = await login(formData);
      if (success) {
        router.push("/home");
      }
    } catch (err) {
      console.error("Login Failed: ", err);
    }
  };

  const togglePasswordVisibility = () => {
    setShowPassword(!showPassword);
  };

  // Show loading state while checking authentication
  if (loading) {
    return <LoadingSpinner size="large" fullPage={true} />;
  }

  return (
    <div className={styles.authContainer}>
      <h1 className="forumName">Notebook</h1>
      <div className={styles.authCard}>
        <h1>Login to Notebook</h1>
        {error && <p className={styles.error}>{error}</p>}
        <form className={styles.authForm} onSubmit={handleSubmit}>
          <input
            type="email"
            name="email"
            placeholder="Email"
            value={formData.email}
            onChange={handleChange}
            required
          />
          <PasswordInput value={formData.password} onChange={handleChange} />
          <button type="submit" className="btn-tertiary">
            Login
          </button>
        </form>
        <p className={styles.authLink}>
          New to Notebook? <Link href="/register">Create Account</Link>
        </p>
      </div>
    </div>
  );
}
