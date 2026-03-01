/**
 * Shared utility functions for tool call display.
 * Used by ToolStatusPill and AssistantMessage.
 */

/**
 * Get user-friendly action text for tool name (present tense / streaming)
 */
export function getToolActionText(toolName: string): string {
  const toolActions: Record<string, string> = {
    search_web: 'Searching the web',
    search_images: 'Finding images',
    fetch_webpage: 'Fetching webpage',
    read_file: 'Reading file',
    write_file: 'Writing file',
    execute_code: 'Running code',
    list_files: 'Listing files',
    search_files: 'Searching files',
    generate_image: 'Generating image',
    edit_image: 'Editing image',
    create_document: 'Creating document',
    create_presentation: 'Creating presentation',
    create_text_file: 'Creating file',
    ask_user: 'Asking the user',
  };
  return toolActions[toolName] || `Using ${toolName}`;
}

/**
 * Get varied action text for repeated tool calls (2nd, 3rd, etc.).
 * Returns null if no special text exists for this tool.
 */
export function getToolRepeatText(toolName: string, occurrence: number): string | null {
  const repeats: Record<string, string[]> = {
    search_web: ['Searching for more results', 'Digging deeper', 'Refining search'],
    search_images: ['Finding more images', 'Searching for more'],
    fetch_webpage: ['Fetching another page', 'Loading more pages'],
    search_files: ['Searching more files', 'Looking further'],
  };

  const variants = repeats[toolName];
  if (!variants) return null;
  // cycle through variants for 2nd, 3rd, 4th, ...
  return variants[Math.min(occurrence - 2, variants.length - 1)];
}

/**
 * Get past-tense action text for tool name (completed state)
 */
export function getToolCompletedText(toolName: string): string {
  const toolCompleted: Record<string, string> = {
    search_web: 'Searched the web',
    search_images: 'Found images',
    fetch_webpage: 'Fetched webpage',
    read_file: 'Read file',
    write_file: 'Wrote file',
    execute_code: 'Ran code',
    list_files: 'Listed files',
    search_files: 'Searched files',
    generate_image: 'Generated image',
    edit_image: 'Edited image',
    create_document: 'Created document',
    create_presentation: 'Created presentation',
    create_text_file: 'Created file',
    ask_user: 'Gathered responses',
  };
  return toolCompleted[toolName] || `Used ${toolName}`;
}

/**
 * Get favicon URL for a given page URL
 */
export function getFaviconUrl(url: string): string {
  try {
    const hostname = new URL(url).hostname;
    return `https://www.google.com/s2/favicons?domain=${hostname}&sz=16`;
  } catch {
    return '';
  }
}

/**
 * Extract hostname from a URL for display
 */
export function getHostname(url: string): string {
  try {
    return new URL(url).hostname.replace('www.', '');
  } catch {
    return url;
  }
}

export interface SearchResultItem {
  title: string;
  url: string;
  domain: string;
  faviconUrl: string;
}

export interface ParsedSearchResult {
  query: string;
  resultCount: string;
  items: SearchResultItem[];
}

/**
 * Parse a single search_web tool's result into structured data for display
 */
export function parseSearchResults(
  result: string | undefined,
  query?: string
): ParsedSearchResult | null {
  if (!result) return null;

  const resultStr = typeof result === 'string' ? result : JSON.stringify(result);
  const items: SearchResultItem[] = [];
  const seenUrls = new Set<string>();

  // Parse citation format: [1]: [Title](url)
  const citationMatches = resultStr.matchAll(/\[\d+\]:\s*\[([^\]]+)\]\(([^)]+)\)/g);
  for (const match of citationMatches) {
    const [, title, url] = match;
    if (url && !seenUrls.has(url)) {
      seenUrls.add(url);
      items.push({
        title: title || getHostname(url),
        url,
        domain: getHostname(url),
        faviconUrl: getFaviconUrl(url),
      });
    }
  }

  // Fallback: numbered results format: [1] Title\n    URL: url
  if (items.length === 0) {
    const oldFormatMatches = resultStr.matchAll(/\[?\d+\]?\.\s*([^\n]+)\n\s*URL:\s*(\S+)/g);
    for (const match of oldFormatMatches) {
      const [, title, url] = match;
      if (url && !seenUrls.has(url)) {
        seenUrls.add(url);
        items.push({
          title: title.trim(),
          url: url.trim(),
          domain: getHostname(url.trim()),
          faviconUrl: getFaviconUrl(url.trim()),
        });
      }
    }
  }

  const resultCount = getResultCount(result, query);

  return {
    query: query || 'web search',
    resultCount,
    items,
  };
}

/**
 * Build a concise completed summary string from tool calls.
 *
 * Keeps it short:
 * - 1 search tool:   "Searched the web"
 * - 2+ search tools: "Searched multiple sites"
 * - 1 other tool:    "Ran code" / "Created a file" etc.
 * - 2+ other tools:  "Used N tools"
 * - Mix:             "Searched the web · Used 3 tools"
 */
export function buildCompletedSummary(
  toolCalls: Array<{ name: string; displayName?: string }>
): string {
  if (toolCalls.length === 0) return '';

  let searchCount = 0;
  let otherCount = 0;
  let singleOtherName = '';

  for (const tool of toolCalls) {
    if (tool.name === 'search_web' || tool.name === 'search_images') {
      searchCount++;
    } else {
      otherCount++;
      if (otherCount === 1) singleOtherName = tool.name;
    }
  }

  const parts: string[] = [];

  // Search part
  if (searchCount === 1) {
    parts.push('Searched the web');
  } else if (searchCount > 1) {
    parts.push('Searched multiple sites');
  }

  // Other tools part — "Ran code", "Used OG Stats", "Used 4 tools"
  if (otherCount === 1) {
    const tool = toolCalls.find(t => t.name === singleOtherName);
    const completed = getToolCompletedText(singleOtherName);
    // Known tool: "Ran code", "Created a file" etc.
    // Unknown tool (starts with "Used"): prefer displayName
    if (completed.startsWith('Used ') && tool?.displayName) {
      parts.push(`Used ${tool.displayName}`);
    } else {
      parts.push(completed);
    }
  } else if (otherCount > 1) {
    parts.push(`Used ${otherCount} tools`);
  }

  if (parts.length === 0) return '';

  return parts.join(' · ');
}

/**
 * Get a default Lucide icon name for a known tool name.
 * Falls back to 'Zap' for unknown tools.
 */
export function getToolIcon(toolName: string): string {
  const icons: Record<string, string> = {
    // Web
    search_web: 'Globe',
    search_images: 'Image',
    fetch_webpage: 'ExternalLink',
    scrape_web: 'ExternalLink',
    // Files
    read_file: 'FileText',
    write_file: 'FileText',
    read_document: 'BookOpen',
    create_document: 'FileText',
    create_presentation: 'FileText',
    create_text_file: 'FileCode',
    list_files: 'FolderOpen',
    search_files: 'Search',
    download_file: 'Download',
    // Code & compute
    execute_code: 'Terminal',
    run_python: 'Terminal',
    calculate_math: 'Calculator',
    analyze_data: 'BarChart2',
    // Media
    generate_image: 'ImagePlus',
    edit_image: 'Pencil',
    describe_image: 'Image',
    transcribe_audio: 'Mic',
    // Interactive
    ask_user: 'MessageCircleQuestion',
  };
  return icons[toolName] || 'Wrench';
}

/**
 * Get result count or summary from result text
 */
export function getResultCount(result: string | unknown, query?: string): string {
  if (result === null || result === undefined) {
    return query || 'Details';
  }

  const resultStr = typeof result === 'string' ? result : JSON.stringify(result);

  // For math results
  if (resultStr.includes('Result:') && resultStr.includes('Expression:')) {
    const resultMatch = resultStr.match(/Result:\s*([^\n]+)/);
    if (resultMatch) return resultMatch[1].trim();
  }

  // Count numbered items
  const numberedItems = resultStr.match(/^\d+\.\s/gm);
  if (numberedItems && numberedItems.length > 0) {
    return `${numberedItems.length} results`;
  }

  // Common patterns
  const patterns = [/found (\d+) results?/i, /(\d+) results?/i, /returned (\d+)/i];
  for (const pattern of patterns) {
    const match = resultStr.match(pattern);
    if (match) return `${match[1]} results`;
  }

  return query || 'Details';
}
