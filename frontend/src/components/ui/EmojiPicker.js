import { useState } from "react";
import POPULAR_EMOJIS, {
  EMOJI_COLLECTIONS,
  EMOJI_CATEGORIES,
} from "@/utils/emojiUtils";
import styles from "@/styles/EmojiPicker.module.css";

export default function EmojiPicker({ onEmojiSelect, onClose }) {
  const [activeCategory, setActiveCategory] = useState("popular");

  const getEmojiCollection = () => {
    if (activeCategory === "popular") return POPULAR_EMOJIS;
    return EMOJI_COLLECTIONS[activeCategory] || POPULAR_EMOJIS;
  };

  return (
    <div className={styles.emojiPicker}>
      <div className={styles.emojiPickerHeader}>
        <h4>Emojis</h4>
        <button
          type="button"
          onClick={onClose}
          className={styles.closeEmojiPicker}
        >
          <i className="fas fa-times"></i>
        </button>
      </div>

      <div className={styles.categoryTabs}>
        <button
          className={`${styles.categoryTab} ${
            activeCategory === "popular" ? styles.active : ""
          }`}
          onClick={() => setActiveCategory("popular")}
        >
          <i className="fas fa-star"></i>
        </button>
        <button
          className={`${styles.categoryTab} ${
            activeCategory === "smileys" ? styles.active : ""
          }`}
          onClick={() => setActiveCategory("smileys")}
        >
          <i className="fas fa-smile"></i>
        </button>
        <button
          className={`${styles.categoryTab} ${
            activeCategory === "hearts" ? styles.active : ""
          }`}
          onClick={() => setActiveCategory("hearts")}
        >
          <i className="fas fa-heart"></i>
        </button>
        <button
          className={`${styles.categoryTab} ${
            activeCategory === "handGestures" ? styles.active : ""
          }`}
          onClick={() => setActiveCategory("handGestures")}
        >
          <i className="fas fa-hand-peace"></i>
        </button>
      </div>

      <div className={styles.emojiGrid}>
        {getEmojiCollection().map((emoji, index) => (
          <button
            key={index}
            type="button"
            onClick={() => onEmojiSelect(emoji)}
            className={styles.emojiButton}
          >
            {emoji}
          </button>
        ))}
      </div>
    </div>
  );
}
