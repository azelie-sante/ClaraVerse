import { useState } from 'react';
import { Sparkles } from 'lucide-react';
import { Modal, Button, Input, Textarea, Select } from '@/components/design-system';
import type { CreateSkillRequest } from '@/services/skillService';

const CATEGORY_OPTIONS = [
  { value: 'research', label: 'Research & Web' },
  { value: 'communication', label: 'Communication' },
  { value: 'project-management', label: 'Project Management' },
  { value: 'data', label: 'Data & Analytics' },
  { value: 'content', label: 'Content Creation' },
  { value: 'productivity', label: 'Productivity' },
  { value: 'sales', label: 'Sales & CRM' },
  { value: 'code', label: 'Code & DevOps' },
  { value: 'database', label: 'Database & Storage' },
];

const MODE_OPTIONS = [
  { value: 'auto', label: 'Auto (route messages automatically)' },
  { value: 'manual', label: 'Manual (activate explicitly)' },
];

const ICON_OPTIONS = [
  { value: 'search', label: 'Search' },
  { value: 'globe', label: 'Globe' },
  { value: 'mail', label: 'Mail' },
  { value: 'message-square', label: 'Message' },
  { value: 'code', label: 'Code' },
  { value: 'database', label: 'Database' },
  { value: 'bar-chart-3', label: 'Chart' },
  { value: 'pen-tool', label: 'Pen' },
  { value: 'calendar', label: 'Calendar' },
  { value: 'briefcase', label: 'Briefcase' },
  { value: 'github', label: 'GitHub' },
  { value: 'zap', label: 'Zap' },
  { value: 'shield', label: 'Shield' },
  { value: 'layers', label: 'Layers' },
  { value: 'send', label: 'Send' },
  { value: 'image', label: 'Image' },
  { value: 'file-text', label: 'File' },
  { value: 'server', label: 'Server' },
  { value: 'wrench', label: 'Wrench' },
];

interface CreateSkillModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: CreateSkillRequest) => Promise<void>;
}

export const CreateSkillModal = ({ isOpen, onClose, onSubmit }: CreateSkillModalProps) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [icon, setIcon] = useState('zap');
  const [category, setCategory] = useState('research');
  const [systemPrompt, setSystemPrompt] = useState('');
  const [keywords, setKeywords] = useState('');
  const [triggerPatterns, setTriggerPatterns] = useState('');
  const [mode, setMode] = useState<'auto' | 'manual'>('auto');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const resetForm = () => {
    setName('');
    setDescription('');
    setIcon('zap');
    setCategory('research');
    setSystemPrompt('');
    setKeywords('');
    setTriggerPatterns('');
    setMode('auto');
    setError(null);
  };

  const handleSubmit = async () => {
    if (!name.trim() || !systemPrompt.trim()) {
      setError('Name and system prompt are required');
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      await onSubmit({
        name: name.trim(),
        description: description.trim(),
        icon,
        category,
        system_prompt: systemPrompt.trim(),
        required_tools: [],
        preferred_servers: [],
        keywords: keywords
          .split(',')
          .map(k => k.trim())
          .filter(Boolean),
        trigger_patterns: triggerPatterns
          .split(',')
          .map(t => t.trim())
          .filter(Boolean),
        mode,
      });
      resetForm();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create skill');
    } finally {
      setSubmitting(false);
    }
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title={
        <span style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <Sparkles size={18} /> Create Custom Skill
        </span>
      }
      size="lg"
      footer={
        <div style={{ display: 'flex', gap: '0.75rem', justifyContent: 'flex-end' }}>
          <Button variant="ghost" onClick={handleClose} disabled={submitting}>
            Cancel
          </Button>
          <Button variant="primary" onClick={handleSubmit} disabled={submitting}>
            {submitting ? 'Creating...' : 'Create Skill'}
          </Button>
        </div>
      }
    >
      <div className="create-skill-form">
        {error && <div className="create-skill-error">{error}</div>}

        <div className="create-skill-row">
          <div className="create-skill-field" style={{ flex: 1 }}>
            <label className="create-skill-label">Name *</label>
            <Input
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="e.g., Code Reviewer"
            />
          </div>
          <div className="create-skill-field" style={{ width: '140px' }}>
            <label className="create-skill-label">Icon</label>
            <Select options={ICON_OPTIONS} value={icon} onChange={setIcon} />
          </div>
        </div>

        <div className="create-skill-field">
          <label className="create-skill-label">Description</label>
          <Input
            value={description}
            onChange={e => setDescription(e.target.value)}
            placeholder="What does this skill do?"
          />
        </div>

        <div className="create-skill-row">
          <div className="create-skill-field" style={{ flex: 1 }}>
            <label className="create-skill-label">Category</label>
            <Select options={CATEGORY_OPTIONS} value={category} onChange={setCategory} />
          </div>
          <div className="create-skill-field" style={{ flex: 1 }}>
            <label className="create-skill-label">Mode</label>
            <Select
              options={MODE_OPTIONS}
              value={mode}
              onChange={v => setMode(v as 'auto' | 'manual')}
            />
          </div>
        </div>

        <div className="create-skill-field">
          <label className="create-skill-label">System Prompt *</label>
          <Textarea
            value={systemPrompt}
            onChange={e => setSystemPrompt(e.target.value)}
            placeholder="Instructions the AI will follow when this skill is active..."
            rows={6}
          />
        </div>

        <div className="create-skill-field">
          <label className="create-skill-label">Keywords</label>
          <Input
            value={keywords}
            onChange={e => setKeywords(e.target.value)}
            placeholder="Comma-separated: review, code, bug, lint"
          />
          <span className="create-skill-hint">Words that trigger this skill in auto mode</span>
        </div>

        <div className="create-skill-field">
          <label className="create-skill-label">Trigger Patterns</label>
          <Input
            value={triggerPatterns}
            onChange={e => setTriggerPatterns(e.target.value)}
            placeholder="Comma-separated: review this, check my code"
          />
          <span className="create-skill-hint">Phrase prefixes for higher-priority matching</span>
        </div>
      </div>
    </Modal>
  );
};
