import { AbsoluteFill, useCurrentFrame, useVideoConfig, interpolate } from 'remotion';
import { colors, font, radius, shadow, toFa } from '../../lib/tokens';
import { Icon, IconName } from '../../components/Icon';
import { Avatar, Chip } from '../../components/mock';
import { enter, pop } from '../../lib/anim';

const PARTICIPANTS = [
  { name: 'استاد محمدی', hue: 0, speaking: true },
  { name: 'سارا', hue: 3 },
  { name: 'امیر', hue: 1 },
  { name: 'نگار', hue: 4 },
];

const Tile: React.FC<{ name: string; hue: number; speaking?: boolean; frame: number }> = ({
  name,
  hue,
  speaking,
  frame,
}) => {
  const glow = speaking ? interpolate(Math.sin(frame / 10), [-1, 1], [0.25, 0.9]) : 0;
  return (
    <div
      style={{
        position: 'relative',
        borderRadius: radius.xxl,
        background: `linear-gradient(150deg, ${colors.zinc800}, ${colors.zinc900})`,
        border: speaking ? `2.5px solid rgba(34,197,94,${glow})` : `1px solid ${colors.zinc700}`,
        boxShadow: speaking ? `0 0 ${28 * glow}px rgba(34,197,94,${glow * 0.5})` : 'none',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        overflow: 'hidden',
      }}
    >
      <Avatar name={name} size={110} hue={hue} />
      <div
        dir="rtl"
        style={{
          position: 'absolute',
          bottom: 14,
          insetInlineStart: 16,
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          padding: '6px 14px',
          borderRadius: 999,
          background: 'rgba(9,9,11,0.6)',
          backdropFilter: 'blur(6px)',
          color: colors.white,
          fontFamily: font,
          fontSize: 15,
          fontWeight: 600,
        }}
      >
        <Icon name="mic" size={14} color={speaking ? colors.green400 : colors.zinc400} />
        {name}
      </div>
    </div>
  );
};

const ControlButton: React.FC<{ icon: IconName; danger?: boolean; active?: boolean }> = ({
  icon,
  danger,
  active,
}) => (
  <div
    style={{
      width: 56,
      height: 56,
      borderRadius: '50%',
      background: danger ? colors.live : active ? colors.green600 : colors.zinc700,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      boxShadow: shadow.lg,
    }}
  >
    <Icon name={icon} size={24} color={colors.white} />
  </div>
);

export const SceneLive: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const roomIn = enter(frame, fps, 5);
  const seconds = 12 * 60 + 3 + Math.floor(frame / fps);
  const timer = `${String(Math.floor(seconds / 60)).padStart(2, '0')}:${String(seconds % 60).padStart(2, '0')}`;
  const dot = interpolate(Math.sin(frame / 8), [-1, 1], [0.35, 1]);

  const handIn = pop(frame, fps, 180);
  const showHand = frame >= 180 && frame < 320;
  const pollIn = pop(frame, fps, 270);
  const showPoll = frame >= 270;
  const bar1 = interpolate(frame, [290, 340], [0, 72], { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' });
  const bar2 = interpolate(frame, [290, 340], [0, 28], { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' });

  const boardIn = enter(frame, fps, 390);
  const showBoard = frame >= 390;
  const draw = interpolate(frame, [405, 480], [0, 1], { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' });
  const draw2 = interpolate(frame, [460, 510], [0, 1], { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' });

  return (
    <AbsoluteFill style={{ background: colors.zinc950, fontFamily: font, overflow: 'hidden' }}>
      <div style={{ opacity: roomIn, height: '100%', display: 'flex', flexDirection: 'column' }}>
        {/* Top bar */}
        <div
          dir="rtl"
          style={{
            height: 84,
            flexShrink: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '0 48px',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 18 }}>
            <div style={{ color: colors.white, fontSize: 24, fontWeight: 700 }}>
              فیزیک پیشرفته — جلسه {toFa(12)}
            </div>
            <Chip color={colors.zinc300} bg={colors.zinc800} icon="users">
              {toFa(24)} نفر
            </Chip>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                padding: '7px 16px',
                borderRadius: 999,
                background: 'rgba(220,38,38,0.16)',
                color: '#f87171',
                fontSize: 15,
                fontWeight: 700,
              }}
            >
              <div style={{ width: 9, height: 9, borderRadius: '50%', background: colors.live, opacity: dot }} />
              زنده
            </div>
            <div style={{ color: colors.zinc400, fontSize: 17, fontWeight: 600 }}>{toFa(timer)}</div>
          </div>
        </div>

        {/* Video grid */}
        <div
          style={{
            flex: 1,
            display: 'grid',
            gridTemplateColumns: '1fr 1fr',
            gridTemplateRows: '1fr 1fr',
            gap: 20,
            padding: '0 48px',
          }}
        >
          {PARTICIPANTS.map((p, i) => (
            <div key={p.name} style={{ opacity: enter(frame, fps, 20 + i * 12), display: 'grid' }}>
              <Tile {...p} frame={frame} />
            </div>
          ))}
        </div>

        {/* Control bar */}
        <div
          style={{
            height: 108,
            flexShrink: 0,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 18,
          }}
        >
          <ControlButton icon="mic" active />
          <ControlButton icon="video" active />
          <ControlButton icon="monitor" />
          <ControlButton icon="chat" />
          <ControlButton icon="phone" danger />
        </div>
      </div>

      {/* Raise-hand toast */}
      {showHand ? (
        <div
          dir="rtl"
          style={{
            position: 'absolute',
            top: 110,
            insetInlineStart: 48,
            transform: `translateY(${(1 - handIn) * -30}px)`,
            opacity: Math.min(handIn, interpolate(frame, [300, 318], [1, 0], { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' })),
            display: 'flex',
            alignItems: 'center',
            gap: 12,
            padding: '14px 22px',
            borderRadius: radius.xl,
            background: colors.zinc800,
            border: `1px solid ${colors.zinc700}`,
            boxShadow: shadow.xl,
            color: colors.white,
            fontSize: 18,
            fontWeight: 600,
          }}
        >
          <Icon name="hand" size={24} color={colors.warning} strokeWidth={2} />
          سارا دست بلند کرد
        </div>
      ) : null}

      {/* Live poll */}
      {showPoll ? (
        <div
          dir="rtl"
          style={{
            position: 'absolute',
            bottom: 150,
            insetInlineEnd: 48,
            width: 400,
            transform: `scale(${pollIn})`,
            transformOrigin: 'bottom right',
            padding: 24,
            borderRadius: radius.xxl,
            background: colors.zinc800,
            border: `1px solid ${colors.zinc700}`,
            boxShadow: shadow.xl,
            display: 'flex',
            flexDirection: 'column',
            gap: 14,
          }}
        >
          <div style={{ color: colors.white, fontSize: 18, fontWeight: 700 }}>
            برای یک کوییز کوتاه آماده‌اید؟
          </div>
          {[
            { label: 'آماده‌ام', pct: bar1, main: true },
            { label: 'یک مثال دیگر', pct: bar2 },
          ].map((o) => (
            <div key={o.label} style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', color: colors.zinc300, fontSize: 15 }}>
                <span>{o.label}</span>
                <span style={{ fontWeight: 700, color: o.main ? colors.green400 : colors.zinc400 }}>
                  ٪{toFa(Math.round(o.pct))}
                </span>
              </div>
              <div style={{ height: 10, borderRadius: 999, background: colors.zinc700, overflow: 'hidden' }}>
                <div
                  style={{
                    width: `${o.pct}%`,
                    height: '100%',
                    borderRadius: 999,
                    background: o.main ? colors.green500 : colors.zinc500,
                  }}
                />
              </div>
            </div>
          ))}
        </div>
      ) : null}

      {/* Whiteboard panel */}
      {showBoard ? (
        <div
          dir="rtl"
          style={{
            position: 'absolute',
            top: 84,
            bottom: 108,
            insetInlineStart: 48,
            width: 700,
            transform: `translateX(${(1 - boardIn) * -80}px)`,
            opacity: boardIn,
            borderRadius: radius.xxl,
            background: colors.white,
            boxShadow: shadow.xl,
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <div
            style={{
              height: 54,
              display: 'flex',
              alignItems: 'center',
              gap: 10,
              paddingInline: 20,
              borderBottom: `1px solid ${colors.zinc200}`,
              color: colors.zinc800,
              fontSize: 17,
              fontWeight: 700,
            }}
          >
            <Icon name="pen" size={17} color={colors.green600} />
            تخته
          </div>
          <svg viewBox="0 0 700 560" style={{ flex: 1 }}>
            <path
              d="M60 420 C 160 420, 180 180, 300 180 S 460 420, 560 420"
              fill="none"
              stroke={colors.green600}
              strokeWidth={6}
              strokeLinecap="round"
              pathLength={1}
              strokeDasharray={1}
              strokeDashoffset={1 - draw}
            />
            <circle
              cx={300}
              cy={180}
              r={34}
              fill="none"
              stroke={colors.info}
              strokeWidth={4}
              pathLength={1}
              strokeDasharray={1}
              strokeDashoffset={1 - draw2}
            />
            <text
              x={340}
              y={110}
              fontFamily={font}
              fontSize={26}
              fontWeight={700}
              fill={colors.zinc700}
              opacity={draw2}
            >
              نقطهٔ اوج
            </text>
          </svg>
        </div>
      ) : null}
    </AbsoluteFill>
  );
};
