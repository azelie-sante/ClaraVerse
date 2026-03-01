import { useEffect, useState } from 'react';
import { Save, AlertCircle, Eye, EyeOff, Trash2, ExternalLink } from 'lucide-react';
import { fetchE2BSettings, updateE2BApiKey } from '@/services/e2bService';
import { toast } from '@/store/useToastStore';

export const E2BSettings = () => {
  const [apiKeySet, setApiKeySet] = useState(false);
  const [apiKeyMasked, setApiKeyMasked] = useState('');
  const [newApiKey, setNewApiKey] = useState('');
  const [showKey, setShowKey] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      setIsLoading(true);
      const settings = await fetchE2BSettings();
      setApiKeySet(settings.api_key_set);
      setApiKeyMasked(settings.api_key_masked);
    } catch (error) {
      console.error('Failed to load E2B settings:', error);
      toast.error('Failed to load E2B settings', 'Error');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSave = async () => {
    if (!newApiKey.trim()) {
      toast.error('Please enter an API key', 'Validation Error');
      return;
    }

    try {
      setIsSaving(true);
      await updateE2BApiKey(newApiKey.trim());
      toast.success('E2B API key saved successfully', 'Success');
      setNewApiKey('');
      await loadSettings();
    } catch (error) {
      console.error('Failed to save E2B settings:', error);
      toast.error('Failed to save E2B API key', 'Error');
    } finally {
      setIsSaving(false);
    }
  };

  const handleClear = async () => {
    try {
      setIsSaving(true);
      await updateE2BApiKey('');
      toast.success('E2B API key removed', 'Success');
      setNewApiKey('');
      await loadSettings();
    } catch (error) {
      console.error('Failed to clear E2B settings:', error);
      toast.error('Failed to remove E2B API key', 'Error');
    } finally {
      setIsSaving(false);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold text-[var(--color-text-primary)]">Code Execution</h1>
        <p className="text-[var(--color-text-secondary)]">Loading...</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-[var(--color-text-primary)]">Code Execution</h1>
        <p className="text-[var(--color-text-secondary)] mt-2">
          Configure E2B for sandboxed Python code execution. E2B provides secure cloud sandboxes for
          running user code.
        </p>
      </div>

      {/* Info Banner */}
      <div
        className="bg-[var(--color-info-bg)] border border-[var(--color-info)] rounded-lg p-4 flex items-start gap-3"
        style={{ backdropFilter: 'blur(20px)' }}
      >
        <AlertCircle size={20} className="text-[var(--color-info)] flex-shrink-0 mt-0.5" />
        <div>
          <p className="text-sm text-[var(--color-text-primary)] font-medium">
            E2B API Key Required
          </p>
          <p className="text-sm text-[var(--color-text-secondary)] mt-1">
            Code execution tools (Python runner, data analyst, ML trainer, API tester) require an
            E2B API key.{' '}
            <a
              href="https://e2b.dev"
              target="_blank"
              rel="noopener noreferrer"
              className="text-[var(--color-accent)] hover:underline inline-flex items-center gap-1"
            >
              Get your free API key at e2b.dev
              <ExternalLink size={12} />
            </a>
          </p>
        </div>
      </div>

      {/* Current Status */}
      <div
        className="bg-[var(--color-surface)] rounded-lg p-6 space-y-6"
        style={{ backdropFilter: 'blur(20px)' }}
      >
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-sm font-medium text-[var(--color-text-primary)]">Current Status</h3>
            <p className="text-xs text-[var(--color-text-tertiary)] mt-1">
              {apiKeySet ? 'API key is configured' : 'No API key set'}
            </p>
          </div>
          <div
            className={`px-3 py-1 rounded-full text-xs font-semibold ${
              apiKeySet
                ? 'bg-green-500/15 text-green-400 border border-green-500/30'
                : 'bg-yellow-500/15 text-yellow-400 border border-yellow-500/30'
            }`}
          >
            {apiKeySet ? 'Active' : 'Not Configured'}
          </div>
        </div>

        {apiKeySet && (
          <div className="flex items-center justify-between p-3 bg-[var(--color-surface-hover)] rounded-lg">
            <div>
              <p className="text-xs text-[var(--color-text-tertiary)]">Current Key</p>
              <p className="text-sm font-mono text-[var(--color-text-primary)] mt-1">
                {apiKeyMasked}
              </p>
            </div>
            <button
              onClick={handleClear}
              disabled={isSaving}
              className="p-2 text-red-400 hover:text-red-300 hover:bg-red-500/10 rounded-lg transition-colors disabled:opacity-50"
              title="Remove API key"
            >
              <Trash2 size={16} />
            </button>
          </div>
        )}

        {/* Set/Update API Key */}
        <div className="space-y-2">
          <label className="block text-sm font-medium text-[var(--color-text-primary)]">
            {apiKeySet ? 'Update API Key' : 'Set API Key'}
          </label>
          <div className="flex gap-3">
            <div className="flex-1 relative">
              <input
                type={showKey ? 'text' : 'password'}
                value={newApiKey}
                onChange={e => setNewApiKey(e.target.value)}
                placeholder="e2b_..."
                className="w-full px-4 py-2 pr-10 bg-[var(--color-surface-hover)] border border-[var(--color-border)] rounded-lg text-[var(--color-text-primary)] placeholder-[var(--color-text-tertiary)] focus:outline-none focus:ring-2 focus:ring-[var(--color-accent)] focus:border-transparent font-mono text-sm"
              />
              <button
                type="button"
                onClick={() => setShowKey(!showKey)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)] transition-colors"
              >
                {showKey ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            </div>
            <button
              onClick={handleSave}
              disabled={isSaving || !newApiKey.trim()}
              className="px-4 py-2 text-sm font-medium text-white bg-[var(--color-accent)] rounded-lg hover:bg-[var(--color-accent-hover)] transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 whitespace-nowrap"
            >
              <Save size={16} />
              {isSaving ? 'Saving...' : 'Save'}
            </button>
          </div>
        </div>
      </div>

      {/* How it works */}
      <div
        className="bg-[var(--color-surface)] rounded-lg p-6"
        style={{ backdropFilter: 'blur(20px)' }}
      >
        <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-3">How it works</h3>
        <ul className="space-y-2 text-sm text-[var(--color-text-secondary)]">
          <li className="flex items-start gap-2">
            <span className="text-[var(--color-accent)] mt-0.5">1.</span>
            When a user runs code (Python, data analysis, etc.), the request goes to the E2B
            microservice.
          </li>
          <li className="flex items-start gap-2">
            <span className="text-[var(--color-accent)] mt-0.5">2.</span>
            E2B creates an isolated cloud sandbox, runs the code safely, and returns results.
          </li>
          <li className="flex items-start gap-2">
            <span className="text-[var(--color-accent)] mt-0.5">3.</span>
            The E2B service must also be running (uncomment it in docker-compose.yml).
          </li>
        </ul>
      </div>
    </div>
  );
};
