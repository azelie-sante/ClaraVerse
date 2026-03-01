import { useState, useEffect, useCallback } from 'react';
import {
  Loader2,
  Eye,
  EyeOff,
  ExternalLink,
  AlertCircle,
  Link as LinkIcon,
  Check,
} from 'lucide-react';
import { Modal, Button, Input } from '@/components/design-system';
import { useCredentialsStore } from '@/store/useCredentialsStore';
import { api } from '@/services/api';
import type { Integration, Credential, IntegrationField } from '@/types/credential';
import { IntegrationIcon } from './IntegrationIcon';
import './CredentialFormModal.css';

interface CredentialFormModalProps {
  integration: Integration;
  credential: Credential | null;
  onClose: () => void;
}

export const CredentialFormModal = ({
  integration,
  credential,
  onClose,
}: CredentialFormModalProps) => {
  const { createCredential, updateCredential } = useCredentialsStore();

  const [name, setName] = useState(credential?.name || `My ${integration.name}`);
  const [formData, setFormData] = useState<Record<string, string>>({});
  const [showSensitive, setShowSensitive] = useState<Record<string, boolean>>({});
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isConnecting, setIsConnecting] = useState(false);

  // Initialize form data with defaults
  useEffect(() => {
    const initialData: Record<string, string> = {};
    integration.fields.forEach(field => {
      initialData[field.key] = field.default || '';
    });
    setFormData(initialData);
  }, [integration]);

  // Validate form
  const validate = useCallback((): boolean => {
    const newErrors: Record<string, string> = {};

    if (!name.trim()) {
      newErrors.name = 'Name is required';
    }

    integration.fields.forEach(field => {
      const value = formData[field.key];
      if (field.required && !value?.trim()) {
        newErrors[field.key] = `${field.label} is required`;
      }

      // URL validation for webhook fields
      if (field.type === 'webhook_url' && value?.trim()) {
        try {
          new URL(value);
        } catch {
          newErrors[field.key] = 'Please enter a valid URL';
        }
      }
    });

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [name, formData, integration.fields]);

  // Handle field change
  const handleFieldChange = (key: string, value: string) => {
    setFormData(prev => ({ ...prev, [key]: value }));
    // Clear error on change
    if (errors[key]) {
      setErrors(prev => {
        const next = { ...prev };
        delete next[key];
        return next;
      });
    }
  };

  // Toggle sensitive field visibility
  const toggleSensitive = (key: string) => {
    setShowSensitive(prev => ({ ...prev, [key]: !prev[key] }));
  };

  // Handle submit
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError(null);

    if (!validate()) return;

    setIsSubmitting(true);
    try {
      if (credential) {
        // Update existing
        await updateCredential(credential.id, {
          name: name.trim(),
          data: formData,
        });
      } else {
        // Create new
        await createCredential({
          name: name.trim(),
          integrationType: integration.id,
          data: formData,
        });
      }
      onClose();
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'Failed to save credential');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle Composio OAuth Connect
  const handleComposioConnect = async () => {
    setIsConnecting(true);
    setSubmitError(null);

    try {
      // Determine OAuth endpoint based on integration type
      // Extract app name from integration ID (e.g., "composio_googlesheets" -> "googlesheets")
      const appName = integration.id.replace('composio_', '');

      // Call backend to initiate OAuth
      const response = await api.get(`/api/integrations/composio/${appName}/authorize`);
      const { authUrl } = response;

      // Open OAuth popup
      const width = 600;
      const height = 700;
      const left = (window.screen.width - width) / 2;
      const top = (window.screen.height - height) / 2;

      const popup = window.open(
        authUrl,
        'Composio OAuth',
        `width=${width},height=${height},left=${left},top=${top}`
      );

      // Poll for popup closure
      const checkPopup = setInterval(() => {
        if (popup?.closed) {
          clearInterval(checkPopup);
          setIsConnecting(false);

          // With Composio /link endpoint, the connection is created automatically
          // Just complete the setup to save the credential in our DB
          handleComposioCompleteSetup();
        }
      }, 500);
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'Failed to initiate OAuth');
      setIsConnecting(false);
    }
  };

  // Complete Composio setup after OAuth
  const handleComposioCompleteSetup = async () => {
    try {
      await api.post('/api/integrations/composio/complete-setup', {
        name: name.trim() || 'Google Sheets',
        integrationType: integration.id,
      });
      onClose();
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : 'Failed to complete setup');
    }
  };

  // Check if integration requires OAuth
  const isComposioIntegration = integration.id.startsWith('composio_');

  // Render a form field based on its type
  const renderField = (field: IntegrationField) => {
    const value = formData[field.key] || '';
    const error = errors[field.key];
    const isSensitive = field.sensitive;
    const isVisible = !isSensitive || showSensitive[field.key];

    return (
      <div key={field.key} className="credential-form-field">
        <label className="field-label">
          {field.label}
          {field.required && <span className="required-mark">*</span>}
        </label>

        <div className="field-input-wrapper">
          {field.type === 'select' ? (
            <select
              value={value}
              onChange={e => handleFieldChange(field.key, e.target.value)}
              className={`field-select ${error ? 'has-error' : ''}`}
            >
              <option value="">Select {field.label.toLowerCase()}...</option>
              {field.options?.map(opt => (
                <option key={opt} value={opt}>
                  {opt}
                </option>
              ))}
            </select>
          ) : field.type === 'json' ? (
            <textarea
              value={value}
              onChange={e => handleFieldChange(field.key, e.target.value)}
              placeholder={field.placeholder}
              className={`field-textarea ${error ? 'has-error' : ''}`}
              rows={4}
            />
          ) : (
            <div className="field-input-container">
              <Input
                type={isVisible ? 'text' : 'password'}
                value={value}
                onChange={e => handleFieldChange(field.key, e.target.value)}
                placeholder={field.placeholder}
                className={error ? 'input-error' : ''}
              />
              {isSensitive && (
                <button
                  type="button"
                  onClick={() => toggleSensitive(field.key)}
                  className="toggle-visibility"
                >
                  {isVisible ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              )}
            </div>
          )}
        </div>

        {error && <span className="field-error">{error}</span>}
        {field.helpText && !error && <span className="field-help">{field.helpText}</span>}
      </div>
    );
  };

  return (
    <Modal
      isOpen={true}
      onClose={onClose}
      title={
        <div className="credential-modal-title">
          <IntegrationIcon integrationId={integration.id} size={24} forceColor="white" />
          <span>
            {credential ? 'Edit' : 'Add'} {integration.name} Credential
          </span>
        </div>
      }
      size="md"
    >
      <form onSubmit={handleSubmit} className="credential-form">
        {submitError && (
          <div className="form-error-banner">
            <AlertCircle size={16} />
            <span>{submitError}</span>
          </div>
        )}

        {isComposioIntegration ? (
          /* OAuth Connect Flow for Composio */
          <>
            <div className="composio-connect-info">
              <div className="info-box">
                <AlertCircle size={20} />
                <div>
                  <p className="info-title">Connect via OAuth</p>
                  <p className="info-text">
                    Click the button below to securely connect your account. No manual credentials
                    needed!
                  </p>
                </div>
              </div>
            </div>

            {/* Credential Name */}
            <div className="credential-form-field">
              <label className="field-label">Credential Name (Optional)</label>
              <Input
                value={name}
                onChange={e => setName(e.target.value)}
                placeholder={`My ${integration.name}`}
              />
              <span className="field-help">A friendly name to identify this credential</span>
            </div>

            {/* Connect Button */}
            <Button
              type="button"
              onClick={handleComposioConnect}
              disabled={isConnecting}
              className="w-full"
            >
              {isConnecting ? (
                <>
                  <Loader2 size={16} className="animate-spin mr-2" />
                  Opening login...
                </>
              ) : (
                <>
                  <LinkIcon size={16} className="mr-2" />
                  Connect Account
                </>
              )}
            </Button>
          </>
        ) : (
          /* Manual Form for Other Integrations */
          <>
            {/* Credential Name */}
            <div className="credential-form-field">
              <label className="field-label">
                Credential Name
                <span className="required-mark">*</span>
              </label>
              <Input
                value={name}
                onChange={e => {
                  setName(e.target.value);
                  if (errors.name) {
                    setErrors(prev => {
                      const next = { ...prev };
                      delete next.name;
                      return next;
                    });
                  }
                }}
                placeholder={`My ${integration.name}`}
                className={errors.name ? 'input-error' : ''}
              />
              {errors.name && <span className="field-error">{errors.name}</span>}
              <span className="field-help">A friendly name to identify this credential</span>
            </div>

            {/* Integration Fields */}
            {integration.fields.map(field => renderField(field))}
          </>
        )}

        {/* Capabilities - Show for Composio integrations */}
        {isComposioIntegration && integration.tools && integration.tools.length > 0 && (
          <div className="capabilities-section">
            <h4 className="capabilities-title">🚀 What you can do with {integration.name}</h4>
            <div className="capabilities-grid">
              {integration.tools.slice(0, 6).map((tool, index) => {
                // Convert snake_case to Title Case and remove prefix
                const toolName = tool
                  .replace(/^[a-z]+_/, '') // Remove prefix like "linkedin_"
                  .split('_')
                  .map(word => word.charAt(0).toUpperCase() + word.slice(1))
                  .join(' ');
                return (
                  <div key={index} className="capability-item">
                    <Check size={14} className="capability-check" />
                    <span>{toolName}</span>
                  </div>
                );
              })}
              {integration.tools.length > 6 && (
                <div className="capability-item capability-more">
                  <span>+ {integration.tools.length - 6} more actions</span>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Docs Link - Show for non-Composio integrations */}
        {!isComposioIntegration && integration.docsUrl && (
          <a
            href={integration.docsUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="docs-link"
          >
            <ExternalLink size={14} />
            View {integration.name} documentation
          </a>
        )}

        {/* Actions */}
        <div className="form-actions">
          <Button variant="ghost" type="button" onClick={onClose}>
            Cancel
          </Button>
          {!isComposioIntegration && (
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting && <Loader2 size={16} className="animate-spin mr-2" />}
              {credential ? 'Update' : 'Save'} Credential
            </Button>
          )}
        </div>
      </form>
    </Modal>
  );
};

export default CredentialFormModal;
