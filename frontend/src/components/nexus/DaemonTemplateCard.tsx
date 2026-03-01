import { memo } from 'react';
import {
  Pencil,
  Trash2,
  CheckCircle2,
  XCircle,
  Code,
  Search,
  Globe,
  PenTool,
  BarChart3,
  Bot,
  Brain,
} from 'lucide-react';
import { Badge } from '@/components/design-system';
import type { DaemonTemplate } from '@/types/nexus';
import styles from './Nexus.module.css';

interface DaemonTemplateCardProps {
  template: DaemonTemplate;
  onClick?: (id: string) => void;
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
  onToggle?: (id: string, active: boolean) => void;
}

const ICON_MAP: Record<string, React.ComponentType<{ size?: number }>> = {
  code: Code,
  search: Search,
  globe: Globe,
  'pen-tool': PenTool,
  'bar-chart-3': BarChart3,
  bot: Bot,
  brain: Brain,
};

export const DaemonTemplateCard = memo(function DaemonTemplateCard({
  template,
  onClick,
  onEdit,
  onDelete,
  onToggle,
}: DaemonTemplateCardProps) {
  const IconComponent = ICON_MAP[template.icon] ?? Bot;
  const isSystem = template.is_default;
  const { stats } = template;
  const learningCount = template.learnings?.length ?? 0;

  return (
    <div
      className={`${styles.daemonCard} ${!template.is_active ? styles.daemonCardDisabled : ''} ${onClick ? styles.daemonCardClickable : ''}`}
      onClick={() => onClick?.(template.id)}
    >
      <div className={styles.daemonCardHeader}>
        <div className={styles.daemonCardIcon} style={{ color: template.color }}>
          <IconComponent size={20} />
        </div>
        <div className={styles.daemonCardInfo}>
          <span className={styles.daemonCardName}>{template.name}</span>
          <span className={styles.daemonCardRole}>{template.role_label}</span>
        </div>
        {!isSystem && (
          <div className={styles.daemonCardActions}>
            {onEdit && (
              <button
                className={styles.routineActionBtn}
                onClick={e => {
                  e.stopPropagation();
                  onEdit(template.id);
                }}
                title="Edit"
              >
                <Pencil size={13} />
              </button>
            )}
            {onDelete && (
              <button
                className={`${styles.routineActionBtn} ${styles.routineActionBtnDanger}`}
                onClick={e => {
                  e.stopPropagation();
                  onDelete(template.id);
                }}
                title="Delete"
              >
                <Trash2 size={13} />
              </button>
            )}
          </div>
        )}
      </div>

      <p className={styles.daemonCardDesc}>{template.description}</p>

      <div className={styles.daemonCardMeta}>
        <span>{(template.default_tools ?? []).length} tools</span>
        <span className={styles.routineCardDot} />
        <span>max {template.max_iterations} iter</span>
        {learningCount > 0 && (
          <>
            <span className={styles.routineCardDot} />
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: '3px' }}>
              <Brain size={10} /> {learningCount} learnings
            </span>
          </>
        )}
      </div>

      {stats.total_runs > 0 && (
        <div className={styles.routineCardStats}>
          <span>{stats.total_runs} runs</span>
          <span
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '2px',
              color: 'var(--color-success, #22c55e)',
            }}
          >
            <CheckCircle2 size={10} /> {stats.successful_runs}
          </span>
          {stats.failed_runs > 0 && (
            <span
              style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '2px',
                color: 'var(--color-error, #ef4444)',
              }}
            >
              <XCircle size={10} /> {stats.failed_runs}
            </span>
          )}
          <span className={styles.routineCardDot} />
          <span>~{stats.avg_iterations.toFixed(1)} avg iter</span>
        </div>
      )}

      {isSystem && (
        <Badge variant="default" className={styles.daemonSystemBadge}>
          System
        </Badge>
      )}

      {!isSystem && onToggle && (
        <div className={styles.routineCardToggle}>
          <label className={styles.routineToggleLabel}>
            <input
              type="checkbox"
              checked={template.is_active}
              onChange={() => onToggle(template.id, !template.is_active)}
              onClick={e => e.stopPropagation()}
              className={styles.routineToggleInput}
            />
            <span className={styles.routineToggleSwitch} />
          </label>
        </div>
      )}
    </div>
  );
});
