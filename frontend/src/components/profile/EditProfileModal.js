import React, { useState, useEffect } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { useRouter } from "next/navigation";
import {
  faCamera,
  faTimes,
  faUser,
  faBriefcase,
  faGraduationCap,
  faEnvelope,
  faPhone,
  faLink,
  faLocationDot,
} from "@fortawesome/free-solid-svg-icons";
import styles from "@/styles/EditProfileModal.module.css";
import { showToast } from "../ui/ToastContainer";

import { useAuth } from "@/context/authcontext";

const EditProfileModal = ({
  isOpen,
  onClose,
  profileData,
  onProfileUpdate,
}) => {
  const router = useRouter();
  // Get auth context with authenticatedFetch
  const { authenticatedFetch, isAuthenticated } = useAuth();
  // Predefined lists of skills and interests
  const predefinedTechSkills = [
    "JavaScript",
    "React",
    "Node.js",
    "Python",
    "Java",
    "SQL",
    "AWS",
  ];
  const predefinedSoftSkills = [
    "Communication",
    "Teamwork",
    "Leadership",
    "Problem Solving",
    "Time Management",
  ];
  const predefinedInterests = [
    "AI",
    "Web Development",
    "Mobile Development",
    "Data Science",
    "Cybersecurity",
  ];

  // Helper function to safely split comma-separated strings
  const safeSplit = (str) => {
    if (!str) return [];
    return str.split(",").filter((item) => item.trim() !== "");
  };

  const [formData, setFormData] = useState({
    bannerImage: null,
    profileImage: null,
    username: "",
    fullName: "",
    bio: "",
    work: "",
    education: "",
    email: "",
    phone: "",
    website: "",
    location: "",
    techSkills: [],
    softSkills: [],
    interests: [],
    isPrivate: false,
  });

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState(null);
  const [bannerPreview, setBannerPreview] = useState("");
  const [profilePreview, setProfilePreview] = useState("");

  // Initialize form data when profileData changes
  useEffect(() => {
    if (profileData) {
      setFormData({
        bannerImage: null, // File objects will be set on change
        profileImage: null,
        username: profileData.username || "",
        fullName: profileData.fullName || "",
        bio: profileData.bio || "",
        work: profileData.work || "",
        education: profileData.education || "",
        email: profileData.email || "",
        phone: profileData.phone || "",
        website: profileData.website || "",
        location: profileData.location || "",
        techSkills: safeSplit(profileData.techSkills),
        softSkills: safeSplit(profileData.softSkills),
        interests: safeSplit(profileData.interests),
        isPrivate: profileData.isPrivate || false,
      });

      setBannerPreview(profileData.bannerUrl || "");
      setProfilePreview(profileData.profileUrl || "");
    }
  }, [profileData]);

  const handleImageChange = (e, type) => {
    const file = e.target.files[0];
    if (!file) return;

    // Validate file type
    const validImageTypes = [
      "image/jpeg",
      "image/png",
      "image/gif",
      "image/webp",
    ];
    if (!validImageTypes.includes(file.type)) {
      setError(`Please select a valid image file (JPEG, PNG, GIF, WEBP)`);
      return;
    }

    // Validate file size (max 5MB)
    const maxSize = 5 * 1024 * 1024; // 5MB
    if (file.size > maxSize) {
      setError(`Image size should not exceed 5MB`);
      return;
    }

    const reader = new FileReader();
    reader.onloadend = () => {
      if (type === "banner") {
        setBannerPreview(reader.result);
        setFormData((prev) => ({ ...prev, bannerImage: file }));
      } else {
        setProfilePreview(reader.result);
        setFormData((prev) => ({ ...prev, profileImage: file }));
      }
    };
    reader.readAsDataURL(file);
  };

  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: type === "checkbox" ? checked : value,
    }));
  };

  const handleSkillSelect = (type, value) => {
    if (!value) return;

    if (!formData[type].includes(value)) {
      setFormData((prev) => ({
        ...prev,
        [type]: [...prev[type], value],
      }));
    }
  };

  const handleRemoveSkill = (type, value) => {
    setFormData((prev) => ({
      ...prev,
      [type]: prev[type].filter((item) => item !== value),
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);

    try {
      const formDataToSend = new FormData();

      // Add text fields
      formDataToSend.append("username", formData.username);
      formDataToSend.append("fullName", formData.fullName);
      formDataToSend.append("bio", formData.bio);
      formDataToSend.append("work", formData.work);
      formDataToSend.append("education", formData.education);
      formDataToSend.append("email", formData.email);
      formDataToSend.append("phone", formData.phone);
      formDataToSend.append("website", formData.website);
      formDataToSend.append("location", formData.location);
      formDataToSend.append("isPrivate", formData.isPrivate);

      // Add skills and interests as comma-separated strings
      formDataToSend.append("techSkills", formData.techSkills.join(","));
      formDataToSend.append("softSkills", formData.softSkills.join(","));
      formDataToSend.append("interests", formData.interests.join(","));

      // Add image files only if they exist
      if (formData.bannerImage instanceof File) {
        formDataToSend.append("bannerImage", formData.bannerImage);
      }

      if (formData.profileImage instanceof File) {
        formDataToSend.append("profileImage", formData.profileImage);
      }
      // Use authenticatedFetch instead of direct fetch
      const response = await authenticatedFetch("users/profile", {
        method: "PUT",
        body: formDataToSend,
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || "Failed to update profile");
      }

      const updatedProfile = await response.json();
      if (updatedProfile?.profile) {
        localStorage.setItem(
          "userData",
          JSON.stringify(updatedProfile.profile)
        );
      } else {
        console.error("Profile data missing in response:", updatedProfile);
      }

      // Call the callback with updated profile data
      if (onProfileUpdate) {
        onProfileUpdate(updatedProfile);
      }
      router.refresh()
      showToast("Profile updated successfully", "success")

      onClose();
    } catch (err) {
      console.error("Error updating profile:", err);
      setError(err.message || "An error occurred while updating your profile");
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className={styles.modalOverlay}>
      <div className={styles.modalContent}>
        <div className={styles.modalHeader}>
          <h2>Edit Profile</h2>
          <button
            type="button"
            onClick={onClose}
            className={styles.closeButton}
            aria-label="Close modal"
          >
            <FontAwesomeIcon icon={faTimes} />
          </button>
        </div>

        {error && <div className={styles.errorMessage}>{error}</div>}

        <form onSubmit={handleSubmit} className={styles.form}>
          {/* Banner Image */}
          <div className={styles.imageSection}>
            <div className={styles.bannerUpload}>
              <img
                src={bannerPreview || "/banner-placeholder.jpg"}
                alt="Banner"
                className={styles.bannerPreview}
              />
              <label className={styles.imageUploadLabel}>
                <FontAwesomeIcon icon={faCamera} />
                <span>Change Banner</span>
                <input
                  type="file"
                  accept="image/jpeg,image/png,image/gif,image/webp"
                  onChange={(e) => handleImageChange(e, "banner")}
                  hidden
                />
              </label>
            </div>

            {/* Profile Image */}
            <div className={styles.profileImageUpload}>
              <img
                src={profilePreview || "/avatar-placeholder.png"}
                alt="Profile"
                className={styles.profilePreview}
              />
              <label className={styles.profileImageLabel}>
                <FontAwesomeIcon icon={faCamera} />
                <input
                  type="file"
                  accept="image/jpeg,image/png,image/gif,image/webp"
                  onChange={(e) => handleImageChange(e, "profile")}
                  hidden
                />
              </label>
            </div>
          </div>

          <div className={styles.formFields}>
            {/* Basic Info */}
            <div className={styles.fieldGroup}>
              <h3>Basic Information</h3>
              <div className={styles.field}>
                <label htmlFor="username">
                  <FontAwesomeIcon icon={faUser} />
                  Username
                </label>
                <input
                  type="text"
                  id="username"
                  name="username"
                  value={formData.username}
                  onChange={handleInputChange}
                  placeholder="Username"
                  maxLength={50}
                />
              </div>
              <div className={styles.field}>
                <label htmlFor="fullName">
                  <FontAwesomeIcon icon={faUser} />
                  Full Name
                </label>
                <input
                  type="text"
                  id="fullName"
                  name="fullName"
                  value={formData.fullName}
                  onChange={handleInputChange}
                  placeholder="Full Name"
                  maxLength={100}
                />
              </div>
            </div>

            <div className={styles.fieldGroup}>
              <h3>Profile Privacy</h3>
              <div className={styles.field}>
                <div className={styles.privacyToggle}>
                  <label
                    className={styles.toggleLabel}
                    htmlFor="profilePrivacy"
                  >
                    <input
                      type="checkbox"
                      id="profilePrivacy"
                      name="isPrivate"
                      checked={formData.isPrivate}
                      onChange={handleInputChange}
                    />
                    <span className={styles.toggleText}>
                      {formData.isPrivate
                        ? "Private Profile"
                        : "Public Profile"}
                    </span>
                  </label>
                </div>
              </div>
            </div>

            {/* Bio */}
            <div className={styles.fieldGroup}>
              <h3>About</h3>
              <div className={styles.field}>
                <label htmlFor="bio">Bio</label>
                <textarea
                  id="bio"
                  name="bio"
                  value={formData.bio}
                  onChange={handleInputChange}
                  placeholder="Write something about yourself..."
                  rows="4"
                  maxLength={500}
                />
                <small className={styles.charCount}>
                  {formData.bio.length}/500
                </small>
              </div>
            </div>

            {/* Work & Education */}
            <div className={styles.fieldGroup}>
              <h3>Work & Education</h3>
              <div className={styles.field}>
                <label htmlFor="work">
                  <FontAwesomeIcon icon={faBriefcase} />
                  Work
                </label>
                <input
                  type="text"
                  id="work"
                  name="work"
                  value={formData.work}
                  onChange={handleInputChange}
                  placeholder="Current work"
                  maxLength={100}
                />
              </div>
              <div className={styles.field}>
                <label htmlFor="education">
                  <FontAwesomeIcon icon={faGraduationCap} />
                  Education
                </label>
                <input
                  type="text"
                  id="education"
                  name="education"
                  value={formData.education}
                  onChange={handleInputChange}
                  placeholder="Education"
                  maxLength={100}
                />
              </div>
            </div>

            {/* Contact Information */}
            <div className={styles.fieldGroup}>
              <h3>Contact Information</h3>
              <div className={styles.field}>
                <label htmlFor="email">
                  <FontAwesomeIcon icon={faEnvelope} />
                  Email
                </label>
                <input
                  type="email"
                  id="email"
                  name="email"
                  value={formData.email}
                  onChange={handleInputChange}
                  placeholder="Email"
                  maxLength={100}
                />
              </div>
              <div className={styles.field}>
                <label htmlFor="phone">
                  <FontAwesomeIcon icon={faPhone} />
                  Phone
                </label>
                <input
                  type="tel"
                  id="phone"
                  name="phone"
                  value={formData.phone}
                  onChange={handleInputChange}
                  placeholder="Phone number"
                  maxLength={20}
                />
              </div>
              <div className={styles.field}>
                <label htmlFor="website">
                  <FontAwesomeIcon icon={faLink} />
                  Website
                </label>
                <input
                  type="url"
                  id="website"
                  name="website"
                  value={formData.website}
                  onChange={handleInputChange}
                  placeholder="Website"
                  pattern="https?://.+"
                  title="Include http:// or https:// in your URL"
                  maxLength={200}
                />
              </div>
              <div className={styles.field}>
                <label htmlFor="location">
                  <FontAwesomeIcon icon={faLocationDot} />
                  Location
                </label>
                <input
                  type="text"
                  id="location"
                  name="location"
                  value={formData.location}
                  onChange={handleInputChange}
                  placeholder="Location"
                  maxLength={100}
                />
              </div>
            </div>

            {/* Skills & Expertise */}
            <div className={styles.fieldGroup}>
              <h3>Skills & Expertise</h3>
              {/* Technical Skills */}
              <div className={styles.field}>
                <label htmlFor="techSkillsSelect">
                  <FontAwesomeIcon icon={faBriefcase} />
                  Technical Skills
                </label>
                <select
                  id="techSkillsSelect"
                  onChange={(e) =>
                    handleSkillSelect("techSkills", e.target.value)
                  }
                  value=""
                >
                  <option value="" disabled>
                    Select a technical skill
                  </option>
                  {predefinedTechSkills
                    .filter((skill) => !formData.techSkills.includes(skill))
                    .map((skill, index) => (
                      <option key={index} value={skill}>
                        {skill}
                      </option>
                    ))}
                </select>
                <div className={styles.tags}>
                  {formData.techSkills.length === 0 && (
                    <span className={styles.noTagsMessage}>
                      No technical skills selected
                    </span>
                  )}
                  {formData.techSkills.map((skill, index) => (
                    <span key={index} className={styles.tag}>
                      {skill}
                      <button
                        type="button"
                        onClick={() => handleRemoveSkill("techSkills", skill)}
                        aria-label={`Remove ${skill}`}
                      >
                        ×
                      </button>
                    </span>
                  ))}
                </div>
              </div>

              {/* Soft Skills */}
              <div className={styles.field}>
                <label htmlFor="softSkillsSelect">
                  <FontAwesomeIcon icon={faBriefcase} />
                  Soft Skills
                </label>
                <select
                  id="softSkillsSelect"
                  onChange={(e) =>
                    handleSkillSelect("softSkills", e.target.value)
                  }
                  value=""
                >
                  <option value="" disabled>
                    Select a soft skill
                  </option>
                  {predefinedSoftSkills
                    .filter((skill) => !formData.softSkills.includes(skill))
                    .map((skill, index) => (
                      <option key={index} value={skill}>
                        {skill}
                      </option>
                    ))}
                </select>
                <div className={styles.tags}>
                  {formData.softSkills.length === 0 && (
                    <span className={styles.noTagsMessage}>
                      No soft skills selected
                    </span>
                  )}
                  {formData.softSkills.map((skill, index) => (
                    <span key={index} className={styles.tag}>
                      {skill}
                      <button
                        type="button"
                        onClick={() => handleRemoveSkill("softSkills", skill)}
                        aria-label={`Remove ${skill}`}
                      >
                        ×
                      </button>
                    </span>
                  ))}
                </div>
              </div>

              {/* Interests */}
              <div className={styles.field}>
                <label htmlFor="interestsSelect">
                  <FontAwesomeIcon icon={faBriefcase} />
                  Interests
                </label>
                <select
                  id="interestsSelect"
                  onChange={(e) =>
                    handleSkillSelect("interests", e.target.value)
                  }
                  value=""
                >
                  <option value="" disabled>
                    Select an interest
                  </option>
                  {predefinedInterests
                    .filter(
                      (interest) => !formData.interests.includes(interest)
                    )
                    .map((interest, index) => (
                      <option key={index} value={interest}>
                        {interest}
                      </option>
                    ))}
                </select>
                <div className={styles.tags}>
                  {formData.interests.length === 0 && (
                    <span className={styles.noTagsMessage}>
                      No interests selected
                    </span>
                  )}
                  {formData.interests.map((interest, index) => (
                    <span key={index} className={styles.tag}>
                      {interest}
                      <button
                        type="button"
                        onClick={() => handleRemoveSkill("interests", interest)}
                        aria-label={`Remove ${interest}`}
                      >
                        ×
                      </button>
                    </span>
                  ))}
                </div>
              </div>
            </div>
          </div>
          <div className={styles.formActions}>
            <button
              type="button"
              onClick={onClose}
              className={styles.cancelButton}
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className={styles.saveButton}
              disabled={isSubmitting}
            >
              {isSubmitting ? "Saving..." : "Save Changes"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default EditProfileModal;
