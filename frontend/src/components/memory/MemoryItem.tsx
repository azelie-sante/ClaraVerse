import React, { useState } from 'react';
import {
  Tag,
  Archive,
  ArchiveRestore,
  Trash2,
  Clock,
  TrendingUp,
  Edit2,
  Check,
  X,
} from 'lucide-react';
import type { Memory } from '@/services/memoryService';

interface MemoryItemProps {
  memory: Memory;
  onArchive: () => void;
  onUnarchive: () => void;
  onDelete: () => void;
  isSelected: boolean;
  onSelect: () => void;
}

const CATEGORY_COLORS: Record<string, string> = {
  personal_info: 'bg-purple-500/20 text-purple-300 border-purple-500/30',
  preferences: 'bg-blue-500/20 text-blue-300 border-blue-500/30',
  context: 'bg-green-500/20 text-green-300 border-green-500/30',
  fact: 'bg-yellow-500/20 text-yellow-300 border-yellow-500/30',
  instruction: 'bg-red-500/20 text-red-300 border-red-500/30',
};

const CATEGORY_LABELS: Record<string, string> = {
  personal_info: 'Personal Info',
  preferences: 'Preferences',
  context: 'Context',
  fact: 'Fact',
  instruction: 'Instruction',
};

export const MemoryItem: React.FC<MemoryItemProps> = ({
  memory,
  onArchive,
  onUnarchive,
  onDelete,
  isSelected,
  onSelect,
}) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const isLongContent = memory.content.length > 200;
  const displayContent = isExpanded ? memory.content : memory.content.slice(0, 200);

  const formatDate = (dateStr: string | null) => {
    if (!dateStr) return 'Never';
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return `${diffDays} days ago`;
    if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`;
    if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`;
    return `${Math.floor(diffDays / 365)} years ago`;
  };

  const getScoreColor = (score: number) => {
    if (score >= 0.7) return 'text-green-400';
    if (score >= 0.4) return 'text-yellow-400';
    return 'text-red-400';
  };

  return (
    <div
      className={`bg-gray-800 rounded-lg p-4 border transition-all ${
        isSelected
          ? 'border-blue-500 shadow-lg shadow-blue-500/20'
          : memory.is_archived
            ? 'border-gray-700 opacity-60'
            : 'border-gray-700 hover:border-gray-600'
      }`}
      onClick={onSelect}
    >
      {/* Header */}
      <div className="flex items-start justify-between gap-3 mb-3">
        {/* Category Badge */}
        <span
          className={`px-2 py-1 rounded text-xs font-medium border ${
            CATEGORY_COLORS[memory.category]
          }`}
        >
          {CATEGORY_LABELS[memory.category]}
        </span>

        {/* Actions */}
        <div className="flex items-center gap-1">
          {memory.is_archived ? (
            <button
              onClick={e => {
                e.stopPropagation();
                onUnarchive();
              }}
              className="p-1.5 text-gray-400 hover:text-blue-400 hover:bg-gray-700 rounded transition-colors"
              title="Restore from archive"
            >
              <ArchiveRestore className="w-4 h-4" />
            </button>
          ) : (
            <button
              onClick={e => {
                e.stopPropagation();
                onArchive();
              }}
              className="p-1.5 text-gray-400 hover:text-yellow-400 hover:bg-gray-700 rounded transition-colors"
              title="Archive memory"
            >
              <Archive className="w-4 h-4" />
            </button>
          )}
          <button
            onClick={e => {
              e.stopPropagation();
              onDelete();
            }}
            className="p-1.5 text-gray-400 hover:text-red-400 hover:bg-gray-700 rounded transition-colors"
            title="Delete permanently"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Content */}
      <p className="text-gray-200 mb-3 leading-relaxed">
        {displayContent}
        {isLongContent && !isExpanded && '...'}
      </p>

      {isLongContent && (
        <button
          onClick={e => {
            e.stopPropagation();
            setIsExpanded(!isExpanded);
          }}
          className="text-sm text-blue-400 hover:text-blue-300 mb-3 transition-colors"
        >
          {isExpanded ? 'Show less' : 'Show more'}
        </button>
      )}

      {/* Tags */}
      {memory.tags.length > 0 && (
        <div className="flex flex-wrap gap-2 mb-3">
          {memory.tags.map(tag => (
            <span
              key={tag}
              className="flex items-center gap-1 px-2 py-1 bg-gray-700 text-gray-300 text-xs rounded"
            >
              <Tag className="w-3 h-3" />
              {tag}
            </span>
          ))}
        </div>
      )}

      {/* Metadata */}
      <div className="flex items-center gap-4 text-xs text-gray-400 pt-3 border-t border-gray-700">
        <div className="flex items-center gap-1" title={`Memory score: ${memory.score.toFixed(3)}`}>
          <TrendingUp className="w-3 h-3" />
          <span className={getScoreColor(memory.score)}>{(memory.score * 100).toFixed(0)}%</span>
        </div>

        <div className="flex items-center gap-1" title="Access count">
          <span>🔄</span>
          <span>{memory.access_count} uses</span>
        </div>

        <div className="flex items-center gap-1" title="Last accessed">
          <Clock className="w-3 h-3" />
          <span>{formatDate(memory.last_accessed_at)}</span>
        </div>

        {memory.is_archived && (
          <span className="ml-auto px-2 py-0.5 bg-gray-700 text-gray-400 rounded text-xs">
            Archived {formatDate(memory.archived_at)}
          </span>
        )}
      </div>

      {/* Engagement Score (if significant) */}
      {memory.source_engagement > 0 && (
        <div className="mt-2 pt-2 border-t border-gray-700">
          <div className="flex items-center gap-2 text-xs text-gray-500">
            <span>Source conversation engagement:</span>
            <div className="flex-1 bg-gray-700 rounded-full h-1.5 overflow-hidden">
              <div
                className="bg-blue-500 h-full transition-all"
                style={{ width: `${memory.source_engagement * 100}%` }}
              />
            </div>
            <span>{(memory.source_engagement * 100).toFixed(0)}%</span>
          </div>
        </div>
      )}
    </div>
  );
};
