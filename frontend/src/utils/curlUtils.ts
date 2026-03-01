/**
 * Curl import/export utilities for HTTP Request blocks.
 */

interface ParsedCurl {
  method?: string;
  url?: string;
  headers?: Record<string, string>;
  body?: string;
  authType?: string;
  authConfig?: Record<string, string>;
}

/**
 * Tokenize a curl command string, handling quotes and line continuations.
 */
function tokenize(input: string): string[] {
  // Normalize line continuations
  const normalized = input.replace(/\\\s*\n/g, ' ').trim();
  const tokens: string[] = [];
  let current = '';
  let inSingle = false;
  let inDouble = false;
  let escaped = false;

  for (let i = 0; i < normalized.length; i++) {
    const ch = normalized[i];

    if (escaped) {
      current += ch;
      escaped = false;
      continue;
    }

    if (ch === '\\' && !inSingle) {
      escaped = true;
      continue;
    }

    if (ch === "'" && !inDouble) {
      inSingle = !inSingle;
      continue;
    }

    if (ch === '"' && !inSingle) {
      inDouble = !inDouble;
      continue;
    }

    if (/\s/.test(ch) && !inSingle && !inDouble) {
      if (current) {
        tokens.push(current);
        current = '';
      }
      continue;
    }

    current += ch;
  }

  if (current) tokens.push(current);
  return tokens;
}

/**
 * Parse a curl command string into HTTP request config fields.
 */
export function parseCurl(curlString: string): ParsedCurl {
  const tokens = tokenize(curlString);
  if (tokens.length === 0) return {};

  // Strip leading "curl" command
  let start = 0;
  if (tokens[0].toLowerCase() === 'curl') start = 1;

  const result: ParsedCurl = {
    method: 'GET',
    headers: {},
  };

  let i = start;
  while (i < tokens.length) {
    const token = tokens[i];

    if (token === '-X' || token === '--request') {
      i++;
      if (i < tokens.length) result.method = tokens[i].toUpperCase();
    } else if (token === '-H' || token === '--header') {
      i++;
      if (i < tokens.length) {
        const header = tokens[i];
        const colonIdx = header.indexOf(':');
        if (colonIdx > 0) {
          const key = header.substring(0, colonIdx).trim();
          const value = header.substring(colonIdx + 1).trim();
          result.headers![key] = value;
        }
      }
    } else if (token === '-d' || token === '--data' || token === '--data-raw') {
      i++;
      if (i < tokens.length) {
        result.body = tokens[i];
        // Imply POST if no explicit method set
        if (result.method === 'GET') result.method = 'POST';
      }
    } else if (token === '-u' || token === '--user') {
      i++;
      if (i < tokens.length) {
        const parts = tokens[i].split(':');
        result.authType = 'basic';
        result.authConfig = {
          username: parts[0] || '',
          password: parts.slice(1).join(':') || '',
        };
      }
    } else if (token === '-A' || token === '--user-agent') {
      i++;
      if (i < tokens.length) {
        result.headers!['User-Agent'] = tokens[i];
      }
    } else if (!token.startsWith('-')) {
      // Bare argument = URL
      result.url = token;
    }
    // Ignore other flags (-s, -k, -L, --compressed, etc.)

    i++;
  }

  // Detect auth from headers
  if (!result.authType && result.headers) {
    const authHeader = Object.entries(result.headers).find(
      ([k]) => k.toLowerCase() === 'authorization'
    );
    if (authHeader) {
      const [headerKey, value] = authHeader;
      if (value.toLowerCase().startsWith('bearer ')) {
        result.authType = 'bearer';
        result.authConfig = { token: value.substring(7) };
        delete result.headers[headerKey];
      } else if (value.toLowerCase().startsWith('basic ')) {
        try {
          const decoded = atob(value.substring(6));
          const [username, ...passParts] = decoded.split(':');
          result.authType = 'basic';
          result.authConfig = { username, password: passParts.join(':') };
          delete result.headers[headerKey];
        } catch {
          // Not valid base64, keep as-is
        }
      }
    }

    // Check for common API key headers
    if (!result.authType) {
      const apiKeyHeader = Object.entries(result.headers).find(([k]) =>
        /^(x-api-key|api-key|apikey)$/i.test(k)
      );
      if (apiKeyHeader) {
        const [headerKey, value] = apiKeyHeader;
        result.authType = 'api_key';
        result.authConfig = { headerName: headerKey, key: value };
        delete result.headers[headerKey];
      }
    }
  }

  return result;
}

/**
 * Export current HTTP request config as a curl command string.
 */
export function exportCurl(config: {
  method?: string;
  url?: string;
  headers?: Record<string, string>;
  body?: string;
  authType?: string;
  authConfig?: Record<string, string>;
}): string {
  const parts: string[] = ['curl'];

  const method = config.method || 'GET';
  if (method !== 'GET') {
    parts.push(`-X ${method}`);
  }

  parts.push(`'${config.url || ''}'`);

  // Headers
  if (config.headers) {
    for (const [key, value] of Object.entries(config.headers)) {
      if (key && value !== undefined) {
        parts.push(`-H '${key}: ${value}'`);
      }
    }
  }

  // Auth
  if (config.authType === 'bearer' && config.authConfig?.token) {
    parts.push(`-H 'Authorization: Bearer ${config.authConfig.token}'`);
  } else if (config.authType === 'basic' && config.authConfig?.username) {
    parts.push(`-u '${config.authConfig.username}:${config.authConfig.password || ''}'`);
  } else if (config.authType === 'api_key' && config.authConfig?.key) {
    const headerName = config.authConfig.headerName || 'X-API-Key';
    parts.push(`-H '${headerName}: ${config.authConfig.key}'`);
  }

  // Body
  if (config.body) {
    parts.push(`-d '${config.body.replace(/'/g, "'\\''")}'`);
  }

  // Multi-line format with continuations
  if (parts.length <= 3) {
    return parts.join(' ');
  }
  return parts.join(' \\\n  ');
}

/**
 * Parse query parameters from a URL string into a KV array.
 */
export function parseQueryParamsFromUrl(url: string): Array<{ key: string; value: string }> {
  try {
    const parsed = new URL(url);
    const params: Array<{ key: string; value: string }> = [];
    parsed.searchParams.forEach((value, key) => {
      params.push({ key, value });
    });
    return params;
  } catch {
    return [];
  }
}

/**
 * Rebuild a URL from base (stripping existing query) + params array.
 */
export function buildUrlWithParams(
  baseUrl: string,
  params: Array<{ key: string; value: string }>
): string {
  try {
    const parsed = new URL(baseUrl);
    // Clear existing params
    parsed.search = '';
    // Add new params
    for (const { key, value } of params) {
      if (key) {
        parsed.searchParams.append(key, value);
      }
    }
    return parsed.toString();
  } catch {
    // URL might contain templates like {{block.field}}, just append manually
    const base = baseUrl.split('?')[0];
    const validParams = params.filter(p => p.key);
    if (validParams.length === 0) return base;
    const qs = validParams
      .map(p => `${encodeURIComponent(p.key)}=${encodeURIComponent(p.value)}`)
      .join('&');
    return `${base}?${qs}`;
  }
}
