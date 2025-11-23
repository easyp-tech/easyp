import type { TocItem, HeadingData } from '../types';

/**
 * Извлекает текст из React children
 * @param children React children (может быть строкой, объектом или массивом)
 * @returns Строка с текстом
 */
function extractTextFromChildren(children: any): string {
  if (typeof children === 'string') {
    return children;
  }

  if (Array.isArray(children)) {
    return children.map(extractTextFromChildren).join('');
  }

  if (children && typeof children === 'object' && children.props) {
    return extractTextFromChildren(children.props.children);
  }

  return '';
}

/**
 * Generates a URL-safe ID from heading text
 * @param text The heading text (может быть строкой или React children)
 * @returns URL-safe ID
 */
export function generateHeadingId(text: any): string {
  const textString = extractTextFromChildren(text);

  return textString
    .toLowerCase()
    .replace(/[^\w\s-]/g, '') // Remove special characters except spaces and hyphens
    .replace(/\s+/g, '-') // Replace spaces with hyphens
    .replace(/-+/g, '-') // Replace multiple hyphens with single hyphen
    .replace(/^-|-$/g, ''); // Remove leading/trailing hyphens
}

/**
 * Extracts headings from markdown content
 * @param content Markdown content
 * @param minLevel Minimum heading level to include (default: 2)
 * @param maxLevel Maximum heading level to include (default: 4)
 * @returns Array of heading data
 */
export function extractHeadings(
  content: string,
  minLevel: number = 2,
  maxLevel: number = 4
): HeadingData[] {
  const headings: HeadingData[] = [];
  const lines = content.split('\n');

  for (const line of lines) {
    // Match markdown headings (### Text)
    const headingMatch = line.match(/^(#{1,6})\s+(.+)$/);

    if (headingMatch) {
      const level = headingMatch[1].length;
      const text = headingMatch[2].trim();

      // Only include headings within the specified level range
      if (level >= minLevel && level <= maxLevel) {
        const id = generateHeadingId(text);

        headings.push({
          id,
          text,
          level,
          element: null as any, // Will be populated when rendered in DOM
        });
      }
    }
  }

  return headings;
}

/**
 * Builds a nested TOC structure from flat headings array
 * @param headings Flat array of headings
 * @returns Nested TOC structure
 */
export function buildTocStructure(headings: HeadingData[]): TocItem[] {
  const toc: TocItem[] = [];
  const stack: TocItem[] = [];

  for (const heading of headings) {
    const tocItem: TocItem = {
      id: heading.id,
      text: heading.text,
      level: heading.level,
      children: [],
    };

    // Find the correct parent level
    while (stack.length > 0 && stack[stack.length - 1].level >= heading.level) {
      stack.pop();
    }

    // Add to parent or root
    if (stack.length === 0) {
      toc.push(tocItem);
    } else {
      const parent = stack[stack.length - 1];
      if (!parent.children) {
        parent.children = [];
      }
      parent.children.push(tocItem);
    }

    stack.push(tocItem);
  }

  return toc;
}

/**
 * Adds IDs to headings in markdown content
 * @param content Markdown content
 * @param headings Array of heading data with IDs
 * @returns Content with heading IDs added
 */
export function addHeadingIds(content: string, headings: HeadingData[]): string {
  let modifiedContent = content;

  for (const heading of headings) {
    // Find the heading in content and add an id
    const headingRegex = new RegExp(`^(#{${heading.level}})\\s+${escapeRegex(heading.text)}$`, 'gm');
    modifiedContent = modifiedContent.replace(
      headingRegex,
      `$1 ${heading.text} {#${heading.id}}`
    );
  }

  return modifiedContent;
}

/**
 * Escapes special regex characters in text
 * @param text Text to escape
 * @returns Escaped text
 */
function escapeRegex(text: string): string {
  return text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

/**
 * Generates TOC from markdown content
 * @param content Markdown content
 * @param minLevel Minimum heading level (default: 2)
 * @param maxLevel Maximum heading level (default: 4)
 * @returns TOC structure
 */
export function generateToc(
  content: string,
  minLevel: number = 2,
  maxLevel: number = 4
): TocItem[] {
  const headings = extractHeadings(content, minLevel, maxLevel);
  return buildTocStructure(headings);
}

/**
 * Flattens nested TOC structure for easier iteration
 * @param toc Nested TOC structure
 * @returns Flat array of TOC items
 */
export function flattenToc(toc: TocItem[]): TocItem[] {
  const flattened: TocItem[] = [];

  function traverse(items: TocItem[]) {
    for (const item of items) {
      flattened.push(item);
      if (item.children && item.children.length > 0) {
        traverse(item.children);
      }
    }
  }

  traverse(toc);
  return flattened;
}
