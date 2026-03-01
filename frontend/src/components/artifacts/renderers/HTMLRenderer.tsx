/**
 * HTMLRenderer Component
 *
 * Renders HTML artifacts in a sandboxed iframe.
 * Supports interactive HTML with CSS and JavaScript.
 * Security is provided by the iframe's sandbox attribute.
 */

import { memo } from 'react';
import styles from './HTMLRenderer.module.css';

interface HTMLRendererProps {
  content: string;
  /** Allow JavaScript execution (default: true) */
  allowScripts?: boolean;
  /** Whether to hide zoom controls (not used in HTMLRenderer, but kept for consistency) */
  hideControls?: boolean;
}

export const HTMLRenderer = memo(function HTMLRenderer({
  content,
  allowScripts = true,
}: HTMLRendererProps) {
  return (
    <iframe
      className={styles.iframe}
      srcDoc={content}
      sandbox={
        allowScripts
          ? 'allow-scripts allow-same-origin allow-forms allow-modals allow-popups allow-popups-to-escape-sandbox allow-downloads'
          : 'allow-same-origin'
      }
      title="HTML Artifact"
    />
  );
});
