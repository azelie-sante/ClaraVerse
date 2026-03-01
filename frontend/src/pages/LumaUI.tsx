import { Link } from 'react-router-dom';
import { ChevronLeft } from 'lucide-react';
import { Typography } from '@/components/design-system';

export const LumaUI = () => {
  return (
    <div
      style={{
        minHeight: '100vh',
        background: 'var(--color-background)',
      }}
    >
      <div
        style={{
          padding: 'var(--space-6) var(--space-8)',
          borderBottom: '1px solid var(--color-border)',
        }}
      >
        <Link
          to="/"
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 'var(--space-2)',
            color: 'var(--color-text-secondary)',
            textDecoration: 'none',
            fontSize: 'var(--text-base)',
            transition: 'color var(--transition-fast)',
          }}
          onMouseOver={e => (e.currentTarget.style.color = 'var(--color-accent)')}
          onMouseOut={e => (e.currentTarget.style.color = 'var(--color-text-secondary)')}
        >
          <ChevronLeft size={20} />
          Home
        </Link>
      </div>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          minHeight: 'calc(100vh - 80px)',
        }}
      >
        <Typography variant="display" gradient align="center">
          Coming Soon
        </Typography>
      </div>
    </div>
  );
};
