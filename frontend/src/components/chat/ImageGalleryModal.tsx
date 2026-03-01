/**
 * ImageGalleryModal Component
 *
 * A mobile-friendly fullscreen image gallery modal with:
 * - Swipe navigation between images
 * - Pinch-to-zoom and pan support
 * - Image counter indicator
 * - Download functionality
 */

import { useState, useEffect, useRef, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { motion, AnimatePresence } from 'framer-motion';
import type { PanInfo } from 'framer-motion';
import { X, Download, ChevronLeft, ChevronRight, ZoomIn, ZoomOut } from 'lucide-react';
import styles from './ImageGalleryModal.module.css';

export interface GalleryImage {
  /** Base64 data or URL */
  src: string;
  /** Image title/caption */
  title?: string;
  /** Image format for download */
  format?: string;
}

interface ImageGalleryModalProps {
  images: GalleryImage[];
  initialIndex?: number;
  isOpen: boolean;
  onClose: () => void;
}

export function ImageGalleryModal({
  images,
  initialIndex = 0,
  isOpen,
  onClose,
}: ImageGalleryModalProps) {
  const [currentIndex, setCurrentIndex] = useState(initialIndex);
  const [scale, setScale] = useState(1);
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const imageRef = useRef<HTMLImageElement>(null);

  // Reset state when modal opens
  useEffect(() => {
    if (isOpen) {
      setCurrentIndex(initialIndex);
      setScale(1);
      setPosition({ x: 0, y: 0 });
    }
  }, [isOpen, initialIndex]);

  // Handle keyboard navigation
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case 'Escape':
          onClose();
          break;
        case 'ArrowLeft':
          goToPrevious();
          break;
        case 'ArrowRight':
          goToNext();
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, currentIndex, images.length]);

  // Prevent body scroll when modal is open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  const goToPrevious = useCallback(() => {
    setScale(1);
    setPosition({ x: 0, y: 0 });
    setCurrentIndex(prev => (prev > 0 ? prev - 1 : images.length - 1));
  }, [images.length]);

  const goToNext = useCallback(() => {
    setScale(1);
    setPosition({ x: 0, y: 0 });
    setCurrentIndex(prev => (prev < images.length - 1 ? prev + 1 : 0));
  }, [images.length]);

  const handleDragEnd = (_: MouseEvent | TouchEvent | PointerEvent, info: PanInfo) => {
    setIsDragging(false);

    // If zoomed in, allow panning
    if (scale > 1) {
      return;
    }

    // Swipe threshold for navigation
    const swipeThreshold = 50;
    const velocityThreshold = 500;

    if (info.offset.x > swipeThreshold || info.velocity.x > velocityThreshold) {
      goToPrevious();
    } else if (info.offset.x < -swipeThreshold || info.velocity.x < -velocityThreshold) {
      goToNext();
    }
  };

  const handleZoomIn = () => {
    setScale(prev => Math.min(prev + 0.5, 3));
  };

  const handleZoomOut = () => {
    const newScale = Math.max(scale - 0.5, 1);
    setScale(newScale);
    if (newScale === 1) {
      setPosition({ x: 0, y: 0 });
    }
  };

  const handleDoubleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (scale > 1) {
      setScale(1);
      setPosition({ x: 0, y: 0 });
    } else {
      setScale(2);
    }
  };

  const handleDownload = () => {
    const image = images[currentIndex];
    if (!image) return;

    const link = document.createElement('a');
    link.href = image.src;
    const format = image.format || 'png';
    const title = image.title?.replace(/[^a-z0-9]/gi, '_').toLowerCase() || 'image';
    link.download = `${title}.${format}`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const handleBackdropClick = (e: React.MouseEvent) => {
    // Only close if clicking the backdrop, not the image
    if (e.target === e.currentTarget && !isDragging) {
      onClose();
    }
  };

  if (!isOpen || images.length === 0) {
    return null;
  }

  const currentImage = images[currentIndex];

  // Use createPortal to render at document body level, escaping any stacking contexts
  return createPortal(
    <AnimatePresence>
      <motion.div
        className={styles.overlay}
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        transition={{ duration: 0.2 }}
        onClick={handleBackdropClick}
      >
        {/* Header */}
        <div className={styles.header}>
          <div className={styles.counter}>
            {currentIndex + 1} / {images.length}
          </div>
          <div className={styles.headerActions}>
            <button
              className={styles.headerButton}
              onClick={handleZoomOut}
              disabled={scale <= 1}
              aria-label="Zoom out"
            >
              <ZoomOut size={20} />
            </button>
            <button
              className={styles.headerButton}
              onClick={handleZoomIn}
              disabled={scale >= 3}
              aria-label="Zoom in"
            >
              <ZoomIn size={20} />
            </button>
            <button
              className={styles.headerButton}
              onClick={handleDownload}
              aria-label="Download image"
            >
              <Download size={20} />
            </button>
            <button className={styles.closeButton} onClick={onClose} aria-label="Close gallery">
              <X size={24} />
            </button>
          </div>
        </div>

        {/* Image Container */}
        <div ref={containerRef} className={styles.imageContainer}>
          <motion.div
            className={styles.imageWrapper}
            drag={scale > 1 ? true : 'x'}
            dragConstraints={
              scale > 1
                ? {
                    left: -((scale - 1) * 150),
                    right: (scale - 1) * 150,
                    top: -((scale - 1) * 150),
                    bottom: (scale - 1) * 150,
                  }
                : { left: 0, right: 0 }
            }
            dragElastic={scale > 1 ? 0.1 : 0.5}
            onDragStart={() => setIsDragging(true)}
            onDragEnd={handleDragEnd}
            style={{
              scale,
              x: position.x,
              y: position.y,
            }}
          >
            <motion.img
              ref={imageRef}
              key={currentIndex}
              src={currentImage.src}
              alt={currentImage.title || `Image ${currentIndex + 1}`}
              className={styles.image}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ duration: 0.2 }}
              onDoubleClick={handleDoubleClick}
              draggable={false}
            />
          </motion.div>
        </div>

        {/* Navigation Arrows (desktop) */}
        {images.length > 1 && (
          <>
            <button
              className={`${styles.navButton} ${styles.navPrev}`}
              onClick={e => {
                e.stopPropagation();
                goToPrevious();
              }}
              aria-label="Previous image"
            >
              <ChevronLeft size={32} />
            </button>
            <button
              className={`${styles.navButton} ${styles.navNext}`}
              onClick={e => {
                e.stopPropagation();
                goToNext();
              }}
              aria-label="Next image"
            >
              <ChevronRight size={32} />
            </button>
          </>
        )}

        {/* Dots indicator */}
        {images.length > 1 && (
          <div className={styles.dotsContainer}>
            {images.map((_, index) => (
              <button
                key={index}
                className={`${styles.dot} ${index === currentIndex ? styles.dotActive : ''}`}
                onClick={e => {
                  e.stopPropagation();
                  setCurrentIndex(index);
                  setScale(1);
                  setPosition({ x: 0, y: 0 });
                }}
                aria-label={`Go to image ${index + 1}`}
              />
            ))}
          </div>
        )}

        {/* Caption */}
        {currentImage.title && <div className={styles.caption}>{currentImage.title}</div>}
      </motion.div>
    </AnimatePresence>,
    document.body
  );
}
