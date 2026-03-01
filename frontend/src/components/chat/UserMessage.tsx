/**
 * UserMessage Component
 *
 * Memoized component for rendering user messages in the chat.
 * Prevents re-renders during streaming since user messages don't change.
 */

import { memo } from 'react';
import { Copy, Check } from 'lucide-react';
import type { Message } from '@/types/chat';
import { MarkdownRenderer } from '@/components/design-system/content/MarkdownRenderer';
import { MessageAttachment } from './MessageAttachment';
import styles from '@/pages/Chat.module.css';

export interface UserMessageProps {
  message: Message;
  userInitials: string;
  copiedMessageId: string | null;
  onCopy: (content: string, id: string) => void;
}

function UserMessageComponent({
  message,
  userInitials,
  copiedMessageId,
  onCopy,
}: UserMessageProps) {
  return (
    <>
      {/* File Attachments - shown above chat bubble */}
      {message.attachments && message.attachments.length > 0 && (
        <div style={{ marginBottom: 'var(--space-3)' }}>
          <MessageAttachment attachments={message.attachments} />
        </div>
      )}
      <div className={styles.userMessageRow}>
        <div className={styles.userMessage}>
          <div className={styles.userBadge} aria-label="User message">
            {userInitials}
          </div>
          <div className={styles.messageText}>
            <MarkdownRenderer content={message.content} />
          </div>
        </div>
        {/* Copy button - inline on mobile */}
        <button
          onClick={() => onCopy(message.content, message.id)}
          className={styles.userCopyButton}
          aria-label={copiedMessageId === message.id ? 'Copied' : 'Copy message'}
        >
          {copiedMessageId === message.id ? (
            <Check size={14} aria-hidden="true" />
          ) : (
            <Copy size={14} aria-hidden="true" />
          )}
        </button>
      </div>
    </>
  );
}

/**
 * Memoized UserMessage - only re-renders when:
 * - message.id changes (new message)
 * - message.content changes (edited)
 * - copiedMessageId changes to/from this message's ID
 */
export const UserMessage = memo(UserMessageComponent, (prevProps, nextProps) => {
  // Re-render if message identity or content changed
  if (prevProps.message.id !== nextProps.message.id) return false;
  if (prevProps.message.content !== nextProps.message.content) return false;

  // Re-render if copy state changed for THIS message
  const prevIsCopied = prevProps.copiedMessageId === prevProps.message.id;
  const nextIsCopied = nextProps.copiedMessageId === nextProps.message.id;
  if (prevIsCopied !== nextIsCopied) return false;

  // Re-render if attachments changed
  if (prevProps.message.attachments?.length !== nextProps.message.attachments?.length) return false;

  // No changes that affect this component
  return true;
});

UserMessage.displayName = 'UserMessage';
