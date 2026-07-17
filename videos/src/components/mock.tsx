import { colors, font, radius, shadow, toFa } from '../lib/tokens';
import { Icon, IconName } from './Icon';

/** Browser chrome around a mock app screen (RTL). */
export const BrowserFrame: React.FC<{
  url?: string;
  width?: number;
  height?: number;
  children: React.ReactNode;
}> = ({ url = 'app.zoora.ir', width = 1560, height = 880, children }) => (
  <div
    style={{
      width,
      height,
      borderRadius: radius.xxl,
      background: colors.white,
      boxShadow: shadow.window,
      border: `1px solid ${colors.zinc200}`,
      overflow: 'hidden',
      display: 'flex',
      flexDirection: 'column',
    }}
  >
    <div
      style={{
        height: 52,
        flexShrink: 0,
        display: 'flex',
        alignItems: 'center',
        gap: 16,
        padding: '0 20px',
        background: colors.zinc100,
        borderBottom: `1px solid ${colors.zinc200}`,
      }}
    >
      <div style={{ display: 'flex', gap: 8 }}>
        {['#f87171', '#fbbf24', '#34d399'].map((c) => (
          <div key={c} style={{ width: 13, height: 13, borderRadius: '50%', background: c }} />
        ))}
      </div>
      <div
        style={{
          flex: 1,
          maxWidth: 460,
          margin: '0 auto',
          height: 32,
          borderRadius: radius.lg,
          background: colors.white,
          border: `1px solid ${colors.zinc200}`,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontFamily: font,
          fontSize: 15,
          color: colors.zinc500,
          direction: 'ltr',
          gap: 7,
        }}
      >
        <Icon name="lock" size={13} color={colors.zinc400} />
        {url}
      </div>
      <div style={{ width: 60 }} />
    </div>
    <div style={{ flex: 1, position: 'relative', overflow: 'hidden' }}>{children}</div>
  </div>
);

export type NavItem = { icon: IconName; label: string; active?: boolean };

/** RTL app sidebar (sits on the right side). */
export const SidebarMock: React.FC<{ items: NavItem[]; title?: string }> = ({
  items,
  title = 'آموزشگاه نمونه',
}) => (
  <div
    dir="rtl"
    style={{
      width: 240,
      height: '100%',
      flexShrink: 0,
      background: colors.zinc50,
      borderLeft: `1px solid ${colors.zinc200}`,
      padding: '18px 12px',
      fontFamily: font,
      display: 'flex',
      flexDirection: 'column',
      gap: 4,
    }}
  >
    <div style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '4px 10px 18px' }}>
      <svg width={30} height={30} viewBox="0 0 48 48">
        <rect width="48" height="48" rx="13" fill={colors.green600} />
        <path
          d="M14 16 H34 L14 32 H34"
          stroke="#fff"
          strokeWidth="5.4"
          strokeLinecap="round"
          strokeLinejoin="round"
          fill="none"
        />
      </svg>
      <span style={{ fontSize: 16, fontWeight: 700, color: colors.zinc900 }}>{title}</span>
    </div>
    {items.map((item) => (
      <div
        key={item.label}
        style={{
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          padding: '9px 12px',
          borderRadius: radius.md,
          fontSize: 15,
          fontWeight: item.active ? 600 : 500,
          color: item.active ? colors.green700 : colors.zinc600,
          background: item.active ? colors.green50 : 'transparent',
        }}
      >
        <Icon name={item.icon} size={17} color={item.active ? colors.green600 : colors.zinc400} />
        {item.label}
      </div>
    ))}
  </div>
);

export const ButtonMock: React.FC<{
  children: React.ReactNode;
  variant?: 'primary' | 'ghost';
  icon?: IconName;
}> = ({ children, variant = 'primary', icon }) => (
  <div
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      gap: 8,
      padding: '10px 20px',
      borderRadius: radius.lg,
      fontFamily: font,
      fontSize: 15,
      fontWeight: 600,
      background: variant === 'primary' ? colors.green600 : colors.white,
      color: variant === 'primary' ? colors.white : colors.zinc700,
      border: variant === 'primary' ? 'none' : `1px solid ${colors.zinc200}`,
      boxShadow: shadow.md,
    }}
  >
    {icon ? <Icon name={icon} size={16} color={variant === 'primary' ? colors.white : colors.zinc500} /> : null}
    {children}
  </div>
);

const AVATAR_HUES = ['#16a34a', '#2563eb', '#d97706', '#db2777', '#7c3aed'];

export const Avatar: React.FC<{ name: string; size?: number; hue?: number }> = ({
  name,
  size = 40,
  hue = 0,
}) => (
  <div
    style={{
      width: size,
      height: size,
      borderRadius: '50%',
      background: AVATAR_HUES[hue % AVATAR_HUES.length],
      color: colors.white,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      fontFamily: font,
      fontSize: size * 0.4,
      fontWeight: 700,
      flexShrink: 0,
    }}
  >
    {name.slice(0, 1)}
  </div>
);

export const Chip: React.FC<{
  children: React.ReactNode;
  color?: string;
  bg?: string;
  icon?: IconName;
}> = ({ children, color = colors.green700, bg = colors.green50, icon }) => (
  <div
    dir="rtl"
    style={{
      display: 'inline-flex',
      alignItems: 'center',
      gap: 7,
      padding: '7px 16px',
      borderRadius: 999,
      background: bg,
      color,
      fontFamily: font,
      fontSize: 15,
      fontWeight: 600,
    }}
  >
    {icon ? <Icon name={icon} size={15} color={color} /> : null}
    {children}
  </div>
);

/** shadcn-style success toast. */
export const ToastMock: React.FC<{ text: string }> = ({ text }) => (
  <div
    dir="rtl"
    style={{
      display: 'flex',
      alignItems: 'center',
      gap: 10,
      padding: '14px 18px',
      borderRadius: radius.lg,
      background: '#f2fbf5',
      border: `1px solid rgba(134, 239, 172, 0.6)`,
      borderRight: `3px solid ${colors.green600}`,
      boxShadow: shadow.lg,
      fontFamily: font,
      fontSize: 15,
      fontWeight: 600,
      color: colors.green900,
    }}
  >
    <div
      style={{
        width: 22,
        height: 22,
        borderRadius: '50%',
        background: colors.green600,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        flexShrink: 0,
      }}
    >
      <Icon name="check" size={13} color={colors.white} strokeWidth={2.6} />
    </div>
    {text}
  </div>
);

/** Class card matching the org dashboard look. */
export const ClassCard: React.FC<{
  title: string;
  students: number;
  next: string;
  hue?: number;
}> = ({ title, students, next, hue = 0 }) => (
  <div
    dir="rtl"
    style={{
      width: 330,
      borderRadius: radius.xl,
      background: colors.white,
      border: `1px solid ${colors.zinc200}`,
      boxShadow: shadow.md,
      overflow: 'hidden',
      fontFamily: font,
    }}
  >
    <div
      style={{
        height: 76,
        background: `linear-gradient(135deg, ${AVATAR_HUES[hue % AVATAR_HUES.length]}22, ${
          AVATAR_HUES[hue % AVATAR_HUES.length]
        }55)`,
        display: 'flex',
        alignItems: 'center',
        paddingInline: 18,
      }}
    >
      <div
        style={{
          width: 44,
          height: 44,
          borderRadius: radius.lg,
          background: colors.white,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          boxShadow: shadow.md,
        }}
      >
        <Icon name="book" size={22} color={AVATAR_HUES[hue % AVATAR_HUES.length]} />
      </div>
    </div>
    <div style={{ padding: 18, display: 'flex', flexDirection: 'column', gap: 10 }}>
      <div style={{ fontSize: 18, fontWeight: 700, color: colors.zinc900 }}>{title}</div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, color: colors.zinc500, fontSize: 14 }}>
        <Icon name="users" size={15} color={colors.zinc400} />
        {toFa(students)} دانش‌آموز
      </div>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, color: colors.zinc500, fontSize: 14 }}>
        <Icon name="clock" size={15} color={colors.zinc400} />
        جلسه بعدی: {next}
      </div>
    </div>
  </div>
);
