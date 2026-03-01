/**
 * ChatStatusIndicator
 * Shows the backend sync status of a chat conversation
 */

import type { ChatStatus } from '@/types/chat';
import { getStatusInfo } from '@/utils/chatStatus';

interface ChatStatusIndicatorProps {
  status: ChatStatus;
  showLabel?: boolean;
  size?: 'sm' | 'md';
}

export const ChatStatusIndicator = ({
  status,
  showLabel = false,
  size = 'sm',
}: ChatStatusIndicatorProps) => {
  const info = getStatusInfo(status);
  const dotSize = size === 'sm' ? 8 : 10;

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
        gap: 'var(--space-2)',
      }}
      title={info.description}
    >
      {/* Status dot */}
      <div
        style={{
          width: `${dotSize}px`,
          height: `${dotSize}px`,
          borderRadius: '50%',
          backgroundColor: info.color,
          flexShrink: 0,
          boxShadow: status === 'active' ? `0 0 8px ${info.color}40` : 'none',
        }}
      />

      {/* Optional label */}
      {showLabel && (
        <span
          style={{
            fontSize: size === 'sm' ? 'var(--text-xs)' : 'var(--text-sm)',
            color: info.color,
            fontWeight: 'var(--font-medium)',
          }}
        >
          {info.label}
        </span>
      )}
    </div>
  );
};
