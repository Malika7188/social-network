"use client";

import React from "react";
import styles from "@/styles/LoadingSpinner.module.css";

export default function LoadingSpinner({
  size = "medium",
  color = "primary",
  fullPage = false,
}) {
  // Determine size class based on prop
  const sizeClass = styles[size] || styles.medium;

  // Determine color class based on prop
  const colorClass = styles[color] || styles.primary;

  // Combine classes
  const spinnerClass = `${styles.spinner} ${sizeClass} ${colorClass}`;

  // If fullPage is true, render a full-page loading overlay
  if (fullPage) {
    return (
      <div className={styles.fullPageOverlay}>
        <div className={spinnerClass}>
          <div className={styles.bounce1}></div>
          <div className={styles.bounce2}></div>
          <div className={styles.bounce3}></div>
        </div>
      </div>
    );
  }

  // Otherwise, render just the spinner
  return (
    <div className={spinnerClass}>
      <div className={styles.bounce1}></div>
      <div className={styles.bounce2}></div>
      <div className={styles.bounce3}></div>
    </div>
  );
}
