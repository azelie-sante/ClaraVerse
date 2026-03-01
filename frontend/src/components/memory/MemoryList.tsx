import React, { useState, useEffect } from 'react';
import {
  Brain,
  Search,
  Filter,
  Archive,
  ArchiveRestore,
  Trash2,
  Plus,
  Loader2,
  AlertCircle,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import memoryService from '@/services/memoryService';
import type { Memory, MemoryListResponse } from '@/services/memoryService';
import { MemoryItem } from './MemoryItem';
import { CreateMemoryModal } from './CreateMemoryModal';

export const MemoryList: React.FC = () => {
  const [memories, setMemories] = useState<Memory[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Pagination
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalMemories, setTotalMemories] = useState(0);
  const pageSize = 20;

  // Filters
  const [searchQuery, setSearchQuery] = useState('');
  const [categoryFilter, setCategoryFilter] = useState<string>('');
  const [showArchived, setShowArchived] = useState(false);

  // UI state
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedMemoryId, setSelectedMemoryId] = useState<string | null>(null);

  useEffect(() => {
    loadMemories();
  }, [currentPage, categoryFilter, showArchived]);

  const loadMemories = async () => {
    setLoading(true);
    setError(null);
    try {
      const response: MemoryListResponse = await memoryService.listMemories({
        category: categoryFilter || undefined,
        includeArchived: showArchived,
        page: currentPage,
        pageSize,
      });
      setMemories(response.memories);
      setTotalPages(response.pagination.total_pages);
      setTotalMemories(response.pagination.total);
    } catch (err) {
      console.error('Failed to load memories:', err);
      setError('Failed to load memories. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleArchive = async (id: string) => {
    try {
      await memoryService.archiveMemory(id);
      loadMemories();
    } catch (err) {
      console.error('Failed to archive memory:', err);
    }
  };

  const handleUnarchive = async (id: string) => {
    try {
      await memoryService.unarchiveMemory(id);
      loadMemories();
    } catch (err) {
      console.error('Failed to unarchive memory:', err);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to permanently delete this memory?')) {
      return;
    }
    try {
      await memoryService.deleteMemory(id);
      loadMemories();
    } catch (err) {
      console.error('Failed to delete memory:', err);
    }
  };

  const handleCreateMemory = () => {
    setShowCreateModal(false);
    loadMemories();
  };

  const filteredMemories = memories.filter(
    memory =>
      searchQuery === '' ||
      memory.content.toLowerCase().includes(searchQuery.toLowerCase()) ||
      memory.tags.some(tag => tag.toLowerCase().includes(searchQuery.toLowerCase()))
  );

  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Brain className="w-8 h-8" />
            Memory Management
          </h1>
          <p className="text-gray-400 mt-1">View and manage your stored memories</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg transition-colors"
        >
          <Plus className="w-4 h-4" />
          Create Memory
        </button>
      </div>

      {/* Stats Bar */}
      <div className="bg-gray-800 rounded-lg p-4 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <div>
            <p className="text-sm text-gray-400">Total</p>
            <p className="text-xl font-bold">{totalMemories}</p>
          </div>
          <div className="w-px h-8 bg-gray-700" />
          <div>
            <p className="text-sm text-gray-400">Active</p>
            <p className="text-xl font-bold text-green-400">
              {memories.filter(m => !m.is_archived).length}
            </p>
          </div>
          <div className="w-px h-8 bg-gray-700" />
          <div>
            <p className="text-sm text-gray-400">Archived</p>
            <p className="text-xl font-bold text-gray-500">
              {memories.filter(m => m.is_archived).length}
            </p>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-gray-800 rounded-lg p-4 space-y-4">
        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search memories..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 bg-gray-900 border border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        {/* Filter Row */}
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Filter className="w-4 h-4 text-gray-400" />
            <select
              value={categoryFilter}
              onChange={e => setCategoryFilter(e.target.value)}
              className="bg-gray-900 border border-gray-700 rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">All Categories</option>
              <option value="personal_info">Personal Info</option>
              <option value="preferences">Preferences</option>
              <option value="context">Context</option>
              <option value="fact">Facts</option>
              <option value="instruction">Instructions</option>
            </select>
          </div>

          <button
            onClick={() => setShowArchived(!showArchived)}
            className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm transition-colors ${
              showArchived
                ? 'bg-blue-600 text-white'
                : 'bg-gray-900 text-gray-400 hover:bg-gray-700'
            }`}
          >
            {showArchived ? (
              <ArchiveRestore className="w-4 h-4" />
            ) : (
              <Archive className="w-4 h-4" />
            )}
            {showArchived ? 'Show Active' : 'Show Archived'}
          </button>

          <button
            onClick={loadMemories}
            disabled={loading}
            className="ml-auto flex items-center gap-2 px-3 py-1.5 bg-gray-900 hover:bg-gray-700 rounded-lg text-sm transition-colors disabled:opacity-50"
          >
            {loading ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
              </svg>
            )}
            Refresh
          </button>
        </div>
      </div>

      {/* Error State */}
      {error && (
        <div className="flex items-center gap-2 p-4 bg-red-500/10 border border-red-500/20 rounded-lg">
          <AlertCircle className="w-5 h-5 text-red-400 flex-shrink-0" />
          <p className="text-red-300">{error}</p>
        </div>
      )}

      {/* Memory List */}
      <div className="space-y-3">
        {loading && memories.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-gray-500">
            <Loader2 className="w-8 h-8 animate-spin mb-3" />
            <p>Loading memories...</p>
          </div>
        ) : filteredMemories.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-gray-500">
            <Brain className="w-16 h-16 mb-3 opacity-50" />
            <p className="text-lg font-semibold">No memories found</p>
            <p className="text-sm mt-1">
              {searchQuery
                ? 'Try adjusting your search or filters'
                : 'Memories will appear here as they are extracted from conversations'}
            </p>
          </div>
        ) : (
          filteredMemories.map(memory => (
            <MemoryItem
              key={memory.id}
              memory={memory}
              onArchive={() => handleArchive(memory.id)}
              onUnarchive={() => handleUnarchive(memory.id)}
              onDelete={() => handleDelete(memory.id)}
              isSelected={selectedMemoryId === memory.id}
              onSelect={() =>
                setSelectedMemoryId(memory.id === selectedMemoryId ? null : memory.id)
              }
            />
          ))
        )}
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between bg-gray-800 rounded-lg p-4">
          <p className="text-sm text-gray-400">
            Page {currentPage} of {totalPages}
          </p>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
              disabled={currentPage === 1 || loading}
              className="p-2 bg-gray-900 hover:bg-gray-700 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <button
              onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
              disabled={currentPage === totalPages || loading}
              className="p-2 bg-gray-900 hover:bg-gray-700 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </div>
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <CreateMemoryModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={handleCreateMemory}
        />
      )}
    </div>
  );
};
