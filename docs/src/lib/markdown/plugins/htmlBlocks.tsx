import type { MarkdownProcessorOptions } from '../types';
import { parseCustomBlockDiv } from '../components/CustomBlock';

/**
 * HTML Blocks Plugin for markdown processing
 * Handles custom block divs like tip, warning, danger, details
 */

/**
 * Processes HTML blocks in markdown content
 * @param content Markdown content
 * @param options Processing options
 * @returns Processed content with HTML blocks handled
 */
export function processHtmlBlocks(
  content: string,
  options?: MarkdownProcessorOptions
): string {
  if (!options?.enableHtmlBlocks) {
    return content;
  }

  // Process custom block divs
  return content.replace(
    /<div\s+class="([^"]*custom-block[^"]*)"[^>]*>([\s\S]*?)<\/div>/gi,
    (_match, className, innerContent) => {
      // Normalize the custom block structure
      return normalizeCustomBlock(className, innerContent);
    }
  );
}

/**
 * Normalizes custom block HTML structure
 * @param className CSS class string
 * @param content Inner content
 * @returns Normalized HTML structure
 */
function normalizeCustomBlock(className: string, content: string): string {
  const classes = className.split(' ');

  // Detect block type
  let blockType = '';
  if (classes.includes('tip')) blockType = 'tip';
  else if (classes.includes('warning')) blockType = 'warning';
  else if (classes.includes('danger')) blockType = 'danger';
  else if (classes.includes('details')) blockType = 'details';

  if (!blockType) {
    return `<div class="${className}">${content}</div>`;
  }

  // Extract and clean content
  const cleanContent = content.trim();

  // Add data attributes for easier processing
  return `<div class="${className}" data-block-type="${blockType}">${cleanContent}</div>`;
}

/**
 * Checks if content contains HTML blocks
 * @param content Markdown content
 * @returns True if contains custom blocks
 */
export function hasHtmlBlocks(content: string): boolean {
  return /class="[^"]*custom-block[^"]*"/i.test(content);
}

/**
 * Pre-processes markdown content for HTML blocks
 * @param content Raw markdown content
 * @param options Processing options
 * @returns Pre-processed content
 */
export function preprocessHtmlBlocks(
  content: string,
  options?: MarkdownProcessorOptions
): string {
  if (!options?.enableHtmlBlocks) {
    return content;
  }

  return processHtmlBlocks(content, options);
}

/**
 * Generates markdown overrides for HTML blocks processing
 * @param options Processing options
 * @returns Markdown-to-jsx overrides
 */
export function getHtmlBlocksOverrides(options?: MarkdownProcessorOptions) {
  if (!options?.enableHtmlBlocks) {
    return {};
  }

  return {
    div: parseCustomBlockDiv,
  };
}

/**
 * Parses custom block attributes from HTML
 * @param htmlString HTML string
 * @returns Parsed attributes
 */
export function parseCustomBlockAttributes(htmlString: string): {
  type?: string;
  className?: string;
  style?: string;
} {
  const classMatch = htmlString.match(/class="([^"]*)"/);
  const styleMatch = htmlString.match(/style="([^"]*)"/);
  const typeMatch = htmlString.match(/data-block-type="([^"]*)"/);

  return {
    type: typeMatch?.[1],
    className: classMatch?.[1],
    style: styleMatch?.[1]
  };
}

/**
 * Extracts title from custom block content
 * @param content Block content
 * @returns Title and remaining content
 */
export function extractBlockTitle(content: string): {
  title?: string;
  content: string;
} {
  // Look for title patterns at the beginning
  const titlePatterns = [
    /^<h[3-6][^>]*>([^<]+)<\/h[3-6]>/i,
    /^<strong[^>]*>([^<]+)<\/strong>/i,
    /^<b[^>]*>([^<]+)<\/b>/i,
    /^\*\*([^*]+)\*\*/,
    /^__([^_]+)__/
  ];

  for (const pattern of titlePatterns) {
    const match = content.match(pattern);
    if (match) {
      const title = match[1].trim();
      const remainingContent = content.replace(pattern, '').trim();
      return { title, content: remainingContent };
    }
  }

  return { content };
}

/**
 * Validates HTML block structure
 * @param htmlString HTML string
 * @returns True if valid custom block
 */
export function validateHtmlBlock(htmlString: string): boolean {
  // Check for required custom-block class
  if (!htmlString.includes('custom-block')) {
    return false;
  }

  // Check for proper div structure
  const divPattern = /<div[^>]*>[\s\S]*<\/div>/i;
  if (!divPattern.test(htmlString)) {
    return false;
  }

  // Check for valid block type
  const validTypes = ['tip', 'warning', 'danger', 'details'];
  const hasValidType = validTypes.some(type =>
    htmlString.includes(`class="[^"]*${type}[^"]*"`) ||
    htmlString.includes(`class="[^"]*custom-block[^"]*${type}[^"]*"`)
  );

  return hasValidType;
}

/**
 * Cleans and sanitizes HTML block content
 * @param content Raw content
 * @returns Cleaned content
 */
export function sanitizeBlockContent(content: string): string {
  return content
    .trim()
    .replace(/^\s*[\r\n]/gm, '') // Remove empty lines at start
    .replace(/[\r\n]\s*$/gm, '') // Remove empty lines at end
    .replace(/\s+/g, ' ') // Normalize whitespace
    .trim();
}

/**
 * Transforms legacy VitePress block syntax to standard HTML
 * @param content Markdown content
 * @returns Content with transformed blocks
 */
export function transformLegacyBlocks(content: string): string {
  // Transform VitePress-style blocks
  return content
    .replace(/::: tip\s*([^\n]*)\n([\s\S]*?):::/gi,
      (_match, title, body) =>
        `<div class="tip custom-block">${title ? `<strong>${title}</strong>\n` : ''}${body.trim()}</div>`
    )
    .replace(/::: warning\s*([^\n]*)\n([\s\S]*?):::/gi,
      (_match, title, body) =>
        `<div class="warning custom-block">${title ? `<strong>${title}</strong>\n` : ''}${body.trim()}</div>`
    )
    .replace(/::: danger\s*([^\n]*)\n([\s\S]*?):::/gi,
      (_match, title, body) =>
        `<div class="danger custom-block">${title ? `<strong>${title}</strong>\n` : ''}${body.trim()}</div>`
    )
    .replace(/::: details\s*([^\n]*)\n([\s\S]*?):::/gi,
      (_match, title, body) =>
        `<div class="details custom-block">${title ? `<strong>${title}</strong>\n` : ''}${body.trim()}</div>`
    );
}

export default {
  processHtmlBlocks,
  hasHtmlBlocks,
  preprocessHtmlBlocks,
  getHtmlBlocksOverrides,
  parseCustomBlockAttributes,
  extractBlockTitle,
  validateHtmlBlock,
  sanitizeBlockContent,
  transformLegacyBlocks
};
