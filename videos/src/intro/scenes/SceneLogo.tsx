import { AbsoluteFill, useCurrentFrame, useVideoConfig, interpolate } from 'remotion';
import { colors, font } from '../../lib/tokens';
import { Logo } from '../../components/Logo';
import { pop, enter } from '../../lib/anim';

export const Aurora: React.FC = () => {
  const frame = useCurrentFrame();
  const t = frame / 90;
  return (
    <>
      <div
        style={{
          position: 'absolute',
          width: 900,
          height: 900,
          borderRadius: '50%',
          left: 260 + Math.sin(t) * 60,
          top: -320 + Math.cos(t * 0.8) * 40,
          background: 'radial-gradient(circle, rgba(22,163,74,0.35), transparent 65%)',
          filter: 'blur(60px)',
        }}
      />
      <div
        style={{
          position: 'absolute',
          width: 800,
          height: 800,
          borderRadius: '50%',
          right: 140 + Math.cos(t * 0.7) * 50,
          bottom: -300 + Math.sin(t * 0.9) * 50,
          background: 'radial-gradient(circle, rgba(37,99,235,0.22), transparent 65%)',
          filter: 'blur(70px)',
        }}
      />
    </>
  );
};

export const SceneLogo: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const logoIn = pop(frame, fps, 8);
  const nameIn = enter(frame, fps, 25);
  const tagIn = enter(frame, fps, 45);
  const glow = interpolate(Math.sin(frame / 18), [-1, 1], [0.35, 0.7]);

  return (
    <AbsoluteFill style={{ background: colors.zinc950, overflow: 'hidden' }}>
      <Aurora />
      <AbsoluteFill style={{ alignItems: 'center', justifyContent: 'center', gap: 34 }}>
        <div
          style={{
            transform: `scale(${logoIn})`,
            filter: `drop-shadow(0 0 ${40 * glow}px rgba(34,197,94,${glow}))`,
          }}
        >
          <Logo size={168} />
        </div>
        <div
          dir="rtl"
          style={{
            fontFamily: font,
            fontSize: 110,
            fontWeight: 800,
            color: colors.white,
            opacity: nameIn,
            transform: `translateY(${(1 - nameIn) * 30}px)`,
            lineHeight: 1.2,
          }}
        >
          زورا
        </div>
        <div
          dir="rtl"
          style={{
            fontFamily: font,
            fontSize: 34,
            fontWeight: 500,
            color: colors.zinc400,
            opacity: tagIn,
            transform: `translateY(${(1 - tagIn) * 24}px)`,
          }}
        >
          کلاس زنده، آزمون و پیام‌رسانی — همه در یک پلتفرم
        </div>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};
