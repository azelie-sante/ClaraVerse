import { useState, useEffect } from 'react';
import { getCachedImageUrl } from '@/services/imageCache';

/**
 * React hook to get a cached image from IndexedDB
 * Falls back to backend URL if not cached
 *
 * @param file_id - UUID of the cached image
 * @param fallbackUrl - Backend URL to use if image not in cache
 * @returns Object with imageUrl (string or null) and loading state
 */
export function useCachedImage(file_id: string | undefined, fallbackUrl?: string) {
  const [imageUrl, setImageUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let isMounted = true;
    let createdObjectUrl: string | null = null;

    async function loadImage() {
      if (!file_id) {
        setImageUrl(fallbackUrl || null);
        setLoading(false);
        return;
      }

      try {
        setLoading(true);

        // Try to get from cache first
        const cachedUrl = await getCachedImageUrl(file_id);

        if (isMounted) {
          if (cachedUrl) {
            // Image found in cache
            setImageUrl(cachedUrl);
            createdObjectUrl = cachedUrl;
          } else if (fallbackUrl) {
            // Not in cache, use fallback URL
            setImageUrl(fallbackUrl);
          } else {
            setImageUrl(null);
          }
          setLoading(false);
        }
      } catch (error) {
        console.error('[useCachedImage] Failed to load image:', error);
        if (isMounted) {
          // On error, fallback to backend URL
          setImageUrl(fallbackUrl || null);
          setLoading(false);
        }
      }
    }

    loadImage();

    // Cleanup: Revoke ObjectURL to prevent memory leaks
    return () => {
      isMounted = false;
      if (createdObjectUrl) {
        URL.revokeObjectURL(createdObjectUrl);
      }
    };
  }, [file_id, fallbackUrl]);

  return { imageUrl, loading };
}
