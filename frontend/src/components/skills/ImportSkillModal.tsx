import { useState } from 'react';
import { Upload, Github } from 'lucide-react';
import { Modal, Button, Input, Textarea, Tabs } from '@/components/design-system';
import type { Skill } from '@/services/skillService';

interface ImportSkillModalProps {
  isOpen: boolean;
  onClose: () => void;
  onImported: (skill: Skill) => void;
  importFromSkillMD: (content: string) => Promise<Skill>;
  importFromGitHub: (url: string) => Promise<Skill>;
}

export const ImportSkillModal = ({
  isOpen,
  onClose,
  onImported,
  importFromSkillMD,
  importFromGitHub,
}: ImportSkillModalProps) => {
  const [activeTab, setActiveTab] = useState('paste');
  const [content, setContent] = useState('');
  const [githubUrl, setGithubUrl] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const resetForm = () => {
    setContent('');
    setGithubUrl('');
    setError(null);
  };

  const handleImportPaste = async () => {
    if (!content.trim()) {
      setError('Paste SKILL.md content');
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      const skill = await importFromSkillMD(content.trim());
      onImported(skill);
      resetForm();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Import failed');
    } finally {
      setSubmitting(false);
    }
  };

  const handleImportGitHub = async () => {
    if (!githubUrl.trim()) {
      setError('Enter a GitHub URL');
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      const skill = await importFromGitHub(githubUrl.trim());
      onImported(skill);
      resetForm();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Import failed');
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
          <Upload size={18} /> Import Skill
        </span>
      }
      size="lg"
    >
      <div className="import-skill-modal">
        <Tabs
          tabs={[
            { id: 'paste', label: 'Paste SKILL.md', icon: <Upload size={14} /> },
            { id: 'github', label: 'GitHub URL', icon: <Github size={14} /> },
          ]}
          activeTab={activeTab}
          onChange={setActiveTab}
        />

        {error && <div className="create-skill-error">{error}</div>}

        {activeTab === 'paste' && (
          <div className="import-skill-tab">
            <p className="import-skill-hint">
              Paste the contents of a SKILL.md file. This is the open standard format used by Claude
              Code, Codex CLI, Gemini CLI, and 27+ AI tools.
            </p>
            <Textarea
              value={content}
              onChange={e => setContent(e.target.value)}
              placeholder={`---\nname: my-skill\ndescription: What this skill does\nallowed-tools: Read Grep\n---\n\nYour instructions here...`}
              rows={12}
            />
            <Button
              variant="primary"
              onClick={handleImportPaste}
              disabled={submitting || !content.trim()}
              style={{ alignSelf: 'flex-end' }}
            >
              {submitting ? 'Importing...' : 'Import Skill'}
            </Button>
          </div>
        )}

        {activeTab === 'github' && (
          <div className="import-skill-tab">
            <p className="import-skill-hint">
              Enter a GitHub URL pointing to a skill directory (e.g.,{' '}
              <code>github.com/anthropics/skills/tree/main/pdf</code>).
            </p>
            <Input
              value={githubUrl}
              onChange={e => setGithubUrl(e.target.value)}
              placeholder="https://github.com/anthropics/skills/tree/main/pdf"
            />
            <Button
              variant="primary"
              onClick={handleImportGitHub}
              disabled={submitting || !githubUrl.trim()}
              style={{ alignSelf: 'flex-end' }}
            >
              {submitting ? 'Importing...' : 'Import from GitHub'}
            </Button>
          </div>
        )}
      </div>
    </Modal>
  );
};
