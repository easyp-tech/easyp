import React from 'react';
import type { MarkdownProcessorOptions } from '../types';

/**
 * Code Focus Plugin for markdown processing
 * Handles [!code focus] annotations in code blocks
 */

/**
 * Processes code focus annotations in markdown content
 * @param content Markdown content
 * @param options Processing options
 * @returns Processed content with focus annotations handled
 */
export function processCodeFocus(
  content: string,
  options?: MarkdownProcessorOptions
): string {
  if (!options?.enableCodeFocus) {
    return content;
  }

  // Process code blocks and extract focus information
  return content.replace(/```(\w+)?\n([\s\S]*?)```/g, (match, language, code) => {
    const { cleanCode, hasFocus } = extractFocusFromCode(code);

    if (hasFocus) {
      // Add a data attribute to indicate this block has focus
      const langPart = language ? language : '';
      return `\`\`\`${langPart} focus\n${cleanCode}\`\`\``;
    }

    return match;
  });
}

/**
 * Extracts focus information from code content
 * @param code Raw code content
 * @returns Clean code and focus status
 */
export function extractFocusFromCode(code: string): {
  cleanCode: string;
  hasFocus: boolean;
  focusLines: number[];
} {
  const lines = code.split('\n');
  const focusLines: number[] = [];
  const cleanLines: string[] = [];
  let hasFocus = false;

  lines.forEach((line, index) => {
    const lineNumber = index + 1;

    // Check for various focus comment patterns
    const focusPatterns = [
      /\/\/\s*\[!code\s+focus\]/i,           // JavaScript-style
      /\/\*\s*\[!code\s+focus\]\s*\*\//i,   // Multi-line JS
      /#\s*\[!code\s+focus\]/i,             // Python, shell, YAML
      /<!--\s*\[!code\s+focus\]\s*-->/i,     // HTML/XML comments
      /\{\{.*\[!code\s+focus\].*\}\}/i,     // Template comments
      /\{\/\*\s*\[!code\s+focus\]\s*\*\/\}/i, // JSX comments
    ];

    let cleanLine = line;

    // Check each pattern and clean the line if focus comment is found
    for (const pattern of focusPatterns) {
      if (pattern.test(line)) {
        hasFocus = true;
        focusLines.push(lineNumber);

        // Remove the focus comment, preserving indentation
        cleanLine = line.replace(pattern, '').trimEnd();

        // If the line becomes empty after removing comment, keep original structure
        if (cleanLine.trim() === '' && line.trim() !== '') {
          cleanLine = line.replace(pattern, '').replace(/\s+$/, '');
        }
        break;
      }
    }

    cleanLines.push(cleanLine);
  });

  return {
    cleanCode: cleanLines.join('\n'),
    hasFocus,
    focusLines
  };
}

/**
 * Checks if a code block contains focus annotations
 * @param code Code content
 * @returns True if contains focus annotations
 */
export function hasCodeFocus(code: string): boolean {
  const focusPattern = /\[!code\s+focus\]/i;
  return focusPattern.test(code);
}

/**
 * Parses focus lines from code block metadata
 * @param metadata Code block metadata (e.g., "javascript focus")
 * @returns Whether block has focus and focus lines
 */
export function parseFocusMetadata(metadata: string): {
  hasFocus: boolean;
  language?: string;
} {
  const parts = metadata.split(/\s+/);
  const language = parts.find(part => !['focus', 'highlight', 'dim'].includes(part));
  const hasFocus = parts.includes('focus');

  return {
    hasFocus,
    language
  };
}

/**
 * Pre-processes markdown content for code focus
 * @param content Raw markdown content
 * @param options Processing options
 * @returns Pre-processed content
 */
export function preprocessCodeFocus(
  content: string,
  options?: MarkdownProcessorOptions
): string {
  if (!options?.enableCodeFocus) {
    return content;
  }

  return processCodeFocus(content, options);
}

/**
 * Generates markdown overrides for code focus processing
 * @param options Processing options
 * @returns Markdown-to-jsx overrides
 */
export function getCodeFocusOverrides(options?: MarkdownProcessorOptions) {
  if (!options?.enableCodeFocus) {
    return {};
  }

  return {
    pre: (props: any) => {
      const { children, ...rest } = props;

      // Handle code blocks with focus
      if (React.isValidElement(children) && children.type === 'code') {
        const codeProps = children.props as any;
        const className = codeProps.className || '';

        // Check if this is a focused code block
        const hasFocus = className.includes('focus');

        return (
          <div className={`code-block ${hasFocus ? 'code-block--with-focus' : ''}`}>
            <pre {...rest} className={`${className} code-pre`}>
              {children}
            </pre>
          </div>
        );
      }

      return <pre {...rest}>{children}</pre>;
    },

    code: (props: any) => {
      const { className, children, ...rest } = props;

      // Handle inline code
      if (!className) {
        return <code className="inline-code" {...rest}>{children}</code>;
      }

      // Handle code blocks
      const hasFocus = className.includes('focus');

      if (hasFocus) {
        const { cleanCode, focusLines } = extractFocusFromCode(children);

        return (
          <code
            {...rest}
            className={className}
            data-focus="true"
            data-focus-lines={focusLines.join(',')}
          >
            {cleanCode}
          </code>
        );
      }

      return <code {...rest} className={className}>{children}</code>;
    }
  };
}

/**
 * Validates code focus configuration
 * @param code Code content
 * @param focusLines Focus line numbers
 * @returns True if valid configuration
 */
export function validateCodeFocus(code: string, focusLines: number[]): boolean {
  const lines = code.split('\n');
  const totalLines = lines.length;

  return focusLines.every(line => line >= 1 && line <= totalLines);
}

/**
 * Normalizes focus line numbers (removes duplicates, sorts)
 * @param focusLines Array of line numbers
 * @returns Normalized array
 */
export function normalizeFocusLines(focusLines: number[]): number[] {
  return [...new Set(focusLines)]
    .filter(line => typeof line === 'number' && line > 0)
    .sort((a, b) => a - b);
}

export default {
  processCodeFocus,
  extractFocusFromCode,
  hasCodeFocus,
  parseFocusMetadata,
  preprocessCodeFocus,
  getCodeFocusOverrides,
  validateCodeFocus,
  normalizeFocusLines
};
