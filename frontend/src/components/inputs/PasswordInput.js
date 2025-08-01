'use client';

import { useState } from 'react';
import styles from '@/styles/auth.module.css';

export default function PasswordInput({ value, onChange, placeholder = "Password", name = "password" }) {
  const [showPassword, setShowPassword] = useState(false);

  return (
    <div className={styles.passwordContainer}>
      <input 
        type={showPassword ? "text" : "password"}
        name={name}
        placeholder={placeholder} 
        value={value}
        onChange={onChange}
        required 
      />
      <button 
        type="button"
        className={styles.passwordToggle}
        onClick={() => setShowPassword(!showPassword)}
        aria-label={showPassword ? "Hide password" : "Show password"}
      >
        {showPassword ? (
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
            <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
            <line x1="1" y1="1" x2="23" y2="23"></line>
          </svg>
        ) : (
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
            <circle cx="12" cy="12" r="3"></circle>
          </svg>
        )}
      </button>
    </div>
  );
} 