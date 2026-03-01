import { memo, useMemo, useRef, useState, useEffect } from 'react';
import { Zap, GitBranch, Trash2, Bot, Eye, RotateCcw, Square } from 'lucide-react';
import { draggable } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { setCustomNativeDragPreview } from '@atlaskit/pragmatic-drag-and-drop/element/set-custom-native-drag-preview';
import { pointerOutsideOfPreview } from '@atlaskit/pragmatic-drag-and-drop/element/pointer-outside-of-preview';
import { Badge } from '@/components/design-system';
import type { NexusTask, Daemon } from '@/types/nexus';
import styles from './Nexus.module.css';

interface KanbanTaskCardProps {
  task: NexusTask;
  daemons: Partial<Daemon>[];
  selected: boolean;
  onSelect: (id: string) => void;
  onDelete?: (id: string) => void;
  onViewLive?: (taskId: string) => void;
  onRetryTask?: (taskId: string) => void;
  onCancelTask?: (taskId: string) => void;
}

const modeLabels: Record<string, { icon: React.ReactNode; label: string }> = {
  quick: { icon: <Zap size={10} />, label: 'Quick' },
  daemon: { icon: <Bot size={10} />, label: 'Daemon' },
  multi_daemon: { icon: <GitBranch size={10} />, label: `Multi` },
};

function relativeTime(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const seconds = Math.floor(diff / 1000);
  if (seconds < 60) return 'just now';
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

const MAX_TOOL_PILLS = 3;

export const KanbanTaskCard = memo(function KanbanTaskCard({
  task,
  daemons,
  selected,
  onSelect,
  onDelete,
  onViewLive,
  onRetryTask,
  onCancelTask,
}: KanbanTaskCardProps) {
  const cardRef = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);

  const isTerminal =
    task.status === 'completed' || task.status === 'failed' || task.status === 'cancelled';
  const isRunning = task.status === 'executing';
  const mode = modeLabels[task.mode] ?? modeLabels.daemon;

  // Status dot color for footer
  const statusDotColor: Record<string, string> = {
    pending: 'rgba(255, 255, 255, 0.3)',
    executing: 'var(--color-accent)',
    waiting_input: 'var(--color-warning, #ff9800)',
    completed: 'var(--color-success)',
    failed: 'var(--color-error)',
    cancelled: 'rgba(255, 152, 0, 0.6)',
  };

  // Set up draggable
  useEffect(() => {
    const el = cardRef.current;
    if (!el) return;
    return draggable({
      element: el,
      getInitialData: () => ({ taskId: task.id, status: task.status }),
      onGenerateDragPreview: ({ nativeSetDragImage }) => {
        setCustomNativeDragPreview({
          nativeSetDragImage,
          getOffset: pointerOutsideOfPreview({ x: '16px', y: '8px' }),
          render: ({ container }) => {
            const root = document.documentElement;
            const cs = getComputedStyle(root);
            const bg = cs.getPropertyValue('--color-surface').trim() || '#1e1e2e';
            const fg = cs.getPropertyValue('--color-text-primary').trim() || '#e0e0e0';
            const border = cs.getPropertyValue('--color-border').trim() || 'rgba(255,255,255,0.15)';
            const preview = document.createElement('div');
            preview.textContent = (task.goal || task.prompt || '').slice(0, 60);
            preview.style.cssText =
              `padding:8px 12px;background:${bg};color:${fg};border:1px solid ${border};` +
              'border-radius:8px;font-size:12px;max-width:220px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;' +
              'box-shadow:0 8px 24px rgba(0,0,0,0.4);';
            container.appendChild(preview);
          },
        });
      },
      onDragStart: () => setIsDragging(true),
      onDrop: () => setIsDragging(false),
    });
  }, [task.id, task.status, task.goal, task.prompt]);

  const avgProgress = useMemo(() => {
    if (daemons.length === 0) return 0;
    const total = daemons.reduce((sum, d) => sum + (d.progress ?? 0), 0);
    return total / daemons.length;
  }, [daemons]);

  // Get the first active daemon's current action for a live status line
  const liveAction = useMemo(() => {
    const active = daemons.find(d => d.status === 'executing');
    return active?.current_action;
  }, [daemons]);

  // Single daemon role label
  const singleDaemonRole = useMemo(() => {
    if (daemons.length === 1) {
      return daemons[0].role_label || daemons[0].role;
    }
    return null;
  }, [daemons]);

  // Collect assigned tools from daemons
  const toolPills = useMemo(() => {
    const tools = new Set<string>();
    for (const d of daemons) {
      if (d.assigned_tools) {
        for (const t of d.assigned_tools) tools.add(t);
      }
    }
    return Array.from(tools);
  }, [daemons]);

  // Check if there are any action buttons to show
  const hasActions =
    (isRunning && (onCancelTask || onViewLive)) ||
    ((task.status === 'failed' || task.status === 'cancelled') &&
      onRetryTask &&
      (task.manual_retry_count ?? 0) < 3) ||
    (isTerminal && onDelete);

  return (
    <div
      ref={cardRef}
      className={[
        styles.kanbanTaskCard,
        selected && styles.kanbanTaskCardSelected,
        isRunning && styles.kanbanTaskCardRunning,
        isDragging && styles.kanbanTaskCardDragging,
      ]
        .filter(Boolean)
        .join(' ')}
      onClick={() => onSelect(task.id)}
    >
      {/* Title — task prompt/goal */}
      <div className={styles.kanbanTaskCardPrompt}>{task.goal || task.prompt}</div>

      {/* Meta row — mode (non-terminal only) + daemon count + tools */}
      <div className={styles.kanbanTaskCardMeta}>
        {!isTerminal && (
          <Badge variant="default" icon={mode.icon}>
            {mode.label}
          </Badge>
        )}
        {!isTerminal && daemons.length > 0 && (
          <span className={styles.kanbanTaskDaemonCount}>
            <Bot size={10} /> {daemons.length}
          </span>
        )}
        {toolPills.length > 0 && (
          <>
            {toolPills.slice(0, MAX_TOOL_PILLS).map(t => (
              <span key={t} className={styles.kanbanTaskToolPill}>
                {t}
              </span>
            ))}
            {toolPills.length > MAX_TOOL_PILLS && (
              <span className={styles.kanbanTaskToolPill}>
                +{toolPills.length - MAX_TOOL_PILLS}
              </span>
            )}
          </>
        )}
      </div>

      {/* Daemon role label — only for active single-daemon tasks */}
      {!isTerminal && singleDaemonRole && (
        <div className={styles.kanbanTaskRoleLine}>{singleDaemonRole}</div>
      )}

      {/* Live action line when running */}
      {isRunning && liveAction && <div className={styles.kanbanTaskLiveAction}>{liveAction}</div>}

      {/* Error preview for failed tasks */}
      {task.status === 'failed' && task.error && (
        <div className={styles.kanbanTaskError}>{task.error}</div>
      )}

      {/* Result preview for completed tasks */}
      {task.status === 'completed' && task.result?.summary && (
        <div className={styles.kanbanTaskResultPreview}>{task.result.summary}</div>
      )}

      {isRunning && daemons.length > 0 && (
        <div className={styles.kanbanTaskProgress}>
          <div
            className={styles.kanbanTaskProgressFill}
            style={{ width: `${Math.round(avgProgress * 100)}%` }}
          />
        </div>
      )}

      {/* Footer — status dot + time + actions */}
      <div className={styles.kanbanTaskCardFooter}>
        <div className={styles.kanbanTaskCardTimeRow}>
          <span
            className={styles.kanbanTaskCardStatusDot}
            style={{ backgroundColor: statusDotColor[task.status] ?? 'rgba(255,255,255,0.3)' }}
          />
          <span className={styles.kanbanTaskCardTime}>{relativeTime(task.created_at)}</span>
        </div>
        {hasActions && (
          <div className={styles.kanbanTaskCardActions}>
            {isRunning && onViewLive && (
              <button
                className={styles.viewLiveBtn}
                onClick={e => {
                  e.stopPropagation();
                  onViewLive(task.id);
                }}
                title="View live"
              >
                <Eye size={12} />
              </button>
            )}
            {isRunning && onCancelTask && (
              <button
                className={styles.kanbanTaskStopBtn}
                onClick={e => {
                  e.stopPropagation();
                  onCancelTask(task.id);
                }}
                title="Stop task"
              >
                <Square size={10} fill="currentColor" />
              </button>
            )}
            {(task.status === 'failed' || task.status === 'cancelled') &&
              onRetryTask &&
              (task.manual_retry_count ?? 0) < 3 && (
                <button
                  className={styles.kanbanTaskRetryBtn}
                  onClick={e => {
                    e.stopPropagation();
                    onRetryTask(task.id);
                  }}
                  title={`Retry (${task.manual_retry_count ?? 0}/3)`}
                >
                  <RotateCcw size={12} />
                </button>
              )}
            {isTerminal && onDelete && (
              <button
                className={styles.kanbanTaskDeleteBtn}
                onClick={e => {
                  e.stopPropagation();
                  onDelete(task.id);
                }}
                title="Remove card"
              >
                <Trash2 size={12} />
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
});
