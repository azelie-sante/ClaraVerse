import { openDB, type DBSchema, type IDBPDatabase } from 'idb';

// Database schema
interface CachedImage {
  file_id: string; // Primary key (UUID)
  blob: Blob; // Binary image data
  mime_type: string; // e.g., "image/jpeg"
  filename: string; // Original filename
  size: number; // File size in bytes
  cached_at: number; // Timestamp (for cleanup)
  url: string; // Original backend URL (fallback)
}

interface ImageCacheDB extends DBSchema {
  images: {
    key: string; // file_id
    value: CachedImage;
    indexes: { 'by-cached-at': number };
  };
}

const DB_NAME = 'image-cache';
const DB_VERSION = 1;
const STORE_NAME = 'images';

// Initialize database
let dbPromise: Promise<IDBPDatabase<ImageCacheDB>> | null = null;

function getDB(): Promise<IDBPDatabase<ImageCacheDB>> {
  if (!dbPromise) {
    dbPromise = openDB<ImageCacheDB>(DB_NAME, DB_VERSION, {
      upgrade(db) {
        // Create images store if it doesn't exist
        if (!db.objectStoreNames.contains(STORE_NAME)) {
          const store = db.createObjectStore(STORE_NAME, {
            keyPath: 'file_id',
          });
          // Create index for cleanup queries
          store.createIndex('by-cached-at', 'cached_at');
        }
      },
    });
  }
  return dbPromise;
}

/**
 * Store an image in IndexedDB
 * @param file_id - UUID of the file
 * @param blob - Binary image data
 * @param metadata - Additional metadata (filename, mime_type, size, url)
 */
export async function storeImage(
  file_id: string,
  blob: Blob,
  metadata: {
    filename: string;
    mime_type: string;
    size: number;
    url: string;
  }
): Promise<void> {
  try {
    const db = await getDB();
    const cachedImage: CachedImage = {
      file_id,
      blob,
      mime_type: metadata.mime_type,
      filename: metadata.filename,
      size: metadata.size,
      url: metadata.url,
      cached_at: Date.now(),
    };

    await db.put(STORE_NAME, cachedImage);
    console.log(`[ImageCache] Stored image: ${file_id}`);
  } catch (error) {
    console.error('[ImageCache] Failed to store image:', error);
    throw error;
  }
}

/**
 * Get a cached image by file_id
 * @param file_id - UUID of the file
 * @returns CachedImage or undefined if not found
 */
export async function getImage(file_id: string): Promise<CachedImage | undefined> {
  try {
    const db = await getDB();
    const image = await db.get(STORE_NAME, file_id);
    if (image) {
      console.log(`[ImageCache] Retrieved image: ${file_id}`);
    }
    return image;
  } catch (error) {
    console.error('[ImageCache] Failed to get image:', error);
    return undefined;
  }
}

/**
 * Get an ObjectURL for a cached image
 * @param file_id - UUID of the file
 * @returns ObjectURL string or null if not found
 */
export async function getCachedImageUrl(file_id: string): Promise<string | null> {
  try {
    const image = await getImage(file_id);
    if (image) {
      return URL.createObjectURL(image.blob);
    }
    return null;
  } catch (error) {
    console.error('[ImageCache] Failed to create ObjectURL:', error);
    return null;
  }
}

/**
 * Delete a cached image
 * @param file_id - UUID of the file
 */
export async function deleteImage(file_id: string): Promise<void> {
  try {
    const db = await getDB();
    await db.delete(STORE_NAME, file_id);
    console.log(`[ImageCache] Deleted image: ${file_id}`);
  } catch (error) {
    console.error('[ImageCache] Failed to delete image:', error);
    throw error;
  }
}

/**
 * Cleanup old cached images
 * @param daysOld - Delete images older than this many days (default: 7)
 */
export async function cleanup(daysOld: number = 7): Promise<number> {
  try {
    const db = await getDB();
    const cutoffTime = Date.now() - daysOld * 24 * 60 * 60 * 1000;

    // Get all images
    const allImages = await db.getAll(STORE_NAME);

    // Filter and delete old images
    let deletedCount = 0;
    for (const image of allImages) {
      if (image.cached_at < cutoffTime) {
        await db.delete(STORE_NAME, image.file_id);
        deletedCount++;
      }
    }

    console.log(`[ImageCache] Cleaned up ${deletedCount} old images`);
    return deletedCount;
  } catch (error) {
    console.error('[ImageCache] Failed to cleanup:', error);
    return 0;
  }
}

/**
 * Get cache statistics
 */
export async function getCacheStats(): Promise<{
  count: number;
  totalSize: number;
}> {
  try {
    const db = await getDB();
    const allImages = await db.getAll(STORE_NAME);

    const totalSize = allImages.reduce((sum, img) => sum + img.size, 0);

    return {
      count: allImages.length,
      totalSize,
    };
  } catch (error) {
    console.error('[ImageCache] Failed to get stats:', error);
    return { count: 0, totalSize: 0 };
  }
}

/**
 * Clear all cached images
 */
export async function clearAll(): Promise<void> {
  try {
    const db = await getDB();
    await db.clear(STORE_NAME);
    console.log('[ImageCache] Cleared all images');
  } catch (error) {
    console.error('[ImageCache] Failed to clear all:', error);
    throw error;
  }
}

// Run cleanup on initialization (remove images older than 7 days)
if (typeof window !== 'undefined') {
  cleanup(7).catch(err => console.error('[ImageCache] Initial cleanup failed:', err));
}
