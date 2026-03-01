import { memo, useState, useEffect, useCallback, useMemo } from 'react';
import {
  Bookmark,
  Trash2,
  ExternalLink,
  Tag,
  Search,
  X,
  Edit3,
  Check,
  FileText,
} from 'lucide-react';
import { MarkdownRenderer } from '@/components/design-system/content/MarkdownRenderer';
import { useNexusStore } from '@/store/useNexusStore';
import { nexusService } from '@/services/nexusService';
import type { NexusSave } from '@/types/nexus';
import styles from './Nexus.module.css';

export const SavesView = memo(function SavesView() {
  const saves = useNexusStore(s => s.saves);
  const setSaves = useNexusStore(s => s.setSaves);
  const updateSaveInStore = useNexusStore(s => s.updateSaveInStore);
  const removeSave = useNexusStore(s => s.removeSave);
  const tasks = useNexusStore(s => s.tasks);
  const setSelectedTaskId = useNexusStore(s => s.setSelectedTaskId);
  const setActiveView = useNexusStore(s => s.setActiveView);
  const setActiveProjectId = useNexusStore(s => s.setActiveProjectId);

  const [loading, setLoading] = useState(false);
  const [expandedId, setExpandedId] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editTitle, setEditTitle] = useState('');
  const [editContent, setEditContent] = useState('');
  const [editTags, setEditTags] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [filterTag, setFilterTag] = useState<string | null>(null);

  // Fetch saves on mount
  useEffect(() => {
    setLoading(true);
    nexusService
      .listSaves({ limit: 100 })
      .then(fetched => setSaves(fetched))
      .catch(err => console.error('Failed to fetch saves:', err))
      .finally(() => setLoading(false));
  }, [setSaves]);

  // Collect all unique tags
  const allTags = useMemo(() => {
    const tagSet = new Set<string>();
    for (const s of saves) {
      if (s.tags) s.tags.forEach(t => tagSet.add(t));
    }
    return Array.from(tagSet).sort();
  }, [saves]);

  // Filter saves
  const filtered = useMemo(() => {
    let result = saves;
    if (filterTag) {
      result = result.filter(s => s.tags?.includes(filterTag));
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      result = result.filter(
        s => s.title.toLowerCase().includes(q) || s.content.toLowerCase().includes(q)
      );
    }
    return result;
  }, [saves, filterTag, searchQuery]);

  const handleDelete = useCallback(
    async (id: string) => {
      try {
        await nexusService.deleteSave(id);
        removeSave(id);
        if (expandedId === id) setExpandedId(null);
        if (editingId === id) setEditingId(null);
      } catch (err) {
        console.error('Failed to delete save:', err);
      }
    },
    [removeSave, expandedId, editingId]
  );

  const startEdit = useCallback((save: NexusSave) => {
    setEditingId(save.id);
    setEditTitle(save.title);
    setEditContent(save.content);
    setEditTags(save.tags?.join(', ') ?? '');
    setExpandedId(save.id);
  }, []);

  const cancelEdit = useCallback(() => {
    setEditingId(null);
  }, []);

  const saveEdit = useCallback(
    async (id: string) => {
      const tags = editTags
        .split(',')
        .map(t => t.trim())
        .filter(Boolean);
      try {
        await nexusService.updateSave(id, {
          title: editTitle,
          content: editContent,
          tags: tags.length > 0 ? tags : undefined,
        });
        updateSaveInStore(id, {
          title: editTitle,
          content: editContent,
          tags: tags.length > 0 ? tags : undefined,
        });
        setEditingId(null);
      } catch (err) {
        console.error('Failed to update save:', err);
      }
    },
    [editTitle, editContent, editTags, updateSaveInStore]
  );

  const navigateToSourceTask = useCallback(
    (save: NexusSave) => {
      if (!save.source_task_id) return;
      const task = tasks.find(t => t.id === save.source_task_id);
      if (task?.project_id) {
        setActiveProjectId(task.project_id);
      }
      setActiveView('project');
      setSelectedTaskId(save.source_task_id);
    },
    [tasks, setActiveProjectId, setActiveView, setSelectedTaskId]
  );

  if (loading && saves.length === 0) {
    return (
      <div className={styles.savesViewContainer}>
        <div className={styles.savesViewHeader}>
          <Bookmark size={18} />
          <h2>Saved</h2>
        </div>
        <div className={styles.detailEmpty}>Loading saves...</div>
      </div>
    );
  }

  return (
    <div className={styles.savesViewContainer}>
      <div className={styles.savesViewHeader}>
        <Bookmark size={18} />
        <h2>Saved</h2>
        <span className={styles.savesCount}>{saves.length}</span>
      </div>

      {/* Search + tag filter */}
      <div className={styles.savesToolbar}>
        <div className={styles.savesSearchWrap}>
          <Search size={14} />
          <input
            className={styles.savesSearchInput}
            placeholder="Search saves..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
          />
          {searchQuery && (
            <button className={styles.savesSearchClear} onClick={() => setSearchQuery('')}>
              <X size={12} />
            </button>
          )}
        </div>
        {allTags.length > 0 && (
          <div className={styles.savesTagBar}>
            {filterTag && (
              <button
                className={`${styles.saveTagChip} ${styles.saveTagChipActive}`}
                onClick={() => setFilterTag(null)}
              >
                {filterTag} <X size={10} />
              </button>
            )}
            {!filterTag &&
              allTags.map(tag => (
                <button key={tag} className={styles.saveTagChip} onClick={() => setFilterTag(tag)}>
                  {tag}
                </button>
              ))}
          </div>
        )}
      </div>

      {/* Save cards */}
      {filtered.length === 0 ? (
        <div className={styles.detailEmpty}>
          {saves.length === 0
            ? 'No saves yet. Save a task result to see it here.'
            : 'No saves match your search.'}
        </div>
      ) : (
        <div className={styles.savesList}>
          {filtered.map(save => {
            const isExpanded = expandedId === save.id;
            const isEditing = editingId === save.id;

            return (
              <div key={save.id} className={styles.saveCard}>
                <div
                  className={styles.saveCardHeader}
                  onClick={() => setExpandedId(isExpanded ? null : save.id)}
                >
                  <FileText size={14} className={styles.saveCardIcon} />
                  <div className={styles.saveCardTitleWrap}>
                    <span className={styles.saveCardTitle}>{save.title || 'Untitled'}</span>
                    <span className={styles.saveCardDate}>
                      {new Date(save.created_at).toLocaleDateString()}
                    </span>
                  </div>
                  <div className={styles.saveCardActions}>
                    {save.source_task_id && (
                      <button
                        className={styles.saveCardActionBtn}
                        onClick={e => {
                          e.stopPropagation();
                          navigateToSourceTask(save);
                        }}
                        title="Go to source task"
                      >
                        <ExternalLink size={13} />
                      </button>
                    )}
                    <button
                      className={styles.saveCardActionBtn}
                      onClick={e => {
                        e.stopPropagation();
                        startEdit(save);
                      }}
                      title="Edit"
                    >
                      <Edit3 size={13} />
                    </button>
                    <button
                      className={`${styles.saveCardActionBtn} ${styles.saveCardActionBtnDanger}`}
                      onClick={e => {
                        e.stopPropagation();
                        handleDelete(save.id);
                      }}
                      title="Delete"
                    >
                      <Trash2 size={13} />
                    </button>
                  </div>
                </div>

                {!isExpanded && (
                  <div className={styles.saveCardPreview}>
                    {save.content.slice(0, 120)}
                    {save.content.length > 120 ? '...' : ''}
                  </div>
                )}

                {save.tags && save.tags.length > 0 && (
                  <div className={styles.saveCardTags}>
                    <Tag size={10} />
                    {save.tags.map(tag => (
                      <span key={tag} className={styles.saveTagChip}>
                        {tag}
                      </span>
                    ))}
                  </div>
                )}

                {isExpanded && (
                  <div className={styles.saveCardBody}>
                    {isEditing ? (
                      <div className={styles.saveEditForm}>
                        <input
                          className={styles.saveEditTitle}
                          value={editTitle}
                          onChange={e => setEditTitle(e.target.value)}
                          placeholder="Title"
                        />
                        <textarea
                          className={styles.saveEditContent}
                          value={editContent}
                          onChange={e => setEditContent(e.target.value)}
                          placeholder="Content (markdown)"
                          rows={8}
                        />
                        <input
                          className={styles.saveEditTags}
                          value={editTags}
                          onChange={e => setEditTags(e.target.value)}
                          placeholder="Tags (comma-separated)"
                        />
                        <div className={styles.saveEditActions}>
                          <button
                            className={styles.saveEditSaveBtn}
                            onClick={() => saveEdit(save.id)}
                          >
                            <Check size={14} /> Save
                          </button>
                          <button className={styles.saveEditCancelBtn} onClick={cancelEdit}>
                            Cancel
                          </button>
                        </div>
                      </div>
                    ) : (
                      <MarkdownRenderer content={save.content} />
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
});
