/**
 * ArtifactPreviewModal Component
 *
 * A modal dialog for previewing artifacts in the gallery.
 * Features:
 * - Full preview with renderers (HTML, SVG, Mermaid, Image)
 * - Download options (Source, PNG)
 * - Open in Chat navigation
 * - Preview/Code toggle
 * - Responsive design for mobile and desktop
 */

import { useState, useRef, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  X,
  Download,
  FileCode,
  Image,
  GitBranch,
  ChevronDown,
  ImageIcon,
  Eye,
  Code,
  MessageSquare,
  Maximize2,
  Minimize2,
} from 'lucide-react';
import { toPng } from 'html-to-image';
import { Tooltip } from '@/components/design-system/Tooltip/Tooltip';
import { CodeBlock } from '@/components/design-system/content/CodeBlock/CodeBlock';
import { getRenderer, ImageRenderer } from './renderers';
import { getArtifactExtension, getArtifactMimeType } from '@/utils/artifactParser';
import type { Artifact, ArtifactType } from '@/types/artifact';
import styles from './ArtifactPreviewModal.module.css';

const ARTIFACT_ICONS: Record<ArtifactType, React.ComponentType<{ size?: number }>> = {
  html: FileCode,
  svg: Image,
  mermaid: GitBranch,
  image: ImageIcon,
};

interface ArtifactPreviewModalProps {
  artifact: Artifact | null;
  chatId?: string;
  chatTitle?: string;
  isOpen: boolean;
  onClose: () => void;
  onNavigateToChat?: (chatId: string) => void;
}

export function ArtifactPreviewModal({
  artifact,
  chatId,
  chatTitle,
  isOpen,
  onClose,
  onNavigateToChat,
}: ArtifactPreviewModalProps) {
  const [viewMode, setViewMode] = useState<'preview' | 'code'>('preview');
  const [showDownloadMenu, setShowDownloadMenu] = useState(false);
  const [hideControls, setHideControls] = useState(false);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const downloadMenuRef = useRef<HTMLDivElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);

  const RendererComponent = artifact ? getRenderer(artifact.type) : null;
  const IconComponent = artifact ? ARTIFACT_ICONS[artifact.type] : FileCode;
  const supportsCodeView = artifact && artifact.type !== 'image';

  // Reset state when modal opens/closes and lock body scroll
  useEffect(() => {
    if (isOpen) {
      setViewMode('preview');
      setShowDownloadMenu(false);
      setIsFullscreen(false);
      // Lock body scroll when modal is open
      document.body.style.overflow = 'hidden';
    } else {
      // Restore body scroll when modal closes
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

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

  // Handle escape key
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (isFullscreen) {
          setIsFullscreen(false);
        } else {
          onClose();
        }
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      return () => document.removeEventListener('keydown', handleEscape);
    }
  }, [isOpen, isFullscreen, onClose]);

  const handleDownloadSource = () => {
    if (!artifact) return;

    const blob = new Blob([artifact.content], {
      type: getArtifactMimeType(artifact.type),
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${artifact.title}${getArtifactExtension(artifact.type)}`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    setShowDownloadMenu(false);
  };

  const handleDownloadPNG = async () => {
    if (!artifact) return;

    try {
      // For image artifacts, directly download the base64 image data
      if (artifact.type === 'image' && artifact.images && artifact.images.length > 0) {
        const currentImage = artifact.images[0]; // Get first image
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
          link.download = `${artifact.title}.${currentImage.format || 'png'}`;
          link.href = url;
          document.body.appendChild(link);
          link.click();
          document.body.removeChild(link);
          URL.revokeObjectURL(url);

          setShowDownloadMenu(false);
          return;
        }
      }

      // For non-image artifacts (SVG, Mermaid), use html-to-image
      if (!contentRef.current) return;

      setHideControls(true);
      await new Promise(resolve => setTimeout(resolve, 100));

      const dataUrl = await toPng(contentRef.current, {
        quality: 1.0,
        pixelRatio: 2,
        backgroundColor: '#ffffff',
        cacheBust: true,
      });

      const link = document.createElement('a');
      link.download = `${artifact.title}.png`;
      link.href = dataUrl;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      setShowDownloadMenu(false);
    } catch (error) {
      console.error('Error exporting PNG:', error);
      alert(`Failed to export PNG: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setHideControls(false);
    }
  };

  const handleOpenInChat = () => {
    if (chatId && onNavigateToChat) {
      onNavigateToChat(chatId);
      onClose();
    }
  };

  if (!isOpen || !artifact) {
    return null;
  }

  return (
    <AnimatePresence>
      <div className={styles.overlay} onClick={onClose}>
        <motion.div
          className={`${styles.modal} ${isFullscreen ? styles.fullscreen : ''}`}
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.95 }}
          transition={{ duration: 0.2 }}
          onClick={e => e.stopPropagation()}
        >
          {/* Header */}
          <div className={styles.header}>
            <div className={styles.headerLeft}>
              <IconComponent size={18} className={styles.artifactIcon} />
              <div className={styles.titleGroup}>
                <h3 className={styles.title}>{artifact.title}</h3>
                {chatTitle && <span className={styles.chatTitle}>from {chatTitle}</span>}
              </div>
            </div>

            <div className={styles.headerActions}>
              {/* Preview/Code Toggle */}
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
                    {artifact.type !== 'image' && (
                      <button className={styles.downloadOption} onClick={handleDownloadSource}>
                        <FileCode size={16} />
                        <span>Download Source</span>
                      </button>
                    )}
                    {(artifact.type === 'svg' ||
                      artifact.type === 'mermaid' ||
                      artifact.type === 'image') && (
                      <button className={styles.downloadOption} onClick={handleDownloadPNG}>
                        <Image size={16} />
                        <span>Download PNG</span>
                      </button>
                    )}
                  </div>
                )}
              </div>

              {/* Open in Chat */}
              {chatId && onNavigateToChat && (
                <Tooltip content="Open in Chat" position="bottom">
                  <button
                    className={styles.openInChatButton}
                    onClick={handleOpenInChat}
                    aria-label="Open in chat"
                  >
                    <MessageSquare size={16} />
                    <span className={styles.openInChatText}>Open in Chat</span>
                  </button>
                </Tooltip>
              )}

              {/* Fullscreen Toggle */}
              <Tooltip content={isFullscreen ? 'Exit Fullscreen' : 'Fullscreen'} position="bottom">
                <button
                  className={styles.actionButton}
                  onClick={() => setIsFullscreen(!isFullscreen)}
                  aria-label={isFullscreen ? 'Exit fullscreen' : 'Enter fullscreen'}
                >
                  {isFullscreen ? <Minimize2 size={18} /> : <Maximize2 size={18} />}
                </button>
              </Tooltip>

              {/* Close */}
              <Tooltip content="Close" position="bottom">
                <button className={styles.actionButton} onClick={onClose} aria-label="Close">
                  <X size={18} />
                </button>
              </Tooltip>
            </div>
          </div>

          {/* Content */}
          <div ref={contentRef} className={styles.content}>
            {viewMode === 'code' && supportsCodeView ? (
              <div className={styles.codeView}>
                <CodeBlock
                  code={artifact.content}
                  language={artifact.type === 'mermaid' ? 'markdown' : artifact.type}
                  showLineNumbers={true}
                />
              </div>
            ) : artifact.type === 'image' ? (
              <ImageRenderer artifact={artifact} hideControls={hideControls} />
            ) : RendererComponent ? (
              <RendererComponent content={artifact.content} hideControls={hideControls} />
            ) : (
              <div className={styles.noRenderer}>
                <FileCode size={48} />
                <p>No renderer available for this artifact type</p>
              </div>
            )}
          </div>
        </motion.div>
      </div>
    </AnimatePresence>
  );
}
