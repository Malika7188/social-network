import React, { useState } from "react";
import styles from "@/styles/ProfileAbout.module.css";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faGlobe,
  faEllipsisH,
  faBriefcase,
  faGraduationCap,
  faEnvelope,
  faLink,
  faPhone,
  faHouse,
  faLocationDot,
  faLock,
  faUser,
  faShare,
  faPenToSquare,
  faTrash,
  faCode,
  faBook,
} from "@fortawesome/free-solid-svg-icons";
import ProfilePhotosGrid from "./ProfilePhotosGrid";
import ContactsList from "../contacts/ContactsList";

const ProfileAbout = ({
  photos = [],
  totalPhotos = 0,
  isLoading = false,
  error = null,
  BASE_URL = "",
  userData,
}) => {
  const [showPrivacyPopup, setShowPrivacyPopup] = useState(false);
  const [showActionsPopup, setShowActionsPopup] = useState(false);

  // Display privacy status based on user data
  const privacyStatus = userData?.isPublic ? "Public" : "Private";
  const privacyIcon = userData?.isPublic ? faGlobe : faLock;

  // Parse skills into arrays if they exist
  const techSkills = userData?.techSkills
    ? userData.techSkills.split(",").map((skill) => skill.trim())
    : [];
  const softSkills = userData?.softSkills
    ? userData.softSkills.split(",").map((skill) => skill.trim())
    : [];
  const interestsList = userData?.interests
    ? userData.interests.split(",").map((interest) => interest.trim())
    : [];

  return (
    <div className={styles.aboutContainer}>
      <div className={styles.mainContent}>
        {/* Overview Section */}
        {(userData?.work || userData?.education || userData?.location) && (
          <div className={styles.overviewSection}>
            <h2>Overview</h2>
            <div className={styles.highlightCards}>
              {userData?.work && (
                <div className={styles.highlightCard}>
                  <FontAwesomeIcon
                    icon={faBriefcase}
                    className={styles.highlightIcon}
                  />
                  <div className={styles.highlightContent}>
                    <h3>Work</h3>
                    <p>{userData.work}</p>
                  </div>
                </div>
              )}

              {userData?.education && (
                <div className={styles.highlightCard}>
                  <FontAwesomeIcon
                    icon={faGraduationCap}
                    className={styles.highlightIcon}
                  />
                  <div className={styles.highlightContent}>
                    <h3>Education</h3>
                    <p>{userData.education}</p>
                  </div>
                </div>
              )}

              {userData?.location && (
                <div className={styles.highlightCard}>
                  <FontAwesomeIcon
                    icon={faLocationDot}
                    className={styles.highlightIcon}
                  />
                  <div className={styles.highlightContent}>
                    <h3>Location</h3>
                    <p>{userData.location}</p>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Bio Section */}
        <div className={styles.bioSection}>
          <div className={styles.sectionHeader}>
            <h2>About</h2>
            <div className={styles.actionButtons}>
              <button
                className={styles.privacyButton}
                onClick={() => setShowPrivacyPopup(!showPrivacyPopup)}
              >
                <FontAwesomeIcon icon={privacyIcon} />
                <span>{privacyStatus}</span>
              </button>
              <button
                className={styles.moreButton}
                onClick={() => setShowActionsPopup(!showActionsPopup)}
              >
                <FontAwesomeIcon icon={faEllipsisH} />
              </button>

              {showPrivacyPopup && (
                <div className={styles.popup}>
                  <button className={styles.popupItem}>
                    <FontAwesomeIcon icon={faGlobe} />
                    <span>Public</span>
                  </button>
                  <button className={styles.popupItem}>
                    <FontAwesomeIcon icon={faLock} />
                    <span>Private</span>
                  </button>
                  <button className={styles.popupItem}>
                    <FontAwesomeIcon icon={faUser} />
                    <span>Only Me</span>
                  </button>
                </div>
              )}

              {showActionsPopup && (
                <div className={styles.popup}>
                  <button className={styles.popupItem}>
                    <FontAwesomeIcon icon={faPenToSquare} />
                    <span>Edit</span>
                  </button>
                  <button className={styles.popupItem}>
                    <FontAwesomeIcon icon={faTrash} />
                    <span>Delete</span>
                  </button>
                </div>
              )}
            </div>
          </div>

          {userData?.aboutMe ? (
            <p className={styles.bioText}>"{userData.aboutMe}"</p>
          ) : (
            <p className={styles.bioText}>No bio provided</p>
          )}

          {/* Contact Information */}
          {(userData?.contactEmail ||
            userData?.phone ||
            userData?.website ||
            userData?.location) && (
            <div className={styles.contactGrid}>
              {userData?.contactEmail && (
                <div className={styles.contactItem}>
                  <FontAwesomeIcon
                    icon={faEnvelope}
                    className={styles.contactIcon}
                  />
                  <div className={styles.contactInfo}>
                    <label>Email</label>
                    <span>{userData.contactEmail}</span>
                  </div>
                </div>
              )}

              {userData?.phone && (
                <div className={styles.contactItem}>
                  <FontAwesomeIcon
                    icon={faPhone}
                    className={styles.contactIcon}
                  />
                  <div className={styles.contactInfo}>
                    <label>Phone</label>
                    <span>{userData.phone}</span>
                  </div>
                </div>
              )}

              {userData?.website && (
                <div className={styles.contactItem}>
                  <FontAwesomeIcon
                    icon={faLink}
                    className={styles.contactIcon}
                  />
                  <div className={styles.contactInfo}>
                    <label>Website</label>
                    <span>{userData.website}</span>
                  </div>
                </div>
              )}

              {userData?.location && (
                <div className={styles.contactItem}>
                  <FontAwesomeIcon
                    icon={faLocationDot}
                    className={styles.contactIcon}
                  />
                  <div className={styles.contactInfo}>
                    <label>Address</label>
                    <span>{userData.location}</span>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Work Experience Section - Only displayed if work exists */}
        {userData?.work && (
          <div className={styles.experienceSection}>
            <h2>Work Experience</h2>
            <div className={styles.timelineList}>
              <div className={styles.timelineItem}>
                <div className={styles.timelineIcon}>
                  <FontAwesomeIcon icon={faBriefcase} />
                </div>
                <div className={styles.timelineContent}>
                  <h3>{userData.work}</h3>
                  {/* Additional details would come from an expanded work history API */}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Education Section - Only displayed if education exists */}
        {userData?.education && (
          <div className={styles.educationSection}>
            <h2>Education</h2>
            <div className={styles.timelineList}>
              <div className={styles.timelineItem}>
                <div className={styles.timelineIcon}>
                  <FontAwesomeIcon icon={faGraduationCap} />
                </div>
                <div className={styles.timelineContent}>
                  <h3>{userData.education}</h3>
                  {/* Additional details would come from an expanded education history API */}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Skills Section - Only displayed if skills exist */}
        {(techSkills.length > 0 || softSkills.length > 0) && (
          <div className={styles.skillsSection}>
            <h2>Skills & Expertise</h2>
            <div className={styles.skillsGrid}>
              {techSkills.length > 0 && (
                <div className={styles.skillCategory}>
                  <h3>Technical Skills</h3>
                  <div className={styles.skillTags}>
                    {techSkills.map((skill, index) => (
                      <span key={`tech-${index}`} className={styles.skillTag}>
                        {skill}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {softSkills.length > 0 && (
                <div className={styles.skillCategory}>
                  <h3>Soft Skills</h3>
                  <div className={styles.skillTags}>
                    {softSkills.map((skill, index) => (
                      <span key={`soft-${index}`} className={styles.skillTag}>
                        {skill}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Interests Section - Only displayed if interests exist */}
        {interestsList.length > 0 && (
          <div className={styles.interestsSection}>
            <h2>Interests</h2>
            <div className={styles.interestsList}>
              {interestsList.map((interest, index) => (
                <div key={`interest-${index}`} className={styles.interestItem}>
                  {/* Using faCode as default icon, could be mapped to different icons based on interest */}
                  <FontAwesomeIcon
                    icon={faCode}
                    className={styles.interestIcon}
                  />
                  <span>{interest}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* If no data exists for any section, show a message */}
        {!userData?.work &&
          !userData?.education &&
          !userData?.location &&
          !userData?.aboutMe &&
          !userData?.contactEmail &&
          !userData?.phone &&
          !userData?.website &&
          techSkills.length === 0 &&
          softSkills.length === 0 &&
          interestsList.length === 0 && (
            <div className={styles.noDataMessage}>
              <p>This profile hasn't been completed yet.</p>
            </div>
          )}
      </div>

      <div className={styles.sidebar}>
        <ProfilePhotosGrid
          photos={photos}
          totalPhotos={photos.length}
          isLoading={isLoading}
          error={error}
          BASE_URL={BASE_URL}
        />
        <ContactsList />
      </div>
    </div>
  );
};

export default ProfileAbout;
