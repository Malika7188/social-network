import React, { useState } from "react";
import Image from "next/image";
import styles from "@/styles/ProfileBanner.module.css";
import EditProfileModal from "./EditProfileModal";

const ProfileBanner = ({
  userData,
  onNavClick,
  activeSection = "posts",
  isOwnProfile = true,
  isFollowing = false,
  onFollow,
  onUnfollow,
  BASE_URL = "",
}) => {
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

  const handleNavClick = (e, section) => {
    e.preventDefault();
    onNavClick(section);
  };

  const fullName = userData.fullName || "";
  const nickname = userData.nickname || "";

  // Use bannerImage or default for banner
  const bannerUrl = userData.bannerImage
    ? `${BASE_URL}/uploads/${userData.bannerImage}`
    : "/banner.png";

  // Use the same approach as in your original code for profile image
  const profileUrl = userData.avatar
    ? `${BASE_URL}/uploads/${userData.avatar}`
    : "/default-avatar.png";
  // Use isPrivate from userData or default to false
  const isPrivate = userData.isPrivate || false;

  return (
    <>
      <div className={styles.bannerContainer}>
        <div className={styles.bannerImage}>
          <Image
            src={bannerUrl}
            alt="Profile Banner"
            fill
            style={{ objectFit: "cover" }}
          />
        </div>

        <div className={styles.profileInfo}>
          <div className={styles.leftSection}>
            <div className={styles.profilePicture}>
              <Image
                src={profileUrl}
                alt="Profile Picture"
                width={120}
                height={120}
              />
            </div>

            <div className={styles.userInfo}>
              <h2>{fullName} {nickname && `(${nickname})`}</h2>
              <div className={styles.privacyBadge}>
                {isPrivate ? (
                  <span className={`${styles.badge} ${styles.private}`}>
                    🔒 Private Profile
                  </span>
                ) : (
                  <span className={`${styles.badge} ${styles.public}`}>
                    🌐 Public Profile
                  </span>
                )}
              </div>
              <div className={styles.followInfo}>
                <div className={styles.followers}>
                  <div className={styles.avatarGroup}>
                    {/* Placeholder for follower avatars */}
                    <div className={styles.avatarStack}>
                      {[1, 2, 3].map((_, index) => (
                        <div key={index} className={styles.avatarWrapper}>
                          <Image
                            src="/avatar.png"
                            alt="Follower"
                            width={24}
                            height={24}
                          />
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
                <div className={styles.following}>
                  <span>| {userData.followersCount || 0} Followers</span>
                  <span> | {userData.followingCount || 0} Following</span>
                </div>
              </div>
            </div>
          </div>

          {isOwnProfile && (
            <div className={styles.rightSection}>
              <button
                className={styles.editButton}
                onClick={() => setIsEditModalOpen(true)}
              >
                Edit Profile
              </button>
            </div>
          )}
        </div>

        <div className={styles.profileNav}>
          <nav>
            <a
              href="#"
              className={activeSection === "posts" ? styles.active : ""}
              onClick={(e) => handleNavClick(e, "posts")}
            >
              Posts
            </a>
            <a
              href="#"
              className={activeSection === "about" ? styles.active : ""}
              onClick={(e) => handleNavClick(e, "about")}
            >
              About
            </a>
            <a
              href="#"
              className={activeSection === "photos" ? styles.active : ""}
              onClick={(e) => handleNavClick(e, "photos")}
            >
              Photos
            </a>
            <a
              href="#"
              className={activeSection === "groups" ? styles.active : ""}
              onClick={(e) => handleNavClick(e, "groups")}
            >
              Groups
            </a>
            <a
              href="#"
              className={activeSection === "connections" ? styles.active : ""}
              onClick={(e) => handleNavClick(e, "connections")}
            >
              Connections
            </a>
          </nav>
        </div>
      </div>

      {isOwnProfile && (
        <EditProfileModal
          isOpen={isEditModalOpen}
          onClose={() => setIsEditModalOpen(false)}
          profileData={{
            bannerUrl,
            profileUrl,
            fullName,
            isPrivate,
          }}
        />
      )}
    </>
  );
};

export default ProfileBanner;
