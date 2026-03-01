/**
 * Utility function to merge class names
 * Useful for combining Tailwind classes conditionally
 */
export function cn(...classes: (string | undefined | null | false)[]): string {
  return classes.filter(Boolean).join(' ');
}
