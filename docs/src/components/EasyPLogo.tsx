import React from 'react';

interface EasyPLogoProps {
  size?: number;
  className?: string;
  showGlow?: boolean;
  variant?: 'default' | 'mono' | 'gradient';
}

const EasyPLogo: React.FC<EasyPLogoProps> = ({
  size = 32,
  className = '',
  showGlow = true,
  variant = 'gradient'
}) => {
  const id = `easyp-logo-${Math.random().toString(36).substr(2, 9)}`;

  const fillColor = variant === 'mono'
    ? '#e2e8f0'
    : variant === 'gradient'
    ? `url(#${id}-gradient)`
    : '#3b82f6';

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 32 32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
      role="img"
      aria-label="EasyP Logo"
    >
      <defs>
        <linearGradient id={`${id}-gradient`} x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#60a5fa" stopOpacity={1} />
          <stop offset="100%" stopColor="#3b82f6" stopOpacity={1} />
        </linearGradient>
        <linearGradient id={`${id}-bg-gradient`} x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#3b82f6" stopOpacity={0.2} />
          <stop offset="100%" stopColor="#8b5cf6" stopOpacity={0.1} />
        </linearGradient>
        {showGlow && (
          <filter id={`${id}-glow`}>
            <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
            <feMerge>
              <feMergeNode in="coloredBlur"/>
              <feMergeNode in="SourceGraphic"/>
            </feMerge>
          </filter>
        )}
        <filter id={`${id}-shadow`}>
          <feDropShadow dx="0" dy="1" stdDeviation="2" floodOpacity="0.2"/>
        </filter>
      </defs>

      {/* Background rounded square */}
      <rect
        x="2"
        y="2"
        width="28"
        height="28"
        rx="6"
        ry="6"
        fill="#1e293b"
        stroke={`url(#${id}-gradient)`}
        strokeWidth="1.5"
        opacity="0.9"
        filter={`url(#${id}-shadow)`}
      />

      {/* Inner gradient overlay */}
      <rect
        x="2"
        y="2"
        width="28"
        height="28"
        rx="6"
        ry="6"
        fill={`url(#${id}-bg-gradient)`}
      />

      {/* Letter E - modernized design */}
      <path
        d="M 8 10 L 13 10 Q 13 10 13 10.5 L 13 11 Q 13 11.5 12.5 11.5 L 9.5 11.5 L 9.5 15 L 12 15 Q 12.5 15 12.5 15.5 L 12.5 16 Q 12.5 16.5 12 16.5 L 9.5 16.5 L 9.5 20 L 13 20 Q 13 20 13 20.5 L 13 21 Q 13 21.5 12.5 21.5 L 8 21.5 Q 8 21.5 8 21 L 8 10.5 Q 8 10 8 10 Z"
        fill={fillColor}
        filter={showGlow ? `url(#${id}-glow)` : undefined}
      />

      {/* Letter P - modernized design */}
      <path
        d="M 17 10 L 22 10 Q 24 10 24 12 L 24 14 Q 24 16 22 16 L 18.5 16 L 18.5 21 Q 18.5 21.5 18 21.5 L 17.5 21.5 Q 17 21.5 17 21 L 17 10.5 Q 17 10 17 10 Z M 18.5 11.5 L 18.5 14.5 L 21.5 14.5 Q 22.5 14.5 22.5 13.5 L 22.5 12.5 Q 22.5 11.5 21.5 11.5 L 18.5 11.5 Z"
        fill={fillColor}
        filter={showGlow ? `url(#${id}-glow)` : undefined}
      />

      {/* Accent dot in corner */}
      <circle
        cx="26"
        cy="6"
        r="1.5"
        fill="#60a5fa"
        opacity="0.8"
        filter={showGlow ? `url(#${id}-glow)` : undefined}
      >
        <animate
          attributeName="opacity"
          values="0.8;1;0.8"
          dur="2s"
          repeatCount="indefinite"
        />
      </circle>

      {/* Secondary accent dot */}
      <circle
        cx="6"
        cy="26"
        r="1"
        fill="#8b5cf6"
        opacity="0.6"
      >
        <animate
          attributeName="opacity"
          values="0.6;0.8;0.6"
          dur="2s"
          begin="1s"
          repeatCount="indefinite"
        />
      </circle>
    </svg>
  );
};

export default EasyPLogo;
