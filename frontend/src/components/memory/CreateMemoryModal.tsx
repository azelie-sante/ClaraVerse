import React, { useState } from 'react';
import { X, Plus, Loader2, Tag as TagIcon } from 'lucide-react';
import memoryService from '@/services/memoryService';

interface CreateMemoryModalProps {
  onClose: () => void;
  onSuccess: () => void;
}

const CATEGORIES = [
  { value: 'personal_info', label: 'Personal Info', description: 'Information about the user' },
  { value: 'preferences', label: 'Preferences', description: 'User preferences and settings' },
  { value: 'context', label: 'Context', description: 'Contextual information' },
  { value: 'fact', label: 'Fact', description: 'Factual information' },
  { value: 'instruction', label: 'Instruction', description: 'User instructions or guidelines' },
];

export const CreateMemoryModal: React.FC<CreateMemoryModalProps> = ({ onClose, onSuccess }) => {
  const [content, setContent] = useState('');
  const [category, setCategory] = useState<string>('context');
  const [tagInput, setTagInput] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleAddTag = () => {
    const trimmedTag = tagInput.trim().toLowerCase();
    if (trimmedTag && !tags.includes(trimmedTag)) {
      setTags([...tags, trimmedTag]);
      setTagInput('');
    }
  };

  const handleRemoveTag = (tagToRemove: string) => {
    setTags(tags.filter(tag => tag !== tagToRemove));
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleAddTag();
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!content.trim()) {
      setError('Content is required');
      return;
    }

    setIsSubmitting(true);

    try {
      await memoryService.createMemory({
        content: content.trim(),
        category: category as any,
        tags,
      });
      onSuccess();
    } catch (err) {
      console.error('Failed to create memory:', err);
      setError('Failed to create memory. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-700">
          <h2 className="text-xl font-bold">Create New Memory</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white transition-colors"
            disabled={isSubmitting}
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Content */}
          <div>
            <label htmlFor="content" className="block text-sm font-medium text-gray-300 mb-2">
              Content <span className="text-red-400">*</span>
            </label>
            <textarea
              id="content"
              value={content}
              onChange={e => setContent(e.target.value)}
              placeholder="Enter the memory content..."
              className="w-full px-4 py-3 bg-gray-900 border border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 min-h-[120px] resize-y"
              disabled={isSubmitting}
            />
          </div>

          {/* Category */}
          <div>
            <label htmlFor="category" className="block text-sm font-medium text-gray-300 mb-2">
              Category <span className="text-red-400">*</span>
            </label>
            <select
              id="category"
              value={category}
              onChange={e => setCategory(e.target.value)}
              className="w-full px-4 py-3 bg-gray-900 border border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={isSubmitting}
            >
              {CATEGORIES.map(cat => (
                <option key={cat.value} value={cat.value}>
                  {cat.label} - {cat.description}
                </option>
              ))}
            </select>
          </div>

          {/* Tags */}
          <div>
            <label htmlFor="tags" className="block text-sm font-medium text-gray-300 mb-2">
              Tags (optional)
            </label>
            <div className="flex gap-2 mb-2">
              <input
                id="tags"
                type="text"
                value={tagInput}
                onChange={e => setTagInput(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Add a tag..."
                className="flex-1 px-4 py-2 bg-gray-900 border border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                disabled={isSubmitting}
              />
              <button
                type="button"
                onClick={handleAddTag}
                disabled={!tagInput.trim() || isSubmitting}
                className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-700 disabled:text-gray-500 rounded-lg transition-colors flex items-center gap-2"
              >
                <Plus className="w-4 h-4" />
                Add
              </button>
            </div>

            {/* Tag List */}
            {tags.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {tags.map(tag => (
                  <span
                    key={tag}
                    className="flex items-center gap-2 px-3 py-1 bg-gray-700 text-gray-300 rounded-full text-sm"
                  >
                    <TagIcon className="w-3 h-3" />
                    {tag}
                    <button
                      type="button"
                      onClick={() => handleRemoveTag(tag)}
                      className="text-gray-400 hover:text-red-400 transition-colors"
                      disabled={isSubmitting}
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </span>
                ))}
              </div>
            )}
          </div>

          {/* Error Message */}
          {error && (
            <div className="p-4 bg-red-500/10 border border-red-500/20 rounded-lg text-red-300">
              {error}
            </div>
          )}

          {/* Help Text */}
          <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
            <p className="text-sm text-blue-300">
              <strong>Tip:</strong> Memories you create manually will be included in the memory
              selection process just like automatically extracted ones. They help personalize your
              conversations.
            </p>
          </div>
        </form>

        {/* Footer */}
        <div className="flex items-center justify-end gap-3 p-6 border-t border-gray-700">
          <button
            type="button"
            onClick={onClose}
            disabled={isSubmitting}
            className="px-4 py-2 text-gray-300 hover:text-white transition-colors disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            disabled={isSubmitting || !content.trim()}
            className="px-6 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-700 disabled:text-gray-500 rounded-lg transition-colors flex items-center gap-2"
          >
            {isSubmitting ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Creating...
              </>
            ) : (
              <>
                <Plus className="w-4 h-4" />
                Create Memory
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  );
};
