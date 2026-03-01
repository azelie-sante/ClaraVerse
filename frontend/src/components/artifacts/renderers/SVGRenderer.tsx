/**
 * SVGRenderer Component
 *
 * Renders SVG artifacts in a scrollable container.
 * Validates and displays scalable vector graphics.
 */

import { useState, useEffect, memo } from 'react';
import { AlertCircle } from 'lucide-react';
import styles from './SVGRenderer.module.css';

interface SVGRendererProps {
  content: string;
  /** Whether to hide zoom controls (not used in SVGRenderer, but kept for consistency) */
  hideControls?: boolean;
}

export const SVGRenderer = memo(function SVGRenderer({ content }: SVGRendererProps) {
  const [isValid, setIsValid] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Validate SVG content
    try {
      const parser = new DOMParser();
      const doc = parser.parseFromString(content, 'image/svg+xml');
      const parserError = doc.querySelector('parsererror');

      if (parserError) {
        setIsValid(false);
        setError('Invalid SVG markup');
      } else {
        setIsValid(true);
        setError(null);
      }
    } catch (err) {
      setIsValid(false);
      setError(err instanceof Error ? err.message : 'Failed to parse SVG');
    }
  }, [content]);

  if (!isValid || error) {
    return (
      <div className={styles.error}>
        <AlertCircle size={24} />
        <p>Failed to render SVG</p>
        <span>{error}</span>
      </div>
    );
  }

  return (
    <div className={styles.svgContainer}>
      <div className={styles.svgWrapper} dangerouslySetInnerHTML={{ __html: content }} />
    </div>
  );
});
