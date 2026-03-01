import { forwardRef } from 'react';
import { Link } from 'react-router-dom';
import type { LucideProps } from 'lucide-react';
import './AppCard.css';

export interface AppCardProps {
  icon: React.ComponentType<LucideProps>;
  title: string;
  description: string;
  href: string;
  className?: string;
  /** Compact mode for mobile phone-like grid */
  compact?: boolean;
}

export const AppCard = forwardRef<HTMLAnchorElement, AppCardProps>(
  ({ icon: Icon, title, description, href, className = '', compact = false }, ref) => {
    if (compact) {
      // Phone-style app icon layout
      return (
        <Link ref={ref} to={href} className={`app-icon ${className}`}>
          <div className="app-icon-container">
            <Icon size={28} strokeWidth={1.5} />
          </div>
          <span className="app-icon-label">{title}</span>
        </Link>
      );
    }

    return (
      <Link ref={ref} to={href} className={`app-card ${className}`}>
        <div className="app-card-icon">
          <Icon size={40} strokeWidth={1.5} />
        </div>
        <h3 className="app-card-title">{title}</h3>
        <p className="app-card-description">{description}</p>
      </Link>
    );
  }
);

AppCard.displayName = 'AppCard';
