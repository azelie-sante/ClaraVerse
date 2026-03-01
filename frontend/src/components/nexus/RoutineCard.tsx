import { memo, useState, useCallback } from 'react';
import {
  Pencil,
  Play,
  Trash2,
  Loader2,
  Send,
  HardDrive,
  CheckCircle2,
  XCircle,
  ChevronDown,
  ChevronUp,
  Clock,
} from 'lucide-react';
import { Badge } from '@/components/design-system';
import { fetchRoutineRuns, type Routine, type RoutineRun } from '@/services/clawService';
import styles from './Nexus.module.css';

interface RoutineCardProps {
  routine: Routine;
  triggering?: boolean;
  onEdit: (id: string) => void;
  onDelete: (id: string) => void;
  onTrigger: (id: string) => void;
  onToggle: (id: string, enabled: boolean) => void;
}

function cronToHuman(cron: string): string {
  const parts = cron.split(' ');
  if (parts.length !== 5) return cron;
  const [min, hour, , , dow] = parts;
  if (hour !== '*' && min !== '*') {
    const h = parseInt(hour, 10);
    const m = parseInt(min, 10);
    const time = `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}`;
    if (dow === '*') return `Daily at ${time}`;
    if (dow === '1-5') return `Weekdays at ${time}`;
    return `${time} (${cron})`;
  }
  if (hour.startsWith('*/')) return `Every ${hour.slice(2)}h`;
  if (min.startsWith('*/')) return `Every ${min.slice(2)}m`;
  return cron;
}

function timeAgo(dateStr?: string): string {
  if (!dateStr) return 'never';
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins}m ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  return `${Math.floor(hrs / 24)}d ago`;
}

const statusIcon: Record<string, React.ReactNode> = {
  completed: <CheckCircle2 size={11} style={{ color: 'var(--color-success, #22c55e)' }} />,
  failed: <XCircle size={11} style={{ color: 'var(--color-error, #ef4444)' }} />,
  executing: <Loader2 size={11} className={styles.spin} />,
  pending: <Clock size={11} style={{ opacity: 0.5 }} />,
};

export const RoutineCard = memo(function RoutineCard({
  routine,
  triggering,
  onEdit,
  onDelete,
  onTrigger,
  onToggle,
}: RoutineCardProps) {
  const [expanded, setExpanded] = useState(false);
  const [runs, setRuns] = useState<RoutineRun[] | null>(null);
  const [loadingRuns, setLoadingRuns] = useState(false);

  const toggleHistory = useCallback(async () => {
    if (expanded) {
      setExpanded(false);
      return;
    }
    setExpanded(true);
    if (runs === null) {
      setLoadingRuns(true);
      try {
        const data = await fetchRoutineRuns(routine.id, 10);
        setRuns(data);
      } catch {
        setRuns([]);
      } finally {
        setLoadingRuns(false);
      }
    }
  }, [expanded, runs, routine.id]);

  return (
    <div className={`${styles.routineCard} ${!routine.enabled ? styles.routineCardDisabled : ''}`}>
      <div className={styles.routineCardHeader}>
        <div className={styles.routineCardTitle}>
          <span className={styles.routineCardName}>{routine.name}</span>
          <Badge variant={routine.enabled ? 'success' : 'default'}>
            {routine.enabled ? 'Active' : 'Paused'}
          </Badge>
        </div>
        <div className={styles.routineCardActions}>
          <button
            className={styles.routineActionBtn}
            onClick={() => onTrigger(routine.id)}
            title="Run now"
            disabled={triggering}
          >
            {triggering ? <Loader2 size={13} className={styles.spin} /> : <Play size={13} />}
          </button>
          <button
            className={styles.routineActionBtn}
            onClick={() => onEdit(routine.id)}
            title="Edit"
          >
            <Pencil size={13} />
          </button>
          <button
            className={`${styles.routineActionBtn} ${styles.routineActionBtnDanger}`}
            onClick={() => onDelete(routine.id)}
            title="Delete"
          >
            <Trash2 size={13} />
          </button>
        </div>
      </div>

      <div className={styles.routineCardMeta}>
        <span>{cronToHuman(routine.cronExpression)}</span>
        <span className={styles.routineCardDot} />
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: '3px' }}>
          {routine.deliveryMethod === 'telegram' ? <Send size={10} /> : <HardDrive size={10} />}
          {routine.deliveryMethod === 'telegram' ? 'Telegram' : 'In-app'}
        </span>
        {routine.enabledTools?.length > 0 && (
          <>
            <span className={styles.routineCardDot} />
            <span>{routine.enabledTools.length} tools</span>
          </>
        )}
      </div>

      {routine.totalRuns > 0 && (
        <div className={styles.routineCardStats}>
          <span>{routine.totalRuns} runs</span>
          <span
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '2px',
              color: 'var(--color-success, #22c55e)',
            }}
          >
            <CheckCircle2 size={10} /> {routine.successfulRuns}
          </span>
          {routine.failedRuns > 0 && (
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '2px',
                color: 'var(--color-error, #ef4444)',
              }}
            >
              <XCircle size={10} /> {routine.failedRuns}
            </span>
          )}
          <span className={styles.routineCardDot} />
          <span>Last: {timeAgo(routine.lastRunAt)}</span>
        </div>
      )}

      {/* Run history toggle */}
      {routine.totalRuns > 0 && (
        <button className={styles.routineHistoryToggle} onClick={toggleHistory}>
          {expanded ? <ChevronUp size={12} /> : <ChevronDown size={12} />}
          {expanded ? 'Hide history' : 'View history'}
        </button>
      )}

      {/* Expandable run history */}
      {expanded && (
        <div className={styles.routineRunHistory}>
          {loadingRuns && (
            <div className={styles.routineRunLoading}>
              <Loader2 size={14} className={styles.spin} /> Loading...
            </div>
          )}
          {!loadingRuns && runs && runs.length === 0 && (
            <div className={styles.routineRunEmpty}>No runs recorded yet</div>
          )}
          {!loadingRuns &&
            runs &&
            runs.map(run => (
              <div key={run.id} className={styles.routineRunItem}>
                <div className={styles.routineRunItemHeader}>
                  {statusIcon[run.status] ?? statusIcon.pending}
                  <span className={styles.routineRunItemTime}>{timeAgo(run.created_at)}</span>
                  <Badge
                    variant={
                      run.status === 'completed'
                        ? 'success'
                        : run.status === 'failed'
                          ? 'error'
                          : 'default'
                    }
                  >
                    {run.status}
                  </Badge>
                </div>
                {run.summary && (
                  <div className={styles.routineRunSummary}>
                    {run.summary.length > 200 ? run.summary.slice(0, 200) + '...' : run.summary}
                  </div>
                )}
                {run.error && <div className={styles.routineRunError}>{run.error}</div>}
              </div>
            ))}
        </div>
      )}

      <div className={styles.routineCardToggle}>
        <label className={styles.routineToggleLabel}>
          <input
            type="checkbox"
            checked={routine.enabled}
            onChange={() => onToggle(routine.id, !routine.enabled)}
            className={styles.routineToggleInput}
          />
          <span className={styles.routineToggleSwitch} />
        </label>
      </div>
    </div>
  );
});
