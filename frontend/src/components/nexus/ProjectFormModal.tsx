import { memo, useState, useCallback } from 'react';
import {
  X,
  Trash2,
  Folder,
  Globe,
  Code,
  BookOpen,
  ShoppingCart,
  Briefcase,
  Heart,
  Star,
  Rocket,
  Lightbulb,
  Music,
  Camera,
  Gamepad2,
  Palette,
  Cpu,
  Database,
  Film,
  FlaskConical,
  GraduationCap,
  Hammer,
  Home,
  Layers,
  MessageCircle,
  Shield,
  Terminal,
  Wrench,
  Zap,
  Bug,
  Cloud,
  Compass,
  Crown,
  FileText,
  Gift,
  Headphones,
  Map,
  Megaphone,
  Puzzle,
  Target,
  Truck,
  Users,
  Wallet,
} from 'lucide-react';
import type { NexusProject } from '@/types/nexus';
import styles from './Nexus.module.css';

const ICON_OPTIONS = [
  { name: 'folder', Icon: Folder },
  { name: 'globe', Icon: Globe },
  { name: 'code', Icon: Code },
  { name: 'terminal', Icon: Terminal },
  { name: 'cpu', Icon: Cpu },
  { name: 'database', Icon: Database },
  { name: 'layers', Icon: Layers },
  { name: 'bug', Icon: Bug },
  { name: 'cloud', Icon: Cloud },
  { name: 'shield', Icon: Shield },
  { name: 'zap', Icon: Zap },
  { name: 'wrench', Icon: Wrench },
  { name: 'hammer', Icon: Hammer },
  { name: 'flask-conical', Icon: FlaskConical },
  { name: 'rocket', Icon: Rocket },
  { name: 'lightbulb', Icon: Lightbulb },
  { name: 'target', Icon: Target },
  { name: 'compass', Icon: Compass },
  { name: 'book-open', Icon: BookOpen },
  { name: 'graduation-cap', Icon: GraduationCap },
  { name: 'file-text', Icon: FileText },
  { name: 'briefcase', Icon: Briefcase },
  { name: 'wallet', Icon: Wallet },
  { name: 'shopping-cart', Icon: ShoppingCart },
  { name: 'truck', Icon: Truck },
  { name: 'megaphone', Icon: Megaphone },
  { name: 'users', Icon: Users },
  { name: 'message-circle', Icon: MessageCircle },
  { name: 'palette', Icon: Palette },
  { name: 'camera', Icon: Camera },
  { name: 'film', Icon: Film },
  { name: 'music', Icon: Music },
  { name: 'headphones', Icon: Headphones },
  { name: 'gamepad-2', Icon: Gamepad2 },
  { name: 'home', Icon: Home },
  { name: 'map', Icon: Map },
  { name: 'heart', Icon: Heart },
  { name: 'star', Icon: Star },
  { name: 'crown', Icon: Crown },
  { name: 'gift', Icon: Gift },
  { name: 'puzzle', Icon: Puzzle },
];

const COLOR_OPTIONS = [
  '#2196F3',
  '#E91E63',
  '#4CAF50',
  '#FF9800',
  '#9C27B0',
  '#00BCD4',
  '#F44336',
  '#607D8B',
];

interface ProjectFormModalProps {
  project?: NexusProject | null;
  onSave: (data: Partial<NexusProject>) => void;
  onDelete?: () => void;
  onClose: () => void;
}

export const ProjectFormModal = memo(function ProjectFormModal({
  project,
  onSave,
  onDelete,
  onClose,
}: ProjectFormModalProps) {
  const [name, setName] = useState(project?.name ?? '');
  const [description, setDescription] = useState(project?.description ?? '');
  const [systemInstruction, setSystemInstruction] = useState(project?.system_instruction ?? '');
  const [icon, setIcon] = useState(project?.icon ?? 'folder');
  const [color, setColor] = useState(project?.color ?? '#2196F3');

  const handleSave = useCallback(() => {
    const trimmed = name.trim();
    if (!trimmed) return;
    onSave({
      name: trimmed,
      description: description.trim() || undefined,
      system_instruction: systemInstruction.trim() || undefined,
      icon,
      color,
    });
  }, [name, description, systemInstruction, icon, color, onSave]);

  return (
    <div className={styles.daemonBuilderOverlay} onClick={onClose}>
      <div className={styles.daemonBuilder} onClick={e => e.stopPropagation()}>
        {/* Header */}
        <div className={styles.daemonBuilderHeader}>
          <h3>{project ? 'Edit Project' : 'New Project'}</h3>
          <button className={styles.detailCloseBtn} onClick={onClose}>
            <X size={16} />
          </button>
        </div>

        {/* Body */}
        <div className={styles.daemonBuilderBody}>
          {/* Name */}
          <div className={styles.daemonBuilderField}>
            <label>Name</label>
            <input
              className={styles.daemonBuilderInput}
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="Project name"
              autoFocus
            />
          </div>

          {/* Description */}
          <div className={styles.daemonBuilderField}>
            <label>Description</label>
            <textarea
              className={styles.daemonBuilderTextarea}
              value={description}
              onChange={e => setDescription(e.target.value)}
              placeholder="Optional description..."
              rows={2}
            />
          </div>

          {/* System Instruction */}
          <div className={styles.daemonBuilderField}>
            <label>System Instruction</label>
            <textarea
              className={styles.daemonBuilderTextarea}
              value={systemInstruction}
              onChange={e => setSystemInstruction(e.target.value)}
              placeholder="Instructions applied to every task in this project (e.g. paths, language, conventions)..."
              rows={4}
            />
          </div>

          {/* Icon */}
          <div className={styles.daemonBuilderField}>
            <label>Icon</label>
            <div className={styles.projectIconGrid}>
              {ICON_OPTIONS.map(opt => (
                <button
                  key={opt.name}
                  className={`${styles.daemonIconBtn} ${icon === opt.name ? styles.daemonIconBtnActive : ''}`}
                  onClick={() => setIcon(opt.name)}
                  type="button"
                  title={opt.name}
                >
                  <opt.Icon size={16} />
                </button>
              ))}
            </div>
          </div>

          {/* Color */}
          <div className={styles.daemonBuilderField}>
            <label>Color</label>
            <div className={styles.projectColorRow}>
              {COLOR_OPTIONS.map(c => (
                <button
                  key={c}
                  className={`${styles.projectColorSwatch} ${color === c ? styles.projectColorSwatchActive : ''}`}
                  style={{ backgroundColor: c }}
                  onClick={() => setColor(c)}
                  type="button"
                />
              ))}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className={styles.daemonBuilderFooter}>
          {project && onDelete ? (
            <button className={styles.projectDeleteBtn} onClick={onDelete} type="button">
              <Trash2 size={14} /> Delete
            </button>
          ) : (
            <button className={styles.daemonBuilderCancelBtn} onClick={onClose}>
              Cancel
            </button>
          )}
          <button
            className={styles.daemonBuilderSaveBtn}
            onClick={handleSave}
            disabled={!name.trim()}
          >
            {project ? 'Save' : 'Create'}
          </button>
        </div>
      </div>
    </div>
  );
});
