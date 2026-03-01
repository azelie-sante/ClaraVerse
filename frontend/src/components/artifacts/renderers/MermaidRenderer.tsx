/**
 * MermaidRenderer Component
 *
 * Renders Mermaid diagrams in a scrollable container.
 * Supports flowcharts, sequence diagrams, gantt charts, and more.
 */

import { useEffect, useRef, useState, memo } from 'react';
import mermaid from 'mermaid';
import { AlertCircle } from 'lucide-react';
import { ZoomPanContainer } from '../ZoomPanContainer';
import styles from './MermaidRenderer.module.css';

interface MermaidRendererProps {
  content: string;
  /** Whether to hide zoom controls (for capturing clean screenshots) */
  hideControls?: boolean;
}

// Initialize Mermaid with dark theme
mermaid.initialize({
  startOnLoad: false,
  theme: 'dark',
  themeVariables: {
    primaryColor: '#e91e63',
    primaryTextColor: '#fff',
    primaryBorderColor: '#e91e63',
    lineColor: '#e91e63',
    secondaryColor: '#1a1a1a',
    tertiaryColor: '#2a2a2a',
    background: '#0a0a0a',
    mainBkg: '#1a1a1a',
    secondBkg: '#2a2a2a',
    clusterBkg: '#2a2a2a',
    clusterBorder: '#e91e63',
    textColor: '#ffffff',
    fontSize: '16px',
  },
});

let diagramIdCounter = 0;

export const MermaidRenderer = memo(function MermaidRenderer({
  content,
  hideControls = false,
}: MermaidRendererProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isRendering, setIsRendering] = useState(true);
  const [svgContent, setSvgContent] = useState<string>('');

  useEffect(() => {
    if (!content) return;

    const diagramId = `mermaid-diagram-${++diagramIdCounter}`;

    const renderDiagram = async () => {
      setIsRendering(true);
      setError(null);
      setSvgContent(''); // Clear previous content

      try {
        // Remove any existing error elements from previous renders
        const existingErrors = document.querySelectorAll(`#${diagramId}`);
        existingErrors.forEach(el => el.remove());

        const { svg } = await mermaid.render(diagramId, content);
        setSvgContent(svg);
        setIsRendering(false);
      } catch (err) {
        console.error('Mermaid rendering error:', err);

        // Clean up any error SVG elements that Mermaid might have created
        const errorElements = document.querySelectorAll(`#${diagramId}`);
        errorElements.forEach(el => el.remove());

        // Extract meaningful error message
        let errorMessage = 'Failed to render Mermaid diagram';
        if (err instanceof Error) {
          // Clean up the error message to be more user-friendly
          errorMessage = err.message
            .replace(/^Error: /, '')
            .replace(/Parse error on line \d+:\n/, 'Syntax error: ')
            .trim();
        }

        setError(errorMessage);
        setIsRendering(false);
      }
    };

    renderDiagram();
  }, [content]);

  if (error) {
    return (
      <div className={styles.error}>
        <AlertCircle size={24} />
        <p>Failed to render Mermaid diagram</p>
        <span>{error}</span>
        <details className={styles.errorDetails}>
          <summary>View diagram source</summary>
          <pre>{content}</pre>
        </details>
      </div>
    );
  }

  if (isRendering) {
    return (
      <div className={styles.loading}>
        <div className={styles.spinner} />
        <p>Rendering diagram...</p>
      </div>
    );
  }

  return (
    <div className={styles.mermaidContainer}>
      <ZoomPanContainer minScale={0.5} maxScale={4} hideControls={hideControls}>
        <div
          ref={containerRef}
          className={styles.diagramWrapper}
          dangerouslySetInnerHTML={{ __html: svgContent }}
        />
      </ZoomPanContainer>
    </div>
  );
});
