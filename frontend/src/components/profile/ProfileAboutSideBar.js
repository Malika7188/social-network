import React from "react";
import styles from "@/styles/ProfileAboutSideBar.module.css";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faBriefcase,
  faGraduationCap,
  faEnvelope,
  faGlobe,
  faPhone,
  faHouse,
  faLocationDot,
  faCalendarAlt
} from "@fortawesome/free-solid-svg-icons";

const ProfileAboutSideBar = ({ userData }) => {
  // Check if sections should be displayed based on available data
  console.log("Usersate", userData)
  const hasBio = !!userData?.aboutMe;
  const hasInfo = !!(
    userData?.work ||
    userData?.education ||
    userData?.contactEmail ||
    userData?.website ||
    userData?.phone ||
    userData?.location
  );

  return (
    <div className={styles.aboutSection}>
      {/* About Section - Only shown if bio exists */}
      {hasBio && (
        <div className={styles.about}>
          <h3>About</h3>
          <p className={styles.bio}>{userData.aboutMe}</p>
        </div>
      )}

      {/* Info Section - Only shown if any info exists */}
      {hasInfo && (
        <div className={styles.info}>
          <h3>Info</h3>
          <ul className={styles.infoList}>
            {userData?.work && (
              <li>
                <FontAwesomeIcon icon={faBriefcase} className={styles.icon} />
                <span>{userData.work}</span>
              </li>
            )}

            {userData?.education && (
              <li>
                <FontAwesomeIcon
                  icon={faGraduationCap}
                  className={styles.icon}
                />
                <span>{userData.education}</span>
              </li>
            )}

            {userData?.contactEmail && (
              <li>
                <FontAwesomeIcon icon={faEnvelope} className={styles.icon} />
                <span>{userData.contactEmail}</span>
              </li>
            )}

            {userData?.website && (
              <li>
                <FontAwesomeIcon icon={faGlobe} className={styles.icon} />
                <span>{userData.website}</span>
              </li>
            )}

            {userData?.phone && (
              <li>
                <FontAwesomeIcon icon={faPhone} className={styles.icon} />
                <span>{userData.phone}</span>
              </li>
            )}

            {/* Split location into parts if it contains both country and address */}
            {userData?.location && userData.location.includes(",") && (
              <>
                <li>
                  <FontAwesomeIcon icon={faHouse} className={styles.icon} />
                  <span>{userData.location.split(",")[1].trim()}</span>
                </li>
                <li>
                  <FontAwesomeIcon
                    icon={faLocationDot}
                    className={styles.icon}
                  />
                  <span>{userData.location.split(",")[0].trim()}</span>
                </li>
              </>
            )}

            {/* If location doesn't have a comma, just show it as is */}
            {userData?.location && !userData.location.includes(",") && (
              <li>
                <FontAwesomeIcon icon={faLocationDot} className={styles.icon} />
                <span>{userData.location}</span>
              </li>
            )}
            {userData?.dateOfBirth &&  (
              <li>
                <FontAwesomeIcon icon={faCalendarAlt} className={styles.icon} />
                <span>{userData.dateOfBirth}</span>
              </li>
            )}
          </ul>
        </div>
      )}

      {/* Show message if no data available */}
      {!hasBio && !hasInfo && (
        <div className={styles.noData}>
          <p>No profile information available.</p>
        </div>
      )}
    </div>
  );
};

export default ProfileAboutSideBar;
