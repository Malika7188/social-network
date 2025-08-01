'use client';

import { useState } from 'react';
import styles from '@/styles/Header.module.css';

export default function SearchBar() {
  const [searchQuery, setSearchQuery] = useState('');

  const handleSearch = (e) => {
    e.preventDefault();
    // Add search logic here
  };

  return (
    <form className={styles.searchContainer} onSubmit={handleSearch}>
      <div className={styles.searchWrapper}>
        <i className="fas fa-search"></i>
        <input
          type="text"
          placeholder="Search Notebook"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className={styles.searchInput}
        />
      </div>
    </form>
  );
} 