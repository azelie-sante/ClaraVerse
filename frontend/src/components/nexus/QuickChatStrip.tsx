import { memo, useMemo, useRef, useEffect } from 'react';
import { Brain, User } from 'lucide-react';
import { MarkdownRenderer } from '@/components/design-system/content/MarkdownRenderer';
import { useNexusStore } from '@/store/useNexusStore';
import styles from './Nexus.module.css';

/**
 * QuickChatStrip — shows inline chat bubbles for quick-mode responses.
 * Daemon/multi-daemon tasks only appear on the kanban board.
 * This gives users a conversational feel for simple questions.
 */
export const QuickChatStrip = memo(function QuickChatStrip() {
  const conversation = useNexusStore((s) => s.conversation);
  const tasks = useNexusStore((s) => s.tasks);
  const isProcessing = useNexusStore((s) => s.isProcessing);
  const scrollRef = useRef<HTMLDivElement>(null);

  // Build a set of task IDs that are quick-mode
  const quickTaskIds = useMemo(
    () => new Set(tasks.filter((t) => t.mode === 'quick').map((t) => t.id)),
    [tasks]
  );

  // Filter conversation to only quick-mode exchanges:
  // - user_message items (always show — we don't know the mode yet when they type)
  // - cortex_response with a taskId that's in quickTaskIds, OR no taskId (direct quick response)
  // - cortex_thinking (brief thinking indicator)
  // Exclude: daemon_activity, task_result, error (those belong to the kanban)
  const quickItems = useMemo(() => {
    const items = conversation.filter((item) => {
      if (item.type === 'user_message') return true;
      if (item.type === 'cortex_thinking') return true;
      if (item.type === 'cortex_response') {
        // No taskId means it's a direct quick response
        if (!item.taskId) return true;
        // If it has a taskId, only show if that task is quick mode
        return quickTaskIds.has(item.taskId);
      }
      return false;
    });

    // Only keep the last few exchanges to avoid clutter
    // Walk backwards to find up to 3 user messages and their responses
    const result: typeof items = [];
    let userCount = 0;
    for (let i = items.length - 1; i >= 0 && userCount < 3; i--) {
      result.unshift(items[i]);
      if (items[i].type === 'user_message') userCount++;
    }
    return result;
  }, [conversation, quickTaskIds]);

  // Auto-scroll to bottom on new items
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [quickItems.length]);

  // Don't render at all if no quick items
  if (quickItems.length === 0 && !isProcessing) return null;

  return (
    <div className={styles.quickChatStrip} ref={scrollRef}>
      {quickItems.map((item) => {
        if (item.type === 'user_message') {
          return (
            <div key={item.id} className={styles.quickChatBubbleRow}>
              <div className={styles.quickChatUserBubble}>
                <User size={12} />
                <span>{item.content}</span>
              </div>
            </div>
          );
        }

        if (item.type === 'cortex_thinking') {
          return (
            <div key={item.id} className={styles.quickChatBubbleRow}>
              <div className={styles.quickChatThinking}>
                <Brain size={12} />
                <span>{item.content}</span>
              </div>
            </div>
          );
        }

        if (item.type === 'cortex_response') {
          return (
            <div key={item.id} className={styles.quickChatBubbleRow}>
              <div className={styles.quickChatClaraBubble}>
                <Brain size={12} />
                <div className={styles.quickChatMd}>
                  <MarkdownRenderer content={item.content} />
                </div>
              </div>
            </div>
          );
        }

        return null;
      })}

      {isProcessing && quickItems.length > 0 && (
        <div className={styles.quickChatBubbleRow}>
          <div className={styles.quickChatThinking}>
            <Brain size={12} />
            <div className={styles.quickChatDots}>
              <span className={styles.thinkingDot} />
              <span className={styles.thinkingDot} />
              <span className={styles.thinkingDot} />
            </div>
          </div>
        </div>
      )}
    </div>
  );
});
