// Zoora Z-mark, recreated from frontend/public/favicon.svg
export const Logo: React.FC<{ size?: number }> = ({ size = 96 }) => (
  <svg width={size} height={size} viewBox="0 0 48 48" fill="none">
    <defs>
      <linearGradient id="zoora-z" x1="6" y1="4" x2="42" y2="46" gradientUnits="userSpaceOnUse">
        <stop offset="0" stopColor="#2fbd68" />
        <stop offset="0.55" stopColor="#16a34a" />
        <stop offset="1" stopColor="#15803d" />
      </linearGradient>
    </defs>
    <rect width="48" height="48" rx="13" fill="url(#zoora-z)" />
    <path
      d="M14 16 H34 L14 32 H34"
      stroke="#ffffff"
      strokeWidth="5.4"
      strokeLinecap="round"
      strokeLinejoin="round"
      fill="none"
    />
  </svg>
);
