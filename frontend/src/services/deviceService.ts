/**
 * Device Authorization Service
 * Handles OAuth 2.0 Device Authorization Grant (RFC 8628)
 */

import { getApiBaseUrl } from '@/lib/config';
import { authClient } from '@/lib/auth';

const API_BASE_URL = `${getApiBaseUrl()}/api`;

// Get JWT token from the auth client (same source as api.ts)
function getAuthToken(): string | null {
  return authClient.getAccessToken();
}

function createAuthHeaders(): HeadersInit {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  const token = getAuthToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return headers;
}

export interface DeviceInfo {
  device_id: string;
  name: string;
  platform: string;
  client_version: string;
  is_active: boolean;
  is_current: boolean;
  last_active_at: string;
  last_ip: string;
  last_location: string;
  created_at: string;
}

export interface DeviceCodeInfo {
  user_code: string;
  client_info: {
    platform: string;
    version: string;
  };
  ip_address: string;
  created_at: string;
}

export interface DeviceAuthorizeResponse {
  message: string;
  device_info: {
    device_id: string;
    platform: string;
    client_version: string;
  };
}

export interface DeviceListResponse {
  devices: DeviceInfo[];
}

/**
 * Authorize a device using the user code
 * Called when user enters the code in the browser
 */
export async function authorizeDevice(userCode: string): Promise<DeviceAuthorizeResponse> {
  const response = await fetch(`${API_BASE_URL}/device/authorize`, {
    method: 'POST',
    headers: createAuthHeaders(),
    body: JSON.stringify({ user_code: userCode }),
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || error.error || 'Failed to authorize device');
  }

  return response.json();
}

/**
 * List all devices for the current user
 */
export async function listDevices(): Promise<DeviceInfo[]> {
  const response = await fetch(`${API_BASE_URL}/devices`, {
    method: 'GET',
    headers: createAuthHeaders(),
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || error.error || 'Failed to list devices');
  }

  const data: DeviceListResponse = await response.json();
  return data.devices || [];
}

/**
 * Rename a device
 */
export async function renameDevice(deviceId: string, name: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/devices/${deviceId}`, {
    method: 'PUT',
    headers: createAuthHeaders(),
    body: JSON.stringify({ name }),
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || error.error || 'Failed to rename device');
  }
}

/**
 * Revoke a device
 */
export async function revokeDevice(deviceId: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/devices/${deviceId}`, {
    method: 'DELETE',
    headers: createAuthHeaders(),
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.message || error.error || 'Failed to revoke device');
  }
}

/**
 * Format a user code for display (adds dash in middle)
 */
export function formatUserCode(code: string): string {
  // Normalize: uppercase and remove any existing dashes/spaces
  const normalized = code.toUpperCase().replace(/[-\s]/g, '');
  if (normalized.length === 8) {
    return `${normalized.slice(0, 4)}-${normalized.slice(4)}`;
  }
  return normalized;
}

/**
 * Validate user code format
 */
export function isValidUserCode(code: string): boolean {
  const normalized = code.toUpperCase().replace(/[-\s]/g, '');
  // 8 alphanumeric characters (excluding confusable ones)
  return /^[BCDFGHJKMNPQRSTVWXYZ23456789]{8}$/.test(normalized);
}

/**
 * Get platform display name
 */
export function getPlatformName(platform: string): string {
  switch (platform.toLowerCase()) {
    case 'darwin':
      return 'macOS';
    case 'linux':
      return 'Linux';
    case 'windows':
      return 'Windows';
    default:
      return platform || 'Unknown';
  }
}

/**
 * Get platform emoji
 */
export function getPlatformEmoji(platform: string): string {
  switch (platform.toLowerCase()) {
    case 'darwin':
      return '\u{1F34E}';
    case 'linux':
      return '\u{1F427}';
    case 'windows':
      return '\u{1FA9F}';
    default:
      return '\u{1F4BB}';
  }
}

/**
 * Format relative time
 */
export function formatRelativeTime(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins} minute${diffMins === 1 ? '' : 's'} ago`;
  if (diffHours < 24) return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`;
  if (diffDays < 30) return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`;

  return date.toLocaleDateString();
}
