import React from 'react';
import type { MarkdownProcessorOptions, LinkComponentProps } from '../types';
import {
  isExternalUrl,
  isAnchorLink,
  resolveRouterHref,
  isMarkdownFile,
  filePathToUrlPath
} from '../utils/paths';

/**
 * Links Plugin for markdown processing
 * Handles link processing, routing, and external link handling
 */

/**
 * Processes links in markdown content
 * @param content Markdown content
 * @param options Processing options
 * @returns Processed content with links handled
 */
export function processLinks(
  content: string,
  options?: MarkdownProcessorOptions
): string {
  if (!options?.enableLinkProcessing) {
    return content;
  }

  // Process markdown links [text](url)
  return content.replace(
    /\[([^\]]+)\]\(([^)]+)\)/g,
    (_match, text, url) => {
      const processedUrl = processLinkUrl(url, options);
      return `[${text}](${processedUrl})`;
    }
  );
}

/**
 * Processes a single link URL
 * @param url Original URL
 * @param options Processing options
 * @returns Processed URL
 */
function processLinkUrl(url: string, options?: MarkdownProcessorOptions): string {
  // Don't process external URLs or anchors
  if (isExternalUrl(url) || isAnchorLink(url)) {
    return url;
  }

  // Convert markdown file paths to URL paths
  if (isMarkdownFile(url)) {
    return filePathToUrlPath(url);
  }

  // Add base URL if specified
  if (options?.baseUrl && !url.startsWith('/')) {
    return `${options.baseUrl.replace(/\/$/, '')}/${url}`;
  }

  return url;
}

/**
 * Link component for markdown-to-jsx
 */
export function LinkComponent({
  href = '',
  children,
  className = '',
  title,
  ...props
}: LinkComponentProps & React.AnchorHTMLAttributes<HTMLAnchorElement>) {
  const isExternal = isExternalUrl(href);
  const isAnchor = isAnchorLink(href);

  // Handle external links
  if (isExternal) {
    return (
      <a
        href={href}
        className={`markdown-link markdown-link--external ${className}`}
        title={title}
        target="_blank"
        rel="noopener noreferrer"
        {...props}
      >
        {children}
        <ExternalLinkIcon />
      </a>
    );
  }

  // Handle anchor links
  if (isAnchor) {
    const handleAnchorClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
      e.preventDefault();
      const targetId = href.substring(1);
      const element = document.getElementById(targetId);

      if (element) {
        element.scrollIntoView({
          behavior: 'smooth',
          block: 'start'
        });

        // Update URL hash
        if (window.history.pushState) {
          window.history.pushState(null, '', href);
        }
      }
    };

    return (
      <a
        href={href}
        className={`markdown-link markdown-link--anchor ${className}`}
        title={title}
        onClick={handleAnchorClick}
        {...props}
      >
        {children}
      </a>
    );
  }

  // Handle internal links (React Router navigation)
  const handleInternalClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
    // Check if React Router is available
    if (typeof window !== 'undefined' && (window as any).routerNavigate) {
      e.preventDefault();
      (window as any).routerNavigate(href);
    }
    // If React Router is not available, let the browser handle it
  };

  return (
    <a
      href={href}
      className={`markdown-link markdown-link--internal ${className}`}
      title={title}
      onClick={handleInternalClick}
      {...props}
    >
      {children}
    </a>
  );
}

/**
 * External link icon component
 */
function ExternalLinkIcon() {
  return (
    <svg
      className="external-link-icon"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
      <polyline points="15,3 21,3 21,9" />
      <line x1="10" y1="14" x2="21" y2="3" />
    </svg>
  );
}

/**
 * Checks if content contains links
 * @param content Markdown content
 * @returns True if contains links
 */
export function hasLinks(content: string): boolean {
  return /\[([^\]]+)\]\(([^)]+)\)/g.test(content);
}

/**
 * Pre-processes markdown content for links
 * @param content Raw markdown content
 * @param options Processing options
 * @returns Pre-processed content
 */
export function preprocessLinks(
  content: string,
  options?: MarkdownProcessorOptions
): string {
  if (!options?.enableLinkProcessing) {
    return content;
  }

  return processLinks(content, options);
}

/**
 * Generates markdown overrides for link processing
 * @param options Processing options
 * @returns Markdown-to-jsx overrides
 */
export function getLinksOverrides(options?: MarkdownProcessorOptions) {
  if (!options?.enableLinkProcessing) {
    return {};
  }

  return {
    a: (props: any) => <LinkComponent {...props} />
  };
}

/**
 * Extracts all links from markdown content
 * @param content Markdown content
 * @returns Array of link objects
 */
export function extractLinks(content: string): Array<{
  text: string;
  url: string;
  isExternal: boolean;
  isAnchor: boolean;
}> {
  const links: Array<{
    text: string;
    url: string;
    isExternal: boolean;
    isAnchor: boolean;
  }> = [];

  const linkRegex = /\[([^\]]+)\]\(([^)]+)\)/g;
  let match;

  while ((match = linkRegex.exec(content)) !== null) {
    const [, text, url] = match;
    links.push({
      text,
      url,
      isExternal: isExternalUrl(url),
      isAnchor: isAnchorLink(url)
    });
  }

  return links;
}

/**
 * Validates link URLs
 * @param url URL to validate
 * @returns True if valid URL
 */
export function validateLinkUrl(url: string): boolean {
  if (!url || url.trim() === '') {
    return false;
  }

  // Check for external URLs
  if (isExternalUrl(url)) {
    try {
      new URL(url);
      return true;
    } catch {
      return false;
    }
  }

  // Check for anchor links
  if (isAnchorLink(url)) {
    return url.length > 1; // Must have content after #
  }

  // Check for relative/absolute paths
  return url.length > 0;
}

/**
 * Normalizes link URLs
 * @param url Original URL
 * @param baseUrl Base URL for relative links
 * @returns Normalized URL
 */
export function normalizeLinkUrl(url: string, baseUrl?: string): string {
  if (!url) return '';

  // Don't modify external URLs or anchors
  if (isExternalUrl(url) || isAnchorLink(url)) {
    return url;
  }

  // Convert markdown files to URL paths
  if (isMarkdownFile(url)) {
    return filePathToUrlPath(url);
  }

  // Add base URL for relative paths
  if (baseUrl && !url.startsWith('/')) {
    const cleanBaseUrl = baseUrl.replace(/\/$/, '');
    return `${cleanBaseUrl}/${url}`;
  }

  return url;
}

/**
 * Transforms relative links to absolute based on current path
 * @param content Markdown content
 * @param currentPath Current file path
 * @param baseUrl Base URL
 * @returns Content with transformed links
 */
export function transformRelativeLinks(
  content: string,
  currentPath: string,
  baseUrl?: string
): string {
  return content.replace(
    /\[([^\]]+)\]\(([^)]+)\)/g,
    (match, text, url) => {
      if (isExternalUrl(url) || isAnchorLink(url)) {
        return match;
      }

      const resolvedUrl = resolveRouterHref(url, currentPath, baseUrl);
      return `[${text}](${resolvedUrl})`;
    }
  );
}

export default {
  processLinks,
  LinkComponent,
  hasLinks,
  preprocessLinks,
  getLinksOverrides,
  extractLinks,
  validateLinkUrl,
  normalizeLinkUrl,
  transformRelativeLinks
};
