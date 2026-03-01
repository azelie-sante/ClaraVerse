import { useState } from 'react';
import { createPortal } from 'react-dom';
import { Maximize2, Minimize2, Download, ChevronDown, ChevronUp } from 'lucide-react';
import { useIsMobile } from '@/hooks/useIsMobile';

interface ArtifactContainerProps {
  title: string;
  downloadUrl?: string;
  filename?: string;
  children: React.ReactNode | ((isFullscreen: boolean) => React.ReactNode);
  toolbar?: React.ReactNode;
  defaultExpanded?: boolean;
  error?: string | null;
}

export function ArtifactContainer({
  title,
  downloadUrl,
  filename,
  children,
  toolbar,
  defaultExpanded = true,
  error = null,
}: ArtifactContainerProps) {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const isMobile = useIsMobile();

  const handleFullscreen = () => {
    setIsFullscreen(!isFullscreen);
    if (!isFullscreen && !isExpanded) {
      setIsExpanded(true);
    }
  };

  // Mobile: Show download-only UI
  if (isMobile && downloadUrl) {
    // Check if error is due to expired/404 link
    const isExpiredLink =
      error &&
      (error.includes('404') ||
        error.includes('Not Found') ||
        error.includes('Unexpected server response'));

    return (
      <div
        className="artifact-container-mobile p-4 rounded-lg flex items-center justify-between"
        style={{
          background: 'var(--color-surface)',
          backdropFilter: 'var(--backdrop-blur-lg)',
          WebkitBackdropFilter: 'var(--backdrop-blur-lg)',
          borderRadius: 'var(--radius-lg)',
          boxShadow: 'var(--shadow-sm)',
        }}
      >
        <div className="flex items-center gap-3 flex-1 min-w-0">
          <Download className="w-5 h-5 flex-shrink-0" style={{ color: 'var(--color-accent)' }} />
          <div className="flex-1 min-w-0">
            <p
              className="text-sm font-medium truncate"
              style={{ color: 'var(--color-text-primary)' }}
            >
              {title}
            </p>
            <p className="text-xs truncate" style={{ color: 'var(--color-text-secondary)' }}>
              {filename || 'Download file'}
            </p>
          </div>
        </div>
        {isExpiredLink ? (
          <div
            className="px-3 py-1.5 rounded text-xs"
            style={{
              color: 'var(--color-text-secondary)',
              background: 'var(--color-surface-elevated)',
              borderRadius: 'var(--radius-md)',
            }}
          >
            Link expired
          </div>
        ) : (
          <a
            href={downloadUrl}
            download={filename}
            className="px-4 py-2 rounded text-sm font-medium flex items-center gap-2 transition-all flex-shrink-0"
            style={{
              background: 'var(--color-accent)',
              color: 'white',
              borderRadius: 'var(--radius-md)',
            }}
            onMouseEnter={e => {
              e.currentTarget.style.background = 'var(--color-accent-hover)';
              e.currentTarget.style.boxShadow = 'var(--shadow-glow-sm)';
            }}
            onMouseLeave={e => {
              e.currentTarget.style.background = 'var(--color-accent)';
              e.currentTarget.style.boxShadow = 'none';
            }}
          >
            <Download className="w-4 h-4" />
            Download
          </a>
        )}
      </div>
    );
  }

  // Artifact content JSX
  const artifactContent = (
    <div
      className={`
        artifact-container overflow-hidden
        ${isFullscreen ? 'fixed inset-4' : 'max-w-[800px] mx-auto'}
      `}
      style={{
        background: 'var(--color-surface)',
        backdropFilter: 'var(--backdrop-blur-lg)',
        WebkitBackdropFilter: 'var(--backdrop-blur-lg)',
        borderRadius: 'var(--radius-2xl)',
        boxShadow: 'var(--shadow-md)',
        ...(isFullscreen && { zIndex: 9999 }),
      }}
    >
      {/* Header */}
      <div
        className="artifact-header px-3 py-2 flex items-center justify-between gap-2"
        style={{
          background: 'rgba(255, 255, 255, 0.02)',
        }}
      >
        <div className={`flex items-center gap-2 min-w-0 ${isFullscreen ? 'flex-1' : ''}`}>
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="p-1 rounded transition-colors flex-shrink-0"
            style={{
              color: 'var(--color-text-secondary)',
            }}
            onMouseEnter={e => (e.currentTarget.style.background = 'var(--color-surface-hover)')}
            onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
            aria-label={isExpanded ? 'Collapse' : 'Expand'}
          >
            {isExpanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </button>
          <span
            className="truncate text-sm"
            style={{ color: 'var(--color-text-primary)', fontWeight: 500 }}
          >
            {title}
          </span>
        </div>

        <div className="flex items-center gap-1.5 flex-shrink-0">
          {/* Custom toolbar (e.g., page navigation) */}
          {toolbar && <div className="flex items-center gap-1.5">{toolbar}</div>}

          {/* Fullscreen toggle */}
          <button
            onClick={handleFullscreen}
            className="p-1.5 rounded transition-colors"
            style={{
              color: 'var(--color-text-secondary)',
            }}
            onMouseEnter={e => (e.currentTarget.style.background = 'var(--color-surface-hover)')}
            onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
            aria-label={isFullscreen ? 'Exit fullscreen' : 'Fullscreen'}
            title={isFullscreen ? 'Exit fullscreen' : 'Fullscreen'}
          >
            {isFullscreen ? <Minimize2 className="w-4 h-4" /> : <Maximize2 className="w-4 h-4" />}
          </button>

          {/* Download button */}
          {downloadUrl && (
            <a
              href={downloadUrl}
              download={filename}
              className="px-2.5 py-1.5 rounded text-xs font-medium flex items-center gap-1.5 transition-all"
              style={{
                background: 'var(--color-accent)',
                color: 'white',
                borderRadius: 'var(--radius-md)',
              }}
              onMouseEnter={e => {
                e.currentTarget.style.background = 'var(--color-accent-hover)';
                e.currentTarget.style.boxShadow = 'var(--shadow-glow-sm)';
              }}
              onMouseLeave={e => {
                e.currentTarget.style.background = 'var(--color-accent)';
                e.currentTarget.style.boxShadow = 'none';
              }}
            >
              <Download className="w-3.5 h-3.5" />
              Download
            </a>
          )}
        </div>
      </div>

      {/* Content */}
      {isExpanded && (
        <div
          className={`artifact-content ${isFullscreen ? 'h-[calc(100%-3rem)]' : 'p-8'}`}
          style={{
            background: 'rgba(255, 255, 255, 0.02)',
            ...(isFullscreen && {
              overflow: 'hidden',
              display: 'flex',
              flexDirection: 'column',
            }),
          }}
        >
          {typeof children === 'function' ? children(isFullscreen) : children}
        </div>
      )}
    </div>
  );

  // Render with portal when fullscreen to escape stacking context
  if (isFullscreen) {
    return createPortal(
      <>
        {/* Backdrop */}
        <div
          style={{
            position: 'fixed',
            inset: 0,
            background: 'rgba(0, 0, 0, 0.7)',
            backdropFilter: 'blur(4px)',
            zIndex: 9998,
          }}
          onClick={handleFullscreen}
        />
        {/* Artifact Container */}
        {artifactContent}
      </>,
      document.body
    );
  }

  // Normal rendering when not fullscreen
  return artifactContent;
}
