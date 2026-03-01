import { memo, useEffect, useState, useCallback } from 'react';
import { Plus, Wifi, WifiOff, Send, Monitor, CalendarClock } from 'lucide-react';
import { useClawStore } from '@/store/useClawStore';
import { RoutineBuilder } from '@/components/claras-claw/RoutineBuilder';
import { Badge } from '@/components/design-system';
import { RoutineCard } from './RoutineCard';
import styles from './Nexus.module.css';

export const RoutinesView = memo(function RoutinesView() {
  const {
    routines,
    telegramConnected,
    mcpConnected,
    fetchRoutines,
    fetchStatus,
    fetchTools,
    updateRoutine,
    deleteRoutine,
    triggerRoutine,
  } = useClawStore();

  const [showBuilder, setShowBuilder] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [triggeringId, setTriggeringId] = useState<string | null>(null);

  useEffect(() => {
    fetchRoutines();
    fetchStatus();
    fetchTools();
  }, [fetchRoutines, fetchStatus, fetchTools]);

  const handleEdit = useCallback((id: string) => {
    setEditingId(id);
    setShowBuilder(true);
  }, []);

  const handleDelete = useCallback(
    async (id: string) => {
      if (!window.confirm('Delete this routine? This cannot be undone.')) return;
      await deleteRoutine(id);
    },
    [deleteRoutine]
  );

  const handleTrigger = useCallback(
    async (id: string) => {
      setTriggeringId(id);
      try {
        await triggerRoutine(id);
      } finally {
        setTriggeringId(null);
      }
    },
    [triggerRoutine]
  );

  const handleToggle = useCallback(
    async (id: string, enabled: boolean) => {
      await updateRoutine(id, { enabled });
    },
    [updateRoutine]
  );

  const handleCloseBuilder = useCallback(() => {
    setShowBuilder(false);
    setEditingId(null);
    fetchRoutines();
  }, [fetchRoutines]);

  return (
    <div className={styles.routinesView}>
      <div className={styles.routinesHeader}>
        <div className={styles.routinesHeaderLeft}>
          <CalendarClock size={20} />
          <h2 className={styles.routinesTitle}>Routines</h2>
        </div>
        <div className={styles.routinesHeaderRight}>
          <Badge
            variant={mcpConnected ? 'success' : 'default'}
            icon={mcpConnected ? <Monitor size={10} /> : <WifiOff size={10} />}
          >
            MCP
          </Badge>
          <Badge
            variant={telegramConnected ? 'success' : 'default'}
            icon={telegramConnected ? <Wifi size={10} /> : <WifiOff size={10} />}
          >
            <Send size={9} /> Telegram
          </Badge>
          <button
            className={styles.routinesNewBtn}
            onClick={() => {
              setEditingId(null);
              setShowBuilder(true);
            }}
          >
            <Plus size={14} />
            New Routine
          </button>
        </div>
      </div>

      <div className={styles.routinesList}>
        {routines.length === 0 ? (
          <div className={styles.routinesEmpty}>
            <CalendarClock size={32} style={{ opacity: 0.3 }} />
            <p>No routines yet</p>
            <span>Create a routine to run tasks on a schedule — daily briefings, weekly reports, automated monitoring.</span>
            <button
              className={styles.routinesNewBtn}
              onClick={() => {
                setEditingId(null);
                setShowBuilder(true);
              }}
            >
              <Plus size={14} />
              Create your first routine
            </button>
          </div>
        ) : (
          routines.map((routine) => (
            <RoutineCard
              key={routine.id}
              routine={routine}
              triggering={triggeringId === routine.id}
              onEdit={handleEdit}
              onDelete={handleDelete}
              onTrigger={handleTrigger}
              onToggle={handleToggle}
            />
          ))
        )}
      </div>

      {showBuilder && (
        <RoutineBuilder editingId={editingId} onClose={handleCloseBuilder} />
      )}
    </div>
  );
});
