import { useState, useEffect, useMemo, useCallback } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ChevronLeft, Plus, Search, Upload, Layers, Globe } from 'lucide-react';
import { Button, SearchInput, Spinner, Tabs } from '@/components/design-system';
import {
  SkillCard,
  SkillDetailModal,
  ImportSkillModal,
  CommunityBrowser,
} from '@/components/skills';
import { useSkillStore } from '@/store/useSkillStore';
import type { Skill } from '@/services/skillService';
import './Skills.css';

const CATEGORIES = [
  { key: '', label: 'All' },
  { key: 'research', label: 'Research' },
  { key: 'communication', label: 'Communication' },
  { key: 'project-management', label: 'Project Mgmt' },
  { key: 'data', label: 'Data' },
  { key: 'content', label: 'Content' },
  { key: 'productivity', label: 'Productivity' },
  { key: 'sales', label: 'Sales' },
  { key: 'code', label: 'Code' },
  { key: 'database', label: 'Database' },
  { key: 'ecommerce', label: 'E-Commerce' },
  { key: 'writing', label: 'Writing' },
];

const PAGE_TABS = [
  { id: 'my-skills', label: 'My Skills', icon: <Layers size={14} /> },
  { id: 'community', label: 'Community', icon: <Globe size={14} /> },
];

export const Skills = () => {
  const navigate = useNavigate();
  const {
    skills,
    loading,
    enabledSkillIds,
    fetchSkills,
    fetchMySkills,
    enableSkill,
    disableSkill,
    importFromSkillMD,
    importFromGitHub,
    exportSkillMD,
  } = useSkillStore();

  const [activeTab, setActiveTab] = useState('my-skills');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('');
  const [importModalOpen, setImportModalOpen] = useState(false);
  const [detailSkill, setDetailSkill] = useState<Skill | null>(null);

  useEffect(() => {
    fetchSkills();
    fetchMySkills();
  }, [fetchSkills, fetchMySkills]);

  const handleToggle = useCallback(
    async (skillId: string, enabled: boolean) => {
      if (enabled) {
        await enableSkill(skillId);
      } else {
        await disableSkill(skillId);
      }
    },
    [enableSkill, disableSkill]
  );

  const handleExport = useCallback(
    async (skillId: string) => {
      try {
        const md = await exportSkillMD(skillId);
        await navigator.clipboard.writeText(md);
      } catch {
        // Silently fail — user can retry
      }
    },
    [exportSkillMD]
  );

  const handleImported = useCallback(
    async (skill: { id: string }) => {
      await enableSkill(skill.id);
      fetchSkills();
      fetchMySkills();
    },
    [enableSkill, fetchSkills, fetchMySkills]
  );

  const handleCardClick = useCallback((skill: Skill) => {
    setDetailSkill(skill);
  }, []);

  const filteredSkills = useMemo(() => {
    let result = skills;

    if (selectedCategory) {
      result = result.filter(s => s.category === selectedCategory);
    }

    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      result = result.filter(
        s =>
          s.name.toLowerCase().includes(q) ||
          s.description.toLowerCase().includes(q) ||
          s.keywords?.some(k => k.toLowerCase().includes(q))
      );
    }

    return result;
  }, [skills, selectedCategory, searchQuery]);

  // Group by category for display
  const grouped = useMemo(() => {
    if (selectedCategory) {
      return [{ category: selectedCategory, skills: filteredSkills }];
    }

    const map = new Map<string, typeof filteredSkills>();
    for (const skill of filteredSkills) {
      const list = map.get(skill.category) ?? [];
      list.push(skill);
      map.set(skill.category, list);
    }

    return Array.from(map.entries())
      .map(([category, skills]) => ({ category, skills }))
      .sort((a, b) => a.category.localeCompare(b.category));
  }, [filteredSkills, selectedCategory]);

  const categoryLabel = (key: string) =>
    CATEGORIES.find(c => c.key === key)?.label ?? key.replace(/-/g, ' ');

  const enabledCount = enabledSkillIds.size;

  return (
    <div className="skills-page">
      {/* Header */}
      <div className="skills-header">
        <div className="skills-header__top">
          <div className="skills-header__left">
            <Link to="/" className="skills-header__back">
              <ChevronLeft size={18} />
            </Link>
            <h1 className="skills-header__title">Skills</h1>
            <span className="skills-header__subtitle">
              {enabledCount} enabled &middot; {skills.length} available
            </span>
          </div>

          <div className="skills-header__actions">
            <Button variant="secondary" size="sm" onClick={() => setImportModalOpen(true)}>
              <Upload size={14} />
              Import
            </Button>
            <Button variant="primary" size="sm" onClick={() => navigate('/skills/new')}>
              <Plus size={14} />
              Create Skill
            </Button>
          </div>
        </div>

        <Tabs tabs={PAGE_TABS} activeTab={activeTab} onChange={setActiveTab} />

        {activeTab === 'my-skills' && (
          <div className="skills-toolbar">
            <div className="skills-search">
              <SearchInput
                value={searchQuery}
                onChange={setSearchQuery}
                placeholder="Search skills..."
              />
            </div>

            <div className="skills-categories">
              {CATEGORIES.map(cat => (
                <button
                  key={cat.key}
                  className={`skills-category-tab ${selectedCategory === cat.key ? 'active' : ''}`}
                  onClick={() => setSelectedCategory(cat.key)}
                >
                  {cat.label}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Content */}
      <div className="skills-content">
        {activeTab === 'my-skills' ? (
          <>
            {loading && skills.length === 0 ? (
              <div className="skills-loading">
                <Spinner size="md" />
                <span>Loading skills...</span>
              </div>
            ) : filteredSkills.length === 0 ? (
              <div className="skills-empty">
                <Search size={40} className="skills-empty__icon" />
                <p>No skills found matching your search</p>
              </div>
            ) : (
              grouped.map(group => (
                <div key={group.category} className="skills-section">
                  <h2 className="skills-section__title">
                    {categoryLabel(group.category)} ({group.skills.length})
                  </h2>
                  <div className="skills-grid">
                    {group.skills.map(skill => (
                      <SkillCard
                        key={skill.id}
                        skill={skill}
                        enabled={enabledSkillIds.has(skill.id)}
                        onToggle={handleToggle}
                        onExport={handleExport}
                        onClick={handleCardClick}
                      />
                    ))}
                  </div>
                </div>
              ))
            )}
          </>
        ) : (
          <CommunityBrowser />
        )}
      </div>

      {/* Modals */}
      <ImportSkillModal
        isOpen={importModalOpen}
        onClose={() => setImportModalOpen(false)}
        onImported={handleImported}
        importFromSkillMD={importFromSkillMD}
        importFromGitHub={importFromGitHub}
      />

      <SkillDetailModal
        skill={detailSkill}
        isOpen={detailSkill !== null}
        onClose={() => setDetailSkill(null)}
        enabled={detailSkill ? enabledSkillIds.has(detailSkill.id) : false}
        onToggle={handleToggle}
        onExport={handleExport}
      />
    </div>
  );
};
