import React from 'react';
import { MarkdownContent } from './index';
import './styles.css';

const demoMarkdown = `---
title: "EasyP Markdown Library Demo"
description: "Complete demonstration of all markdown features"
---

# EasyP Markdown Library Demo

[[toc]]

Welcome to the **EasyP Markdown Library**! This demo showcases all the powerful features available for processing markdown content in React applications.

<div class="tip custom-block">

**🎉 Features Included**: Table of Contents, Code Focus, Custom HTML Blocks, Smart Link Processing, and much more!

</div>

## Table of Contents Features

The table of contents is automatically generated from your document headings. Simply add \`[[toc]]\` anywhere in your markdown, and it will be replaced with a beautiful, interactive table of contents.

### How TOC Works

- Extracts headings from H2 to H4 by default
- Generates unique anchor IDs
- Provides smooth scrolling navigation
- Highlights active section while scrolling

### TOC Configuration

You can customize the TOC behavior:

\`\`\`typescript
const options = {
  enableToc: true,
  tocDepth: { min: 2, max: 4 } // Include h2, h3, h4
};
\`\`\`

## Code Focus Feature

One of the most powerful features is the ability to highlight specific lines in code blocks using special comments.

### JavaScript Example

\`\`\`javascript
function processUser(userData) {
  // Validate input data
  if (!userData || !userData.email) {
    throw new Error('Invalid user data'); // [!code focus]
  }

  // Process the user
  const user = {
    id: generateId(),
    email: userData.email.toLowerCase(), // [!code focus]
    name: userData.name || 'Anonymous',
    createdAt: new Date() // [!code focus]
  };

  // Save to database
  return saveUser(user);
}
\`\`\`

### Python Example

\`\`\`python
def calculate_metrics(data):
    """Calculate performance metrics from data."""
    total = 0
    count = 0

    for item in data:
        if item.is_valid():  # [!code focus]
            total += item.value  # [!code focus]
            count += 1  # [!code focus]

    return {
        'average': total / count if count > 0 else 0,
        'total': total,
        'count': count
    }
\`\`\`

### Supported Comment Styles

The library supports multiple comment syntaxes:

- \`// [!code focus]\` - JavaScript, C++, Java, etc.
- \`# [!code focus]\` - Python, Shell, YAML, etc.
- \`<!-- [!code focus] -->\` - HTML, XML
- \`/* [!code focus] */\` - CSS, multi-line comments

## Custom HTML Blocks

Create visually appealing information blocks using HTML divs with special classes.

### Tip Block

<div class="tip custom-block">

**💡 Pro Tip**: Use tip blocks to provide helpful hints and best practices to your readers.

This block can contain **any markdown content**, including:
- Lists
- Links
- Code snippets
- And more!

</div>

### Warning Block

<div class="warning custom-block">

**⚠️ Important Warning**: Always backup your data before making significant changes to your system configuration.

Use warning blocks for information that could prevent issues or mistakes.

</div>

### Danger Block

<div class="danger custom-block">

**🚨 Critical Alert**: This operation will permanently delete all data and cannot be undone.

Reserve danger blocks for truly critical information that could cause data loss or system damage.

</div>

### Details Block

<div class="details custom-block">

**📋 Additional Details**: Click to expand this section for more detailed information.

This is perfect for optional information that doesn't need to be visible by default. You can include:

1. Extended explanations
2. Advanced configuration options
3. Troubleshooting steps
4. Reference materials

</div>

## Link Processing

The library intelligently handles different types of links:

### Internal Links
- [Documentation Overview](./README.md) - Automatically converted for React Router
- [Getting Started Guide](./getting-started.md) - Works with relative paths
- [API Reference](../api/index.md) - Handles directory traversal

### External Links
- [EasyP GitHub Repository](https://github.com/easyp-tech/easyp) - Opens in new tab with icon
- [React Documentation](https://react.dev) - External link indicator
- [TypeScript Handbook](https://www.typescriptlang.org/docs/) - Secure external links

### Anchor Links
- [Jump to Code Focus Section](#code-focus-feature) - Smooth scrolling navigation
- [Go to Custom Blocks](#custom-html-blocks) - Internal page anchors
- [Back to Table of Contents](#table-of-contents-features) - Section navigation

## Advanced Formatting

### Lists and Structure

#### Unordered Lists
- **Primary item** with emphasis
- *Secondary item* with italics
- \`Code item\` with inline code
- Regular item with [internal link](./example.md)
- Item with [external link](https://example.com)

#### Ordered Lists
1. First step in the process
2. Second step with **important** details
3. Third step with \`code example\`
4. Final step with completion

#### Task Lists
- [x] Implement Table of Contents generation
- [x] Add Code Focus highlighting
- [x] Create Custom HTML blocks
- [x] Implement Link processing
- [ ] Add mathematical formula support
- [ ] Implement diagram rendering

### Tables

| Feature | Status | Description | Version |
|---------|--------|-------------|---------|
| TOC Generation | ✅ Complete | Automatic from headings | 1.0.0 |
| Code Focus | ✅ Complete | Line highlighting | 1.0.0 |
| HTML Blocks | ✅ Complete | Custom info blocks | 1.0.0 |
| Link Processing | ✅ Complete | Smart link handling | 1.0.0 |
| Math Support | 🔄 Planned | LaTeX formulas | 1.1.0 |
| Diagrams | 🔄 Planned | Mermaid integration | 1.2.0 |

### Blockquotes

> **"The best documentation is the one that doesn't exist because the code is so clear."**
>
> However, when documentation is necessary, it should be comprehensive, accessible, and maintainable.
>
> — Software Engineering Wisdom

### Code Examples

#### Inline Code
Use \`const variable = "value"\` for variable assignments, \`npm install package\` for commands, and \`<Component />\` for JSX elements.

#### Configuration Example

\`\`\`yaml
# easyp.yaml configuration
version: v1alpha
lint:
  use: [MINIMAL, BASIC, DEFAULT] # [!code focus]
  enum_zero_value_suffix: "UNSPECIFIED" # [!code focus]
  service_suffix: "Service" # [!code focus]
  ignore:
    - "vendor/"
    - "third_party/"
\`\`\`

#### Usage Example

\`\`\`tsx
import { MarkdownContent } from '@/lib/markdown';
import '@/lib/markdown/styles.css';

export function DocumentationPage({ content }: { content: string }) {
  return (
    <div className="documentation">
      <MarkdownContent  // [!code focus]
        content={content}  // [!code focus]
        options={{  // [!code focus]
          enableToc: true,  // [!code focus]
          enableCodeFocus: true,  // [!code focus]
          enableHtmlBlocks: true,  // [!code focus]
          enableLinkProcessing: true,  // [!code focus]
          baseUrl: '/docs'  // [!code focus]
        }}  // [!code focus]
      />  // [!code focus]
    </div>
  );
}
\`\`\`

## Performance and Optimization

The library is built with performance in mind:

### Key Optimizations
- **Memoized Processing**: Results are cached to avoid recomputation
- **Efficient Rendering**: Only re-renders when content changes
- **Lazy Loading**: Large documents load progressively
- **Code Splitting**: Syntax highlighting loads on demand

### Memory Usage
- Minimal memory footprint
- Efficient garbage collection
- Optimized for large documents
- Smart caching strategies

## Browser Compatibility

| Browser | Minimum Version | Notes |
|---------|----------------|-------|
| Chrome | 88+ | Full support |
| Firefox | 85+ | Full support |
| Safari | 14+ | Full support |
| Edge | 88+ | Full support |

---

## Conclusion

The **EasyP Markdown Library** provides a comprehensive solution for rendering markdown content in React applications. With features like:

1. ✅ **Automatic Table of Contents**
2. ✅ **Code Focus Highlighting**
3. ✅ **Custom HTML Blocks**
4. ✅ **Smart Link Processing**
5. ✅ **TypeScript Support**
6. ✅ **Customizable Styling**

You can create professional documentation sites with ease.

<div class="tip custom-block">

**🚀 Ready to Get Started?** Check out the [README](./README.md) for installation and usage instructions, or explore the [example files](./example.md) for more detailed examples.

</div>

For more information about the EasyP project, visit our [GitHub repository](https://github.com/easyp-tech/easyp) or read the [full documentation](https://docs.easyp.tech).
`;

export function MarkdownDemo() {
  return (
    <div style={{
      maxWidth: '1200px',
      margin: '0 auto',
      padding: '20px',
      fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif'
    }}>
      <MarkdownContent
        content={demoMarkdown}
        options={{
          enableToc: true,
          enableCodeFocus: true,
          enableHtmlBlocks: true,
          enableLinkProcessing: true,
          tocDepth: { min: 2, max: 4 },
          baseUrl: '/docs'
        }}
        className="demo-content"
      />
    </div>
  );
}

export default MarkdownDemo;
