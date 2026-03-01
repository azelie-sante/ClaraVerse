import { useState, useEffect } from 'react';
import { ArtifactContainer } from './ArtifactContainer';

interface HtmlArtifactProps {
  content?: string;
  url?: string;
  title?: string;
  filename?: string;
}

export function HtmlArtifact({ content, url, title, filename }: HtmlArtifactProps) {
  const [htmlContent, setHtmlContent] = useState<string>(content || '');
  const [loading, setLoading] = useState(!!url && !content);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (url && !content) {
      // Fetch HTML content from URL
      fetch(url)
        .then(res => {
          if (!res.ok) throw new Error(`Failed to fetch: ${res.statusText}`);
          return res.text();
        })
        .then(html => {
          setHtmlContent(html);
          setLoading(false);
        })
        .catch(err => {
          setError(err.message);
          setLoading(false);
        });
    }
  }, [url, content]);

  // Check if error is due to expired/404 link
  const isExpiredLink = error && (error.includes('404') || error.includes('Not Found'));

  return (
    <ArtifactContainer
      title={title || filename || 'HTML Preview'}
      downloadUrl={url}
      filename={filename}
      error={error}
    >
      <div
        className={`flex items-center justify-center ${isExpiredLink ? 'min-h-[60px]' : 'min-h-[400px]'}`}
      >
        {loading && (
          <div
            className="flex flex-col items-center gap-2"
            style={{ color: 'var(--color-text-secondary)' }}
          >
            <div
              className="animate-spin rounded-full h-8 w-8"
              style={{
                border: '2px solid transparent',
                borderTopColor: 'var(--color-accent)',
                borderRightColor: 'var(--color-accent)',
              }}
            ></div>
            <span className="text-sm">Loading HTML...</span>
          </div>
        )}

        {error &&
          (isExpiredLink ? (
            // Compact message for expired links
            <div
              className="px-3 py-1.5 rounded text-xs"
              style={{
                color: 'var(--color-text-secondary)',
                background: 'var(--color-surface-elevated)',
                borderRadius: 'var(--radius-md)',
              }}
            >
              Download link expired
            </div>
          ) : (
            // Full error message for other errors
            <div
              className="rounded-lg p-4 max-w-md"
              style={{
                color: 'var(--color-error)',
                background: 'var(--color-error-light)',
                borderRadius: 'var(--radius-lg)',
              }}
            >
              <p className="font-medium">Failed to load HTML</p>
              <p className="text-sm mt-1">{error}</p>
            </div>
          ))}

        {!loading && !error && htmlContent && (
          <iframe
            srcDoc={htmlContent}
            sandbox="allow-scripts allow-same-origin"
            className="w-full h-[600px] rounded"
            style={{
              borderRadius: 'var(--radius-lg)',
              boxShadow: 'var(--shadow-sm)',
              background: 'white',
            }}
            title={title || 'HTML Preview'}
          />
        )}
      </div>
    </ArtifactContainer>
  );
}
