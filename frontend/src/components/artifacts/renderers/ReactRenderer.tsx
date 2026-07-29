/**
 * ReactRenderer Component
 *
 * Renders a single React component (the artifact content) in a sandboxed iframe.
 * The component is transpiled in-browser (Babel) and mounted with React + Tailwind
 * + charts/icons loaded from pinned CDNs — see buildReactSrcDoc.
 *
 * Security: the iframe uses `sandbox="allow-scripts"` WITHOUT `allow-same-origin`,
 * so the frame is origin-null and cannot read the parent's cookies, DOM, or
 * localStorage. (This is stricter than the HTML renderer on purpose — React
 * artifacts are model-generated code.)
 */

import { memo, useMemo } from 'react';
import { buildReactSrcDoc } from '@/utils/reactSrcDoc';
import styles from './HTMLRenderer.module.css';

interface ReactRendererProps {
  content: string;
  /** Kept for signature parity with the other renderers (unused). */
  allowScripts?: boolean;
  hideControls?: boolean;
}

export const ReactRenderer = memo(function ReactRenderer({ content }: ReactRendererProps) {
  const srcDoc = useMemo(() => buildReactSrcDoc(content), [content]);
  return (
    <iframe
      className={styles.iframe}
      srcDoc={srcDoc}
      // No allow-same-origin: the frame is origin-null and can't touch the parent.
      sandbox="allow-scripts allow-popups allow-popups-to-escape-sandbox"
      title="React Artifact"
    />
  );
});
