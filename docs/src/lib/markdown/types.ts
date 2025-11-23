export interface TocItem {
  id: string;
  text: string;
  level: number;
  children?: TocItem[];
}

export interface MarkdownProcessorOptions {
  enableToc?: boolean;
  enableCodeFocus?: boolean;
  enableHtmlBlocks?: boolean;
  enableLinkProcessing?: boolean;
  baseUrl?: string;
  className?: string;
  tocDepth?: {
    min: number;
    max: number;
  };
}

export interface ProcessedMarkdown {
  content: React.ReactNode;
  toc?: TocItem[];
  frontmatter?: Record<string, any>;
}

export interface HeadingData {
  id: string;
  text: string;
  level: number;
  element: HTMLElement;
}

export interface CodeBlockProps {
  children?: string;
  className?: string;
  language?: string;
  hasFocus?: boolean;
  focusLines?: number[];
}

export interface CustomBlockProps {
  type: 'tip' | 'warning' | 'danger' | 'details';
  title?: string;
  children: React.ReactNode;
  className?: string;
}

export interface LinkComponentProps {
  href?: string;
  children: React.ReactNode;
  className?: string;
  title?: string;
}

export interface HeadingComponentProps {
  level: number;
  id?: string;
  children: React.ReactNode;
  className?: string;
}

export interface MarkdownContentProps {
  content: string;
  options?: MarkdownProcessorOptions;
  className?: string;
}

export interface FrontmatterData {
  [key: string]: any;
  title?: string;
  description?: string;
  toc?: boolean;
}

export interface ParsedMarkdown {
  content: string;
  frontmatter: FrontmatterData;
  hasToc: boolean;
}
