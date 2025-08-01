/**
 * Formats a date to a relative time string (e.g., "just now", "2 hours ago", "3 days ago")
 * @param {string|Date} dateInput - The date to format (ISO string or Date object)
 * @returns {string} - Formatted relative time string
 */
export function formatRelativeTime(dateInput) {
  const date = dateInput instanceof Date ? dateInput : new Date(dateInput);
  const now = new Date();
  const diffInSeconds = Math.floor((now - date) / 1000);

  // Check for invalid date
  if (isNaN(date.getTime())) {
    return "Invalid date";
  }

  // Just now: less than a minute ago
  if (diffInSeconds < 60) {
    return "Just now";
  }

  // Minutes
  const diffInMinutes = Math.floor(diffInSeconds / 60);
  if (diffInMinutes < 60) {
    return `${diffInMinutes}m`;
  }

  // Hours
  const diffInHours = Math.floor(diffInMinutes / 60);
  if (diffInHours < 24) {
    return `${diffInHours}h`;
  }

  // Days
  const diffInDays = Math.floor(diffInHours / 24);
  if (diffInDays < 7) {
    return `${diffInDays}d`;
  }

  // For older messages, show the actual date
  return date.toLocaleDateString();
}
