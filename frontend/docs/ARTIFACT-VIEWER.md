# Artifact Viewer System

## Overview

The artifact viewer system allows inline rendering of HTML and PDF files directly in the chat interface, providing a rich preview experience similar to modern chat applications.

## Components

### 1. ArtifactContainer

Base wrapper component that provides common controls for all artifact types.

**Features:**
- Expand/collapse toggle
- Fullscreen mode
- Download button
- Custom toolbar support (e.g., page navigation)

**Usage:**
```tsx
<ArtifactContainer
  title="My Document"
  downloadUrl="/api/files/123"
  filename="document.pdf"
  toolbar={<PageNavigation />}
>
  <YourContentHere />
</ArtifactContainer>
```

### 2. PdfArtifact

Renders PDF files using react-pdf with page navigation.

**Features:**
- Page-by-page navigation
- Previous/Next buttons
- Page counter (e.g., "3 / 10")
- Download integration
- Loading states
- Error handling

**Usage:**
```tsx
<PdfArtifact
  url="http://localhost:3001/api/files/123?code=abc"
  title="Presentation"
  filename="slides.pdf"
/>
```

### 3. HtmlArtifact

Renders HTML content in a sandboxed iframe.

**Features:**
- Sandboxed iframe (allow-scripts, allow-same-origin)
- Can accept content directly or fetch from URL
- Loading states
- Error handling

**Usage:**
```tsx
// With direct content
<HtmlArtifact
  content="<html>...</html>"
  title="HTML Preview"
/>

// Or fetch from URL
<HtmlArtifact
  url="http://localhost:3001/api/files/123"
  title="HTML Preview"
  filename="index.html"
/>
```

### 4. ArtifactRenderer

Smart component that detects file types and renders the appropriate artifact.

**Usage in AssistantMessage:**
```tsx
const renderDownloadTile = () => {
  const downloadTool = message.toolCalls?.find(/* ... */);
  if (!downloadTool) return null;

  // Try to render as artifact
  const artifactRenderer = (
    <ArtifactRenderer
      toolResult={downloadTool.result}
      backendUrl={backendUrl}
    />
  );

  if (artifactRenderer) {
    return <div>{artifactRenderer}</div>;
  }

  // Fall back to download tile
  return <DownloadTile />;
};
```

## Integration

The artifact viewer automatically detects and renders:

1. **PDF files** (create_presentation, create_document tools)
   - Presentations with page navigation
   - Documents with scroll view

2. **HTML files** (future HTML export tools)
   - Interactive HTML content
   - Sandboxed for security

## File Type Detection

Detection is based on:
1. `file_type` field in tool result JSON
2. File extension from `filename` field

```typescript
const isPdf =
  parsedResult.file_type === 'pdf' ||
  filename.toLowerCase().endsWith('.pdf');

const isHtml =
  parsedResult.file_type === 'html' ||
  filename.toLowerCase().endsWith('.html');
```

## Dependencies

```json
{
  "dependencies": {
    "react-pdf": "^9.2.1",
    "pdfjs-dist": "^4.9.155"
  }
}
```

## PDF.js Worker Configuration

The PdfArtifact component configures the PDF.js worker using a CDN:

```typescript
import { pdfjs } from 'react-pdf';

pdfjs.GlobalWorkerOptions.workerSrc =
  `//unpkg.com/pdfjs-dist@${pdfjs.version}/build/pdf.worker.min.js`;
```

## Styling

All artifact components use CSS variables from the design system:

- `var(--space-*)` for spacing
- `var(--color-*)` for colors
- `var(--radius-*)` for border radius
- `var(--shadow-*)` for shadows
- `var(--transition-*)` for animations

## Future Enhancements

- **Image artifacts** - Gallery view for image collections
- **Code artifacts** - Syntax-highlighted code preview
- **Data artifacts** - Table/chart viewers for CSV/JSON
- **3D artifacts** - Viewers for 3D models (GLB, OBJ)
- **Audio/Video artifacts** - Media players

## Example Tool Result

```json
{
  "success": true,
  "file_id": "abc-123",
  "filename": "presentation.pdf",
  "download_url": "http://localhost:3001/api/files/abc-123?code=xyz",
  "file_type": "pdf",
  "size": 45678,
  "page_count": 10,
  "expires_at": "2024-02-15"
}
```

The artifact system will automatically detect this as a PDF and render the PdfArtifact component with page navigation.
