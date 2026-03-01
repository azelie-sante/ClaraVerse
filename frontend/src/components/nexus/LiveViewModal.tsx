import { memo, useMemo, useState, useCallback, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { X, Send, Loader2, Bot, Wrench } from 'lucide-react';
import { Badge } from '@/components/design-system';
import { DaemonActivityBlock } from './DaemonActivityBlock';
import { useNexusStore } from '@/store/useNexusStore';
import type { NexusTaskStatus } from '@/types/nexus';
import styles from './Nexus.module.css';

interface LiveViewModalProps {
  taskId: string;
  onClose: () => void;
  onSendFollowUp?: (content: string, taskId: string) => void;
}

const statusConfig: Record<NexusTaskStatus, { variant: 'default' | 'accent' | 'success' | 'warning' | 'error'; label: string }> = {
  pending: { variant: 'default', label: 'Queued' },
  executing: { variant: 'accent', label: 'Running' },
  waiting_input: { variant: 'warning', label: 'Waiting' },
  completed: { variant: 'success', label: 'Done' },
  failed: { variant: 'error', label: 'Failed' },
  cancelled: { variant: 'default', label: 'Cancelled' },
};

export const LiveViewModal = memo(function LiveViewModal({
  taskId,
  onClose,
  onSendFollowUp,
}: LiveViewModalProps) {
  const tasks = useNexusStore((s) => s.tasks);
  const daemons = useNexusStore((s) => s.daemons);
  const conversation = useNexusStore((s) => s.conversation);
  const [replyText, setReplyText] = useState('');

  const task = useMemo(() => tasks.find((t) => t.id === taskId), [tasks, taskId]);
  const taskDaemons = useMemo(
    () => Object.values(daemons).filter((d) => d.task_id === taskId),
    [daemons, taskId]
  );

  const daemonIds = useMemo(() => new Set(taskDaemons.map((d) => d.id).filter(Boolean)), [taskDaemons]);
  const taskConversation = useMemo(
    () => conversation.filter((item) => item.taskId === taskId || (item.daemonId && daemonIds.has(item.daemonId))),
    [conversation, taskId, daemonIds]
  );

  const groupedByDaemon = useMemo(() => {
    const groups: Record<string, typeof taskConversation> = {};
    for (const item of taskConversation) {
      if (item.daemonId) {
        if (!groups[item.daemonId]) groups[item.daemonId] = [];
        groups[item.daemonId].push(item);
      }
    }
    return groups;
  }, [taskConversation]);

  const handleSendReply = useCallback(() => {
    if (!replyText.trim() || !onSendFollowUp) return;
    onSendFollowUp(replyText.trim(), taskId);
    setReplyText('');
  }, [replyText, taskId, onSendFollowUp]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        handleSendReply();
      }
    },
    [handleSendReply]
  );

  // Close on Escape
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    document.addEventListener('keydown', handler);
    return () => document.removeEventListener('keydown', handler);
  }, [onClose]);

  if (!task) return null;

  const cfg = statusConfig[task.status] ?? statusConfig.pending;
  const isRunning = task.status === 'executing';
  const isWaiting = task.status === 'waiting_input';

  return createPortal(
    <div className={styles.liveViewOverlay} onClick={onClose}>
      <div className={styles.liveViewModal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.liveViewHeader}>
          <div className={styles.liveViewHeaderLeft}>
            <Badge variant={cfg.variant}>{cfg.label}</Badge>
            <span className={styles.liveViewTitle}>
              {task.goal || task.prompt}
            </span>
          </div>
          <button className={styles.liveViewCloseBtn} onClick={onClose}>
            <X size={18} />
          </button>
        </div>

        <div className={styles.liveViewBody}>
          {taskDaemons.length > 0 ? (
            taskDaemons.map((daemon) => {
              const items = groupedByDaemon[daemon.id!] ?? [];
              if (items.length > 0) {
                return (
                  <DaemonActivityBlock
                    key={daemon.id}
                    daemonId={daemon.id!}
                    daemonRole={daemon.role_label || daemon.role || 'Daemon'}
                    items={items}
                  />
                );
              }
              return (
                <div key={daemon.id} className={styles.liveViewDaemonLive}>
                  <div className={styles.liveViewDaemonHeader}>
                    <Bot size={14} />
                    <span>{daemon.role_label || daemon.role || 'Daemon'}</span>
                    <Badge
                      variant={daemon.status === 'executing' ? 'accent' : daemon.status === 'completed' ? 'success' : daemon.status === 'failed' ? 'error' : 'default'}
                    >
                      {daemon.status === 'executing' ? 'Running' : daemon.status ?? 'Idle'}
                    </Badge>
                  </div>
                  {daemon.task_summary && (
                    <div className={styles.liveViewDaemonTask}>{daemon.task_summary}</div>
                  )}
                  {daemon.current_action && daemon.status === 'executing' && (
                    <div className={styles.liveViewDaemonAction}>
                      <Wrench size={11} />
                      <span>{daemon.current_action}</span>
                    </div>
                  )}
                  {daemon.status === 'executing' && (
                    <div className={styles.liveViewDaemonIndicator}>
                      <span className={styles.thinkingDot} />
                      <span className={styles.thinkingDot} />
                      <span className={styles.thinkingDot} />
                    </div>
                  )}
                </div>
              );
            })
          ) : isRunning ? (
            <div className={styles.liveViewWaiting}>
              <Loader2 size={20} className={styles.spin} />
              <span>Setting up daemons...</span>
            </div>
          ) : (
            <div className={styles.liveViewEmpty}>No activity recorded yet.</div>
          )}
        </div>

        {onSendFollowUp && (isRunning || isWaiting) && (
          <div className={styles.liveViewFooter}>
            <input
              className={styles.liveViewReplyInput}
              value={replyText}
              onChange={(e) => setReplyText(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={isWaiting ? 'Respond to daemon...' : 'Ask a follow-up...'}
            />
            <button
              className={styles.liveViewReplyBtn}
              onClick={handleSendReply}
              disabled={!replyText.trim()}
            >
              <Send size={14} />
            </button>
          </div>
        )}
      </div>
    </div>,
    document.body
  );
});
