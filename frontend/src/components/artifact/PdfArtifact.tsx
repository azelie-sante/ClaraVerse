import { useState } from 'react';
import { Document, Page, pdfjs } from 'react-pdf';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { ArtifactContainer } from './ArtifactContainer';

// Import required CSS for react-pdf
import 'react-pdf/dist/Page/AnnotationLayer.css';
import 'react-pdf/dist/Page/TextLayer.css';

// Configure PDF.js worker
pdfjs.GlobalWorkerOptions.workerSrc = `https://unpkg.com/pdfjs-dist@${pdfjs.version}/legacy/build/pdf.worker.min.mjs`;

interface PdfArtifactProps {
  url: string;
  title?: string;
  filename?: string;
}

export function PdfArtifact({ url, title, filename }: PdfArtifactProps) {
  const [numPages, setNumPages] = useState<number>(0);
  const [pageNumber, setPageNumber] = useState<number>(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const onDocumentLoadSuccess = ({ numPages }: { numPages: number }) => {
    setNumPages(numPages);
    setLoading(false);
    setError(null);
  };

  const onDocumentLoadError = (error: Error) => {
    setError(error.message);
    setLoading(false);
  };

  // Check if error is due to expired/404 link
  const isExpiredLink =
    error && (error.includes('404') || error.includes('Unexpected server response'));

  const goToPrevPage = () => {
    setPageNumber(prev => Math.max(1, prev - 1));
  };

  const goToNextPage = () => {
    setPageNumber(prev => Math.min(numPages, prev + 1));
  };

  // Page navigation toolbar
  const toolbar = numPages > 0 && (
    <>
      <button
        onClick={goToPrevPage}
        disabled={pageNumber <= 1}
        className="p-1.5 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        style={{
          color: 'var(--color-text-secondary)',
        }}
        onMouseEnter={e =>
          !e.currentTarget.disabled &&
          (e.currentTarget.style.background = 'var(--color-surface-hover)')
        }
        onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
        aria-label="Previous page"
      >
        <ChevronLeft className="w-4 h-4" />
      </button>
      <span
        className="text-sm font-medium min-w-[80px] text-center"
        style={{ color: 'var(--color-text-primary)' }}
      >
        {pageNumber} / {numPages}
      </span>
      <button
        onClick={goToNextPage}
        disabled={pageNumber >= numPages}
        className="p-1.5 rounded transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        style={{
          color: 'var(--color-text-secondary)',
        }}
        onMouseEnter={e =>
          !e.currentTarget.disabled &&
          (e.currentTarget.style.background = 'var(--color-surface-hover)')
        }
        onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
        aria-label="Next page"
      >
        <ChevronRight className="w-4 h-4" />
      </button>
    </>
  );

  return (
    <ArtifactContainer
      title={title || filename || 'PDF Preview'}
      downloadUrl={url}
      filename={filename}
      toolbar={toolbar}
      error={error}
    >
      {isFullscreen => (
        <div
          className={`flex flex-col items-center justify-center ${isFullscreen ? 'h-full' : ''} ${isExpiredLink ? 'min-h-[60px]' : ''}`}
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
              <span className="text-sm">Loading PDF...</span>
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
                <p className="font-medium">Failed to load PDF</p>
                <p className="text-sm mt-1">{error}</p>
              </div>
            ))}

          <Document
            file={url}
            onLoadSuccess={onDocumentLoadSuccess}
            onLoadError={onDocumentLoadError}
            loading=""
            error=""
          >
            <Page
              pageNumber={pageNumber}
              renderTextLayer={true}
              renderAnnotationLayer={true}
              className="shadow-lg"
              width={isFullscreen ? Math.min(window.innerWidth - 100, 1400) : 750}
            />
          </Document>
        </div>
      )}
    </ArtifactContainer>
  );
}
