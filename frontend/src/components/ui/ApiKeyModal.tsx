import React, { useState } from 'react';
import { X, Key, Shield } from 'lucide-react';
import { Input, Button } from '@/components/design-system';
import styles from './ApiKeyModal.module.css';

interface ApiKeyModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (apiKey: string, rememberSession: boolean) => void;
  providerName: string;
}

export const ApiKeyModal: React.FC<ApiKeyModalProps> = ({
  isOpen,
  onClose,
  onSubmit,
  providerName,
}) => {
  const [apiKey, setApiKey] = useState('');
  const [rememberSession, setRememberSession] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (apiKey.trim()) {
      onSubmit(apiKey.trim(), rememberSession);
      setApiKey('');
      setRememberSession(false);
    }
  };

  return (
    <div className={styles.overlay}>
      <div className={styles.modal}>
        <div className={styles.header}>
          <div className={styles.titleContainer}>
            <div className={styles.iconWrapper}>
              <Key size={20} />
            </div>
            <h2 className={styles.title}>API Key Required</h2>
          </div>
          <button onClick={onClose} className={styles.closeButton}>
            <X size={20} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className={styles.content}>
          <p className={styles.description}>
            Enter your API key for <strong>{providerName}</strong> to continue.
          </p>

          <div className={styles.inputGroup}>
            <Input
              value={apiKey}
              onChange={e => setApiKey(e.target.value)}
              placeholder="sk-..."
              type="password"
              autoFocus
            />
          </div>

          <div className={styles.options}>
            <label className={styles.checkboxLabel}>
              <input
                type="checkbox"
                checked={rememberSession}
                onChange={e => setRememberSession(e.target.checked)}
              />
              <span>Remember for this session only</span>
            </label>
          </div>

          <div className={styles.securityNote}>
            <Shield size={14} />
            <span>
              {rememberSession
                ? 'Key will be cleared when you close the browser tab.'
                : 'Key will be used once and not stored.'}
            </span>
          </div>

          <div className={styles.actions}>
            <Button variant="ghost" onClick={onClose} type="button">
              Cancel
            </Button>
            <Button variant="primary" type="submit" disabled={!apiKey.trim()}>
              Continue
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};
