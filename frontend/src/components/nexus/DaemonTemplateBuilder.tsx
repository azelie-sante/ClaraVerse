import { memo, useState, useCallback, useEffect, useMemo } from 'react';
import {
  X,
  Save,
  Copy,
  Code,
  Search,
  Globe,
  PenTool,
  BarChart3,
  Bot,
  Brain,
  Wrench,
  ChevronDown,
  ChevronRight,
} from 'lucide-react';
import { nexusService } from '@/services/nexusService';
import { useNexusStore } from '@/store/useNexusStore';
import { useClawStore } from '@/store/useClawStore';
import type { DaemonTemplate } from '@/types/nexus';
import styles from './Nexus.module.css';

interface DaemonTemplateBuilderProps {
  editingTemplate?: DaemonTemplate | null;
  onClose: () => void;
}

const ICON_OPTIONS = [
  { value: 'code', label: 'Code', Icon: Code },
  { value: 'search', label: 'Search', Icon: Search },
  { value: 'globe', label: 'Globe', Icon: Globe },
  { value: 'pen-tool', label: 'Pen', Icon: PenTool },
  { value: 'bar-chart-3', label: 'Chart', Icon: BarChart3 },
  { value: 'bot', label: 'Bot', Icon: Bot },
  { value: 'brain', label: 'Brain', Icon: Brain },
  { value: 'wrench', label: 'Wrench', Icon: Wrench },
];

const COLOR_OPTIONS = [
  '#2196F3',
  '#4CAF50',
  '#FF9800',
  '#9C27B0',
  '#F44336',
  '#00BCD4',
  '#FF5722',
  '#607D8B',
  '#E91E63',
  '#3F51B5',
];

const ROLE_OPTIONS = [
  { value: 'coder', label: 'Coder' },
  { value: 'researcher', label: 'Researcher' },
  { value: 'browser', label: 'Browser Agent' },
  { value: 'writer', label: 'Writer' },
  { value: 'analyst', label: 'Analyst' },
  { value: 'custom', label: 'Custom' },
];

export const DaemonTemplateBuilder = memo(function DaemonTemplateBuilder({
  editingTemplate,
  onClose,
}: DaemonTemplateBuilderProps) {
  const addDaemonTemplate = useNexusStore(s => s.addDaemonTemplate);
  const updateDaemonTemplateStore = useNexusStore(s => s.updateDaemonTemplate);
  const { toolCategories, toolsLoading, fetchTools } = useClawStore();

  const isSystemTemplate = editingTemplate?.is_default ?? false;

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [role, setRole] = useState('custom');
  const [roleLabel, setRoleLabel] = useState('Custom Agent');
  const [persona, setPersona] = useState('');
  const [instructions, setInstructions] = useState('');
  const [constraints, setConstraints] = useState('');
  const [outputFormat, setOutputFormat] = useState('');
  const [defaultTools, setDefaultTools] = useState<Set<string>>(new Set());
  const [icon, setIcon] = useState('bot');
  const [color, setColor] = useState('#2196F3');
  const [maxIterations, setMaxIterations] = useState(25);
  const [maxRetries, setMaxRetries] = useState(3);
  const [saving, setSaving] = useState(false);

  // Tool picker state
  const [toolSearch, setToolSearch] = useState('');
  const [expandedCategories, setExpandedCategories] = useState<Set<string>>(new Set());

  // Fetch tools on mount
  useEffect(() => {
    if (toolCategories.length === 0) fetchTools();
  }, [toolCategories.length, fetchTools]);

  useEffect(() => {
    if (editingTemplate) {
      setName(isSystemTemplate ? `${editingTemplate.name} (Copy)` : editingTemplate.name);
      setDescription(editingTemplate.description);
      setRole(editingTemplate.role);
      setRoleLabel(editingTemplate.role_label);
      setPersona(editingTemplate.persona);
      setInstructions(editingTemplate.instructions);
      setConstraints(editingTemplate.constraints);
      setOutputFormat(editingTemplate.output_format);
      setDefaultTools(new Set(editingTemplate.default_tools ?? []));
      setIcon(editingTemplate.icon);
      setColor(editingTemplate.color);
      setMaxIterations(editingTemplate.max_iterations);
      setMaxRetries(editingTemplate.max_retries);
    }
  }, [editingTemplate, isSystemTemplate]);

  const handleRoleChange = useCallback((newRole: string) => {
    setRole(newRole);
    const match = ROLE_OPTIONS.find(r => r.value === newRole);
    if (match && newRole !== 'custom') {
      setRoleLabel(match.label);
    }
  }, []);

  // Tool picker helpers
  const filteredCategories = useMemo(() => {
    if (!toolSearch.trim()) return toolCategories;
    const q = toolSearch.toLowerCase();
    return toolCategories
      .map(cat => ({
        ...cat,
        tools: cat.tools.filter(
          t =>
            t.display_name.toLowerCase().includes(q) ||
            t.name.toLowerCase().includes(q) ||
            t.description?.toLowerCase().includes(q)
        ),
      }))
      .filter(cat => cat.tools.length > 0);
  }, [toolCategories, toolSearch]);

  const toggleCategory = useCallback((catName: string) => {
    setExpandedCategories(prev => {
      const next = new Set(prev);
      if (next.has(catName)) next.delete(catName);
      else next.add(catName);
      return next;
    });
  }, []);

  const toggleTool = useCallback((toolName: string) => {
    setDefaultTools(prev => {
      const next = new Set(prev);
      if (next.has(toolName)) next.delete(toolName);
      else next.add(toolName);
      return next;
    });
  }, []);

  const handleSave = useCallback(async () => {
    if (!name.trim()) return;
    setSaving(true);
    try {
      const slug = name
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '_')
        .replace(/^_|_$/g, '');
      const payload: Partial<DaemonTemplate> = {
        name: name.trim(),
        slug,
        description: description.trim(),
        role,
        role_label: roleLabel.trim(),
        persona: persona.trim(),
        instructions: instructions.trim(),
        constraints: constraints.trim(),
        output_format: outputFormat.trim(),
        default_tools: Array.from(defaultTools),
        icon,
        color,
        max_iterations: maxIterations,
        max_retries: maxRetries,
      };

      if (editingTemplate && !isSystemTemplate) {
        // Edit existing user template
        await nexusService.updateDaemonTemplate(editingTemplate.id, payload);
        updateDaemonTemplateStore(editingTemplate.id, payload);
      } else {
        // Create new (or fork from system template)
        const created = await nexusService.createDaemonTemplate(payload);
        addDaemonTemplate(created);
      }
      onClose();
    } catch (err) {
      console.error('Failed to save daemon template:', err);
    } finally {
      setSaving(false);
    }
  }, [
    name,
    description,
    role,
    roleLabel,
    persona,
    instructions,
    constraints,
    outputFormat,
    defaultTools,
    icon,
    color,
    maxIterations,
    maxRetries,
    editingTemplate,
    isSystemTemplate,
    onClose,
    addDaemonTemplate,
    updateDaemonTemplateStore,
  ]);

  const title = isSystemTemplate
    ? 'Customize System Daemon'
    : editingTemplate
      ? 'Edit Daemon Template'
      : 'New Daemon Template';

  const saveLabel = isSystemTemplate ? 'Save as Copy' : editingTemplate ? 'Update' : 'Create';
  const SaveIcon = isSystemTemplate ? Copy : Save;

  return (
    <div className={styles.daemonBuilderOverlay} onClick={onClose}>
      <div className={styles.daemonBuilder} onClick={e => e.stopPropagation()}>
        <div className={styles.daemonBuilderHeader}>
          <h3>{title}</h3>
          <button className={styles.routineActionBtn} onClick={onClose}>
            <X size={16} />
          </button>
        </div>

        <div className={styles.daemonBuilderBody}>
          {/* Name & Description */}
          <div className={styles.daemonBuilderField}>
            <label>Name</label>
            <input
              type="text"
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="My Custom Daemon"
              className={styles.daemonBuilderInput}
            />
          </div>

          <div className={styles.daemonBuilderField}>
            <label>Description</label>
            <input
              type="text"
              value={description}
              onChange={e => setDescription(e.target.value)}
              placeholder="What this daemon specializes in..."
              className={styles.daemonBuilderInput}
            />
          </div>

          {/* Role & Role Label */}
          <div className={styles.daemonBuilderRow}>
            <div className={styles.daemonBuilderField}>
              <label>Role</label>
              <select
                value={role}
                onChange={e => handleRoleChange(e.target.value)}
                className={styles.daemonBuilderInput}
              >
                {ROLE_OPTIONS.map(r => (
                  <option key={r.value} value={r.value}>
                    {r.label}
                  </option>
                ))}
              </select>
            </div>
            <div className={styles.daemonBuilderField}>
              <label>Role Label</label>
              <input
                type="text"
                value={roleLabel}
                onChange={e => setRoleLabel(e.target.value)}
                placeholder="Display label"
                className={styles.daemonBuilderInput}
              />
            </div>
          </div>

          {/* Persona */}
          <div className={styles.daemonBuilderField}>
            <label>Persona / System Prompt</label>
            <textarea
              value={persona}
              onChange={e => setPersona(e.target.value)}
              placeholder="You are an expert..."
              className={styles.daemonBuilderTextarea}
              rows={3}
            />
          </div>

          {/* Instructions */}
          <div className={styles.daemonBuilderField}>
            <label>Instructions (Workflow)</label>
            <textarea
              value={instructions}
              onChange={e => setInstructions(e.target.value)}
              placeholder="Step-by-step workflow..."
              className={styles.daemonBuilderTextarea}
              rows={4}
            />
          </div>

          {/* Constraints */}
          <div className={styles.daemonBuilderField}>
            <label>Constraints</label>
            <textarea
              value={constraints}
              onChange={e => setConstraints(e.target.value)}
              placeholder="Rules and guardrails..."
              className={styles.daemonBuilderTextarea}
              rows={3}
            />
          </div>

          {/* Output Format */}
          <div className={styles.daemonBuilderField}>
            <label>Output Format</label>
            <textarea
              value={outputFormat}
              onChange={e => setOutputFormat(e.target.value)}
              placeholder="Expected output structure..."
              className={styles.daemonBuilderTextarea}
              rows={2}
            />
          </div>

          {/* Tools — real tool picker */}
          <div className={styles.daemonBuilderField}>
            <label>Default Tools {defaultTools.size > 0 && `(${defaultTools.size})`}</label>
            <div className={styles.toolPicker}>
              <div className={styles.toolSearchRow}>
                <Search size={12} />
                <input
                  className={styles.toolSearchInput}
                  value={toolSearch}
                  onChange={e => setToolSearch(e.target.value)}
                  placeholder="Search tools..."
                />
              </div>
              <div className={styles.toolList}>
                {toolsLoading ? (
                  <div className={styles.toolEmpty}>Loading tools...</div>
                ) : filteredCategories.length === 0 ? (
                  <div className={styles.toolEmpty}>No tools found</div>
                ) : (
                  filteredCategories.map(cat => (
                    <div key={cat.name}>
                      <button
                        className={styles.toolCategoryBtn}
                        onClick={() => toggleCategory(cat.name)}
                      >
                        {expandedCategories.has(cat.name) ? (
                          <ChevronDown size={12} />
                        ) : (
                          <ChevronRight size={12} />
                        )}
                        <span>{cat.name}</span>
                        <span className={styles.toolCategoryCount}>{cat.tools.length}</span>
                      </button>
                      {expandedCategories.has(cat.name) &&
                        cat.tools.map(tool => (
                          <label key={tool.name} className={styles.toolItem}>
                            <input
                              type="checkbox"
                              checked={defaultTools.has(tool.name)}
                              onChange={() => toggleTool(tool.name)}
                            />
                            <span className={styles.toolItemName}>
                              {tool.display_name || tool.name}
                            </span>
                          </label>
                        ))}
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>

          {/* Icon & Color */}
          <div className={styles.daemonBuilderRow}>
            <div className={styles.daemonBuilderField}>
              <label>Icon</label>
              <div className={styles.daemonBuilderChips}>
                {ICON_OPTIONS.map(({ value, Icon }) => (
                  <button
                    key={value}
                    className={`${styles.daemonIconBtn} ${icon === value ? styles.daemonIconBtnActive : ''}`}
                    onClick={() => setIcon(value)}
                    title={value}
                  >
                    <Icon size={16} />
                  </button>
                ))}
              </div>
            </div>
            <div className={styles.daemonBuilderField}>
              <label>Color</label>
              <div className={styles.daemonBuilderChips}>
                {COLOR_OPTIONS.map(c => (
                  <button
                    key={c}
                    className={`${styles.daemonColorBtn} ${color === c ? styles.daemonColorBtnActive : ''}`}
                    style={{ background: c }}
                    onClick={() => setColor(c)}
                  />
                ))}
              </div>
            </div>
          </div>

          {/* Behavior */}
          <div className={styles.daemonBuilderRow}>
            <div className={styles.daemonBuilderField}>
              <label>Max Iterations</label>
              <input
                type="number"
                value={maxIterations}
                onChange={e => setMaxIterations(Number(e.target.value))}
                min={1}
                max={50}
                className={styles.daemonBuilderInput}
              />
            </div>
            <div className={styles.daemonBuilderField}>
              <label>Max Retries</label>
              <input
                type="number"
                value={maxRetries}
                onChange={e => setMaxRetries(Number(e.target.value))}
                min={0}
                max={10}
                className={styles.daemonBuilderInput}
              />
            </div>
          </div>
        </div>

        <div className={styles.daemonBuilderFooter}>
          <button className={styles.daemonBuilderCancelBtn} onClick={onClose}>
            Cancel
          </button>
          <button
            className={styles.daemonBuilderSaveBtn}
            onClick={handleSave}
            disabled={saving || !name.trim()}
          >
            <SaveIcon size={14} />
            {saving ? 'Saving...' : saveLabel}
          </button>
        </div>
      </div>
    </div>
  );
});
