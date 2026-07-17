import { interpolate, useCurrentFrame, Easing } from 'remotion';
import { colors } from '../lib/tokens';

export type CursorKey = { frame: number; x: number; y: number; click?: boolean };

/**
 * Animated pointer that glides between keyframes and ripples on clicks.
 * Coordinates are relative to the parent (position: absolute).
 */
export const Cursor: React.FC<{ keys: CursorKey[] }> = ({ keys }) => {
  const frame = useCurrentFrame();
  const frames = keys.map((k) => k.frame);
  const ease = Easing.inOut(Easing.cubic);
  const x = interpolate(frame, frames, keys.map((k) => k.x), {
    easing: ease,
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const y = interpolate(frame, frames, keys.map((k) => k.y), {
    easing: ease,
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const clicks = keys.filter((k) => k.click);
  // Press-down squash near a click
  const press = clicks.reduce((acc, c) => {
    const p = interpolate(frame, [c.frame - 4, c.frame, c.frame + 6], [1, 0.82, 1], {
      extrapolateLeft: 'clamp',
      extrapolateRight: 'clamp',
    });
    return Math.min(acc, p);
  }, 1);

  const appear = interpolate(frame, [frames[0] - 10, frames[0]], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  return (
    <>
      {clicks.map((c, i) => {
        const t = frame - c.frame;
        if (t < 0 || t > 22) return null;
        const r = interpolate(t, [0, 22], [6, 34]);
        const o = interpolate(t, [0, 22], [0.5, 0]);
        return (
          <div
            key={i}
            style={{
              position: 'absolute',
              left: c.x - r,
              top: c.y - r,
              width: r * 2,
              height: r * 2,
              borderRadius: '50%',
              border: `3px solid ${colors.green600}`,
              opacity: o,
            }}
          />
        );
      })}
      <svg
        width={30}
        height={30}
        viewBox="0 0 24 24"
        style={{
          position: 'absolute',
          left: x,
          top: y,
          opacity: appear,
          transform: `scale(${press})`,
          transformOrigin: '4px 3px',
          filter: 'drop-shadow(0 2px 4px rgb(0 0 0 / 0.35))',
        }}
      >
        <path
          d="M4 3l7 17 2.4-6.8L20 10.5z"
          fill={colors.zinc950}
          stroke={colors.white}
          strokeWidth={1.6}
          strokeLinejoin="round"
        />
      </svg>
    </>
  );
};
