import { useState } from 'react';
import {
  Search,
  Mail,
  Github,
  Database,
  BarChart3,
  PenTool,
  Calendar,
  MessageSquare,
  Code,
  Globe,
  Zap,
  Shield,
  Briefcase,
  FileText,
  Image,
  Send,
  Layers,
  Server,
  Wrench,
  Download,
  type LucideIcon,
} from 'lucide-react';
import { Badge, Switch, Tooltip } from '@/components/design-system';
import type { Skill } from '@/services/skillService';

// Map icon names to lucide components
const iconMap: Record<string, LucideIcon> = {
  search: Search,
  globe: Globe,
  shield: Shield,
  'bar-chart-3': BarChart3,
  mail: Mail,
  'message-square': MessageSquare,
  send: Send,
  zap: Zap,
  github: Github,
  layers: Layers,
  calendar: Calendar,
  briefcase: Briefcase,
  database: Database,
  code: Code,
  'pen-tool': PenTool,
  'file-text': FileText,
  image: Image,
  server: Server,
  wrench: Wrench,
};

function getIcon(iconName: string): LucideIcon {
  return iconMap[iconName] || Zap;
}

const categoryColors: Record<string, string> = {
  research: 'var(--color-info, #007aff)',
  communication: 'var(--color-success, #30d158)',
  'project-management': 'var(--color-warning, #ff9f0a)',
  data: 'var(--color-accent, #bf5af2)',
  content: 'var(--color-primary)',
  productivity: 'var(--color-info, #007aff)',
  sales: 'var(--color-warning, #ff9f0a)',
  code: 'var(--color-success, #30d158)',
  database: 'var(--color-accent, #bf5af2)',
};

interface SkillCardProps {
  skill: Skill;
  enabled: boolean;
  onToggle: (skillId: string, enabled: boolean) => void;
  onExport?: (skillId: string) => void;
  onClick?: (skill: Skill) => void;
}

export const SkillCard = ({ skill, enabled, onToggle, onExport, onClick }: SkillCardProps) => {
  const [toggling, setToggling] = useState(false);
  const Icon = getIcon(skill.icon);
  const accentColor = categoryColors[skill.category] || 'var(--color-primary)';

  const handleToggle = async (checked: boolean) => {
    setToggling(true);
    try {
      await onToggle(skill.id, checked);
    } finally {
      setToggling(false);
    }
  };

  return (
    <div className={`skill-card ${enabled ? 'skill-card--enabled' : ''}`}>
      <div className="skill-card__header">
        <div
          className="skill-card__icon"
          style={{ background: accentColor + '18', color: accentColor }}
        >
          <Icon size={22} />
        </div>
        <div className="skill-card__toggle">
          <Switch checked={enabled} onChange={handleToggle} disabled={toggling} size="sm" />
        </div>
      </div>

      <div
        className="skill-card__body"
        onClick={onClick ? () => onClick(skill) : undefined}
        role={onClick ? 'button' : undefined}
        tabIndex={onClick ? 0 : undefined}
        onKeyDown={
          onClick
            ? (e: React.KeyboardEvent) => {
                if (e.key === 'Enter' || e.key === ' ') onClick(skill);
              }
            : undefined
        }
      >
        <h3 className="skill-card__name">{skill.name}</h3>
        <p className="skill-card__description">{skill.description}</p>
      </div>

      <div className="skill-card__footer">
        <Badge variant={skill.mode === 'auto' ? 'accent' : 'default'}>
          {skill.mode === 'auto' ? 'Auto' : 'Manual'}
        </Badge>
        {skill.required_tools?.length > 0 && (
          <span className="skill-card__tool-count">
            <Wrench size={12} />
            {skill.required_tools.length} tool{skill.required_tools.length !== 1 ? 's' : ''}
          </span>
        )}
        {onExport && (
          <Tooltip content="Export as SKILL.md">
            <button
              className="skill-card__export-btn"
              onClick={() => onExport(skill.id)}
              aria-label="Export skill"
            >
              <Download size={12} />
            </button>
          </Tooltip>
        )}
      </div>
    </div>
  );
};
