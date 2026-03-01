// Centralized runtime configuration
// Reads from window.__CLARA_CONFIG__ (set by Docker entrypoint) first,
// then falls back to Vite build-time env vars, then sensible defaults.

interface ClaraConfig {
  API_BASE_URL?: string;
  WS_URL?: string;
  APP_NAME?: string;
}

declare global {
  interface Window {
    __CLARA_CONFIG__?: ClaraConfig;
  }
}

/**
 * Get the API base URL for HTTP requests.
 * In all-in-one Docker mode, returns empty string (same-origin requests).
 * In dev mode, uses VITE_API_BASE_URL from .env.
 */
export function getApiBaseUrl(): string {
  return (
    window.__CLARA_CONFIG__?.API_BASE_URL ??
    import.meta.env.VITE_API_BASE_URL ??
    ''
  );
}

/**
 * Get the WebSocket URL.
 * In all-in-one Docker mode, auto-detects from window.location.
 * In dev mode, uses VITE_WS_URL from .env.
 */
export function getWsUrl(): string {
  const configured =
    window.__CLARA_CONFIG__?.WS_URL ?? import.meta.env.VITE_WS_URL ?? '';

  if (configured) return configured;

  // Auto-detect from current page URL (for all-in-one Docker mode)
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.host}`;
}
