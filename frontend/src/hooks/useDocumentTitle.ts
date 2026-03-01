import { useEffect } from 'react';

/**
 * Custom hook to update the document title
 * @param title - The title to set (will be appended to "Claraverse | ")
 * @param baseTitle - Optional base title (defaults to "Claraverse")
 */
export function useDocumentTitle(title?: string, baseTitle = 'Claraverse') {
  useEffect(() => {
    const previousTitle = document.title;

    if (title) {
      document.title = `${baseTitle} | ${title}`;
    } else {
      document.title = baseTitle;
    }

    // Cleanup function to restore previous title
    return () => {
      document.title = previousTitle;
    };
  }, [title, baseTitle]);
}
