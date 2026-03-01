import { memo, useRef, useState, useEffect, useCallback } from 'react';
import { ChevronDown, ChevronRight } from 'lucide-react';
import { dropTargetForElements } from '@atlaskit/pragmatic-drag-and-drop/element/adapter';
import { KanbanTaskCard } from './KanbanTaskCard';
import type { NexusTask, Daemon, NexusTaskStatus } from '@/types/nexus';
import styles from './Nexus.module.css';

/** Allowed status transitions for drag-and-drop */
const ALLOWED_DROPS: Record<string, NexusTaskStatus[]> = {
  draft: ['pending', 'cancelled'],
  pending: ['draft', 'executing', 'cancelled'],
  executing: ['cancelled'],
  waiting_input: ['cancelled'],
  completed: ['pending'],
  failed: ['pending'],
  cancelled: ['pending', 'draft'],
};

interface KanbanColumnProps {
  title: string;
  icon: React.ReactNode;
  accentColor: string;
  columnStatus: NexusTaskStatus;
  tasks: NexusTask[];
  daemons: Record<string, Partial<Daemon>>;
  selectedTaskId: string | null;
  onSelectTask: (id: string | null) => void;
  onDeleteTask?: (id: string) => void;
  onViewLive?: (taskId: string) => void;
  onRetryTask?: (taskId: string) => void;
  onCancelTask?: (taskId: string) => void;
  isMobile?: boolean;
  isExpanded?: boolean;
  onToggle?: () => void;
}

export const KanbanColumn = memo(function KanbanColumn({
  title,
  icon,
  accentColor,
  columnStatus,
  tasks,
  daemons,
  selectedTaskId,
  onSelectTask,
  onDeleteTask,
  onViewLive,
  onRetryTask,
  onCancelTask,
  isMobile,
  isExpanded = true,
  onToggle,
}: KanbanColumnProps) {
  const columnRef = useRef<HTMLDivElement>(null);
  const [isOver, setIsOver] = useState(false);

  useEffect(() => {
    const el = columnRef.current;
    if (!el) return;
    return dropTargetForElements({
      element: el,
      getData: () => ({ columnStatus }),
      canDrop: ({ source }) => {
        const sourceStatus = source.data.status as string;
        if (sourceStatus === columnStatus) return false;
        const allowed = ALLOWED_DROPS[sourceStatus] ?? [];
        return allowed.includes(columnStatus);
      },
      onDragEnter: () => setIsOver(true),
      onDragLeave: () => setIsOver(false),
      onDrop: () => setIsOver(false),
    });
  }, [columnStatus]);

  const daemonsForTask = useCallback(
    (taskId: string): Partial<Daemon>[] => {
      return Object.values(daemons).filter(d => d.task_id === taskId);
    },
    [daemons]
  );

  const showBody = isMobile ? isExpanded : true;

  return (
    <div
      ref={columnRef}
      className={[
        styles.kanbanColumn,
        isOver && styles.kanbanColumnDropTarget,
        isMobile && styles.kanbanColumnMobile,
      ]
        .filter(Boolean)
        .join(' ')}
    >
      <div
        className={styles.kanbanColumnHeader}
        onClick={isMobile ? onToggle : undefined}
        role={isMobile ? 'button' : undefined}
        aria-expanded={isMobile ? isExpanded : undefined}
      >
        <span className={styles.kanbanColumnDot} style={{ background: accentColor }} />
        {icon}
        <span>{title}</span>
        <span className={styles.kanbanColumnCount}>{tasks.length}</span>
        {isMobile && (
          <span className={styles.kanbanColumnChevron}>
            {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
          </span>
        )}
      </div>
      {showBody && (
        <div className={styles.kanbanColumnBody}>
          {tasks.length === 0 ? (
            <div className={styles.kanbanColumnEmpty}>No tasks</div>
          ) : (
            tasks.map(task => (
              <KanbanTaskCard
                key={task.id}
                task={task}
                daemons={daemonsForTask(task.id)}
                selected={selectedTaskId === task.id}
                onSelect={onSelectTask}
                onDelete={onDeleteTask}
                onViewLive={onViewLive}
                onRetryTask={onRetryTask}
                onCancelTask={onCancelTask}
              />
            ))
          )}
        </div>
      )}
    </div>
  );
});
