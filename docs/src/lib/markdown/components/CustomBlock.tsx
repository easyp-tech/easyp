import React from 'react';
import type { CustomBlockProps } from '../types';

export function CustomBlock({
  type,
  title,
  children,
  className = ''
}: CustomBlockProps) {
  const blockClassName = `custom-block custom-block--${type} ${className}`;

  const getDefaultTitle = () => {
    switch (type) {
      case 'tip':
        return 'TIP';
      case 'warning':
        return 'WARNING';
      case 'danger':
        return 'DANGER';
      case 'details':
        return 'DETAILS';
      default:
        return '';
    }
  };

  const getIcon = () => {
    switch (type) {
      case 'tip':
        return (
          <svg className="custom-block__icon" viewBox="0 0 24 24" fill="none">
            <path
              d="M12 2L2 7V17L12 22L22 17V7L12 2Z"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path
              d="M12 8V12"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <circle cx="12" cy="16" r="1" fill="currentColor" />
          </svg>
        );
      case 'warning':
        return (
          <svg className="custom-block__icon" viewBox="0 0 24 24" fill="none">
            <path
              d="M10.29 3.86L1.82 18A2 2 0 003.54 21H20.46A2 2 0 0022.18 18L13.71 3.86A2 2 0 0010.29 3.86Z"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path
              d="M12 9V13"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <circle cx="12" cy="17" r="1" fill="currentColor" />
          </svg>
        );
      case 'danger':
        return (
          <svg className="custom-block__icon" viewBox="0 0 24 24" fill="none">
            <circle
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="2"
            />
            <path
              d="M15 9L9 15"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path
              d="M9 9L15 15"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        );
      case 'details':
        return (
          <svg className="custom-block__icon" viewBox="0 0 24 24" fill="none">
            <circle
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="2"
            />
            <path
              d="M12 16V12"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <circle cx="12" cy="8" r="1" fill="currentColor" />
          </svg>
        );
      default:
        return null;
    }
  };

  const displayTitle = title || getDefaultTitle();

  if (type === 'details') {
    return (
      <details className={blockClassName}>
        <summary className="custom-block__summary">
          {getIcon()}
          <span className="custom-block__title">{displayTitle}</span>
        </summary>
        <div className="custom-block__content">
          {children}
        </div>
      </details>
    );
  }

  return (
    <div className={blockClassName}>
      {displayTitle && (
        <div className="custom-block__header">
          {getIcon()}
          <span className="custom-block__title">{displayTitle}</span>
        </div>
      )}
      <div className="custom-block__content">
        {children}
      </div>
    </div>
  );
}

/**
 * Parses custom block div elements from markdown
 * @param props HTML div props
 * @returns CustomBlock component or regular div
 */
export function parseCustomBlockDiv(props: any) {
  const { className, children, style, ...rest } = props;

  // Check if this is a custom block div
  if (typeof className === 'string' && className.includes('custom-block')) {
    const classes = className.split(' ');

    // Extract block type from className
    let blockType: CustomBlockProps['type'] | null = null;

    if (classes.includes('tip')) blockType = 'tip';
    else if (classes.includes('warning')) blockType = 'warning';
    else if (classes.includes('danger')) blockType = 'danger';
    else if (classes.includes('details')) blockType = 'details';

    if (blockType) {
      // Extract title from children if present
      let title: string | undefined;
      let content = children;

      // Check if first child is a title element or text
      if (React.Children.count(children) > 0) {
        const firstChild = React.Children.toArray(children)[0];

        if (React.isValidElement(firstChild) &&
            (firstChild.type === 'h3' || firstChild.type === 'h4' || firstChild.type === 'strong')) {
          const props = firstChild.props as any;
          title = typeof props.children === 'string'
            ? props.children
            : undefined;
          content = React.Children.toArray(children).slice(1);
        }
      }

      return (
        <CustomBlock
          type={blockType}
          title={title}
          className={className.replace('custom-block', '').replace(blockType, '').trim()}
          {...rest}
        >
          {content}
        </CustomBlock>
      );
    }
  }

  // Return regular div for non-custom blocks
  return <div className={className} style={style} {...rest}>{children}</div>;
}

/**
 * Creates a specific custom block component
 */
export function createCustomBlock(type: CustomBlockProps['type']) {
  return function CustomBlockComponent({ title, children, className, ...props }: Omit<CustomBlockProps, 'type'>) {
    return (
      <CustomBlock
        type={type}
        title={title}
        className={className}
        {...props}
      >
        {children}
      </CustomBlock>
    );
  };
}

// Pre-made components for each type
export const TipBlock = createCustomBlock('tip');
export const WarningBlock = createCustomBlock('warning');
export const DangerBlock = createCustomBlock('danger');
export const DetailsBlock = createCustomBlock('details');

export default CustomBlock;
