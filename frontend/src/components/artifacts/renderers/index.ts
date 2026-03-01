/**
 * Artifact Renderers - Barrel Export
 *
 * Provides a centralized registry of artifact renderers using the plug-and-play pattern.
 * To add a new artifact type, simply:
 * 1. Create a new renderer component
 * 2. Import it here
 * 3. Add it to the RENDERERS registry
 */

import { HTMLRenderer } from './HTMLRenderer';
import { SVGRenderer } from './SVGRenderer';
import { MermaidRenderer } from './MermaidRenderer';
import { ImageRenderer } from './ImageRenderer';
import type { ArtifactType, Artifact } from '@/types/artifact';

/**
 * Renderer component props interface (for content-based renderers)
 */
export interface RendererProps {
  content: string;
  /** Whether to hide zoom controls (for capturing clean screenshots) */
  hideControls?: boolean;
}

/**
 * Image renderer props (for artifact-based renderers)
 */
export interface ImageRendererProps {
  artifact: Artifact;
  hideControls?: boolean;
}

/**
 * Content-based renderer component type
 */
export type ContentRendererComponent = React.ComponentType<RendererProps>;

/**
 * Renderer Registry - Maps artifact types to their renderer components
 *
 * Note: ImageRenderer is handled separately in ArtifactPane because it takes
 * an artifact prop instead of content. Only content-based renderers are here.
 */
const CONTENT_RENDERERS: Partial<Record<ArtifactType, ContentRendererComponent>> = {
  html: HTMLRenderer,
  svg: SVGRenderer,
  mermaid: MermaidRenderer,
} as const;

/**
 * Get content-based renderer component for an artifact type
 * Returns null for image type (handled separately in ArtifactPane)
 */
export function getRenderer(type: ArtifactType): ContentRendererComponent | null {
  return CONTENT_RENDERERS[type] || null;
}

/**
 * Check if a content-based renderer exists for the given type
 */
export function hasRenderer(type: string): boolean {
  return type in CONTENT_RENDERERS;
}

// Export individual renderers for direct usage if needed
export { HTMLRenderer, SVGRenderer, MermaidRenderer, ImageRenderer };
