/**
 * RetryDropdown Component
 *
 * Dropdown menu for retry options in assistant messages.
 * Shows options like "Try again", "Add more details", "More concise", etc.
 * Opens up or down based on available viewport space.
 */

import { useState, useRef, useEffect, useCallback } from 'react';
import { createPortal } from 'react-dom';
import {
  RotateCcw,
  ChevronDown,
  FileText,
  Scissors,
  Globe,
  Brain,
  type LucideIcon,
} from 'lucide-react';
import type { RetryType } from '@/types/chat';

interface RetryOption {
  id: RetryType;
  label: string;
  icon: LucideIcon;
  description?: string;
}

const RETRY_OPTIONS: RetryOption[] = [
  { id: 'regenerate', label: 'Try again', icon: RotateCcw },
  {
    id: 'add_details',
    label: 'Add more details',
    icon: FileText,
    description: 'Get a more detailed response',
  },
  {
    id: 'more_concise',
    label: 'More concise',
    icon: Scissors,
    description: 'Get a shorter response',
  },
  {
    id: 'no_search',
    label: "Don't search the web",
    icon: Globe,
    description: 'Answer without web search',
  },
  {
    id: 'think_longer',
    label: 'Think longer',
    icon: Brain,
    description: 'Take more time to reason',
  },
];

interface RetryDropdownProps {
  onRetry: (type: RetryType) => void;
  disabled?: boolean;
}

export function RetryDropdown({ onRetry, disabled }: RetryDropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [openDirection, setOpenDirection] = useState<'up' | 'down'>('down');
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const buttonRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const [menuPosition, setMenuPosition] = useState({ top: 0, left: 0 });

  // Calculate menu position and direction
  const updateMenuPosition = useCallback(() => {
    if (!buttonRef.current) return;

    const buttonRect = buttonRef.current.getBoundingClientRect();
    const viewportHeight = window.innerHeight;
    const menuHeight = 280; // Approximate menu height

    // Check if there's more space above or below
    const spaceBelow = viewportHeight - buttonRect.bottom;
    const spaceAbove = buttonRect.top;

    if (spaceBelow < menuHeight && spaceAbove > spaceBelow) {
      setOpenDirection('up');
      setMenuPosition({
        top: buttonRect.top - menuHeight - 4,
        left: buttonRect.left,
      });
    } else {
      setOpenDirection('down');
      setMenuPosition({
        top: buttonRect.bottom + 4,
        left: buttonRect.left,
      });
    }
  }, []);

  // Handle opening/closing
  const toggleOpen = useCallback(() => {
    if (disabled) return;

    if (!isOpen) {
      updateMenuPosition();
    }
    setIsOpen(!isOpen);
    setFocusedIndex(-1);
  }, [isOpen, disabled, updateMenuPosition]);

  // Handle option selection
  const handleSelect = useCallback(
    (type: RetryType) => {
      setIsOpen(false);
      onRetry(type);
    },
    [onRetry]
  );

  // Keyboard navigation
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (!isOpen) {
        if (e.key === 'Enter' || e.key === ' ' || e.key === 'ArrowDown') {
          e.preventDefault();
          toggleOpen();
        }
        return;
      }

      switch (e.key) {
        case 'Escape':
          e.preventDefault();
          setIsOpen(false);
          buttonRef.current?.focus();
          break;
        case 'ArrowDown':
          e.preventDefault();
          setFocusedIndex(prev => (prev + 1) % RETRY_OPTIONS.length);
          break;
        case 'ArrowUp':
          e.preventDefault();
          setFocusedIndex(prev => (prev - 1 + RETRY_OPTIONS.length) % RETRY_OPTIONS.length);
          break;
        case 'Enter':
        case ' ':
          e.preventDefault();
          if (focusedIndex >= 0) {
            handleSelect(RETRY_OPTIONS[focusedIndex].id);
          }
          break;
        case 'Tab':
          setIsOpen(false);
          break;
      }
    },
    [isOpen, focusedIndex, toggleOpen, handleSelect]
  );

  // Close on click outside
  useEffect(() => {
    if (!isOpen) return;

    const handleClickOutside = (e: MouseEvent) => {
      if (
        buttonRef.current &&
        !buttonRef.current.contains(e.target as Node) &&
        menuRef.current &&
        !menuRef.current.contains(e.target as Node)
      ) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  // Update position on scroll/resize
  useEffect(() => {
    if (!isOpen) return;

    const handleUpdate = () => updateMenuPosition();
    window.addEventListener('scroll', handleUpdate, true);
    window.addEventListener('resize', handleUpdate);

    return () => {
      window.removeEventListener('scroll', handleUpdate, true);
      window.removeEventListener('resize', handleUpdate);
    };
  }, [isOpen, updateMenuPosition]);

  return (
    <>
      <button
        ref={buttonRef}
        onClick={toggleOpen}
        onKeyDown={handleKeyDown}
        disabled={disabled}
        aria-label="Retry options"
        aria-haspopup="menu"
        aria-expanded={isOpen}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: '2px',
          padding: '4px 6px',
          background: 'transparent',
          border: 'none',
          borderRadius: 'var(--radius-sm)',
          color: 'var(--color-text-secondary)',
          cursor: disabled ? 'not-allowed' : 'pointer',
          opacity: disabled ? 0.5 : 1,
          transition: 'all 0.15s',
        }}
        onMouseEnter={e => {
          if (!disabled) {
            e.currentTarget.style.background = 'rgba(255, 255, 255, 0.1)';
            e.currentTarget.style.color = 'var(--color-text-primary)';
          }
        }}
        onMouseLeave={e => {
          e.currentTarget.style.background = 'transparent';
          e.currentTarget.style.color = 'var(--color-text-secondary)';
        }}
      >
        <RotateCcw size={16} />
        <ChevronDown
          size={12}
          style={{
            transform: isOpen ? 'rotate(180deg)' : 'rotate(0deg)',
            transition: 'transform 0.15s',
          }}
        />
      </button>

      {isOpen &&
        createPortal(
          <div
            ref={menuRef}
            role="menu"
            aria-orientation="vertical"
            style={{
              position: 'fixed',
              top: `${menuPosition.top}px`,
              left: `${menuPosition.left}px`,
              minWidth: '200px',
              maxWidth: '280px',
              background: 'rgba(32, 32, 32, 0.98)',
              backdropFilter: 'blur(20px)',
              WebkitBackdropFilter: 'blur(20px)',
              border: '1px solid rgba(255, 255, 255, 0.1)',
              borderRadius: 'var(--radius-lg)',
              boxShadow: '0 10px 40px rgba(0, 0, 0, 0.5)',
              padding: '4px',
              zIndex: 1060,
              transformOrigin: openDirection === 'up' ? 'bottom left' : 'top left',
              animation: 'dropdownFadeIn 0.15s ease-out',
            }}
          >
            {RETRY_OPTIONS.map((option, index) => {
              const Icon = option.icon;
              const isFocused = index === focusedIndex;

              return (
                <button
                  key={option.id}
                  role="menuitem"
                  onClick={() => handleSelect(option.id)}
                  onMouseEnter={() => setFocusedIndex(index)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '10px',
                    width: '100%',
                    padding: '10px 12px',
                    background: isFocused ? 'rgba(255, 255, 255, 0.08)' : 'transparent',
                    border: 'none',
                    borderRadius: 'var(--radius-md)',
                    color: 'var(--color-text-primary)',
                    fontSize: '14px',
                    textAlign: 'left',
                    cursor: 'pointer',
                    transition: 'background 0.1s',
                  }}
                >
                  <Icon size={16} style={{ color: 'var(--color-text-secondary)', flexShrink: 0 }} />
                  <span style={{ flex: 1 }}>{option.label}</span>
                </button>
              );
            })}
          </div>,
          document.body
        )}

      <style>
        {`
          @keyframes dropdownFadeIn {
            from {
              opacity: 0;
              transform: scale(0.95) translateY(${openDirection === 'up' ? '8px' : '-8px'});
            }
            to {
              opacity: 1;
              transform: scale(1) translateY(0);
            }
          }
        `}
      </style>
    </>
  );
}
