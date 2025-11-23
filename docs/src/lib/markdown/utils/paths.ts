/**
 * Checks if a URL is external (absolute URL with protocol)
 * @param url URL to check
 * @returns True if external, false if internal
 */
export function isExternalUrl(url: string): boolean {
  try {
    new URL(url);
    return true;
  } catch {
    return false;
  }
}

/**
 * Checks if a path is relative (starts with ./ or ../)
 * @param path Path to check
 * @returns True if relative, false otherwise
 */
export function isRelativePath(path: string): boolean {
  return path.startsWith('./') || path.startsWith('../');
}

/**
 * Checks if a path is an anchor link (starts with #)
 * @param path Path to check
 * @returns True if anchor link, false otherwise
 */
export function isAnchorLink(path: string): boolean {
  return path.startsWith('#');
}

/**
 * Normalizes a path by removing double slashes and resolving relative paths
 * @param path Path to normalize
 * @returns Normalized path
 */
export function normalizePath(path: string): string {
  // Remove double slashes
  let normalized = path.replace(/\/+/g, '/');

  // Remove trailing slash unless it's the root
  if (normalized.length > 1 && normalized.endsWith('/')) {
    normalized = normalized.slice(0, -1);
  }

  return normalized;
}

/**
 * Joins path segments with proper slash handling
 * @param segments Path segments to join
 * @returns Joined path
 */
export function joinPaths(...segments: string[]): string {
  const filtered = segments.filter(segment => segment && segment !== '.');

  if (filtered.length === 0) {
    return '';
  }

  let joined = filtered[0];

  for (let i = 1; i < filtered.length; i++) {
    const segment = filtered[i];

    // Add slash if needed
    if (!joined.endsWith('/') && !segment.startsWith('/')) {
      joined += '/';
    }

    // Remove duplicate slash
    if (joined.endsWith('/') && segment.startsWith('/')) {
      joined += segment.slice(1);
    } else {
      joined += segment;
    }
  }

  return normalizePath(joined);
}

/**
 * Resolves a relative markdown path to an absolute path
 * @param currentPath Current page path
 * @param relativePath Relative path from markdown link
 * @param baseUrl Base URL for the application
 * @returns Resolved absolute path
 */
export function resolveMarkdownPath(
  currentPath: string,
  relativePath: string,
  baseUrl?: string
): string {
  // Handle anchor links
  if (isAnchorLink(relativePath)) {
    return relativePath;
  }

  // Handle external URLs
  if (isExternalUrl(relativePath)) {
    return relativePath;
  }

  // Handle absolute paths (starting with /)
  if (relativePath.startsWith('/')) {
    return baseUrl ? joinPaths(baseUrl, relativePath) : relativePath;
  }

  // Handle relative paths
  const currentDir = currentPath.split('/').slice(0, -1).join('/');
  const resolvedPath = joinPaths(currentDir, relativePath);

  return baseUrl ? joinPaths(baseUrl, resolvedPath) : resolvedPath;
}

/**
 * Converts a file path to a URL path (removes .md extension, handles index files)
 * @param filePath File path (e.g., "guide/introduction/what-is.md")
 * @returns URL path (e.g., "/guide/introduction/what-is")
 */
export function filePathToUrlPath(filePath: string): string {
  let urlPath = filePath;

  // Remove .md extension
  if (urlPath.endsWith('.md')) {
    urlPath = urlPath.slice(0, -3);
  }

  // Handle index files (convert "folder/index" to "folder/")
  if (urlPath.endsWith('/index')) {
    urlPath = urlPath.slice(0, -6);
  } else if (urlPath === 'index') {
    urlPath = '';
  }

  // Ensure starts with /
  if (!urlPath.startsWith('/') && urlPath !== '') {
    urlPath = '/' + urlPath;
  }

  // Handle empty path (root)
  if (urlPath === '') {
    urlPath = '/';
  }

  return urlPath;
}

/**
 * Extracts the directory path from a file path
 * @param filePath File path
 * @returns Directory path
 */
export function getDirectoryPath(filePath: string): string {
  const segments = filePath.split('/');
  segments.pop(); // Remove filename
  return segments.join('/') || '/';
}

/**
 * Extracts filename from a path
 * @param path File path
 * @returns Filename
 */
export function getFileName(path: string): string {
  return path.split('/').pop() || '';
}

/**
 * Extracts file extension from a path
 * @param path File path
 * @returns File extension (without dot) or empty string
 */
export function getFileExtension(path: string): string {
  const fileName = getFileName(path);
  const dotIndex = fileName.lastIndexOf('.');

  if (dotIndex === -1 || dotIndex === 0) {
    return '';
  }

  return fileName.slice(dotIndex + 1);
}

/**
 * Checks if a path points to a markdown file
 * @param path File path
 * @returns True if markdown file, false otherwise
 */
export function isMarkdownFile(path: string): boolean {
  const extension = getFileExtension(path);
  return extension === 'md' || extension === 'markdown';
}

/**
 * Resolves a link href for React Router navigation
 * @param href Original href
 * @param currentPath Current page path
 * @param baseUrl Base URL
 * @returns Resolved href for React Router
 */
export function resolveRouterHref(
  href: string,
  currentPath: string,
  baseUrl?: string
): string {
  // External links and anchors stay as is
  if (isExternalUrl(href) || isAnchorLink(href)) {
    return href;
  }

  // Resolve the path
  let resolvedPath = resolveMarkdownPath(currentPath, href, baseUrl);

  // Convert to URL path if it's a markdown file
  if (isMarkdownFile(resolvedPath)) {
    resolvedPath = filePathToUrlPath(resolvedPath);
  }

  return resolvedPath;
}
