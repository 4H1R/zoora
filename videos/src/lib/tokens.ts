// Design tokens mirrored from frontend/src/styles.css (zinc neutrals + green-600 brand)
export const colors = {
  // Brand green scale (tailwind green)
  green50: '#f0fdf4',
  green100: '#dcfce7',
  green200: '#bbf7d0',
  green300: '#86efac',
  green400: '#4ade80',
  green500: '#22c55e',
  green600: '#16a34a',
  green700: '#15803d',
  green800: '#166534',
  green900: '#14532d',
  // Neutral scale (zinc)
  zinc50: '#fafafa',
  zinc100: '#f4f4f5',
  zinc200: '#e4e4e7',
  zinc300: '#d4d4d8',
  zinc400: '#a1a1aa',
  zinc500: '#71717a',
  zinc600: '#52525b',
  zinc700: '#3f3f46',
  zinc800: '#27272a',
  zinc900: '#18181b',
  zinc950: '#09090b',
  // Semantic
  white: '#ffffff',
  live: '#dc2626',
  warning: '#d97706',
  info: '#2563eb',
} as const;

export const font = "'Vazirmatn Variable', sans-serif";

export const radius = {
  sm: 6,
  md: 8,
  lg: 10,
  xl: 14,
  xxl: 18,
  xxxl: 22,
} as const;

export const shadow = {
  md: '0 4px 6px -1px rgb(0 0 0 / 0.06), 0 2px 4px -2px rgb(0 0 0 / 0.05)',
  lg: '0 10px 15px -3px rgb(0 0 0 / 0.08), 0 4px 6px -4px rgb(0 0 0 / 0.06)',
  xl: '0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.08)',
  window: '0 40px 80px -20px rgb(0 0 0 / 0.35)',
} as const;

const FA_DIGITS = '۰۱۲۳۴۵۶۷۸۹';
export const toFa = (value: string | number): string =>
  String(value).replace(/\d/g, (d) => FA_DIGITS[Number(d)]);
