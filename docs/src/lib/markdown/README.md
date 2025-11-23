# EasyP Markdown Library

A comprehensive React-based markdown processing library with advanced features for documentation sites.

## Features

- 🔗 **Table of Contents** - Automatic generation from headings with `[[toc]]` placeholder
- 🎯 **Code Focus** - Highlight specific lines using `// [!code focus]` comments
- 📦 **Custom HTML Blocks** - Support for tip, warning, danger, and details blocks
- 🔗 **Link Processing** - Smart handling of internal/external links with React Router integration
- ⚡ **React Integration** - Built for React with TypeScript support
- 🎨 **Customizable Styling** - CSS classes for complete style control

## Installation

```bash
# The library is part of the EasyP docs project
# No separate installation needed
```

## Quick Start

```tsx
import { MarkdownContent } from '@/lib/markdown';

function MyDocPage() {
  const markdownContent = `
# My Document

[[toc]]

## Introduction

This is a sample document with various features.

<div class="tip custom-block">
**Tip**: This is a helpful tip!
</div>

## Code Example

\`\`\`javascript
function example() {
    console.log("Hello"); // [!code focus]
    return "world"; // [!code focus]
}
\`\`\`
  `;

  return (
    <MarkdownContent
      content={markdownContent}
      options={{
        enableToc: true,
        enableCodeFocus: true,
        enableHtmlBlocks: true,
        enableLinkProcessing: true
      }}
    />
  );
}
```

## Configuration

### MarkdownProcessorOptions

```typescript
interface MarkdownProcessorOptions {
  enableToc?: boolean;           // Enable TOC generation (default: true)
  enableCodeFocus?: boolean;     // Enable code focus (default: true)
  enableHtmlBlocks?: boolean;    // Enable custom blocks (default: true)
  enableLinkProcessing?: boolean; // Enable link processing (default: true)
  baseUrl?: string;              // Base URL for relative links
  className?: string;            // Additional CSS class
  tocDepth?: {                   // TOC depth configuration
    min: number;                 // Minimum heading level (default: 2)
    max: number;                 // Maximum heading level (default: 4)
  };
}
```

## Features Guide

### Table of Contents

Add `[[toc]]` anywhere in your markdown to generate an automatic table of contents:

```markdown
# My Document

[[toc]]

## Section 1
### Subsection 1.1
### Subsection 1.2

## Section 2
```

Configure TOC depth:

```tsx
<MarkdownContent
  content={content}
  options={{
    tocDepth: { min: 2, max: 4 } // Include h2, h3, h4
  }}
/>
```

### Code Focus

Highlight specific lines in code blocks using `[!code focus]` comments:

```markdown
\`\`\`javascript
function example() {
    const setup = "preliminary"; 
    const result = process(); // [!code focus]
    return result; // [!code focus]
}
\`\`\`
```

Supports multiple comment styles:
- `// [!code focus]` (JavaScript, C++, etc.)
- `# [!code focus]` (Python, Shell, etc.)  
- `<!-- [!code focus] -->` (HTML, XML)
- `/* [!code focus] */` (CSS, multiline)

### Custom HTML Blocks

Create informational blocks using HTML divs:

```markdown
<div class="tip custom-block">
**Tip**: This is helpful information!
</div>

<div class="warning custom-block">
**Warning**: Be careful with this!
</div>

<div class="danger custom-block">
**Danger**: This could be harmful!
</div>

<div class="details custom-block">
**Details**: Additional information here.
</div>
```

### Link Processing

Links are automatically processed for React Router navigation:

```markdown
- [Internal link](./other-page.md) → Converted to React Router navigation
- [External link](https://example.com) → Opens in new tab with icon
- [Anchor link](#section) → Smooth scrolling to section
```

## Components

### MarkdownContent

Main component for rendering markdown content:

```tsx
<MarkdownContent
  content={markdownString}
  options={processingOptions}
  className="custom-styles"
/>
```

### TableOfContents

Standalone TOC component:

```tsx
import { TableOfContents, generateToc } from '@/lib/markdown';

const toc = generateToc(markdownContent);

<TableOfContents
  toc={toc}
  title="On This Page"
  activeId={activeHeadingId}
/>
```

### CodeBlock

Standalone code block component:

```tsx
import { CodeBlock } from '@/lib/markdown';

<CodeBlock
  language="javascript"
  hasFocus={true}
  focusLines={[2, 3]}
>
{`function example() {
  console.log("hello");
  return "world";
}`}
</CodeBlock>
```

### CustomBlock

Standalone custom block component:

```tsx
import { CustomBlock } from '@/lib/markdown';

<CustomBlock type="tip" title="Pro Tip">
  This is a helpful tip!
</CustomBlock>
```

## Utilities

### Processing Functions

```tsx
import { 
  processMarkdown, 
  validateMarkdown, 
  extractMarkdownMetadata 
} from '@/lib/markdown';

// Process markdown
const result = processMarkdown(content, options);
// { content: ReactNode, toc: TocItem[], frontmatter: object }

// Validate markdown
const validation = validateMarkdown(content);
// { isValid: boolean, errors: string[], warnings: string[] }

// Extract metadata
const metadata = extractMarkdownMetadata(content);
// { frontmatter, headingCount, wordCount, readingTime, ... }
```

### Path Utilities

```tsx
import { 
  isExternalUrl, 
  resolveMarkdownPath,
  filePathToUrlPath 
} from '@/lib/markdown';

isExternalUrl('https://example.com'); // true
resolveMarkdownPath('./doc.md', '/current/path'); // '/current/doc'
filePathToUrlPath('guide/intro.md'); // '/guide/intro'
```

## Styling

### CSS Classes

The library generates semantic CSS classes for styling:

```css
/* Content containers */
.markdown-content
.markdown-content--with-toc
.markdown-body

/* Headings */
.markdown-heading
.markdown-heading--level-{1-6}
.markdown-heading--with-anchor
.heading-anchor

/* Table of Contents */
.toc-container
.toc-title
.toc-list
.toc-item
.toc-item--level-{2-4}
.toc-link
.toc-link--active

/* Code blocks */
.code-block
.code-block--with-focus
.code-pre
.code-line
.line-focus
.line-dimmed
.inline-code

/* Custom blocks */
.custom-block
.custom-block--{tip|warning|danger|details}
.custom-block__header
.custom-block__icon
.custom-block__title
.custom-block__content

/* Links */
.markdown-link
.markdown-link--{external|internal|anchor}
.external-link-icon
```

### Import Styles

```tsx
import '@/lib/markdown/styles.css';
```

Or customize with your own CSS using the provided classes.

## Advanced Usage

### Custom Processor

```tsx
import { MarkdownProcessor } from '@/lib/markdown';

const processor = new MarkdownProcessor({
  enableToc: true,
  enableCodeFocus: true,
  tocDepth: { min: 1, max: 6 }
});

const result = processor.process(markdownContent);
```

### Frontmatter Support

```markdown
---
title: "My Document"
description: "A great document"
toc: true
---

# My Document

Content here...
```

```tsx
const { frontmatter, content, toc } = processMarkdown(markdown);
// frontmatter.title === "My Document"
```

### React Router Integration

Set up navigation helper:

```tsx
// In your app root
window.routerNavigate = navigate; // from useNavigate()
```

Links in markdown will automatically use React Router navigation.

## Browser Support

- Chrome/Edge 88+
- Firefox 85+
- Safari 14+

## TypeScript

Full TypeScript support with exported types:

```tsx
import type {
  MarkdownProcessorOptions,
  ProcessedMarkdown,
  TocItem,
  CodeBlockProps,
  CustomBlockProps
} from '@/lib/markdown';
```

## Performance

- Efficient processing with memoized results
- Lazy loading for large documents
- Optimized re-renders with React hooks
- Syntax highlighting via Prism (code splitting supported)

## Contributing

The library is part of the EasyP project. See the main repository for contribution guidelines.

## License

MIT License - see the main EasyP project for details.