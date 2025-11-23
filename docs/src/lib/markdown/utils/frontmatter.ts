import matter from 'gray-matter';
import type { FrontmatterData, ParsedMarkdown } from '../types';

/**
 * Parses frontmatter from markdown content
 * @param markdown The raw markdown content
 * @returns Parsed markdown with frontmatter and content separated
 */
export function parseFrontmatter(markdown: string): ParsedMarkdown {
  try {
    const parsed = matter(markdown);

    const frontmatter: FrontmatterData = {
      ...parsed.data,
    };

    // Check if TOC should be enabled based on frontmatter or [[toc]] presence
    const hasTocInFrontmatter = frontmatter.toc === true;
    const hasTocInContent = parsed.content.includes('[[toc]]');
    const hasToc = hasTocInFrontmatter || hasTocInContent;

    return {
      content: parsed.content,
      frontmatter,
      hasToc,
    };
  } catch (error) {
    console.warn('Failed to parse frontmatter:', error);

    // Fallback: treat entire content as markdown without frontmatter
    const hasToc = markdown.includes('[[toc]]');

    return {
      content: markdown,
      frontmatter: {},
      hasToc,
    };
  }
}

/**
 * Removes the [[toc]] placeholder from markdown content
 * @param content Markdown content
 * @returns Content with [[toc]] removed
 */
export function removeTocPlaceholder(content: string): string {
  return content.replace(/\[\[toc\]\]/g, '').trim();
}

/**
 * Extracts title from frontmatter or first heading
 * @param frontmatter Parsed frontmatter data
 * @param content Markdown content
 * @returns Extracted title or undefined
 */
export function extractTitle(frontmatter: FrontmatterData, content: string): string | undefined {
  // Try frontmatter first
  if (frontmatter.title && typeof frontmatter.title === 'string') {
    return frontmatter.title;
  }

  // Try to extract from first H1 heading
  const h1Match = content.match(/^#\s+(.+)$/m);
  if (h1Match) {
    return h1Match[1].trim();
  }

  return undefined;
}

/**
 * Extracts description from frontmatter
 * @param frontmatter Parsed frontmatter data
 * @returns Description or undefined
 */
export function extractDescription(frontmatter: FrontmatterData): string | undefined {
  if (frontmatter.description && typeof frontmatter.description === 'string') {
    return frontmatter.description;
  }

  return undefined;
}
