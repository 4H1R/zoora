import { SubtitleLine } from '../lib/Subtitles';

export const FPS = 30;

// Scene durations (frames) — transitions overlap 15f between scenes
export const T = 15;
export const D = {
  logo: 180,
  classes: 420,
  live: 540,
  quiz: 420,
  reach: 420,
  outro: 270,
} as const;

// Absolute start frame of each scene inside the composition
export const START = {
  logo: 0,
  classes: D.logo - T,
  live: D.logo + D.classes - 2 * T,
  quiz: D.logo + D.classes + D.live - 3 * T,
  reach: D.logo + D.classes + D.live + D.quiz - 4 * T,
  outro: D.logo + D.classes + D.live + D.quiz + D.reach - 5 * T,
} as const;

export const INTRO_DURATION = START.outro + D.outro;

/**
 * Narration lines. Burned in as subtitles AND used as the voice-over
 * recording script (see docs/voiceover-intro.md).
 */
export const LINES: SubtitleLine[] = [
  // ── Scene 1: hook
  { from: 20, to: 160, text: 'زورا — کلاس زنده، آزمون و پیام‌رسانی؛ همه در یک پلتفرم.' },
  // ── Scene 2: org dashboard / classes
  { from: 185, to: 330, text: 'با زورا، آموزشگاه شما در چند دقیقه کلاس آنلاین می‌سازد.' },
  { from: 335, to: 480, text: 'کلاس بسازید، دانش‌آموزها را اضافه کنید و برنامهٔ جلسه‌ها را بچینید.' },
  { from: 485, to: 565, text: 'همه‌چیز از یک داشبورد ساده مدیریت می‌شود.' },
  // ── Scene 3: live class
  { from: 585, to: 730, text: 'کلاس‌های زندهٔ کم‌تأخیر، با ویدیو و صدای شفاف.' },
  { from: 735, to: 885, text: 'دست بلند کردن، نظرسنجی زنده و گفت‌وگوی کلاس، درس را تعاملی نگه می‌دارد.' },
  { from: 890, to: 1030, text: 'روی تختهٔ اشتراکی توضیح دهید — همراه اتاق ذخیره می‌شود.' },
  { from: 1035, to: 1090, text: 'و هر جلسه به‌صورت ابری ضبط می‌شود.' },
  // ── Scene 4: quizzes
  { from: 1110, to: 1255, text: 'آزمون‌های آنلاین مطمئن، با زمان‌بندی دقیق و سؤال‌های درهم.' },
  { from: 1260, to: 1405, text: 'بانک سؤال بسازید و در آزمون‌های بعدی دوباره استفاده کنید.' },
  { from: 1410, to: 1495, text: 'نمره‌دهی خودکار انجام می‌شود و کارنامه آماده است.' },
  // ── Scene 5: recordings + notifications + files
  { from: 1515, to: 1650, text: 'هیچ درسی از دست نمی‌رود؛ ضبط جلسه‌ها همیشه در دسترس دانش‌آموزهاست.' },
  { from: 1655, to: 1795, text: 'اطلاعیه‌ها از راه تلگرام، بله، پیامک و اعلان وب واقعاً به دست می‌رسند.' },
  { from: 1800, to: 1895, text: 'فایل‌ها و جزوه‌ها هم همان‌جا، کنار کلاس.' },
  // ── Scene 6: outro
  { from: 1925, to: 2090, text: 'زورا — جایی که تدریس خوب زنده می‌شود. همین امروز رایگان شروع کنید.' },
];
