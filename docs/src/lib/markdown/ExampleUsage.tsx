import { useState } from 'react';
import { MarkdownContent } from './index';
import type { MarkdownProcessorOptions } from './types';

// Example markdown content
const exampleMarkdown = `---
title: "Example Document"
description: "Testing the markdown library"
---

# Example Document

[[toc]]

This is a test document to demonstrate the markdown library features.

<div class="tip custom-block">

**Tip**: This is a custom tip block that shows how HTML content works in markdown.

</div>

## Code Example

Here's some JavaScript code with focus:

\`\`\`javascript
function example() {
    const message = "Hello World"; // [!code focus]
    console.log(message); // [!code focus]
    return message;
}
\`\`\`

## Links

- [Internal link](./other-page.md)
- [External link](https://github.com/easyp-tech/easyp)
- [Anchor link](#code-example)

## More Sections

### Subsection 1

Some content here.

### Subsection 2

More content here.

<div class="warning custom-block">

**Warning**: This is a warning block example.

</div>
`;

export function ExampleUsage() {
  const [options, setOptions] = useState<MarkdownProcessorOptions>({
    enableToc: true,
    enableCodeFocus: true,
    enableHtmlBlocks: true,
    enableLinkProcessing: true,
    tocDepth: { min: 2, max: 4 },
    baseUrl: '/docs'
  });

  const [content, setContent] = useState(exampleMarkdown);

  const toggleOption = (key: keyof MarkdownProcessorOptions) => {
    setOptions(prev => ({
      ...prev,
      [key]: !prev[key]
    }));
  };

  return (
    <div className="example-usage">
      <div className="controls">
        <h2>Markdown Library Demo</h2>

        <div className="options">
          <h3>Options</h3>

          <label>
            <input
              type="checkbox"
              checked={options.enableToc}
              onChange={() => toggleOption('enableToc')}
            />
            Enable Table of Contents
          </label>

          <label>
            <input
              type="checkbox"
              checked={options.enableCodeFocus}
              onChange={() => toggleOption('enableCodeFocus')}
            />
            Enable Code Focus
          </label>

          <label>
            <input
              type="checkbox"
              checked={options.enableHtmlBlocks}
              onChange={() => toggleOption('enableHtmlBlocks')}
            />
            Enable HTML Blocks
          </label>

          <label>
            <input
              type="checkbox"
              checked={options.enableLinkProcessing}
              onChange={() => toggleOption('enableLinkProcessing')}
            />
            Enable Link Processing
          </label>
        </div>

        <div className="editor">
          <h3>Edit Markdown</h3>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            rows={15}
            cols={80}
            style={{
              fontFamily: 'monospace',
              width: '100%',
              padding: '8px',
              border: '1px solid #ccc',
              borderRadius: '4px'
            }}
          />
        </div>
      </div>

      <div className="preview">
        <h3>Preview</h3>
        <div className="markdown-preview">
          <MarkdownContent
            content={content}
            options={options}
            className="example-content"
          />
        </div>
      </div>

      <style>{`
        .example-usage {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 20px;
          max-width: 1400px;
          margin: 0 auto;
          padding: 20px;
        }

        .controls {
          border: 1px solid #e0e0e0;
          border-radius: 8px;
          padding: 20px;
          background: #f9f9f9;
        }

        .options {
          margin-bottom: 20px;
        }

        .options label {
          display: block;
          margin-bottom: 8px;
          cursor: pointer;
        }

        .options input {
          margin-right: 8px;
        }

        .preview {
          border: 1px solid #e0e0e0;
          border-radius: 8px;
          padding: 20px;
          background: white;
        }

        .markdown-preview {
          max-height: 600px;
          overflow-y: auto;
          border: 1px solid #eee;
          padding: 20px;
          border-radius: 4px;
        }

        h2, h3 {
          margin-top: 0;
          color: #333;
        }

        @media (max-width: 768px) {
          .example-usage {
            grid-template-columns: 1fr;
          }
        }
      `}</style>
    </div>
  );
}

export default ExampleUsage;
