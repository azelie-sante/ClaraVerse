/**
 * Image Gallery Store
 *
 * Manages the state for the mobile image gallery modal.
 * Collects images from chat artifacts and allows opening the gallery at any image.
 */

import { create } from 'zustand';
import type { GalleryImage } from '@/components/chat/ImageGalleryModal';

interface ImageGalleryState {
  /** Whether the gallery modal is open */
  isOpen: boolean;

  /** All images available in the gallery */
  images: GalleryImage[];

  /** Index of the initially selected image */
  initialIndex: number;

  /** Open the gallery with a specific set of images */
  openGallery: (images: GalleryImage[], initialIndex?: number) => void;

  /** Close the gallery */
  closeGallery: () => void;

  /** Set images without opening */
  setImages: (images: GalleryImage[]) => void;
}

export const useImageGalleryStore = create<ImageGalleryState>(set => ({
  isOpen: false,
  images: [],
  initialIndex: 0,

  openGallery: (images, initialIndex = 0) =>
    set({
      isOpen: true,
      images,
      initialIndex,
    }),

  closeGallery: () =>
    set({
      isOpen: false,
    }),

  setImages: images =>
    set({
      images,
    }),
}));
