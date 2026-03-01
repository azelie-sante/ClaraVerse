/**
 * Chat status utilities
 * Determines the backend sync status of conversations
 */

import type { Chat, ChatStatus } from '@/types/chat';
import { isConversationStale } from '@/services/conversationService';

/**
 * Compute the current status of a chat
 * - local-only: New chat, no messages sent to backend yet (only 1 user message)
 * - active: Recently active (<25 min), likely exists on backend
 * - stale: Inactive for >25 min, might be expired on backend
 * - expired: Confirmed expired by backend (manually set)
 */
export function computeChatStatus(chat: Chat): ChatStatus {
  // If backend status is explicitly set (e.g., after checking API), use it
  if (chat.backendStatus === 'expired') {
    return 'expired';
  }

  // Guard against missing messages array (defensive programming for corrupted data)
  const messages = chat.messages || [];

  // Check if this is a brand new chat (only user message, no assistant response yet)
  const hasAssistantMessage = messages.some(msg => msg.role === 'assistant');
  if (!hasAssistantMessage && messages.length === 1) {
    return 'local-only';
  }

  // If no activity timestamp, assume local-only
  if (!chat.lastActivityAt) {
    return 'local-only';
  }

  // Check if conversation is stale
  if (isConversationStale(chat.lastActivityAt)) {
    return 'stale';
  }

  // Active conversation
  return 'active';
}

/**
 * Get status display info (color, label, description)
 */
export function getStatusInfo(status: ChatStatus): {
  color: string;
  label: string;
  description: string;
} {
  switch (status) {
    case 'local-only':
      return {
        color: '#888888',
        label: 'Local',
        description: 'Not yet sent to server',
      };
    case 'active':
      return {
        color: '#10b981', // green
        label: 'Active',
        description: 'Available on server',
      };
    case 'stale':
      return {
        color: '#f59e0b', // amber/orange
        label: 'Stale',
        description: 'May have expired on server (>25 min inactive)',
      };
    case 'expired':
      return {
        color: '#ef4444', // red
        label: 'Expired',
        description: 'Expired on server (history in local storage only)',
      };
  }
}

/**
 * Get human-readable time since last activity
 */
export function getTimeSinceActivity(lastActivityAt?: Date | string | number): string {
  if (!lastActivityAt) {
    return 'Never';
  }

  const now = new Date().getTime();
  // Ensure lastActivityAt is a proper Date object (handles string/number from JSON/IndexedDB)
  const dateObj = lastActivityAt instanceof Date ? lastActivityAt : new Date(lastActivityAt);
  const lastActivity = dateObj.getTime();
  const diffMs = now - lastActivity;

  const minutes = Math.floor(diffMs / (60 * 1000));
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) {
    return `${days}d ago`;
  }
  if (hours > 0) {
    return `${hours}h ago`;
  }
  if (minutes > 0) {
    return `${minutes}m ago`;
  }
  return 'Just now';
}
