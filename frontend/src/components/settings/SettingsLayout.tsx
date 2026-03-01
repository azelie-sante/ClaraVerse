import React, { useState, useEffect } from 'react';
import { Menu, Check } from 'lucide-react';
import { SettingsSidebar } from './SettingsSidebar';
import styles from './SettingsLayout.module.css';

export type SettingsTab =
  | 'ai'
  | 'api-keys'
  | 'credentials'
  | 'channels'
  | 'devices'
  | 'usage'
  | 'privacy'
  | 'account';

const MOBILE_BREAKPOINT = 768;

const isMobileDevice = () => typeof window !== 'undefined' && window.innerWidth < MOBILE_BREAKPOINT;

export interface SettingsLayoutProps {
  /** Currently active tab */
  activeTab: SettingsTab;
  /** Callback when tab changes */
  onTabChange: (tab: SettingsTab) => void;
  /** Content to render */
  children: React.ReactNode;
  /** Show save indicator */
  showSaveIndicator?: boolean;
}

export const SettingsLayout: React.FC<SettingsLayoutProps> = ({
  activeTab,
  onTabChange,
  children,
  showSaveIndicator = false,
}) => {
  const [isMobile, setIsMobile] = useState(() => isMobileDevice());
  const [sidebarOpen, setSidebarOpen] = useState(() => !isMobileDevice());

  // Listen for window resize
  useEffect(() => {
    const handleResize = () => {
      const mobile = isMobileDevice();
      setIsMobile(mobile);
      if (mobile) {
        setSidebarOpen(false);
      }
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const handleTabChange = (tab: SettingsTab) => {
    onTabChange(tab);
    // Close sidebar on mobile after navigation
    if (isMobile) {
      setSidebarOpen(false);
    }
  };

  const handleBackdropClick = () => {
    if (isMobile && sidebarOpen) {
      setSidebarOpen(false);
    }
  };

  return (
    <div className={styles.container}>
      {/* Backdrop for mobile */}
      <div
        className={`${styles.backdrop} ${sidebarOpen ? styles.visible : ''}`}
        onClick={handleBackdropClick}
        aria-hidden="true"
      />

      {/* Sidebar */}
      <SettingsSidebar
        activeTab={activeTab}
        onTabChange={handleTabChange}
        isOpen={sidebarOpen}
        onOpenChange={setSidebarOpen}
      />

      {/* Main Content Area */}
      <main className={styles.content}>
        {/* Header with menu button and title */}
        <header className={styles.header}>
          <button
            className={styles.menuButton}
            onClick={() => setSidebarOpen(true)}
            aria-label="Open menu"
            type="button"
          >
            <Menu size={20} />
          </button>
          <h1 className={styles.title}>Settings</h1>
          <div className={`${styles.saveIndicator} ${showSaveIndicator ? styles.visible : ''}`}>
            <Check size={16} />
            <span>Saved</span>
          </div>
        </header>

        {/* Content */}
        <div className={styles.contentInner}>{children}</div>
      </main>
    </div>
  );
};
