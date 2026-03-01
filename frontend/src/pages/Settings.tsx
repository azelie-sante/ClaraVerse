import { useState, useEffect, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import {
  SettingsLayout,
  AIConfigSection,
  APIKeysSection,
  UsageSection,
  CredentialsSection,
  ChannelsSection,
  PrivacySection,
  AccountSection,
  PrivacyPolicySidebar,
  DevicesSection,
} from '@/components/settings';
import type { SettingsTab } from '@/components/settings/SettingsLayout';
import './Settings.css';

export const Settings = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const [activeTab, setActiveTab] = useState<SettingsTab>('ai');
  const [saveIndicatorVisible, setSaveIndicatorVisible] = useState(false);

  // Handle URL query params for tab
  useEffect(() => {
    const tab = searchParams.get('tab');

    if (
      tab &&
      [
        'ai',
        'api-keys',
        'credentials',
        'channels',
        'devices',
        'usage',
        'privacy',
        'account',
      ].includes(tab)
    ) {
      setActiveTab(tab as SettingsTab);
    }
  }, [searchParams, setSearchParams]);

  // Show save indicator
  const showSaveIndicator = useCallback(() => {
    setSaveIndicatorVisible(true);
    setTimeout(() => setSaveIndicatorVisible(false), 2000);
  }, []);

  // Handle tab change - update URL
  const handleTabChange = (tab: SettingsTab) => {
    setActiveTab(tab);
    searchParams.set('tab', tab);
    setSearchParams(searchParams, { replace: true });
  };

  return (
    <SettingsLayout
      activeTab={activeTab}
      onTabChange={handleTabChange}
      showSaveIndicator={saveIndicatorVisible}
    >
      {/* AI Configuration Tab */}
      {activeTab === 'ai' && <AIConfigSection onSave={showSaveIndicator} />}

      {/* API Keys Tab */}
      {activeTab === 'api-keys' && (
        <section className="settings-section">
          <APIKeysSection />
        </section>
      )}

      {/* Credentials/Integrations Tab */}
      {activeTab === 'credentials' && (
        <section className="settings-section">
          <h2 className="settings-section-title">Integration Credentials</h2>
          <p className="settings-section-description">
            Securely manage API keys and webhooks for external integrations like Discord, Slack,
            GitHub, and more.
          </p>
          <CredentialsSection />
        </section>
      )}

      {/* Channels Tab */}
      {activeTab === 'channels' && (
        <section className="settings-section">
          <ChannelsSection />
        </section>
      )}

      {/* Devices Tab */}
      {activeTab === 'devices' && (
        <section className="settings-section">
          <DevicesSection />
        </section>
      )}

      {/* Usage Tab */}
      {activeTab === 'usage' && (
        <section className="settings-section">
          <UsageSection />
        </section>
      )}

      {/* Privacy Tab */}
      {activeTab === 'privacy' && (
        <div className="privacy-tab-layout">
          <PrivacySection onSave={showSaveIndicator} />
          {/* Privacy Policy Sidebar - Desktop Only */}
          <PrivacyPolicySidebar />
        </div>
      )}

      {/* Account Tab */}
      {activeTab === 'account' && <AccountSection onSave={showSaveIndicator} />}
    </SettingsLayout>
  );
};
