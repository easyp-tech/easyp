/**
 * EasyP Markdown Library
 *
 * A comprehensive markdown processing library for React applications
 * with support for Table of Contents, code focus, HTML blocks, and link processing.
 */

// Main processor and types
export { MarkdownProcessor, processMarkdown, preprocessMarkdown, validateMarkdown, extractMarkdownMetadata } from './processor';
export type { MarkdownProcessorOptions, ProcessedMarkdown } from './types';
export type {
  TocItem,
  HeadingData,
  CodeBlockProps,
  CustomBlockProps,
  LinkComponentProps,
  HeadingComponentProps,
  MarkdownContentProps,
  FrontmatterData,
  ParsedMarkdown
} from './types';

// Main React components
export { MarkdownContent, useMarkdownProcessor } from './components/MarkdownContent';
export { TableOfContents, useActiveHeading } from './components/TableOfContents';
export { CodeBlock, MarkdownCodeBlock, MarkdownPre } from './components/CodeBlock';
export { CustomBlock, parseCustomBlockDiv, TipBlock, WarningBlock, DangerBlock, DetailsBlock } from './components/CustomBlock';
export { Heading, H1, H2, H3, H4, H5, H6 } from './components/Heading';

// Plugin exports
export { default as tocPlugin } from './plugins/toc';
export { default as codeFocusPlugin } from './plugins/codeFocus';
export { default as htmlBlocksPlugin } from './plugins/htmlBlocks';
export { default as linksPlugin } from './plugins/links';

// Utilities
export {
  parseFrontmatter,
  removeTocPlaceholder,
  extractTitle,
  extractDescription
} from './utils/frontmatter';

export {
  generateHeadingId,
  extractHeadings,
  buildTocStructure,
  addHeadingIds,
  generateToc,
  flattenToc
} from './utils/headings';

export {
  isExternalUrl,
  isRelativePath,
  isAnchorLink,
  normalizePath,
  joinPaths,
  resolveMarkdownPath,
  filePathToUrlPath,
  getDirectoryPath,
  getFileName,
  getFileExtension,
  isMarkdownFile,
  resolveRouterHref
} from './utils/paths';

// Default configuration
const defaultOptions = {
  enableToc: true,
  enableCodeFocus: true,
  enableHtmlBlocks: true,
  enableLinkProcessing: true,
  tocDepth: { min: 2, max: 4 }
} as const;

export { defaultOptions };

// Version info
export const version = '1.0.0';

// Re-export main components for convenience
import { MarkdownProcessor as MP, processMarkdown as pm, validateMarkdown as vm, extractMarkdownMetadata as emm } from './processor';
import { MarkdownContent as MC } from './components/MarkdownContent';

// Convenience wrapper for quick usage
export function createMarkdownProcessor(options?: any) {
  return new MP({ ...defaultOptions, ...options });
}

// Quick processing functions
export function quickProcess(content: string, options?: any) {
  return pm(content, { ...defaultOptions, ...options });
}

export function quickValidate(content: string) {
  return vm(content);
}

export function quickMetadata(content: string) {
  return emm(content);
}

// Default export
const markdownLib = {
  MarkdownProcessor: MP,
  MarkdownContent: MC,
  processMarkdown: pm,
  validateMarkdown: vm,
  defaultOptions,
  version
};

export default markdownLib;
