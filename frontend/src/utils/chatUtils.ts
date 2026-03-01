/**
 * Chat Utilities
 * Helper functions for chat functionality
 */

/**
 * Truncate text to a maximum length
 */
export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
}

/**
 * Get user initials from name
 */
export function getUserInitials(name: string): string {
  const parts = name.trim().split(' ');
  if (parts.length === 0) return '?';
  if (parts.length === 1) return parts[0].charAt(0).toUpperCase();
  return (parts[0].charAt(0) + parts[parts.length - 1].charAt(0)).toUpperCase();
}

/**
 * Format timestamp for display
 */
export function formatTimestamp(date: Date | undefined | null): string {
  if (!date) return '';

  try {
    const dateObj = date instanceof Date ? date : new Date(date);
    if (isNaN(dateObj.getTime())) return '';

    return dateObj.toLocaleTimeString('en-US', {
      hour: 'numeric',
      minute: '2-digit',
      hour12: true,
    });
  } catch {
    return '';
  }
}

/**
 * Check if a message contains code blocks
 */
export function containsCodeBlock(content: string): boolean {
  return /```[\s\S]*?```/.test(content);
}

/**
 * Extract code blocks from message
 */
export function extractCodeBlocks(content: string): Array<{
  language: string;
  code: string;
}> {
  const regex = /```(\w+)?\n([\s\S]*?)```/g;
  const blocks: Array<{ language: string; code: string }> = [];
  let match;

  while ((match = regex.exec(content)) !== null) {
    blocks.push({
      language: match[1] || 'text',
      code: match[2].trim(),
    });
  }

  return blocks;
}

/**
 * Estimate reading time for message
 */
export function estimateReadingTime(content: string): number {
  const wordsPerMinute = 200;
  const words = content.trim().split(/\s+/).length;
  return Math.ceil(words / wordsPerMinute);
}

/**
 * Check if chat is empty (no messages)
 */
export function isChatEmpty(messageCount: number): boolean {
  return messageCount === 0;
}

/**
 * Get chat summary (first N characters)
 */
export function getChatSummary(
  messages: Array<{ content: string }>,
  maxLength: number = 100
): string {
  if (messages.length === 0) return 'No messages';
  const firstMessage = messages[0].content;
  return truncateText(firstMessage, maxLength);
}

/**
 * Count messages by role
 */
export function countMessagesByRole(messages: Array<{ role: 'user' | 'assistant' }>): {
  user: number;
  assistant: number;
} {
  return messages.reduce(
    (acc, msg) => {
      acc[msg.role]++;
      return acc;
    },
    { user: 0, assistant: 0 }
  );
}

/**
 * Calculate chat statistics
 */
export function getChatStats(messages: Array<{ role: 'user' | 'assistant'; content: string }>) {
  const { user, assistant } = countMessagesByRole(messages);
  const totalWords = messages.reduce((acc, msg) => {
    return acc + msg.content.trim().split(/\s+/).length;
  }, 0);

  return {
    totalMessages: messages.length,
    userMessages: user,
    assistantMessages: assistant,
    totalWords,
    averageWordsPerMessage: messages.length > 0 ? Math.round(totalWords / messages.length) : 0,
  };
}
