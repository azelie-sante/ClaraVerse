import {
  Wrench,
  Server,
  Tag,
  Clock,
  User,
  Download,
  Copy,
  Check,
  type LucideIcon,
} from 'lucide-react';
import { useState, useCallback } from 'react';
import { Badge, Modal, Switch, Button } from '@/components/design-system';
import type { Skill } from '@/services/skillService';

// Reuse icon map from SkillCard
import {
  Search,
  Globe,
  Shield,
  BarChart3,
  Mail,
  MessageSquare,
  Send,
  Zap,
  Github,
  Layers,
  Calendar,
  Briefcase,
  Database,
  Code,
  PenTool,
  FileText,
  Image,
} from 'lucide-react';

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

interface SkillDetailModalProps {
  skill: Skill | null;
  isOpen: boolean;
  onClose: () => void;
  enabled: boolean;
  onToggle: (skillId: string, enabled: boolean) => void;
  onExport?: (skillId: string) => void;
}

export const SkillDetailModal = ({
  skill,
  isOpen,
  onClose,
  enabled,
  onToggle,
  onExport,
}: SkillDetailModalProps) => {
  const [toggling, setToggling] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleToggle = useCallback(
    async (checked: boolean) => {
      if (!skill) return;
      setToggling(true);
      try {
        await onToggle(skill.id, checked);
      } finally {
        setToggling(false);
      }
    },
    [skill, onToggle]
  );

  const handleCopyPrompt = useCallback(async () => {
    if (!skill?.system_prompt) return;
    await navigator.clipboard.writeText(skill.system_prompt);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [skill]);

  if (!skill) return null;

  const Icon = getIcon(skill.icon);
  const accentColor = categoryColors[skill.category] || 'var(--color-primary)';
  const tools = skill.required_tools ?? [];
  const keywords = skill.keywords ?? [];
  const servers = skill.preferred_servers ?? [];

  return (
    <Modal isOpen={isOpen} onClose={onClose} size="lg" title={null} showClose>
      <div className="skill-detail">
        {/* Hero */}
        <div className="skill-detail__hero">
          <div
            className="skill-detail__icon"
            style={{ background: accentColor + '18', color: accentColor }}
          >
            <Icon size={28} />
          </div>
          <div className="skill-detail__hero-info">
            <h2 className="skill-detail__name">{skill.name}</h2>
            <p className="skill-detail__description">{skill.description}</p>
          </div>
          <div className="skill-detail__toggle">
            <span className="skill-detail__toggle-label">{enabled ? 'Enabled' : 'Disabled'}</span>
            <Switch checked={enabled} onChange={handleToggle} disabled={toggling} />
          </div>
        </div>

        {/* Meta badges */}
        <div className="skill-detail__meta">
          <Badge variant={skill.mode === 'auto' ? 'accent' : 'default'}>
            {skill.mode === 'auto' ? 'Auto' : 'Manual'}
          </Badge>
          <Badge variant="info">{skill.category.replace(/-/g, ' ')}</Badge>
          {skill.is_builtin && <Badge variant="default">Built-in</Badge>}
          {skill.version && (
            <span className="skill-detail__meta-item">
              <Tag size={12} /> v{skill.version}
            </span>
          )}
          {skill.author_id && (
            <span className="skill-detail__meta-item">
              <User size={12} /> {skill.author_id}
            </span>
          )}
          {skill.created_at && (
            <span className="skill-detail__meta-item">
              <Clock size={12} /> {new Date(skill.created_at).toLocaleDateString()}
            </span>
          )}
        </div>

        {/* Tools */}
        {tools.length > 0 && (
          <div className="skill-detail__section">
            <h4 className="skill-detail__section-title">
              <Wrench size={14} /> Required Tools
            </h4>
            <div className="skill-detail__tags">
              {tools.map(tool => (
                <span key={tool} className="skill-detail__tag">
                  {tool}
                </span>
              ))}
            </div>
          </div>
        )}

        {/* Preferred Servers */}
        {servers.length > 0 && (
          <div className="skill-detail__section">
            <h4 className="skill-detail__section-title">
              <Server size={14} /> Preferred MCP Servers
            </h4>
            <div className="skill-detail__tags">
              {servers.map(s => (
                <span key={s} className="skill-detail__tag">
                  {s}
                </span>
              ))}
            </div>
          </div>
        )}

        {/* Keywords */}
        {keywords.length > 0 && (
          <div className="skill-detail__section">
            <h4 className="skill-detail__section-title">
              <Tag size={14} /> Keywords
            </h4>
            <div className="skill-detail__tags">
              {keywords.map(kw => (
                <span key={kw} className="skill-detail__tag skill-detail__tag--keyword">
                  {kw}
                </span>
              ))}
            </div>
          </div>
        )}

        {/* System Prompt */}
        {skill.system_prompt && (
          <div className="skill-detail__section">
            <div className="skill-detail__section-header">
              <h4 className="skill-detail__section-title">System Prompt</h4>
              <button
                className="skill-detail__copy-btn"
                onClick={handleCopyPrompt}
                title="Copy to clipboard"
              >
                {copied ? <Check size={14} /> : <Copy size={14} />}
                {copied ? 'Copied' : 'Copy'}
              </button>
            </div>
            <pre className="skill-detail__prompt">{skill.system_prompt}</pre>
          </div>
        )}

        {/* Actions */}
        {onExport && (
          <div className="skill-detail__actions">
            <Button variant="secondary" size="sm" onClick={() => onExport(skill.id)}>
              <Download size={14} />
              Export as SKILL.md
            </Button>
          </div>
        )}
      </div>
    </Modal>
  );
};
