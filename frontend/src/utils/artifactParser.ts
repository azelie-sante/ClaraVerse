/**
 * Artifact Parser Utility
 *
 * Extracts artifact blocks from message content and parses them into structured
 * Artifact objects. Supports two syntax formats:
 *
 * 1. XML-style tags:
 *    <artifact type="html" title="Login Page">
 *      <!DOCTYPE html>...
 *    </artifact>
 *
 * 2. Code fence with metadata:
 *    ```html:artifact:Login Page
 *    <!DOCTYPE html>...
 *    ```
 */

import type { Artifact, ArtifactType, ArtifactExtractionResult } from '@/types/artifact';

/**
 * Generate a unique artifact ID
 */
function generateArtifactId(): string {
  return `artifact_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
}

/**
 * Validate artifact type
 */
function isValidArtifactType(type: string): type is ArtifactType {
  return ['html', 'svg', 'mermaid'].includes(type);
}

/**
 * Extract artifacts using XML-style tag syntax
 * Format: <artifact type="html" title="My Page">content</artifact>
 */
function extractXmlArtifacts(content: string): { artifacts: Artifact[]; cleanedContent: string } {
  const artifacts: Artifact[] = [];
  let cleanedContent = content;

  // Regex to match <artifact> tags with attributes and content
  const artifactRegex =
    /<artifact\s+type=["']([^"']+)["']\s+title=["']([^"']+)["']>([\s\S]*?)<\/artifact>/gi;

  let match;
  while ((match = artifactRegex.exec(content)) !== null) {
    const [fullMatch, type, title, artifactContent] = match;

    if (isValidArtifactType(type)) {
      artifacts.push({
        id: generateArtifactId(),
        type,
        title: title.trim(),
        content: artifactContent.trim(),
        metadata: {
          language: type,
        },
      });

      // Remove the artifact tag from cleaned content
      cleanedContent = cleanedContent.replace(fullMatch, '');
    }
  }

  return { artifacts, cleanedContent };
}

/**
 * Extract artifacts from standard markdown code fences
 * Automatically detects HTML, SVG, and Mermaid code blocks
 * Format: ```html
 *         content
 *         ```
 */
function extractCodeFenceArtifacts(content: string): {
  artifacts: Artifact[];
  cleanedContent: string;
} {
  const artifacts: Artifact[] = [];
  let cleanedContent = content;

  // Regex to match standard code fences
  // - Handles both Unix \n and Windows \r\n line endings
  // - Allows optional whitespace after language identifier
  // - Closing ``` can have optional whitespace/newline before it
  const codeFenceRegex = /```(html|htm|svg|mermaid)[ \t]*[\r\n]+([\s\S]*?)[\r\n]*```/gi;

  let match;
  const processedRanges: Array<{ start: number; end: number }> = [];

  while ((match = codeFenceRegex.exec(content)) !== null) {
    const [fullMatch, language, artifactContent] = match;
    const matchStart = match.index;
    const matchEnd = matchStart + fullMatch.length;

    // Determine artifact type from language (regex already filters to valid types)
    const langLower = language.toLowerCase();
    const type: ArtifactType =
      langLower === 'html' || langLower === 'htm'
        ? 'html'
        : langLower === 'svg'
          ? 'svg'
          : 'mermaid';

    // Check if content looks like actual artifact content (heuristic)
    const trimmedContent = artifactContent.trim();
    const isLikelyArtifact =
      (type === 'html' &&
        (trimmedContent.includes('<!DOCTYPE') ||
          trimmedContent.includes('<html') ||
          trimmedContent.includes('<body') ||
          trimmedContent.includes('<head') ||
          trimmedContent.includes('<div'))) ||
      (type === 'svg' && trimmedContent.includes('<svg')) ||
      (type === 'mermaid' && trimmedContent.length > 20); // Mermaid diagrams are usually substantial

    if (isLikelyArtifact) {
      // Generate a descriptive title
      let title = `${language.toUpperCase()} Document`;

      // Try to extract title from HTML
      if (type === 'html') {
        const titleMatch = trimmedContent.match(/<title>(.*?)<\/title>/i);
        if (titleMatch) {
          title = titleMatch[1].trim();
        }
      }

      artifacts.push({
        id: generateArtifactId(),
        type,
        title,
        content: trimmedContent,
        metadata: {
          language,
          autoFit: type === 'svg' || type === 'mermaid',
        },
      });

      // Mark this range as processed
      processedRanges.push({ start: matchStart, end: matchEnd });
    }
  }

  // Remove processed code fences from content (in reverse order to maintain indices)
  processedRanges.reverse().forEach(({ start, end }) => {
    cleanedContent = cleanedContent.slice(0, start) + cleanedContent.slice(end);
  });

  return { artifacts, cleanedContent };
}

/**
 * Main extraction function - tries both XML and code fence formats
 */
export function extractArtifacts(content: string): ArtifactExtractionResult {
  if (!content || typeof content !== 'string') {
    return {
      artifacts: [],
      cleanedContent: content || '',
    };
  }

  // Try XML-style extraction first
  const xmlResult = extractXmlArtifacts(content);

  // Then try code fence extraction on the cleaned content
  const codeFenceResult = extractCodeFenceArtifacts(xmlResult.cleanedContent);

  // Combine artifacts from both methods
  const allArtifacts = [...xmlResult.artifacts, ...codeFenceResult.artifacts];

  // Clean up any excessive whitespace left after artifact removal
  const finalCleanedContent = codeFenceResult.cleanedContent
    .replace(/\n{3,}/g, '\n\n') // Replace multiple newlines with max 2
    .trim();

  return {
    artifacts: allArtifacts,
    cleanedContent: finalCleanedContent,
  };
}

/**
 * Check if content contains any artifacts
 */
export function hasArtifacts(content: string): boolean {
  if (!content) return false;

  // Quick check for artifact markers (XML tags or artifact-type code fences)
  if (content.includes('<artifact')) return true;

  // Check for standard code fences with artifact languages
  return (
    content.includes('```html') || content.includes('```svg') || content.includes('```mermaid')
  );
}

/**
 * Get file extension for artifact type
 */
export function getArtifactExtension(type: ArtifactType): string {
  const extensions: Record<ArtifactType, string> = {
    html: '.html',
    svg: '.svg',
    mermaid: '.mermaid',
    image: '.png',
  };
  return extensions[type];
}

/**
 * Get MIME type for artifact
 */
export function getArtifactMimeType(type: ArtifactType): string {
  const mimeTypes: Record<ArtifactType, string> = {
    html: 'text/html',
    svg: 'image/svg+xml',
    mermaid: 'text/plain',
    image: 'image/png',
  };
  return mimeTypes[type];
}
