import React from 'react';
import type { HeadingComponentProps } from '../types';

export function Heading({
  level,
  id,
  children,
  className = '',
  ...props
}: HeadingComponentProps & React.HTMLAttributes<HTMLHeadingElement>) {
  const HeadingTag = `h${level}` as any;

  // Generate anchor link icon
  const AnchorIcon = () => (
    <svg
      className="heading-anchor-icon"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
      <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
    </svg>
  );

  const handleAnchorClick = (e: React.MouseEvent) => {
    e.preventDefault();
    if (id) {
      const element = document.getElementById(id);
      if (element) {
        element.scrollIntoView({
          behavior: 'smooth',
          block: 'start'
        });

        // Update URL hash
        if (window.history.pushState) {
          window.history.pushState(null, '', `#${id}`);
        }
      }
    }
  };

  const headingClassName = `
    markdown-heading
    markdown-heading--level-${level}
    ${id ? 'markdown-heading--with-anchor' : ''}
    ${className}
  `.trim();

  const Tag = HeadingTag;

  return (
    <Tag
      id={id}
      className={headingClassName}
      {...props}
    >
      <span className="heading-content">
        {children}
      </span>
      {id && (
        <a
          href={`#${id}`}
          className="heading-anchor"
          onClick={handleAnchorClick}
          aria-label={`Link to heading: ${typeof children === 'string' ? children : 'heading'}`}
          title="Direct link to heading"
        >
          <AnchorIcon />
        </a>
      )}
    </Tag>
  );
}

/**
 * Creates heading components for different levels
 */
export function createHeadingComponent(level: number) {
  return function HeadingComponent(props: Omit<HeadingComponentProps, 'level'>) {
    return <Heading level={level} {...props} />;
  };
}

// Pre-made heading components
export const H1 = createHeadingComponent(1);
export const H2 = createHeadingComponent(2);
export const H3 = createHeadingComponent(3);
export const H4 = createHeadingComponent(4);
export const H5 = createHeadingComponent(5);
export const H6 = createHeadingComponent(6);

export default Heading;
