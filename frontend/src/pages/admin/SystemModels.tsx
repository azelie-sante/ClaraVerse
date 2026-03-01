import { useEffect, useState } from 'react';
import { Settings, Save, AlertCircle } from 'lucide-react';
import { fetchModels } from '@/services/modelService';
import {
  fetchSystemModelAssignments,
  updateSystemModelAssignments,
  type SystemModelAssignments,
} from '@/services/systemModelsService';
import { toast } from '@/store/useToastStore';
import type { Model } from '@/types/websocket';

export const SystemModels = () => {
  const [models, setModels] = useState<Model[]>([]);
  const [assignments, setAssignments] = useState<SystemModelAssignments>({
    tool_selector: '',
    memory_extractor: '',
    title_generator: '',
    workflow_validator: '',
    agent_default: '',
  });
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setIsLoading(true);
      const [modelsData, assignmentsData] = await Promise.all([
        fetchModels(true),
        fetchSystemModelAssignments(),
      ]);
      setModels(modelsData);
      setAssignments(assignmentsData);
    } catch (error) {
      console.error('Failed to load system models:', error);
      toast.error('Failed to load system model assignments', 'Error');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSave = async () => {
    try {
      setIsSaving(true);
      await updateSystemModelAssignments(assignments);
      toast.success('System model assignments updated successfully', 'Success');
    } catch (error) {
      console.error('Failed to save system models:', error);
      toast.error('Failed to save system model assignments', 'Error');
    } finally {
      setIsSaving(false);
    }
  };

  const handleChange = (field: keyof SystemModelAssignments, value: string) => {
    setAssignments(prev => ({ ...prev, [field]: value }));
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-[var(--color-text-primary)]">
          System Model Assignments
        </h1>
        <p className="text-[var(--color-text-secondary)]">Loading...</p>
      </div>
    );
  }

  const assignmentFields: Array<{
    key: keyof SystemModelAssignments;
    label: string;
    description: string;
  }> = [
    {
      key: 'tool_selector',
      label: 'Tool Selector',
      description:
        'Model used for intelligent tool selection in chat. Analyzes user requests and selects relevant tools.',
    },
    {
      key: 'memory_extractor',
      label: 'Memory Extractor',
      description:
        'Model used for extracting important facts from conversations to build user memory.',
    },
    {
      key: 'title_generator',
      label: 'Title Generator',
      description: 'Model used for generating short, descriptive titles for conversations.',
    },
    {
      key: 'workflow_validator',
      label: 'Workflow Validator',
      description: 'Model used for validating and checking workflow configurations before execution.',
    },
    {
      key: 'agent_default',
      label: 'Agent Default',
      description: 'Default model for agent workflows when no specific model is selected.',
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-[var(--color-text-primary)]">
          System Model Assignments
        </h1>
        <p className="text-[var(--color-text-secondary)] mt-2">
          Configure which models are used for different system operations. These assignments override
          user preferences.
        </p>
      </div>

      {/* Info Banner */}
      <div
        className="bg-[var(--color-info-bg)] border border-[var(--color-info)] rounded-lg p-4 flex items-start gap-3"
        style={{ backdropFilter: 'blur(20px)' }}
      >
        <AlertCircle size={20} className="text-[var(--color-info)] flex-shrink-0 mt-0.5" />
        <div>
          <p className="text-sm text-[var(--color-text-primary)] font-medium">Admin Override</p>
          <p className="text-sm text-[var(--color-text-secondary)] mt-1">
            System model assignments take priority over user-specific model preferences. Leave a
            field empty to allow user preferences or use the default model pool.
          </p>
        </div>
      </div>

      {/* Assignment Form */}
      <div
        className="bg-[var(--color-surface)] rounded-lg p-6 space-y-6"
        style={{ backdropFilter: 'blur(20px)' }}
      >
        {assignmentFields.map(field => (
          <div key={field.key} className="space-y-2">
            <label className="block text-sm font-medium text-[var(--color-text-primary)]">
              {field.label}
            </label>
            <p className="text-xs text-[var(--color-text-tertiary)]">{field.description}</p>
            <select
              value={assignments[field.key]}
              onChange={e => handleChange(field.key, e.target.value)}
              className="w-full px-4 py-2 bg-[var(--color-surface-hover)] border border-[var(--color-border)] rounded-lg text-[var(--color-text-primary)] focus:outline-none focus:ring-2 focus:ring-[var(--color-accent)] focus:border-transparent"
            >
              <option value="">-- Use Default Pool --</option>
              {models.map(model => (
                <option key={model.id} value={model.id}>
                  {model.name} ({model.providerName})
                </option>
              ))}
            </select>
          </div>
        ))}
      </div>

      {/* Actions */}
      <div className="flex items-center justify-end gap-4">
        <button
          onClick={loadData}
          disabled={isSaving}
          className="px-4 py-2 text-sm font-medium text-[var(--color-text-primary)] bg-[var(--color-surface-hover)] rounded-lg hover:bg-[var(--color-surface)] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Reset
        </button>
        <button
          onClick={handleSave}
          disabled={isSaving}
          className="px-4 py-2 text-sm font-medium text-white bg-[var(--color-accent)] rounded-lg hover:bg-[var(--color-accent-hover)] transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
        >
          <Save size={16} />
          {isSaving ? 'Saving...' : 'Save Changes'}
        </button>
      </div>
    </div>
  );
};
