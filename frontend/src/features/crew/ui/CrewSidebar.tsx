import React from 'react';
import { Plus, LayoutGrid, Home, MessageSquare, BookOpen } from 'lucide-react';
import { Sidebar, type NavItem, type FooterLink, type RecentChat } from '@/components/ui/Sidebar';

const FOOTER: FooterLink[] = [
  { href: '/', label: 'Home', icon: Home, ariaLabel: 'Navigate to home' },
  { href: '/chat', label: 'Chats', icon: MessageSquare, ariaLabel: 'Navigate to chats' },
];

export interface CrewSidebarProps {
  onNewProject: () => void;
  onProjects: () => void;
  /** Opens the "getting the most out of Crew" playbook panel. */
  onHelp: () => void;
  projectsActive?: boolean;
  /** Recent crew projects as one-click shortcuts. */
  recents?: RecentChat[];
}

/** Crew sidebar — the same shared Sidebar as Chat / Builder / Settings. */
export const CrewSidebar: React.FC<CrewSidebarProps> = ({ onNewProject, onProjects, onHelp, projectsActive, recents = [] }) => {
  const nav: NavItem[] = [
    { id: 'new', label: 'New project', icon: Plus, onClick: onNewProject },
    { id: 'projects', label: 'Projects', icon: LayoutGrid, isActive: projectsActive, onClick: onProjects },
    { id: 'help', label: 'How to use', icon: BookOpen, onClick: onHelp },
  ];
  return <Sidebar brandName="Crew" navItems={nav} recentChats={recents} footerLinks={FOOTER} />;
};
