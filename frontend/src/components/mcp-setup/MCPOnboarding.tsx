import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { CheckCircle, Terminal, LayoutDashboard } from 'lucide-react';
import { Card, Typography, Button } from '@/components/design-system';
import { useAuthStore } from '@/store/useAuthStore';
import { AuthStep } from './steps/AuthStep';
import './MCPOnboarding.css';

export interface MCPOnboardingProps {
  /** Initial code from URL parameter */
  initialCode?: string;
  /** Callback when wizard completes */
  onComplete?: () => void;
}

export interface DeviceInfo {
  device_id: string;
  platform: string;
  client_version: string;
}

export const MCPOnboarding = ({ initialCode, onComplete }: MCPOnboardingProps) => {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const [isAuthorized, setIsAuthorized] = useState(false);
  const [deviceInfo, setDeviceInfo] = useState<DeviceInfo | null>(null);

  const handleAuthComplete = (info: DeviceInfo) => {
    setDeviceInfo(info);
    setIsAuthorized(true);
  };

  const handleGoToDashboard = () => {
    if (onComplete) {
      onComplete();
    } else {
      navigate('/');
    }
  };

  return (
    <div className="mcp-onboarding">
      <Card variant="glass" className="mcp-onboarding-card" hoverable={false}>
        <AnimatePresence mode="wait">
          {!isAuthorized ? (
            <motion.div
              key="auth"
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.2 }}
            >
              <AuthStep
                initialCode={initialCode}
                userEmail={user?.email}
                onAuthorized={handleAuthComplete}
              />
            </motion.div>
          ) : (
            <motion.div
              key="success"
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.3 }}
              className="success-state"
            >
              <div className="success-icon-wrapper">
                <CheckCircle size={48} />
              </div>

              <Typography variant="h4" weight="semibold" className="success-title">
                Device Authorized Successfully!
              </Typography>

              {deviceInfo && (
                <Typography variant="sm" className="success-device-info">
                  {deviceInfo.platform} device connected (v{deviceInfo.client_version})
                </Typography>
              )}

              <div className="success-message">
                <Terminal size={20} />
                <Typography variant="base">
                  You can safely close this window and continue in your terminal.
                </Typography>
              </div>

              <Typography variant="xs" className="success-hint">
                The CLI will automatically detect the authorization and proceed with setup.
              </Typography>

              <Button
                variant="secondary"
                onClick={handleGoToDashboard}
                className="dashboard-button"
              >
                <LayoutDashboard size={18} />
                Go to Dashboard
              </Button>
            </motion.div>
          )}
        </AnimatePresence>
      </Card>
    </div>
  );
};
