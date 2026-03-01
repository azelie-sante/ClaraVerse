import { useState } from 'react';
import { Shield, Smartphone, Cloud, Lock } from 'lucide-react';
import type { ChatPrivacyMode } from '@/store/useSettingsStore';
import styles from './PrivacySettingsModal.module.css';

interface PrivacySettingsModalProps {
  isOpen: boolean;
  onSubmit: (mode: ChatPrivacyMode) => void;
}

export const PrivacySettingsModal: React.FC<PrivacySettingsModalProps> = ({ isOpen, onSubmit }) => {
  const [selectedMode, setSelectedMode] = useState<ChatPrivacyMode>(null);

  const handleSubmit = () => {
    if (selectedMode) {
      onSubmit(selectedMode);
    }
  };

  if (!isOpen) return null;

  return (
    <div className={styles.overlay}>
      <div className={styles.modal}>
        {/* Header */}
        <div className={styles.header}>
          <div className={styles.iconContainer}>
            <Shield size={24} />
          </div>
          <h2 className={styles.title}>Where should we save your chats?</h2>
        </div>

        {/* Options */}
        <div className={styles.options}>
          {/* Local Option */}
          <button
            className={`${styles.option} ${selectedMode === 'local' ? styles.selected : ''}`}
            onClick={() => setSelectedMode('local')}
            type="button"
          >
            <div className={styles.optionIcon}>
              <Smartphone size={20} />
            </div>
            <div className={styles.optionText}>
              <span className={styles.optionTitle}>This device only</span>
              <span className={styles.optionDesc}>Private, but won't sync to other devices</span>
            </div>
            {selectedMode === 'local' && <div className={styles.check}>✓</div>}
          </button>

          {/* Cloud Option */}
          <button
            className={`${styles.option} ${selectedMode === 'cloud' ? styles.selected : ''}`}
            onClick={() => setSelectedMode('cloud')}
            type="button"
          >
            <div className={styles.optionIcon}>
              <Cloud size={20} />
            </div>
            <div className={styles.optionText}>
              <span className={styles.optionTitle}>Sync across devices</span>
              <span className={styles.optionDesc}>
                <Lock size={10} /> Securely saved, access anywhere with your account
              </span>
            </div>
            {selectedMode === 'cloud' && <div className={styles.check}>✓</div>}
          </button>
        </div>

        {/* Button */}
        <button onClick={handleSubmit} className={styles.button} disabled={!selectedMode}>
          Continue
        </button>

        <p className={styles.note}>You can change this later in Settings</p>
      </div>
    </div>
  );
};
