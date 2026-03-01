import { useState, useRef, useEffect, useCallback } from 'react';
import { Link2, Loader2, AlertCircle, CheckCircle } from 'lucide-react';
import { Typography, Alert } from '@/components/design-system';
import { authorizeDevice, formatUserCode, isValidUserCode } from '@/services/deviceService';
import type { DeviceInfo } from '../MCPOnboarding';
import './AuthStep.css';

export interface AuthStepProps {
  initialCode?: string;
  userEmail?: string;
  onAuthorized: (deviceInfo: DeviceInfo) => void;
}

export const AuthStep = ({ initialCode, userEmail, onAuthorized }: AuthStepProps) => {
  const [code, setCode] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  // Pre-fill from URL parameter and auto-submit
  useEffect(() => {
    if (initialCode) {
      const normalized = initialCode.toUpperCase().replace(/[-\s]/g, '');
      setCode(normalized);
      if (isValidUserCode(normalized)) {
        setTimeout(() => handleSubmit(normalized), 500);
      }
    }
  }, [initialCode]);

  const handleSubmit = useCallback(
    async (codeOverride?: string) => {
      const submitCode = codeOverride || code;
      const normalized = submitCode.toUpperCase().replace(/[-\s]/g, '');

      if (!isValidUserCode(normalized)) {
        setError('Invalid code format. The code should be 8 characters.');
        return;
      }

      setIsSubmitting(true);
      setError(null);

      try {
        const formatted = formatUserCode(normalized);
        const response = await authorizeDevice(formatted);

        onAuthorized({
          device_id: response.device_info?.device_id || 'unknown',
          platform: response.device_info?.platform || 'unknown',
          client_version: response.device_info?.client_version || '1.0.0',
        });
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Authorization failed.');
        setIsSubmitting(false);
      }
    },
    [code, onAuthorized],
  );

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const raw = e.target.value.toUpperCase().replace(/[^BCDFGHJKMNPQRSTVWXYZ23456789-]/g, '');
    setCode(raw);
    setError(null);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSubmit();
    }
  };

  const normalized = code.replace(/[-\s]/g, '');
  const isValid = isValidUserCode(normalized);

  return (
    <div className="auth-step">
      <div className="step-header">
        <div className="step-icon-wrapper">
          <Link2 size={24} />
        </div>
        <Typography variant="h4" weight="semibold" className="step-title">
          Authorize Device
        </Typography>
        <Typography variant="sm" className="step-subtitle">
          Enter the code shown in your terminal to connect your device.
        </Typography>
      </div>

      {userEmail && (
        <div className="user-info">
          <Typography variant="xs">
            Authorizing as <strong>{userEmail}</strong>
          </Typography>
        </div>
      )}

      <div className="auth-form">
        <input
          ref={inputRef}
          type="text"
          value={code}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          placeholder="XXXX-XXXX"
          maxLength={9}
          autoFocus={!initialCode}
          disabled={isSubmitting}
          style={{
            width: '100%',
            padding: '14px 16px',
            borderRadius: '8px',
            border: '1px solid var(--color-border, #333)',
            background: 'var(--color-bg-secondary, #1a1a2e)',
            color: 'var(--color-text-primary, #fff)',
            fontSize: '1.5rem',
            fontFamily: 'monospace',
            textAlign: 'center',
            letterSpacing: '0.2em',
            outline: 'none',
          }}
        />

        {error && (
          <Alert variant="error" className="auth-error">
            <AlertCircle size={16} />
            <span>{error}</span>
          </Alert>
        )}

        <button
          type="button"
          onClick={() => handleSubmit()}
          disabled={!isValid || isSubmitting}
          style={{
            width: '100%',
            padding: '12px',
            borderRadius: '8px',
            border: 'none',
            background: isValid && !isSubmitting ? 'var(--color-primary, #6366f1)' : '#444',
            color: 'white',
            fontWeight: 600,
            cursor: isValid && !isSubmitting ? 'pointer' : 'not-allowed',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '8px',
            opacity: isValid && !isSubmitting ? 1 : 0.5,
          }}
        >
          {isSubmitting ? (
            <>
              <Loader2 size={18} className="spin" /> Authorizing...
            </>
          ) : (
            <>
              <CheckCircle size={18} /> Authorize Device
            </>
          )}
        </button>
      </div>

      <Alert
        variant="info"
        message="Run 'clara_companion login' in your terminal to get a device code."
        className="auth-warning"
      />
    </div>
  );
};
