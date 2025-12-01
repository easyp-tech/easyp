import React from 'react';
import type { TocItem, MarkdownProcessorOptions } from '../types';
import { generateToc } from '../utils/headings';
import { removeTocPlaceholder } from '../utils/frontmatter';
import { parseFrontmatter } from '../utils/frontmatter';

/**
 * TOC Plugin for markdown processing
 * Handles [[toc]] placeholder and generates table of contents
 */

/**
 * Processes TOC in markdown content
 * @param content Markdown content
 * @param options Processing options
 * @returns Processed content and TOC data
 */
export function processToc(
  content: string,
  options?: MarkdownProcessorOptions
): {
  content: string;
  toc: TocItem[] | undefined;
  hasToc: boolean;
} {
  const { hasToc } = parseFrontmatter(content);

  if (!hasToc && !options?.enableToc) {
    return {
      content,
      toc: undefined,
      hasToc: false
    };
  }

  // Remove [[toc]] placeholder from content
  const processedContent = removeTocPlaceholder(content);

  // Generate TOC from headings
  const tocDepth = options?.tocDepth || { min: 2, max: 4 };
  const toc = generateToc(processedContent, tocDepth.min, tocDepth.max);

  return {
    content: processedContent,
    toc: toc.length > 0 ? toc : undefined,
    hasToc: true
  };
}

/**
 * Checks if content contains TOC placeholder
 * @param content Markdown content
 * @returns True if contains [[toc]], false otherwise
 */
export function hasTocPlaceholder(content: string): boolean {
  return /\[\[toc\]\]/i.test(content);
}

/**
 * Replaces [[toc]] placeholder with a React component placeholder
 * This is used during markdown-to-jsx processing
 * @param content Markdown content
 * @returns Content with TOC placeholder replaced
 */
export function replaceTocPlaceholder(content: string): string {
  return content.replace(
    /\[\[toc\]\]/gi,
    '<TocPlaceholder />'
  );
}

/**
 * TOC Placeholder component that gets replaced during processing
 */
export function TocPlaceholder() {
  return <div className="toc-placeholder" />;
}

/**
 * Generates markdown overrides for TOC processing
 * @param toc TOC data
 * @returns Markdown-to-jsx overrides
 */
export function getTocOverrides(_toc?: TocItem[]) {
  return {
    TocPlaceholder: () => {
      // This will be replaced by the actual TOC component
      return null;
    }
  };
}

/**
 * Pre-processes markdown content for TOC
 * This runs before markdown-to-jsx processing
 * @param content Raw markdown content
 * @param options Processing options
 * @returns Pre-processed content
 */
export function preprocessToc(
  content: string,
  options?: MarkdownProcessorOptions
): string {
  if (!options?.enableToc && !hasTocPlaceholder(content)) {
    return content;
  }

  // Replace [[toc]] with component placeholder
  return replaceTocPlaceholder(content);
}

/**
 * Post-processes rendered content to inject TOC
 * @param renderedContent Rendered React content
 * @param toc TOC data
 * @returns Content with TOC injected
 */
export function postprocessToc(
  renderedContent: React.ReactNode,
  toc?: TocItem[]
): React.ReactNode {
  if (!toc || toc.length === 0) {
    return renderedContent;
  }

  // This would need to be implemented based on how we want to inject TOC
  // For now, return the original content
  return renderedContent;
}

/**
 * Validates TOC structure
 * @param toc TOC items
 * @returns True if valid, false otherwise
 */
export function validateToc(toc: TocItem[]): boolean {
  if (!Array.isArray(toc)) {
    return false;
  }

  for (const item of toc) {
    if (!item.id || !item.text || typeof item.level !== 'number') {
      return false;
    }

    if (item.children && !validateToc(item.children)) {
      return false;
    }
  }

  return true;
}

/**
 * Filters TOC items by depth
 * @param toc TOC items
 * @param maxDepth Maximum depth to include
 * @returns Filtered TOC items
 */
export function filterTocByDepth(toc: TocItem[], maxDepth: number): TocItem[] {
  return toc.map(item => {
    const filtered: TocItem = {
      id: item.id,
      text: item.text,
      level: item.level
    };

    if (item.children && item.level < maxDepth) {
      filtered.children = filterTocByDepth(item.children, maxDepth);
    }

    return filtered;
  });
}

export default {
  processToc,
  hasTocPlaceholder,
  replaceTocPlaceholder,
  TocPlaceholder,
  getTocOverrides,
  preprocessToc,
  postprocessToc,
  validateToc,
  filterTocByDepth
};
