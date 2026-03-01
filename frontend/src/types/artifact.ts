/**
 * Artifact Type Definitions
 *
 * Artifacts are special content blocks (HTML, SVG, Mermaid diagrams, images) that can be
 * rendered in a separate pane alongside chat messages, similar to Claude's artifact system.
 */

/**
 * Supported artifact types
 * - html: Web pages with embedded CSS and JavaScript
 * - svg: Scalable vector graphics and diagrams
 * - mermaid: Flowcharts, sequence diagrams, gantt charts, etc.
 * - image: Tool-generated images/plots displayed as slides
 */
export type ArtifactType = 'html' | 'svg' | 'mermaid' | 'image';

/**
 * Image data for image artifact type
 */
export interface ArtifactImage {
  /** Base64-encoded image data */
  data: string;
  /** Image format (png, jpg, svg) */
  format: string;
  /** Optional caption for the image */
  caption?: string;
}

/**
 * Artifact metadata and content
 */
export interface Artifact {
  /** Unique identifier for the artifact */
  id: string;

  /** Type of artifact content */
  type: ArtifactType;

  /** Display title/filename (e.g., "Pikachu.svg", "Login Page.html") */
  title: string;

  /** Raw artifact content (for html, svg, mermaid types) */
  content: string;

  /** Array of images (for image type artifacts) */
  images?: ArtifactImage[];

  /** Optional metadata for artifact-specific configuration */
  metadata?: {
    /** Original language from code fence (e.g., "html", "svg", "mermaid") */
    language?: string;

    /** Whether to auto-fit content to viewport (for SVG/Mermaid) */
    autoFit?: boolean;

    /** Custom theme/styling hints */
    theme?: string;

    /** Source tool name (for image artifacts) */
    toolName?: string;

    /** Additional arbitrary metadata */
    [key: string]: unknown;
  };
}

/**
 * Artifact extraction result
 */
export interface ArtifactExtractionResult {
  /** Extracted artifacts from content */
  artifacts: Artifact[];

  /** Original content with artifact markers removed */
  cleanedContent: string;
}
