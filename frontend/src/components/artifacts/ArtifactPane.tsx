/**
 * ArtifactPane Component
 *
 * Main container for displaying artifacts with glassmorphism styling.
 * Features:
 * - Tab navigation for multiple artifacts
 * - Dynamic renderer switching based on artifact type
 * - Header with title, type badge, and action buttons
 * - Download and fullscreen support
 */

import {
  X,
  Download,
  Maximize2,
  Minimize2,
  FileCode,
  Image,
  GitBranch,
  ChevronDown,
  ImageIcon,
  Eye,
  Code,
} from 'lucide-react';
import { useState, useRef, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { toPng } from 'html-to-image';
import { useArtifactStore } from '@/store/useArtifactStore';
import { Tooltip } from '@/components/design-system/Tooltip/Tooltip';
import { CodeBlock } from '@/components/design-system/content/CodeBlock/CodeBlock';
import { getRenderer, ImageRenderer } from './renderers';
import { getArtifactExtension, getArtifactMimeType } from '@/utils/artifactParser';
import type { ArtifactType } from '@/types/artifact';
import styles from './ArtifactPane.module.css';

const ARTIFACT_ICONS: Record<ArtifactType, React.ComponentType<{ size?: number }>> = {
  html: FileCode,
  svg: Image,
  mermaid: GitBranch,
  image: ImageIcon,
};

export function ArtifactPane() {
  const isOpen = useArtifactStore(s => s.isOpen);
  const artifacts = useArtifactStore(s => s.artifacts);
  const selectedIndex = useArtifactStore(s => s.selectedIndex);
  const selectArtifact = useArtifactStore(s => s.selectArtifact);
  const closePane = useArtifactStore(s => s.closePane);

  const [isFullscreen, setIsFullscreen] = useState(false);
  const [showDownloadMenu, setShowDownloadMenu] = useState(false);
  const [hideControls, setHideControls] = useState(false);
  const [viewMode, setViewMode] = useState<'preview' | 'code'>('preview');
  const downloadMenuRef = useRef<HTMLDivElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);

  console.log('🎨 ArtifactPane render:', {
    isOpen,
    artifactCount: artifacts.length,
    selectedIndex,
  });

  const selectedArtifact = artifacts[selectedIndex];
  const RendererComponent = selectedArtifact ? getRenderer(selectedArtifact.type) : null;
  const IconComponent = selectedArtifact ? ARTIFACT_ICONS[selectedArtifact.type] : FileCode;

  // Check if artifact supports code view (not images)
  const supportsCodeView = selectedArtifact && selectedArtifact.type !== 'image';

  // Reset to preview mode when switching artifacts
  useEffect(() => {
    setViewMode('preview');
  }, [selectedIndex]);

  // Close download menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (downloadMenuRef.current && !downloadMenuRef.current.contains(event.target as Node)) {
        setShowDownloadMenu(false);
      }
    };

    if (showDownloadMenu) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [showDownloadMenu]);

  const handleDownloadSource = () => {
    if (!selectedArtifact) return;

    const blob = new Blob([selectedArtifact.content], {
      type: getArtifactMimeType(selectedArtifact.type),
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${selectedArtifact.title}${getArtifactExtension(selectedArtifact.type)}`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    setShowDownloadMenu(false);
  };

  const handleDownloadPNG = async () => {
    if (!selectedArtifact) return;

    try {
      console.log('Starting PNG export for', selectedArtifact.type);

      // For image artifacts, directly download the base64 image data
      if (
        selectedArtifact.type === 'image' &&
        selectedArtifact.images &&
        selectedArtifact.images.length > 0
      ) {
        const currentImage = selectedArtifact.images[0]; // Get first image (or could track selected index)
        if (currentImage && currentImage.data) {
          // Convert base64 to blob for download
          const byteCharacters = atob(currentImage.data);
          const byteNumbers = new Array(byteCharacters.length);
          for (let i = 0; i < byteCharacters.length; i++) {
            byteNumbers[i] = byteCharacters.charCodeAt(i);
          }
          const byteArray = new Uint8Array(byteNumbers);
          const blob = new Blob([byteArray], { type: `image/${currentImage.format || 'png'}` });

          // Create download link
          const url = URL.createObjectURL(blob);
          const link = document.createElement('a');
          link.download = `${selectedArtifact.title}.${currentImage.format || 'png'}`;
          link.href = url;
          document.body.appendChild(link);
          link.click();
          document.body.removeChild(link);
          URL.revokeObjectURL(url);

          console.log('Image download triggered');
          setShowDownloadMenu(false);
          return;
        }
      }

      // For non-image artifacts (SVG, Mermaid), use html-to-image
      if (!contentRef.current) return;

      // Hide zoom controls before capturing
      setHideControls(true);

      // Wait for DOM to update (controls to be hidden)
      await new Promise(resolve => setTimeout(resolve, 100));

      // Use html-to-image to capture the rendered content directly
      const dataUrl = await toPng(contentRef.current, {
        quality: 1.0,
        pixelRatio: 2, // 2x for high quality
        backgroundColor: '#ffffff', // White background
        cacheBust: true, // Prevent caching issues
      });

      console.log('PNG generated, downloading...');

      // Download the PNG
      const link = document.createElement('a');
      link.download = `${selectedArtifact.title}.png`;
      link.href = dataUrl;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);

      console.log('PNG download triggered');
      setShowDownloadMenu(false);
    } catch (error) {
      console.error('Error exporting PNG:', error);
      alert(`Failed to export PNG: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      // Always restore controls visibility
      setHideControls(false);
    }
  };

  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen);
  };

  if (!isOpen || artifacts.length === 0) {
    return null;
  }

  return (
    <AnimatePresence>
      <motion.div
        className={`${styles.container} ${isFullscreen ? styles.fullscreen : ''}`}
        initial={{ opacity: 0, x: 20 }}
        animate={{ opacity: 1, x: 0 }}
        exit={{ opacity: 0, x: 20 }}
        transition={{ duration: 0.2 }}
      >
        {/* Header */}
        <div className={styles.header}>
          <div className={styles.headerLeft}>
            <IconComponent size={18} className={styles.artifactIcon} />
            <h3 className={styles.title}>{selectedArtifact?.title || 'Artifact'}</h3>
          </div>

          <div className={styles.headerActions}>
            {/* Preview/Code Toggle - Only for non-image artifacts */}
            {supportsCodeView && (
              <div className={styles.viewToggle}>
                <Tooltip content="Preview" position="bottom">
                  <button
                    className={`${styles.viewToggleButton} ${viewMode === 'preview' ? styles.viewToggleActive : ''}`}
                    onClick={() => setViewMode('preview')}
                    aria-label="Preview mode"
                  >
                    <Eye size={16} />
                  </button>
                </Tooltip>
                <Tooltip content="Code" position="bottom">
                  <button
                    className={`${styles.viewToggleButton} ${viewMode === 'code' ? styles.viewToggleActive : ''}`}
                    onClick={() => setViewMode('code')}
                    aria-label="Code mode"
                  >
                    <Code size={16} />
                  </button>
                </Tooltip>
              </div>
            )}

            {/* Download Menu */}
            <div className={styles.downloadMenu} ref={downloadMenuRef}>
              <Tooltip content="Download" position="bottom">
                <button
                  className={styles.actionButton}
                  onClick={() => setShowDownloadMenu(!showDownloadMenu)}
                  aria-label="Download options"
                >
                  <Download size={18} />
                  <ChevronDown size={12} className={styles.chevron} />
                </button>
              </Tooltip>

              {showDownloadMenu && (
                <div className={styles.downloadDropdown}>
                  {selectedArtifact?.type !== 'image' && (
                    <button className={styles.downloadOption} onClick={handleDownloadSource}>
                      <FileCode size={16} />
                      <span>Download Source</span>
                    </button>
                  )}
                  {(selectedArtifact?.type === 'svg' ||
                    selectedArtifact?.type === 'mermaid' ||
                    selectedArtifact?.type === 'image') && (
                    <button className={styles.downloadOption} onClick={handleDownloadPNG}>
                      <Image size={16} />
                      <span>Download PNG</span>
                    </button>
                  )}
                </div>
              )}
            </div>

            <Tooltip content={isFullscreen ? 'Exit Fullscreen' : 'Fullscreen'} position="bottom">
              <button
                className={styles.actionButton}
                onClick={toggleFullscreen}
                aria-label={isFullscreen ? 'Exit fullscreen' : 'Enter fullscreen'}
              >
                {isFullscreen ? <Minimize2 size={18} /> : <Maximize2 size={18} />}
              </button>
            </Tooltip>

            <Tooltip content="Close" position="bottom">
              <button
                className={styles.actionButton}
                onClick={closePane}
                aria-label="Close artifact pane"
              >
                <X size={18} />
              </button>
            </Tooltip>
          </div>
        </div>

        {/* Tabs (if multiple artifacts) */}
        {artifacts.length > 1 && (
          <div className={styles.tabs}>
            {artifacts.map((artifact, index) => {
              const TabIcon = ARTIFACT_ICONS[artifact.type];
              return (
                <Tooltip key={artifact.id} content={artifact.title} position="bottom">
                  <button
                    className={`${styles.tab} ${index === selectedIndex ? styles.tabActive : ''}`}
                    onClick={() => selectArtifact(index)}
                  >
                    <TabIcon size={14} />
                    <span className={styles.tabTitle}>{artifact.title}</span>
                  </button>
                </Tooltip>
              );
            })}
          </div>
        )}

        {/* Renderer Content */}
        <div ref={contentRef} className={styles.content}>
          {selectedArtifact ? (
            viewMode === 'code' && supportsCodeView ? (
              // Code view - show source code with syntax highlighting
              <div className={styles.codeView}>
                <CodeBlock
                  code={selectedArtifact.content}
                  language={
                    selectedArtifact.type === 'mermaid' ? 'markdown' : selectedArtifact.type
                  }
                  showLineNumbers={true}
                />
              </div>
            ) : selectedArtifact.type === 'image' ? (
              // Image artifacts use the ImageRenderer with full artifact
              <ImageRenderer artifact={selectedArtifact} hideControls={hideControls} />
            ) : RendererComponent ? (
              // Other artifacts use content-based renderers
              <RendererComponent content={selectedArtifact.content} hideControls={hideControls} />
            ) : (
              <div className={styles.noRenderer}>
                <FileCode size={48} />
                <p>No renderer available for this artifact type</p>
              </div>
            )
          ) : (
            <div className={styles.noRenderer}>
              <FileCode size={48} />
              <p>No artifact selected</p>
            </div>
          )}
        </div>
      </motion.div>
    </AnimatePresence>
  );
}
