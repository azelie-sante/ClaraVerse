/**
 * VersionNavigator Component
 *
 * Shows version navigation for assistant messages with multiple versions.
 * Displays "< 2/3 >" style navigation to browse between response versions.
 * Only renders when there are multiple versions (totalVersions > 1).
 */

import { ChevronLeft, ChevronRight } from 'lucide-react';

interface VersionNavigatorProps {
  currentVersion: number;
  totalVersions: number;
  onNavigate: (direction: 'prev' | 'next') => void;
}

export function VersionNavigator({
  currentVersion,
  totalVersions,
  onNavigate,
}: VersionNavigatorProps) {
  // Don't render if only one version
  if (totalVersions <= 1) return null;

  const canGoPrev = currentVersion > 1;
  const canGoNext = currentVersion < totalVersions;

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: '2px',
        marginTop: 'var(--space-2)',
        marginBottom: 'var(--space-1)',
      }}
    >
      <button
        onClick={() => onNavigate('prev')}
        disabled={!canGoPrev}
        aria-label="Previous version"
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          width: '24px',
          height: '24px',
          padding: 0,
          background: 'transparent',
          border: 'none',
          borderRadius: 'var(--radius-sm)',
          color: canGoPrev ? 'var(--color-text-secondary)' : 'var(--color-text-tertiary)',
          cursor: canGoPrev ? 'pointer' : 'not-allowed',
          opacity: canGoPrev ? 1 : 0.4,
          transition: 'all 0.15s',
        }}
        onMouseEnter={e => {
          if (canGoPrev) {
            e.currentTarget.style.background = 'rgba(255, 255, 255, 0.1)';
            e.currentTarget.style.color = 'var(--color-text-primary)';
          }
        }}
        onMouseLeave={e => {
          e.currentTarget.style.background = 'transparent';
          e.currentTarget.style.color = canGoPrev
            ? 'var(--color-text-secondary)'
            : 'var(--color-text-tertiary)';
        }}
      >
        <ChevronLeft size={16} />
      </button>

      <span
        style={{
          fontSize: '12px',
          fontWeight: 500,
          color: 'var(--color-text-secondary)',
          minWidth: '32px',
          textAlign: 'center',
          userSelect: 'none',
        }}
      >
        {currentVersion}/{totalVersions}
      </span>

      <button
        onClick={() => onNavigate('next')}
        disabled={!canGoNext}
        aria-label="Next version"
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          width: '24px',
          height: '24px',
          padding: 0,
          background: 'transparent',
          border: 'none',
          borderRadius: 'var(--radius-sm)',
          color: canGoNext ? 'var(--color-text-secondary)' : 'var(--color-text-tertiary)',
          cursor: canGoNext ? 'pointer' : 'not-allowed',
          opacity: canGoNext ? 1 : 0.4,
          transition: 'all 0.15s',
        }}
        onMouseEnter={e => {
          if (canGoNext) {
            e.currentTarget.style.background = 'rgba(255, 255, 255, 0.1)';
            e.currentTarget.style.color = 'var(--color-text-primary)';
          }
        }}
        onMouseLeave={e => {
          e.currentTarget.style.background = 'transparent';
          e.currentTarget.style.color = canGoNext
            ? 'var(--color-text-secondary)'
            : 'var(--color-text-tertiary)';
        }}
      >
        <ChevronRight size={16} />
      </button>
    </div>
  );
}
