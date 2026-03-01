/**
 * ZoomPanContainer Component
 *
 * Provides zoom and pan controls for SVG and Mermaid artifacts.
 * Uses react-zoom-pan-pinch for smooth transformations with glassmorphism controls.
 */

import { TransformWrapper, TransformComponent } from 'react-zoom-pan-pinch';
import { ZoomIn, ZoomOut, Maximize2, RotateCcw } from 'lucide-react';
import { Tooltip } from '@/components/design-system/Tooltip/Tooltip';
import styles from './ZoomPanContainer.module.css';

interface ZoomPanContainerProps {
  children: React.ReactNode;
  /** Initial scale (default: 1) */
  initialScale?: number;
  /** Minimum zoom scale (default: 0.1) */
  minScale?: number;
  /** Maximum zoom scale (default: 5) */
  maxScale?: number;
  /** Whether to center content initially (default: true) */
  centerOnInit?: boolean;
  /** Whether to hide zoom controls (for capturing clean screenshots) */
  hideControls?: boolean;
}

export function ZoomPanContainer({
  children,
  initialScale = 1,
  minScale = 0.1,
  maxScale = 5,
  centerOnInit = true,
  hideControls = false,
}: ZoomPanContainerProps) {
  return (
    <TransformWrapper
      initialScale={initialScale}
      minScale={minScale}
      maxScale={maxScale}
      centerOnInit={centerOnInit}
      wheel={{ step: 0.1 }}
      doubleClick={{ mode: 'reset' }}
      panning={{ velocityDisabled: false }}
      limitToBounds={false}
      centerZoomedOut={true}
    >
      {({ zoomIn, zoomOut, resetTransform, centerView }) => (
        <div className={styles.container}>
          {/* Zoom Controls Toolbar */}
          {!hideControls && (
            <div className={styles.controls}>
              <Tooltip content="Zoom In (Ctrl +)" position="bottom">
                <button
                  className={styles.controlButton}
                  onClick={() => zoomIn()}
                  aria-label="Zoom in"
                >
                  <ZoomIn size={18} />
                </button>
              </Tooltip>

              <Tooltip content="Zoom Out (Ctrl -)" position="bottom">
                <button
                  className={styles.controlButton}
                  onClick={() => zoomOut()}
                  aria-label="Zoom out"
                >
                  <ZoomOut size={18} />
                </button>
              </Tooltip>

              <Tooltip content="Fit to Screen" position="bottom">
                <button
                  className={styles.controlButton}
                  onClick={() => centerView()}
                  aria-label="Fit to screen"
                >
                  <Maximize2 size={18} />
                </button>
              </Tooltip>

              <Tooltip content="Reset View (Ctrl 0)" position="bottom">
                <button
                  className={styles.controlButton}
                  onClick={() => resetTransform()}
                  aria-label="Reset view"
                >
                  <RotateCcw size={18} />
                </button>
              </Tooltip>
            </div>
          )}

          {/* Zoomable Content Area */}
          <TransformComponent wrapperClass={styles.wrapper} contentClass={styles.content}>
            {children}
          </TransformComponent>
        </div>
      )}
    </TransformWrapper>
  );
}
