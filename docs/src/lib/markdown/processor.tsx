import type {
  MarkdownProcessorOptions,
  ProcessedMarkdown,
  TocItem,
  FrontmatterData
} from './types';
import { parseFrontmatter } from './utils/frontmatter';
import { generateToc, extractHeadings } from './utils/headings';
import { processToc } from './plugins/toc';
import { processCodeFocus } from './plugins/codeFocus';
import { processHtmlBlocks } from './plugins/htmlBlocks';
import { processLinks } from './plugins/links';
import { processMermaidAsHtml } from './plugins/mermaid';

/**
 * Main markdown processor
 * Orchestrates all plugins and transformations
 */
export class MarkdownProcessor {
  private options: MarkdownProcessorOptions;

  constructor(options: MarkdownProcessorOptions = {}) {
    this.options = {
      enableToc: true,
      enableCodeFocus: true,
      enableHtmlBlocks: true,
      enableLinkProcessing: true,
      tocDepth: { min: 2, max: 4 },
      ...options
    };
  }

  /**
   * Processes markdown content through all enabled plugins
   * @param content Raw markdown content
   * @returns Processed markdown data
   */
  process(content: string): ProcessedMarkdown {
    try {
      // Step 1: Parse frontmatter
      const { content: rawContent, frontmatter, hasToc } = parseFrontmatter(content);

      // Step 2: Pre-process content through plugins
      let processedContent = rawContent;

      // Process Mermaid diagrams FIRST (before any other processing including markdown)
      processedContent = processMermaidAsHtml(processedContent);

      // Process TOC
      if (this.options.enableToc && hasToc) {
        const tocResult = processToc(processedContent, this.options);
        processedContent = tocResult.content;
      }

      // Process code focus
      if (this.options.enableCodeFocus) {
        processedContent = processCodeFocus(processedContent, this.options);
      }

      // Process HTML blocks
      if (this.options.enableHtmlBlocks) {
        processedContent = processHtmlBlocks(processedContent, this.options);
      }

      // Process links
      if (this.options.enableLinkProcessing) {
        processedContent = processLinks(processedContent, this.options);
      }

      // Step 3: Generate TOC if needed
      let toc: TocItem[] | undefined;
      if ((this.options.enableToc && hasToc) || this.options.enableToc) {
        const tocDepth = this.options.tocDepth || { min: 2, max: 4 };
        toc = generateToc(processedContent, tocDepth.min, tocDepth.max);

        // Add heading IDs for anchor links
        if (toc.length > 0) {
          const headings = toc.map(item => ({
            id: item.id,
            text: item.text,
            level: item.level,
            element: null as any
          }));
          // Note: We don't modify the content string here because markdown-to-jsx
          // doesn't support the {#id} syntax and renders it as text.
        }
      }

      // Step 4: Return processed data (content as string, will be rendered later)
      return {
        content: processedContent,
        toc: toc && toc.length > 0 ? toc : undefined,
        frontmatter
      };

    } catch (error) {
      console.error('Error processing markdown:', error);

      // Return fallback data
      return {
        content,
        toc: undefined,
        frontmatter: {}
      };
    }
  }

  /**
   * Updates processor options
   * @param options New options to merge
   */
  updateOptions(options: Partial<MarkdownProcessorOptions>) {
    this.options = { ...this.options, ...options };
  }

  /**
   * Gets current processor options
   * @returns Current options
   */
  getOptions(): MarkdownProcessorOptions {
    return { ...this.options };
  }
}

/**
 * Convenience function for processing markdown
 * @param content Raw markdown content
 * @param options Processing options
 * @returns Processed markdown data
 */
export function processMarkdown(
  content: string,
  options?: MarkdownProcessorOptions
): ProcessedMarkdown {
  const processor = new MarkdownProcessor(options);
  return processor.process(content);
}

/**
 * Pre-processes markdown content for rendering
 * This runs all transformations but doesn't render to React
 * @param content Raw markdown content
 * @param options Processing options
 * @returns Pre-processed content string
 */
export function preprocessMarkdown(
  content: string,
  options?: MarkdownProcessorOptions
): {
  content: string;
  toc?: TocItem[];
  frontmatter: FrontmatterData;
} {
  const processed = processMarkdown(content, options);

  return {
    content: typeof processed.content === 'string' ? processed.content : content,
    toc: processed.toc,
    frontmatter: processed.frontmatter || {}
  };
}

/**
 * Validates markdown content
 * @param content Markdown content to validate
 * @returns Validation result
 */
export function validateMarkdown(content: string): {
  isValid: boolean;
  errors: string[];
  warnings: string[];
} {
  const errors: string[] = [];
  const warnings: string[] = [];

  try {
    // Check if content is empty
    if (!content || content.trim().length === 0) {
      warnings.push('Content is empty');
    }

    // Try to parse frontmatter
    try {
      parseFrontmatter(content);
    } catch (error) {
      errors.push(`Frontmatter parsing error: ${(error as Error).message}`);
    }

    // Check for common markdown issues
    if (content.includes('[[toc]]') && !content.match(/^#{1,6}\s+/m)) {
      warnings.push('TOC placeholder found but no headings detected');
    }

    // Check for unmatched brackets in links
    const linkMatches = content.match(/\[[^\]]*\]/g) || [];
    const urlMatches = content.match(/\([^)]*\)/g) || [];

    if (linkMatches.length !== urlMatches.length) {
      warnings.push('Possible unmatched markdown link syntax');
    }

  } catch (error) {
    errors.push(`Validation error: ${(error as Error).message}`);
  }

  return {
    isValid: errors.length === 0,
    errors,
    warnings
  };
}

/**
 * Extracts metadata from markdown content
 * @param content Markdown content
 * @returns Extracted metadata
 */
export function extractMarkdownMetadata(content: string): {
  frontmatter: FrontmatterData;
  headingCount: number;
  wordCount: number;
  readingTime: number; // in minutes
  hasCodeBlocks: boolean;
  hasLinks: boolean;
  hasToc: boolean;
} {
  const { frontmatter, hasToc } = parseFrontmatter(content);

  // Extract content without frontmatter
  const cleanContent = content.replace(/^---[\s\S]*?---\n?/, '');

  // Count words (approximate)
  const words = cleanContent.match(/\b\w+\b/g) || [];
  const wordCount = words.length;

  // Estimate reading time (average 200 words per minute)
  const readingTime = Math.ceil(wordCount / 200);

  // Count headings
  const headings = extractHeadings(cleanContent);

  // Check for code blocks
  const hasCodeBlocks = /```[\s\S]*?```/g.test(cleanContent);

  // Check for links
  const hasLinks = /\[([^\]]+)\]\(([^)]+)\)/g.test(cleanContent);

  return {
    frontmatter,
    headingCount: headings.length,
    wordCount,
    readingTime,
    hasCodeBlocks,
    hasLinks,
    hasToc
  };
}

export default MarkdownProcessor;
