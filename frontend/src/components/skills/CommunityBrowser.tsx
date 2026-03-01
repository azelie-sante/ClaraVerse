import { useEffect, useMemo, useState, useCallback } from 'react';
import {
  Download,
  Globe,
  ExternalLink,
  Search,
  Loader2,
  AlertCircle,
  type LucideIcon,
} from 'lucide-react';
import { Badge, SearchInput, Spinner } from '@/components/design-system';
import { useSkillStore } from '@/store/useSkillStore';

// Reuse icon mapping from SkillCard
import {
  Search as SearchIcon,
  Mail,
  Briefcase,
  BarChart3,
  PenTool,
  Zap,
  Code,
  Database,
} from 'lucide-react';

const iconMap: Record<string, LucideIcon> = {
  search: SearchIcon,
  mail: Mail,
  briefcase: Briefcase,
  'bar-chart-3': BarChart3,
  'pen-tool': PenTool,
  zap: Zap,
  code: Code,
  database: Database,
};

function getIcon(iconName: string): LucideIcon {
  return iconMap[iconName] || Globe;
}

export const CommunityBrowser = () => {
  const {
    communitySkills,
    communityLoading,
    communityError,
    fetchCommunitySkills,
    importFromGitHub,
    enableSkill,
    fetchSkills,
  } = useSkillStore();

  const [search, setSearch] = useState('');
  const [importing, setImporting] = useState<string | null>(null);
  const [importedNames, setImportedNames] = useState<Set<string>>(new Set());

  useEffect(() => {
    if (communitySkills.length === 0 && !communityLoading) {
      fetchCommunitySkills();
    }
  }, [communitySkills.length, communityLoading, fetchCommunitySkills]);

  const filtered = useMemo(() => {
    if (!search.trim()) return communitySkills;
    const q = search.toLowerCase();
    return communitySkills.filter(
      s =>
        s.name.toLowerCase().includes(q) ||
        s.description.toLowerCase().includes(q) ||
        s.category.toLowerCase().includes(q)
    );
  }, [communitySkills, search]);

  const handleImport = useCallback(
    async (entry: (typeof communitySkills)[0]) => {
      setImporting(entry.name);
      try {
        const skill = await importFromGitHub(entry.raw_url);
        await enableSkill(skill.id);
        setImportedNames(prev => new Set(prev).add(entry.name));
        // Refresh local skills list
        fetchSkills();
      } catch {
        // Error handled by store
      } finally {
        setImporting(null);
      }
    },
    [importFromGitHub, enableSkill, fetchSkills]
  );

  if (communityLoading && communitySkills.length === 0) {
    return (
      <div className="skills-loading">
        <Spinner size="md" />
        <span>Loading community skills from GitHub...</span>
      </div>
    );
  }

  if (communityError && communitySkills.length === 0) {
    return (
      <div className="skills-empty">
        <AlertCircle size={40} className="skills-empty__icon" />
        <p>{communityError}</p>
        <button
          className="skills-category-tab"
          onClick={() => fetchCommunitySkills()}
          style={{ marginTop: '0.5rem' }}
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div>
      <div className="community-header">
        <p className="community-subtitle">
          Skills from the{' '}
          <a
            href="https://github.com/anthropics/skills"
            target="_blank"
            rel="noopener noreferrer"
            className="community-link"
          >
            anthropics/skills <ExternalLink size={12} />
          </a>{' '}
          repository. Compatible with Claude Code, Codex CLI, and 27+ AI tools.
        </p>
        <div className="skills-search" style={{ maxWidth: 280 }}>
          <SearchInput
            value={search}
            onChange={setSearch}
            placeholder="Filter community skills..."
          />
        </div>
      </div>

      {filtered.length === 0 ? (
        <div className="skills-empty">
          <Search size={40} className="skills-empty__icon" />
          <p>No community skills match your filter</p>
        </div>
      ) : (
        <div className="skills-grid">
          {filtered.map(entry => {
            const Icon = getIcon(entry.icon);
            const isImporting = importing === entry.name;
            const isImported = importedNames.has(entry.name);

            return (
              <div key={entry.name} className="skill-card community-card">
                <div className="skill-card__header">
                  <div
                    className="skill-card__icon"
                    style={{
                      background: 'var(--color-primary-light, #e8f0fe)',
                      color: 'var(--color-primary)',
                    }}
                  >
                    <Icon size={20} />
                  </div>
                  <a
                    href={entry.repo_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="community-card__link"
                    title="View on GitHub"
                  >
                    <ExternalLink size={14} />
                  </a>
                </div>

                <div className="skill-card__body">
                  <h3 className="skill-card__name">{entry.name}</h3>
                  <p className="skill-card__description">{entry.description || 'No description'}</p>
                </div>

                <div className="skill-card__footer">
                  <Badge variant="info">{entry.category}</Badge>
                  {entry.author && (
                    <span className="skill-card__tool-count">by {entry.author}</span>
                  )}
                  <div style={{ marginLeft: 'auto' }}>
                    {isImported ? (
                      <Badge variant="success">Imported</Badge>
                    ) : (
                      <button
                        className="community-card__import-btn"
                        onClick={() => handleImport(entry)}
                        disabled={isImporting}
                      >
                        {isImporting ? (
                          <Loader2 size={14} className="community-spin" />
                        ) : (
                          <Download size={14} />
                        )}
                        {isImporting ? 'Importing...' : 'Import'}
                      </button>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
