import { memo, useEffect, useState, useCallback } from 'react';
import { Plus, Bot } from 'lucide-react';
import { nexusService } from '@/services/nexusService';
import { useNexusStore } from '@/store/useNexusStore';
import { DaemonTemplateCard } from './DaemonTemplateCard';
import { DaemonTemplateBuilder } from './DaemonTemplateBuilder';
import type { DaemonTemplate } from '@/types/nexus';
import styles from './Nexus.module.css';

export const DaemonsView = memo(function DaemonsView() {
  const daemonTemplates = useNexusStore(s => s.daemonTemplates);
  const setDaemonTemplates = useNexusStore(s => s.setDaemonTemplates);
  const removeDaemonTemplate = useNexusStore(s => s.removeDaemonTemplate);
  const updateDaemonTemplate = useNexusStore(s => s.updateDaemonTemplate);

  const [showBuilder, setShowBuilder] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<DaemonTemplate | null>(null);

  useEffect(() => {
    nexusService.listDaemonTemplates().then(setDaemonTemplates).catch(console.error);
  }, [setDaemonTemplates]);

  const systemTemplates = daemonTemplates.filter(t => t.is_default);
  const userTemplates = daemonTemplates.filter(t => !t.is_default);

  const handleEdit = useCallback(
    (id: string) => {
      const tmpl = daemonTemplates.find(t => t.id === id);
      if (tmpl) {
        setEditingTemplate(tmpl);
        setShowBuilder(true);
      }
    },
    [daemonTemplates]
  );

  const handleDelete = useCallback(
    async (id: string) => {
      if (!window.confirm('Delete this daemon template? This cannot be undone.')) return;
      try {
        await nexusService.deleteDaemonTemplate(id);
        removeDaemonTemplate(id);
      } catch (err) {
        console.error('Failed to delete template:', err);
      }
    },
    [removeDaemonTemplate]
  );

  const handleToggle = useCallback(
    async (id: string, active: boolean) => {
      try {
        await nexusService.updateDaemonTemplate(id, { is_active: active });
        updateDaemonTemplate(id, { is_active: active });
      } catch (err) {
        console.error('Failed to toggle template:', err);
      }
    },
    [updateDaemonTemplate]
  );

  const handleCloseBuilder = useCallback(() => {
    setShowBuilder(false);
    setEditingTemplate(null);
    // Re-fetch to pick up any changes
    nexusService.listDaemonTemplates().then(setDaemonTemplates).catch(console.error);
  }, [setDaemonTemplates]);

  return (
    <div className={styles.daemonsView}>
      <div className={styles.daemonsHeader}>
        <div className={styles.routinesHeaderLeft}>
          <Bot size={20} />
          <h2 className={styles.routinesTitle}>Daemons</h2>
        </div>
        <div className={styles.routinesHeaderRight}>
          <button
            className={styles.routinesNewBtn}
            onClick={() => {
              setEditingTemplate(null);
              setShowBuilder(true);
            }}
          >
            <Plus size={14} />
            New Daemon
          </button>
        </div>
      </div>

      <div className={styles.daemonsBody}>
        {/* System Templates */}
        {systemTemplates.length > 0 && (
          <section className={styles.daemonsSection}>
            <h3 className={styles.daemonsSectionTitle}>System Daemons</h3>
            <div className={styles.daemonsGrid}>
              {systemTemplates.map(tmpl => (
                <DaemonTemplateCard key={tmpl.id} template={tmpl} onClick={handleEdit} />
              ))}
            </div>
          </section>
        )}

        {/* User Templates */}
        <section className={styles.daemonsSection}>
          <h3 className={styles.daemonsSectionTitle}>Your Daemons</h3>
          {userTemplates.length === 0 ? (
            <div className={styles.routinesEmpty}>
              <Bot size={32} style={{ opacity: 0.3 }} />
              <p>No custom daemons yet</p>
              <span>
                Create a daemon template to define reusable agent presets — specialized workers with
                their own tools, personality, and learned behaviors.
              </span>
              <button
                className={styles.routinesNewBtn}
                onClick={() => {
                  setEditingTemplate(null);
                  setShowBuilder(true);
                }}
              >
                <Plus size={14} />
                Create your first daemon
              </button>
            </div>
          ) : (
            <div className={styles.daemonsGrid}>
              {userTemplates.map(tmpl => (
                <DaemonTemplateCard
                  key={tmpl.id}
                  template={tmpl}
                  onClick={handleEdit}
                  onEdit={handleEdit}
                  onDelete={handleDelete}
                  onToggle={handleToggle}
                />
              ))}
            </div>
          )}
        </section>
      </div>

      {showBuilder && (
        <DaemonTemplateBuilder editingTemplate={editingTemplate} onClose={handleCloseBuilder} />
      )}
    </div>
  );
});
