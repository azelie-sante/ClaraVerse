import type { LucideIcon } from 'lucide-react';
import * as LucideIcons from 'lucide-react';

/**
 * Map of icon names to their Lucide React components
 * This allows the backend to send icon names as strings and the frontend to render them dynamically
 */

// Create a typed map of all Lucide icons
const iconMap: Record<string, LucideIcon> = {
  // Tools
  Calculator: LucideIcons.Calculator,
  Clock: LucideIcons.Clock,
  Search: LucideIcons.Search,
  Globe: LucideIcons.Globe,
  FileText: LucideIcons.FileText,
  File: LucideIcons.File,
  FileCode: LucideIcons.FileCode,
  FileSpreadsheet: LucideIcons.FileSpreadsheet,
  Code: LucideIcons.Code,
  Terminal: LucideIcons.Terminal,
  FolderOpen: LucideIcons.FolderOpen,
  Database: LucideIcons.Database,
  Cloud: LucideIcons.Cloud,
  Download: LucideIcons.Download,
  Upload: LucideIcons.Upload,
  Wrench: LucideIcons.Wrench,
  Settings: LucideIcons.Settings,
  Zap: LucideIcons.Zap,
  GitBranch: LucideIcons.GitBranch,
  Package: LucideIcons.Package,
  Box: LucideIcons.Box,
  Archive: LucideIcons.Archive,
  BookOpen: LucideIcons.BookOpen,
  Book: LucideIcons.Book,
  Bookmark: LucideIcons.Bookmark,
  Link: LucideIcons.Link,
  ExternalLink: LucideIcons.ExternalLink,
  Image: LucideIcons.Image,
  ImagePlus: LucideIcons.ImagePlus,
  Video: LucideIcons.Video,
  Music: LucideIcons.Music,
  Mic: LucideIcons.Mic,
  Camera: LucideIcons.Camera,
  MapPin: LucideIcons.MapPin,
  Map: LucideIcons.Map,
  Compass: LucideIcons.Compass,
  Calendar: LucideIcons.Calendar,
  Mail: LucideIcons.Mail,
  Send: LucideIcons.Send,
  MessageSquare: LucideIcons.MessageSquare,
  MessageCircle: LucideIcons.MessageCircle,
  MessageCircleQuestion: LucideIcons.MessageCircleQuestion,
  Phone: LucideIcons.Phone,
  User: LucideIcons.User,
  Users: LucideIcons.Users,
  Activity: LucideIcons.Activity,
  TrendingUp: LucideIcons.TrendingUp,
  BarChart: LucideIcons.BarChart,
  ChartBar: LucideIcons.ChartBar,
  PieChart: LucideIcons.PieChart,
  Cpu: LucideIcons.Cpu,
  HardDrive: LucideIcons.HardDrive,
  Smartphone: LucideIcons.Smartphone,
  Monitor: LucideIcons.Monitor,
  // Status / Misc
  Sparkles: LucideIcons.Sparkles,
  BarChart2: LucideIcons.BarChart2,
  FilePlus: LucideIcons.FilePlus,
  Pencil: LucideIcons.Pencil,
  // Add more as needed...
};

/**
 * Get a Lucide icon component by name
 * @param iconName - The name of the icon (e.g., "Calculator", "Clock", "Search")
 * @param fallback - Optional fallback icon if the requested icon is not found
 * @returns The Lucide icon component or the fallback
 */
export function getIconByName(
  iconName?: string,
  fallback: LucideIcon = LucideIcons.HelpCircle
): LucideIcon {
  if (!iconName) {
    return fallback;
  }

  // Try exact match first
  if (iconMap[iconName]) {
    return iconMap[iconName];
  }

  // Try case-insensitive match
  const lowerIconName = iconName.toLowerCase();
  const matchedKey = Object.keys(iconMap).find(key => key.toLowerCase() === lowerIconName);

  if (matchedKey) {
    return iconMap[matchedKey];
  }

  // Return fallback if no match found
  console.warn(`Icon "${iconName}" not found in icon map, using fallback`);
  return fallback;
}

/**
 * Check if an icon exists in the map
 * @param iconName - The name of the icon
 * @returns true if the icon exists, false otherwise
 */
export function hasIcon(iconName?: string): boolean {
  if (!iconName) return false;

  const lowerIconName = iconName.toLowerCase();
  return Object.keys(iconMap).some(key => key.toLowerCase() === lowerIconName);
}
