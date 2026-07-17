import { AbsoluteFill, useCurrentFrame, useVideoConfig } from 'remotion';
import { colors, font, radius, shadow } from '../../lib/tokens';
import { Icon } from '../../components/Icon';
import { enter, pop, fadeUp } from '../../lib/anim';

const RECORDINGS = [
  { title: 'جلسه ۱۲ — فیزیک پیشرفته', duration: '۰۱:۱۲:۳۴' },
  { title: 'جلسه ۱۱ — مرور فصل ۳', duration: '۰۰:۵۸:۱۰' },
  { title: 'جلسه ۱۰ — حل تمرین', duration: '۰۱:۰۴:۲۲' },
];

const CHANNELS = [
  { name: 'تلگرام', color: '#2563eb' },
  { name: 'بله', color: '#16a34a' },
  { name: 'پیامک', color: '#71717a' },
  { name: 'اعلان وب', color: '#d97706' },
];

const FILES = ['جزوهٔ فصل ۳.pdf', 'اسلایدهای درس.pptx', 'نمونه‌سؤال.pdf'];

export const SceneReach: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  return (
    <AbsoluteFill
      style={{
        background: `linear-gradient(160deg, ${colors.zinc100}, #e9f5ee)`,
        fontFamily: font,
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <div dir="rtl" style={{ display: 'flex', gap: 44, alignItems: 'flex-start' }}>
        {/* Recordings panel */}
        <div
          style={{
            width: 660,
            borderRadius: radius.xxl,
            background: colors.white,
            border: `1px solid ${colors.zinc200}`,
            boxShadow: shadow.xl,
            padding: 28,
            display: 'flex',
            flexDirection: 'column',
            gap: 16,
            opacity: enter(frame, fps, 10),
            transform: `translateY(${(1 - enter(frame, fps, 10)) * 40}px)`,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, fontSize: 22, fontWeight: 800, color: colors.zinc900 }}>
            <Icon name="video" size={22} color={colors.green600} />
            ضبط جلسه‌ها
          </div>
          {RECORDINGS.map((r, i) => (
            <div key={r.title} style={fadeUp(frame, fps, 50 + i * 16, 24)}>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 16,
                  padding: '14px 18px',
                  borderRadius: radius.lg,
                  border: `1px solid ${colors.zinc200}`,
                  background: i === 0 ? colors.green50 : colors.white,
                }}
              >
                <div
                  style={{
                    width: 42,
                    height: 42,
                    borderRadius: '50%',
                    background: colors.green600,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexShrink: 0,
                  }}
                >
                  <Icon name="play" size={18} color={colors.white} />
                </div>
                <div style={{ flex: 1, fontSize: 17, fontWeight: 700, color: colors.zinc900 }}>{r.title}</div>
                <div style={{ fontSize: 15, fontWeight: 600, color: colors.zinc500, direction: 'ltr' }}>
                  {r.duration}
                </div>
              </div>
            </div>
          ))}
          {/* File chips */}
          <div style={{ display: 'flex', gap: 12, marginTop: 6, flexWrap: 'wrap' }}>
            {FILES.map((f, i) => (
              <div key={f} style={fadeUp(frame, fps, 280 + i * 12, 20)}>
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 8,
                    padding: '9px 16px',
                    borderRadius: 999,
                    background: colors.zinc100,
                    border: `1px solid ${colors.zinc200}`,
                    fontSize: 14.5,
                    fontWeight: 600,
                    color: colors.zinc700,
                  }}
                >
                  <Icon name="file" size={15} color={colors.zinc500} />
                  {f}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Notification delivery stack */}
        <div style={{ width: 480, display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 10,
              fontSize: 22,
              fontWeight: 800,
              color: colors.zinc900,
              opacity: enter(frame, fps, 90),
            }}
          >
            <Icon name="bell" size={22} color={colors.green600} />
            اطلاعیهٔ «جلسهٔ فردا ساعت ۱۶»
          </div>
          {CHANNELS.map((c, i) => {
            const p = pop(frame, fps, 120 + i * 26);
            const delivered = frame >= 120 + i * 26 + 30;
            const checkIn = pop(frame, fps, 120 + i * 26 + 30);
            return (
              <div
                key={c.name}
                style={{
                  opacity: Math.min(1, p),
                  transform: `translateY(${(1 - p) * -30}px)`,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 14,
                  padding: '16px 20px',
                  borderRadius: radius.xl,
                  background: colors.white,
                  border: `1px solid ${colors.zinc200}`,
                  boxShadow: shadow.lg,
                }}
              >
                <div
                  style={{
                    width: 44,
                    height: 44,
                    borderRadius: radius.lg,
                    background: `${c.color}1d`,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexShrink: 0,
                  }}
                >
                  <Icon name={i === 3 ? 'bell' : 'chat'} size={20} color={c.color} />
                </div>
                <div style={{ flex: 1, fontSize: 18, fontWeight: 700, color: colors.zinc900 }}>{c.name}</div>
                {delivered ? (
                  <div
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 6,
                      padding: '6px 14px',
                      borderRadius: 999,
                      background: colors.green50,
                      color: colors.green700,
                      fontSize: 14,
                      fontWeight: 700,
                      transform: `scale(${checkIn})`,
                    }}
                  >
                    <Icon name="check" size={13} color={colors.green700} strokeWidth={2.6} />
                    تحویل شد
                  </div>
                ) : null}
              </div>
            );
          })}
        </div>
      </div>
    </AbsoluteFill>
  );
};
