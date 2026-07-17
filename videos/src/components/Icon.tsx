// Minimal hand-drawn icon set (lucide-flavored) for mock UI at small sizes
export type IconName =
  | 'grid'
  | 'book'
  | 'calendar'
  | 'video'
  | 'users'
  | 'chart'
  | 'gear'
  | 'mic'
  | 'monitor'
  | 'chat'
  | 'phone'
  | 'play'
  | 'file'
  | 'bell'
  | 'check'
  | 'plus'
  | 'clock'
  | 'pen'
  | 'lock'
  | 'hand';

export const Icon: React.FC<{ name: IconName; size?: number; color?: string; strokeWidth?: number }> = ({
  name,
  size = 16,
  color = 'currentColor',
  strokeWidth = 1.8,
}) => {
  const common = {
    width: size,
    height: size,
    viewBox: '0 0 24 24',
    fill: 'none',
    stroke: color,
    strokeWidth,
    strokeLinecap: 'round' as const,
    strokeLinejoin: 'round' as const,
  };
  switch (name) {
    case 'grid':
      return (
        <svg {...common}>
          <rect x="3" y="3" width="7" height="7" rx="1.5" />
          <rect x="14" y="3" width="7" height="7" rx="1.5" />
          <rect x="3" y="14" width="7" height="7" rx="1.5" />
          <rect x="14" y="14" width="7" height="7" rx="1.5" />
        </svg>
      );
    case 'book':
      return (
        <svg {...common}>
          <path d="M12 6c-2-1.8-5-2-8-2v14c3 0 6 .2 8 2 2-1.8 5-2 8-2V4c-3 0-6 .2-8 2z" />
          <path d="M12 6v14" />
        </svg>
      );
    case 'calendar':
      return (
        <svg {...common}>
          <rect x="3" y="5" width="18" height="16" rx="2" />
          <path d="M8 3v4M16 3v4M3 10h18" />
        </svg>
      );
    case 'video':
      return (
        <svg {...common}>
          <rect x="2" y="6" width="13" height="12" rx="2" />
          <path d="M15 10l7-3v10l-7-3" />
        </svg>
      );
    case 'users':
      return (
        <svg {...common}>
          <circle cx="9" cy="8" r="3.5" />
          <path d="M2.5 20c0-3.5 3-5.5 6.5-5.5s6.5 2 6.5 5.5" />
          <path d="M16 5a3.5 3.5 0 0 1 0 6.5M21.5 20c0-2.8-1.8-4.6-4.3-5.3" />
        </svg>
      );
    case 'chart':
      return (
        <svg {...common}>
          <path d="M3 3v18h18" />
          <path d="M8 16v-5M13 16V8M18 16v-3" />
        </svg>
      );
    case 'gear':
      return (
        <svg {...common}>
          <circle cx="12" cy="12" r="3.2" />
          <path d="M12 2.5v3M12 18.5v3M2.5 12h3M18.5 12h3M5 5l2.1 2.1M16.9 16.9L19 19M19 5l-2.1 2.1M7.1 16.9L5 19" />
        </svg>
      );
    case 'mic':
      return (
        <svg {...common}>
          <rect x="9" y="3" width="6" height="11" rx="3" />
          <path d="M5.5 11a6.5 6.5 0 0 0 13 0M12 17.5V21" />
        </svg>
      );
    case 'monitor':
      return (
        <svg {...common}>
          <rect x="2.5" y="4" width="19" height="13" rx="2" />
          <path d="M9 21h6M12 17v4" />
        </svg>
      );
    case 'chat':
      return (
        <svg {...common}>
          <path d="M21 12a8 8 0 0 1-8 8H4l2-3.2A8 8 0 1 1 21 12z" />
        </svg>
      );
    case 'phone':
      return (
        <svg {...common}>
          <path d="M6 3h4l2 5-2.5 1.5a11 11 0 0 0 5 5L16 12l5 2v4a2 2 0 0 1-2 2A16 16 0 0 1 4 5a2 2 0 0 1 2-2z" />
        </svg>
      );
    case 'play':
      return (
        <svg {...common}>
          <path d="M7 5l12 7-12 7z" />
        </svg>
      );
    case 'file':
      return (
        <svg {...common}>
          <path d="M6 2.5h8l4 4V21.5H6z" />
          <path d="M14 2.5v4h4" />
        </svg>
      );
    case 'bell':
      return (
        <svg {...common}>
          <path d="M18 9a6 6 0 1 0-12 0c0 6-2.5 7-2.5 7h17S18 15 18 9z" />
          <path d="M10 20a2.2 2.2 0 0 0 4 0" />
        </svg>
      );
    case 'check':
      return (
        <svg {...common}>
          <path d="M4.5 12.5l5 5 10-11" />
        </svg>
      );
    case 'plus':
      return (
        <svg {...common}>
          <path d="M12 5v14M5 12h14" />
        </svg>
      );
    case 'clock':
      return (
        <svg {...common}>
          <circle cx="12" cy="12" r="9" />
          <path d="M12 7v5l3.5 2" />
        </svg>
      );
    case 'pen':
      return (
        <svg {...common}>
          <path d="M4 20l1-4L17.5 3.5a2.1 2.1 0 0 1 3 3L8 19z" />
        </svg>
      );
    case 'lock':
      return (
        <svg {...common}>
          <rect x="5" y="11" width="14" height="10" rx="2" />
          <path d="M8 11V7a4 4 0 0 1 8 0v4" />
        </svg>
      );
    case 'hand':
      return (
        <svg {...common}>
          <path d="M18 11V6a2 2 0 0 0-4 0v5" />
          <path d="M14 10V4a2 2 0 0 0-4 0v2" />
          <path d="M10 10.5V6a2 2 0 0 0-4 0v8" />
          <path d="M18 8a2 2 0 1 1 4 0v6a8 8 0 0 1-8 8h-2c-2.8 0-4.5-.86-5.99-2.34l-3.6-3.6a2 2 0 0 1 2.83-2.82L7 15" />
        </svg>
      );
  }
};
