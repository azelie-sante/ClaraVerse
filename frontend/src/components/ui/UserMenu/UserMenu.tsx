import { useState, useRef, useEffect } from 'react';
import { LogOut, User, ChevronDown } from 'lucide-react';
import { useAuthStore } from '@/store/useAuthStore';
import { useNavigate } from 'react-router-dom';
import './UserMenu.css';

export const UserMenu = ({ collapsed = false }: { collapsed?: boolean }) => {
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const { user, signOut } = useAuthStore();
  const navigate = useNavigate();

  // Get user initials from display name or email
  const getUserInitials = (): string => {
    if (user?.user_metadata?.display_name) {
      const name = user.user_metadata.display_name as string;
      const parts = name.trim().split(' ');
      if (parts.length >= 2) {
        return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
      }
      return name.substring(0, 2).toUpperCase();
    }
    if (user?.email) {
      return user.email.substring(0, 2).toUpperCase();
    }
    return 'U';
  };

  // Get display name
  const getDisplayName = (): string => {
    if (user?.user_metadata?.display_name) {
      return user.user_metadata.display_name as string;
    }
    if (user?.email) {
      return user.email;
    }
    return 'User';
  };

  // Close menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  const handleSignOut = async () => {
    await signOut();
    setIsOpen(false);
    navigate('/signin');
  };

  if (!user) return null;

  return (
    <div className={`user-menu ${collapsed ? 'collapsed' : ''}`} ref={menuRef}>
      <button
        className="user-menu-trigger"
        onClick={() => setIsOpen(!isOpen)}
        aria-label="User menu"
        aria-expanded={isOpen}
        style={collapsed ? { padding: 'var(--space-2)', justifyContent: 'center' } : undefined}
      >
        <div className="user-avatar">{getUserInitials()}</div>
        {!collapsed && (
          <>
            <span className="user-name">{getDisplayName()}</span>
            <ChevronDown
              size={16}
              className="user-menu-chevron"
              style={{
                transform: isOpen ? 'rotate(180deg)' : 'rotate(0deg)',
              }}
            />
          </>
        )}
      </button>

      {isOpen && (
        <div
          className="user-menu-dropdown"
          style={
            collapsed
              ? { left: '100%', top: 'auto', bottom: 0, marginLeft: 'var(--space-2)' }
              : undefined
          }
        >
          <div className="user-menu-header">
            <div className="user-avatar-large">{getUserInitials()}</div>
            <div className="user-info">
              <div className="user-info-name">{getDisplayName()}</div>
              {user.email && <div className="user-info-email">{user.email}</div>}
            </div>
          </div>

          <div className="user-menu-divider" />

          <button className="user-menu-item" onClick={() => navigate('/settings')}>
            <User size={16} />
            <span>Profile Settings</span>
          </button>

          <div className="user-menu-divider" />

          <button className="user-menu-item user-menu-item-danger" onClick={handleSignOut}>
            <LogOut size={16} />
            <span>Sign Out</span>
          </button>
        </div>
      )}
    </div>
  );
};
