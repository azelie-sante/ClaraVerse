import { memo, useMemo, useState } from 'react';
import { Pencil, Search, X } from 'lucide-react';
import { KanbanBoard } from './KanbanBoard';
import { LiveViewModal } from './LiveViewModal';
import { useNexusStore } from '@/store/useNexusStore';
import type { NexusTaskStatus } from '@/types/nexus';
import styles from './Nexus.module.css';

interface CommandCenterViewProps {
  projectId?: string;
  onCancelDaemon: (id: string) => void;
  onStatusChange: (taskId: string, newStatus: NexusTaskStatus) => void;
  onSendFollowUp?: (content: string, taskId: string) => void;
  onRetryTask?: (taskId: string) => void;
  onCancelTask?: (taskId: string) => void;
  onEditProject?: () => void;
}

export const CommandCenterView = memo(function CommandCenterView({
  projectId,
  onCancelDaemon,
  onStatusChange,
  onSendFollowUp,
  onRetryTask,
  onCancelTask,
  onEditProject,
}: CommandCenterViewProps) {
  const [liveViewTaskId, setLiveViewTaskId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const projects = useNexusStore(s => s.projects);
  const tasks = useNexusStore(s => s.tasks);
  const project = projectId ? projects.find(p => p.id === projectId) : undefined;

  const statusCounts = useMemo(() => {
    if (!projectId) return null;
    const pt = tasks.filter(t => t.project_id === projectId);
    return {
      draft: pt.filter(t => t.status === 'draft').length,
      queued: pt.filter(t => t.status === 'pending').length,
      running: pt.filter(t => t.status === 'executing' || t.status === 'waiting_input').length,
      done: pt.filter(t => t.status === 'completed').length,
      failed: pt.filter(t => t.status === 'failed' || t.status === 'cancelled').length,
    };
  }, [tasks, projectId]);

  const hasAnyCounts =
    statusCounts &&
    (statusCounts.draft > 0 ||
      statusCounts.queued > 0 ||
      statusCounts.running > 0 ||
      statusCounts.done > 0 ||
      statusCounts.failed > 0);

  return (
    <div className={styles.main}>
      {project && (
        <div className={styles.projectHeader}>
          <span className={styles.projectHeaderDot} style={{ backgroundColor: project.color }} />
          <span className={styles.projectHeaderName}>{project.name}</span>
          <div className={styles.projectHeaderActions}>
            <div className={styles.projectHeaderSearch}>
              <Search size={12} className={styles.projectHeaderSearchIcon} />
              <input
                type="text"
                className={styles.projectHeaderSearchInput}
                placeholder="Search tasks..."
                value={searchQuery}
                onChange={e => setSearchQuery(e.target.value)}
              />
              {searchQuery && (
                <button
                  className={styles.projectHeaderSearchClear}
                  onClick={() => setSearchQuery('')}
                  title="Clear search"
                >
                  <X size={10} />
                </button>
              )}
            </div>
            {onEditProject && (
              <button className={styles.daemonIconBtn} onClick={onEditProject} title="Edit project">
                <Pencil size={14} />
              </button>
            )}
          </div>
          {project.system_instruction && (
            <span className={styles.projectHeaderInstruction} title={project.system_instruction}>
              {project.system_instruction.length > 80
                ? project.system_instruction.slice(0, 80) + '...'
                : project.system_instruction}
            </span>
          )}
          {hasAnyCounts && (
            <div className={styles.projectStatusBar}>
              {statusCounts.draft > 0 && (
                <span className={`${styles.projectStatusChip} ${styles.projectStatusChipDraft}`}>
                  {statusCounts.draft} draft
                </span>
              )}
              {statusCounts.queued > 0 && (
                <span className={`${styles.projectStatusChip} ${styles.projectStatusChipQueued}`}>
                  {statusCounts.queued} queued
                </span>
              )}
              {statusCounts.running > 0 && (
                <span className={`${styles.projectStatusChip} ${styles.projectStatusChipRunning}`}>
                  {statusCounts.running} running
                </span>
              )}
              {statusCounts.done > 0 && (
                <span className={`${styles.projectStatusChip} ${styles.projectStatusChipDone}`}>
                  {statusCounts.done} done
                </span>
              )}
              {statusCounts.failed > 0 && (
                <span className={`${styles.projectStatusChip} ${styles.projectStatusChipFailed}`}>
                  {statusCounts.failed} failed
                </span>
              )}
            </div>
          )}
        </div>
      )}
      <KanbanBoard
        projectId={projectId}
        projectName={project?.name}
        searchQuery={searchQuery}
        onCancelDaemon={onCancelDaemon}
        onStatusChange={onStatusChange}
        onViewLive={setLiveViewTaskId}
        onRetryTask={onRetryTask}
        onCancelTask={onCancelTask}
      />
      {liveViewTaskId && (
        <LiveViewModal
          taskId={liveViewTaskId}
          onClose={() => setLiveViewTaskId(null)}
          onSendFollowUp={onSendFollowUp}
        />
      )}
    </div>
  );
});
