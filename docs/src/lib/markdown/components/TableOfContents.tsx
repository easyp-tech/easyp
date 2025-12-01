import React from 'react';
import type { TocItem } from '../types';

interface TableOfContentsProps {
  toc: TocItem[];
  title?: string;
  className?: string;
  maxDepth?: number;
  activeId?: string;
}

export function TableOfContents({
  toc,
  title = "Table of Contents",
  className = "",
  maxDepth = 4,
  activeId
}: TableOfContentsProps) {
  if (!toc || toc.length === 0) {
    return null;
  }

  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>, id: string) => {
    e.preventDefault();

    const element = document.getElementById(id);
    if (element) {
      element.scrollIntoView({
        behavior: 'smooth',
        block: 'start'
      });

      // Update URL hash without triggering navigation
      if (window.history.pushState) {
        window.history.pushState(null, '', `#${id}`);
      }
    }
  };

  const renderTocItem = (item: TocItem, depth: number = 0) => {
    if (depth >= maxDepth) {
      return null;
    }

    const isActive = activeId === item.id;
    const hasChildren = item.children && item.children.length > 0;

    return (
      <li key={item.id} className={`toc-item toc-item--level-${item.level}`}>
        <a
          href={`#${item.id}`}
          className={`toc-link ${isActive ? 'toc-link--active' : ''}`}
          onClick={(e) => handleClick(e, item.id)}
          title={item.text}
        >
          {item.text}
        </a>
        {hasChildren && (
          <ul className={`toc-list toc-list--level-${item.level + 1}`}>
            {item.children!.map(child => renderTocItem(child, depth + 1))}
          </ul>
        )}
      </li>
    );
  };

  return (
    <nav className={`toc-container ${className}`} aria-label="Table of contents">
      <h3 className="toc-title">{title}</h3>
      <ul className="toc-list toc-list--root">
        {toc.map(item => renderTocItem(item))}
      </ul>
    </nav>
  );
}

// Hook to track active heading based on scroll position
export function useActiveHeading(toc: TocItem[]) {
  const [activeId, setActiveId] = React.useState<string>('');

  React.useEffect(() => {
    if (!toc || toc.length === 0) return;

    // Helper to flatten TOC for scroll tracking
    const getAllHeadings = (items: TocItem[]): TocItem[] => {
      return items.reduce((acc: TocItem[], item) => {
        acc.push(item);
        if (item.children && item.children.length > 0) {
          acc.push(...getAllHeadings(item.children));
        }
        return acc;
      }, []);
    };

    const allHeadings = getAllHeadings(toc);

    const handleScroll = () => {
      const headings = allHeadings.map(item => {
        const element = document.getElementById(item.id);
        if (element) {
          const rect = element.getBoundingClientRect();
          return {
            id: item.id,
            top: rect.top,
            element
          };
        }
        return null;
      }).filter(Boolean);

      if (headings.length === 0) return;

      // Find the heading that's currently in view
      const threshold = 100; // Distance from top of viewport
      let currentActive = '';

      for (const heading of headings) {
        if (heading!.top <= threshold) {
          currentActive = heading!.id;
        } else {
          break;
        }
      }

      // If no heading is above the threshold, use the first one
      if (!currentActive && headings.length > 0) {
        currentActive = headings[0]!.id;
      }

      setActiveId(currentActive);
    };

    // Initial check
    handleScroll();

    // Add scroll listener
    window.addEventListener('scroll', handleScroll, { passive: true });

    // Also check on resize in case layout changes
    window.addEventListener('resize', handleScroll, { passive: true });

    return () => {
      window.removeEventListener('scroll', handleScroll);
      window.removeEventListener('resize', handleScroll);
    };
  }, [toc]);

  return activeId;
}

export default TableOfContents;
