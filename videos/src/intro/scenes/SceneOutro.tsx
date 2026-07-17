import { AbsoluteFill, useCurrentFrame, useVideoConfig, interpolate } from 'remotion';
import { colors, font } from '../../lib/tokens';
import { Logo } from '../../components/Logo';
import { Aurora } from './SceneLogo';
import { pop, enter } from '../../lib/anim';

export const SceneOutro: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const logoIn = pop(frame, fps, 10);
  const titleIn = enter(frame, fps, 28);
  const ctaIn = pop(frame, fps, 55);
  const urlIn = enter(frame, fps, 80);
  const sheen = interpolate((frame - 70) % 90, [0, 90], [-140, 340]);

  return (
    <AbsoluteFill style={{ background: colors.zinc950, overflow: 'hidden', fontFamily: font }}>
      <Aurora />
      <AbsoluteFill style={{ alignItems: 'center', justifyContent: 'center', gap: 30 }}>
        <div style={{ transform: `scale(${logoIn})` }}>
          <Logo size={132} />
        </div>
        <div
          dir="rtl"
          style={{
            opacity: titleIn,
            transform: `translateY(${(1 - titleIn) * 26}px)`,
            textAlign: 'center',
            display: 'flex',
            flexDirection: 'column',
            gap: 14,
          }}
        >
          <div style={{ fontSize: 62, fontWeight: 800, color: colors.white, lineHeight: 1.4 }}>
            جایی که تدریس خوب زنده می‌شود
          </div>
        </div>
        <div
          dir="rtl"
          style={{
            transform: `scale(${ctaIn})`,
            position: 'relative',
            overflow: 'hidden',
            padding: '18px 52px',
            borderRadius: 999,
            background: `linear-gradient(135deg, ${colors.green500}, ${colors.green700})`,
            color: colors.white,
            fontSize: 30,
            fontWeight: 800,
            boxShadow: '0 18px 50px -12px rgba(34,197,94,0.55)',
          }}
        >
          رایگان شروع کنید
          <div
            style={{
              position: 'absolute',
              top: 0,
              bottom: 0,
              left: sheen,
              width: 60,
              background: 'linear-gradient(105deg, transparent, rgba(255,255,255,0.35), transparent)',
              transform: 'skewX(-14deg)',
            }}
          />
        </div>
        <div style={{ opacity: urlIn, color: colors.zinc500, fontSize: 24, fontWeight: 600, direction: 'ltr' }}>
          zoora.ir
        </div>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};
