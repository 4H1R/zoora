import { SubtitleLine } from '../lib/Subtitles';

export const FPS = 30;

// Scene durations (frames) — transitions overlap 15f between scenes
export const T = 15;
export const D = {
  // 135 so the crossfade into the showcase starts at frame 120 — the music's 4s drop
  logo: 135,
  classes: 420,
  live: 540,
  quiz: 420,
  reach: 420,
  // 293 (not 270): pads the video to 2153 frames so the music arrangement
  // ends on the track's natural final chord — see intro/Music.tsx
  outro: 293,
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
  { from: 20, to: 115, text: 'زورا — کلاس زنده، آزمون و پیام‌رسانی؛ همه در یک پلتفرم.' },
  // ── Scene 2: org dashboard / classes
  { from: 140, to: 285, text: 'با زورا، آموزشگاه شما در چند دقیقه کلاس آنلاین می‌سازد.' },
  { from: 290, to: 435, text: 'کلاس بسازید، دانش‌آموزها را اضافه کنید و برنامهٔ جلسه‌ها را بچینید.' },
  { from: 440, to: 520, text: 'همه‌چیز از یک داشبورد ساده مدیریت می‌شود.' },
  // ── Scene 3: live class
  { from: 540, to: 685, text: 'کلاس‌های زندهٔ کم‌تأخیر، با ویدیو و صدای شفاف.' },
  { from: 690, to: 840, text: 'دست بلند کردن، نظرسنجی زنده و گفت‌وگوی کلاس، درس را تعاملی نگه می‌دارد.' },
  { from: 845, to: 985, text: 'روی تختهٔ اشتراکی توضیح دهید — همراه اتاق ذخیره می‌شود.' },
  { from: 990, to: 1045, text: 'و هر جلسه به‌صورت ابری ضبط می‌شود.' },
  // ── Scene 4: quizzes
  { from: 1065, to: 1210, text: 'آزمون‌های آنلاین مطمئن، با زمان‌بندی دقیق و سؤال‌های درهم.' },
  { from: 1215, to: 1360, text: 'بانک سؤال بسازید و در آزمون‌های بعدی دوباره استفاده کنید.' },
  { from: 1365, to: 1450, text: 'نمره‌دهی خودکار انجام می‌شود و کارنامه آماده است.' },
  // ── Scene 5: recordings + notifications + files
  { from: 1470, to: 1605, text: 'هیچ درسی از دست نمی‌رود؛ ضبط جلسه‌ها همیشه در دسترس دانش‌آموزهاست.' },
  { from: 1610, to: 1750, text: 'اطلاعیه‌ها از راه تلگرام، بله، پیامک و اعلان وب واقعاً به دست می‌رسند.' },
  { from: 1755, to: 1850, text: 'فایل‌ها و جزوه‌ها هم همان‌جا، کنار کلاس.' },
  // ── Scene 6: outro
  { from: 1880, to: 2045, text: 'زورا — جایی که تدریس خوب زنده می‌شود. همین امروز رایگان شروع کنید.' },
];
