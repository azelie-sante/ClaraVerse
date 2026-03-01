/**
 * Converts markdown text to clean, formatted plain text suitable for emails.
 * Preserves structure and readability while removing markdown syntax.
 */

export function markdownToPlainText(markdown: string): string {
  let text = markdown;

  // Remove code blocks but keep content with indentation
  text = text.replace(/```[\w]*\n([\s\S]*?)```/g, (_, code) => {
    return code
      .split('\n')
      .map((line: string) => '    ' + line)
      .join('\n');
  });

  // Remove inline code backticks
  text = text.replace(/`([^`]+)`/g, '$1');

  // Convert headers to plain text with underlines for h1/h2
  text = text.replace(/^#{1}\s+(.+)$/gm, (_, content) => {
    return `${content}\n${'='.repeat(content.length)}`;
  });
  text = text.replace(/^#{2}\s+(.+)$/gm, (_, content) => {
    return `${content}\n${'-'.repeat(content.length)}`;
  });
  text = text.replace(/^#{3,6}\s+(.+)$/gm, '$1');

  // Convert bold and italic
  text = text.replace(/\*\*\*(.+?)\*\*\*/g, '$1'); // bold italic
  text = text.replace(/\*\*(.+?)\*\*/g, '$1'); // bold
  text = text.replace(/\*(.+?)\*/g, '$1'); // italic
  text = text.replace(/___(.+?)___/g, '$1'); // bold italic alt
  text = text.replace(/__(.+?)__/g, '$1'); // bold alt
  text = text.replace(/_(.+?)_/g, '$1'); // italic alt

  // Convert links: [text](url) -> text (url)
  text = text.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '$1 ($2)');

  // Convert images: ![alt](url) -> [Image: alt]
  text = text.replace(/!\[([^\]]*)\]\([^)]+\)/g, '[Image: $1]');

  // Convert unordered lists (preserve structure with bullets)
  text = text.replace(/^[\s]*[-*+]\s+/gm, '• ');

  // Convert ordered lists (keep numbers)
  text = text.replace(/^[\s]*(\d+)\.\s+/gm, '$1. ');

  // Convert blockquotes
  text = text.replace(/^>\s*/gm, '│ ');

  // Convert horizontal rules
  text = text.replace(/^[-*_]{3,}$/gm, '────────────────────');

  // Convert tables to simple format
  text = text.replace(/\|([^|]+)\|/g, match => {
    return match.replace(/\|/g, ' │ ').trim();
  });
  // Remove table alignment rows
  text = text.replace(/^[\s│:-]+$/gm, '');

  // Remove HTML tags
  text = text.replace(/<[^>]+>/g, '');

  // Clean up extra whitespace
  text = text.replace(/\n{3,}/g, '\n\n'); // Max 2 newlines
  text = text.replace(/[ \t]+$/gm, ''); // Trailing spaces

  return text.trim();
}

/**
 * Converts markdown to HTML for rich text clipboard (preserves formatting in email clients)
 */
export function markdownToHtml(markdown: string): string {
  let html = markdown;

  // Escape HTML entities first
  html = html.replace(/&/g, '&amp;');
  html = html.replace(/</g, '&lt;');
  html = html.replace(/>/g, '&gt;');

  // Code blocks
  html = html.replace(
    /```[\w]*\n([\s\S]*?)```/g,
    '<pre style="background:#f4f4f4;padding:12px;border-radius:4px;overflow-x:auto;font-family:monospace;">$1</pre>'
  );

  // Inline code
  html = html.replace(
    /`([^`]+)`/g,
    '<code style="background:#f4f4f4;padding:2px 6px;border-radius:3px;font-family:monospace;">$1</code>'
  );

  // Headers
  html = html.replace(/^#{1}\s+(.+)$/gm, '<h1 style="margin:16px 0 8px 0;font-size:24px;">$1</h1>');
  html = html.replace(/^#{2}\s+(.+)$/gm, '<h2 style="margin:14px 0 6px 0;font-size:20px;">$1</h2>');
  html = html.replace(/^#{3}\s+(.+)$/gm, '<h3 style="margin:12px 0 4px 0;font-size:16px;">$1</h3>');
  html = html.replace(
    /^#{4,6}\s+(.+)$/gm,
    '<h4 style="margin:10px 0 4px 0;font-size:14px;">$1</h4>'
  );

  // Bold and italic
  html = html.replace(/\*\*\*(.+?)\*\*\*/g, '<strong><em>$1</em></strong>');
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');
  html = html.replace(/___(.+?)___/g, '<strong><em>$1</em></strong>');
  html = html.replace(/__(.+?)__/g, '<strong>$1</strong>');
  html = html.replace(/_(.+?)_/g, '<em>$1</em>');

  // Links
  html = html.replace(
    /\[([^\]]+)\]\(([^)]+)\)/g,
    '<a href="$2" style="color:#0066cc;text-decoration:underline;">$1</a>'
  );

  // Images
  html = html.replace(
    /!\[([^\]]*)\]\(([^)]+)\)/g,
    '<img src="$2" alt="$1" style="max-width:100%;">'
  );

  // Unordered lists
  html = html.replace(/^[\s]*[-*+]\s+(.+)$/gm, '<li style="margin:4px 0;">$1</li>');

  // Ordered lists
  html = html.replace(/^[\s]*\d+\.\s+(.+)$/gm, '<li style="margin:4px 0;">$1</li>');

  // Wrap consecutive li elements in ul/ol
  html = html.replace(
    /(<li[^>]*>[\s\S]*?<\/li>\n?)+/g,
    '<ul style="margin:8px 0;padding-left:24px;">$&</ul>'
  );

  // Blockquotes
  html = html.replace(
    /^&gt;\s*(.+)$/gm,
    '<blockquote style="border-left:3px solid #ccc;margin:8px 0;padding-left:12px;color:#666;">$1</blockquote>'
  );

  // Horizontal rules
  html = html.replace(
    /^[-*_]{3,}$/gm,
    '<hr style="border:none;border-top:1px solid #ddd;margin:16px 0;">'
  );

  // Paragraphs (double newlines)
  html = html.replace(/\n\n+/g, '</p><p style="margin:8px 0;">');

  // Single newlines to <br>
  html = html.replace(/\n/g, '<br>');

  // Wrap in paragraph
  html = `<div style="font-family:sans-serif;font-size:14px;line-height:1.6;color:#333;"><p style="margin:8px 0;">${html}</p></div>`;

  // Clean up empty paragraphs
  html = html.replace(/<p[^>]*><\/p>/g, '');
  html = html.replace(/<p[^>]*><br><\/p>/g, '');

  return html;
}

/**
 * Copies text to clipboard with both plain text and HTML formats.
 * This allows email clients to paste with formatting.
 */
export async function copyAsFormattedText(markdown: string): Promise<void> {
  const plainText = markdownToPlainText(markdown);
  const htmlText = markdownToHtml(markdown);

  // Use ClipboardItem API for rich text support
  if (navigator.clipboard && typeof ClipboardItem !== 'undefined') {
    try {
      const clipboardItem = new ClipboardItem({
        'text/plain': new Blob([plainText], { type: 'text/plain' }),
        'text/html': new Blob([htmlText], { type: 'text/html' }),
      });
      await navigator.clipboard.write([clipboardItem]);
      return;
    } catch {
      // Fall back to plain text if ClipboardItem fails
    }
  }

  // Fallback: copy plain text only
  await navigator.clipboard.writeText(plainText);
}
