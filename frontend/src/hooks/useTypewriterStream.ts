import { useEffect, useRef, useState } from 'react';

interface StreamBuffer {
  pendingBuffer: string; // Chunks waiting to be displayed
  displayedContent: string; // Content currently visible to user
  isFinished: boolean; // Whether backend finished streaming
}

interface UseTypewriterStreamReturn {
  addChunk: (messageId: string, chunk: string) => void;
  getDisplayedContent: (messageId: string) => string;
  finishStreaming: (messageId: string) => void;
  clearMessage: (messageId: string) => void;
  isTyping: (messageId: string) => boolean;
}

const TYPING_INTERVAL = 50; // 50ms per character = 20 chars/second

/**
 * Hook to create smooth typewriter effect for streaming messages
 * Buffers incoming chunks and displays them character-by-character at consistent speed
 */
export function useTypewriterStream(): UseTypewriterStreamReturn {
  const [, forceUpdate] = useState({});
  const buffersRef = useRef<Map<string, StreamBuffer>>(new Map());
  const intervalRef = useRef<number | null>(null);

  // Start typing interval when component mounts
  useEffect(() => {
    intervalRef.current = window.setInterval(() => {
      let hasChanges = false;

      buffersRef.current.forEach((buffer, _messageId) => {
        // If there's content in pending buffer, move characters to displayed
        if (buffer.pendingBuffer.length > 0) {
          // Adaptive character display based on buffer size
          // Small buffer: 1 char/tick (smooth)
          // Medium buffer: 2-3 chars/tick (faster)
          // Large buffer: 5+ chars/tick (catch up)
          let charsToDisplay = 1;

          const bufferSize = buffer.pendingBuffer.length;
          if (bufferSize > 200) {
            charsToDisplay = 8; // Very fast catch-up
          } else if (bufferSize > 100) {
            charsToDisplay = 5; // Fast catch-up
          } else if (bufferSize > 50) {
            charsToDisplay = 3; // Medium speed
          } else if (bufferSize > 20) {
            charsToDisplay = 2; // Slightly faster
          }
          // else: 1 char (default smooth speed)

          // Extract and display characters
          const charsToShow = Math.min(charsToDisplay, buffer.pendingBuffer.length);
          const nextChars = buffer.pendingBuffer.substring(0, charsToShow);
          buffer.pendingBuffer = buffer.pendingBuffer.slice(charsToShow);
          buffer.displayedContent += nextChars;
          hasChanges = true;
        }
      });

      // Force re-render if any content was updated
      if (hasChanges) {
        forceUpdate({});
      }
    }, TYPING_INTERVAL);

    // Cleanup interval on unmount
    return () => {
      if (intervalRef.current) {
        window.clearInterval(intervalRef.current);
      }
    };
  }, []);

  const addChunk = (messageId: string, chunk: string) => {
    let buffer = buffersRef.current.get(messageId);

    if (!buffer) {
      buffer = {
        pendingBuffer: '',
        displayedContent: '',
        isFinished: false,
      };
      buffersRef.current.set(messageId, buffer);
    }

    // Add chunk to pending buffer
    buffer.pendingBuffer += chunk;
  };

  const getDisplayedContent = (messageId: string): string => {
    const buffer = buffersRef.current.get(messageId);
    return buffer?.displayedContent || '';
  };

  const finishStreaming = (messageId: string) => {
    const buffer = buffersRef.current.get(messageId);
    if (buffer) {
      buffer.isFinished = true;
    }
  };

  const clearMessage = (messageId: string) => {
    buffersRef.current.delete(messageId);
  };

  const isTyping = (messageId: string): boolean => {
    const buffer = buffersRef.current.get(messageId);
    if (!buffer) return false;

    // Still typing if there's content in pending buffer
    return buffer.pendingBuffer.length > 0;
  };

  return {
    addChunk,
    getDisplayedContent,
    finishStreaming,
    clearMessage,
    isTyping,
  };
}
