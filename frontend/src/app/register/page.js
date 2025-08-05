"use client";

import styles from "@/styles/auth.module.css";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, useEffect } from "react";
import { useAuth } from "@/context/authcontext";
import PasswordInput from "@/components/inputs/PasswordInput";
import LoadingSpinner from "@/components/ui/LoadingSpinner";

export default function Register() {
  const router = useRouter();
  const [formData, setFormData] = useState({
    firstName: "",
    lastName: "",
    email: "",
    password: "",
    dateOfBirth: "",
    nickname: "",
    aboutMe: "",
    avatar: null,
  });
  const [formErrors, setFormErrors] = useState({});
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [avatarPreview, setAvatarPreview] = useState(null);
  const [showPassword, setShowPassword] = useState(false);
  const { signUp, isAuthenticated, loading } = useAuth();

  // Redirect if user is already authenticated
  useEffect(() => {
    if (!loading && isAuthenticated) {
      router.push("/home");
    }
  }, [isAuthenticated, loading, router]);

  const handleChange = (e) => {
    const { name, value, type, files } = e.target;
    if (type === "file") {
      const file = files[0];
      if (file) {
        setFormData({
          ...formData,
          avatar: file,
        });
        setAvatarPreview(URL.createObjectURL(file));
      }
    } else {
      setFormData({
        ...formData,
        [name]: value,
      });
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setIsLoading(true);
    // Trim and validate required fields
    const trimmedData = {
      firstName: formData.firstName.trim(),
      lastName: formData.lastName.trim(),
      email: formData.email.trim(),
      password: formData.password.trim(),
      dateOfBirth: formData.dateOfBirth.trim(),
      nickname: formData.nickname.trim(),
      aboutMe: formData.aboutMe.trim(),
      avatar: formData.avatar,
    };

    const newErrors = {};
    if (!trimmedData.firstName) newErrors.firstName = "First name is required";
    if (!trimmedData.lastName) newErrors.lastName = "Last name is required";
    if (!trimmedData.email) newErrors.email = "Email is required";
    if (!trimmedData.password) newErrors.password = "Password is required";
    if (!trimmedData.dateOfBirth) {
      newErrors.dateOfBirth = "Date of Birth is required";
    } else {
      const selectedDate = new Date(trimmedData.dateOfBirth);
      const today = new Date();
      today.setHours(0, 0, 0, 0); // Normalize time for accuracy

      if (selectedDate > today) {
        newErrors.dateOfBirth = "Date of Birth cannot be in the future";
      }
    }

    setFormErrors(newErrors);

    if (Object.keys(newErrors).length > 0) {
      setIsLoading(false);
      return; // Don't proceed if there are errors
    }

    try {
      const success = await signUp(formData);
      if (success) {
        router.push("/home");
      } else {
        setError("Registration failed. Please try again.");
      }
    } catch (err) {
      setError(err.message || "An error occurred during registration");
    } finally {
      setIsLoading(false);
    }
  };

  const togglePasswordVisibility = () => {
    setShowPassword(!showPassword);
  };

  if (loading) {
    return <LoadingSpinner size="large" fullPage={true} />;
  }

  return (
    <div className={styles.authContainer}>
      <h1 className="brandName">Vibes</h1>
      <div className={styles.authCard}>
        <h1 id="auth-title">Create an Account</h1>
        {error && <p className={styles.error}>{error}</p>}
        <form className={styles.authForm} onSubmit={handleSubmit}>
          <input
            type="text"
            name="firstName"
            placeholder="First Name"
            value={formData.firstName}
            onChange={handleChange}
            required
          />
          {formErrors.firstName && (
            <span className={styles.fieldError}>{formErrors.firstName}</span>
          )}
          <input
            type="text"
            name="lastName"
            placeholder="Last Name"
            value={formData.lastName}
            onChange={handleChange}
            required
          />
          {formErrors.lastName && (
                <span className={styles.fieldError}>{formErrors.lastName}</span>
              )}
          <input
            type="email"
            name="email"
            placeholder="Email"
            value={formData.email}
            onChange={handleChange}
            required
          />
          {formErrors.email && (
              <span className={styles.fieldError}>{formErrors.email}</span>
            )}
          <PasswordInput
            name="password"
            placeholder="Password"
            value={formData.password}
            onChange={handleChange}
          />
          {formErrors.password && (
              <span className={styles.fieldError}>{formErrors.password}</span>
            )}
          <PasswordInput
            name="confirmPassword"
            placeholder="Confirm Password"
            value={formData.confirmPassword}
            onChange={handleChange}
          />
          <div className={styles.formGroup}>
            <input
              type="date"
              name="dateOfBirth"
              placeholder="Date of Birth"
              value={formData.dateOfBirth}
              onChange={handleChange}
              max={new Date().toISOString().split("T")[0]}
              required
            />


            {formErrors.dateOfBirth && (
              <span className={styles.fieldError}>
                {formErrors.dateOfBirth}
              </span>
            )}
          </div>

          <div className={styles.formGroup}>
            <input
              type="text"
              name="nickname"
              placeholder="Nickname (optional)"
              value={formData.nickname}
              onChange={handleChange}
            />
          </div>

          <div className={styles.formGroup}>
            <textarea
              name="aboutMe"
              placeholder="About Me (optional)"
              value={formData.aboutMe}
              onChange={handleChange}
            />
          </div>

          <div className={styles.formGroup}>
            <label htmlFor="avatar" className={styles.avatarLabel}>
              Profile Picture (optional)
            </label>
            <input
              type="file"
              id="avatar"
              name="avatar"
              accept="image/*"
              onChange={handleChange}
              className={styles.fileInput}
            />
            {avatarPreview && (
              <div className={styles.avatarPreview}>
                <img
                  src={avatarPreview}
                  alt="Avatar preview"
                  className={styles.previewImage}
                />
              </div>
            )}
          </div>
          <button type="submit" className="btn-tertiary">
            Register
          </button>
        </form>
        <p className={styles.authLink}>
          Already have an account? <Link href="/">Login</Link>
        </p>
      </div>
    </div>
  );
}
