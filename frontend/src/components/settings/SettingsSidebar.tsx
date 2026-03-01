import React from 'react';
import {
  Bot,
  KeyRound,
  Key,
  Activity,
  Shield,
  User,
  Home,
  MessageSquare,
  Radio,
  Smartphone,
} from 'lucide-react';
import { Sidebar, type NavItem, type FooterLink } from '@/components/ui/Sidebar';
import type { SettingsTab } from './SettingsLayout';

/** Footer links for settings - Home and Chats */
const SETTINGS_FOOTER_LINKS: FooterLink[] = [
  { href: '/', label: 'Home', icon: Home, ariaLabel: 'Navigate to home' },
  { href: '/chat', label: 'Chats', icon: MessageSquare, ariaLabel: 'Navigate to chats' },
];

export interface SettingsSidebarProps {
  /** Currently active tab */
  activeTab: SettingsTab;
  /** Callback when tab changes */
  onTabChange: (tab: SettingsTab) => void;
  /** External control: is sidebar open */
  isOpen?: boolean;
  /** External control: callback when sidebar should open/close */
  onOpenChange?: (open: boolean) => void;
}

/**
 * Settings-specific sidebar wrapper around the base Sidebar component.
 * Configures navigation items for the 6 settings tabs.
 * Home and Chats are in the sidebar footer (bottom buttons).
 */
export const SettingsSidebar: React.FC<SettingsSidebarProps> = ({
  activeTab,
  onTabChange,
  isOpen,
  onOpenChange,
}) => {
  // Define navigation items for settings tabs only
  // Home and Chats are handled by the Sidebar footer
  const navItems: NavItem[] = [
    {
      id: 'ai',
      label: 'AI Configuration',
      icon: Bot,
      isActive: activeTab === 'ai',
      onClick: () => onTabChange('ai'),
    },
    {
      id: 'api-keys',
      label: 'API Keys',
      icon: KeyRound,
      isActive: activeTab === 'api-keys',
      onClick: () => onTabChange('api-keys'),
    },
    {
      id: 'credentials',
      label: 'Integrations',
      icon: Key,
      isActive: activeTab === 'credentials',
      onClick: () => onTabChange('credentials'),
    },
    {
      id: 'channels',
      label: 'Channels',
      icon: Radio,
      isActive: activeTab === 'channels',
      onClick: () => onTabChange('channels'),
    },
    {
      id: 'devices',
      label: 'Devices',
      icon: Smartphone,
      isActive: activeTab === 'devices',
      onClick: () => onTabChange('devices'),
    },
    {
      id: 'usage',
      label: 'Usage',
      icon: Activity,
      isActive: activeTab === 'usage',
      onClick: () => onTabChange('usage'),
    },
    {
      id: 'privacy',
      label: 'Privacy',
      icon: Shield,
      isActive: activeTab === 'privacy',
      onClick: () => onTabChange('privacy'),
    },
    {
      id: 'account',
      label: 'Account',
      icon: User,
      isActive: activeTab === 'account',
      onClick: () => onTabChange('account'),
    },
  ];

  return (
    <Sidebar
      brandName="Settings"
      navItems={navItems}
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      footerLinks={SETTINGS_FOOTER_LINKS}
    />
  );
};
