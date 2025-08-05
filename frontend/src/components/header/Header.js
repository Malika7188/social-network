// src/components/header/Header.js
'use client';

import styles from '@/styles/Header.module.css';
import Link from 'next/link';
import { useRouter, usePathname } from 'next/navigation';
import ProfileDropdown from './ProfileDropdown';
import SearchBar from './SearchBar';
import { useAuth } from '@/context/authcontext';
import NotificationDropdown from './NotificationDropdown';
import { User, Store, Home, Paperclip, Bell, MessageCircle } from 'lucide-react';

export default function Header() {
  const router = useRouter();
  const pathname = usePathname();
  const { logout } = useAuth();

  const handleLogout = () => {
    logout();
  };

  return (
    <header className={styles.header}>
      <div className={styles.headerContent}>
        <div className={styles.leftSection}>
          <Link href="/home">
            <h1 className="headerBrandName">Vibes</h1>
          </Link>
          <SearchBar />
        </div>
        <nav className={styles.nav}>
          <Link 
            href="/home" 
            className={`${styles.iconLink} ${pathname === '/home' ? styles.active : ''}`}
          >
            {/* <i className="fas fa-home"></i> */}
            <Home />
          </Link>
          <Link 
            href="/group-feeds" 
            className={`${styles.iconLink} ${pathname === '/group-feeds' ? styles.active : ''}`}
          >
            {/* <i className="fas fa-newspaper"></i> */}
            <Paperclip />
          </Link>
          <Link 
            href="/profile" 
            className={`${styles.iconLink} ${pathname.startsWith('/profile') ? styles.active : ''}`}
          >
            {/* <i className="fas fa-user"></i> */}
            <User />
          </Link>
          <Link 
            href="/messages" 
            className={`${styles.iconLink} ${pathname === '/messages' ? styles.active : ''}`}
          >
            {/* <i className="fas fa-envelope"></i> */}
            <MessageCircle />
          </Link>
          <NotificationDropdown />
          <ProfileDropdown />
        </nav>
      </div>
    </header>
  );
}