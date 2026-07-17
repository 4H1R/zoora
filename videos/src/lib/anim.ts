import { interpolate, spring, Easing } from 'remotion';

// Smooth entrance without overshoot
export const enter = (frame: number, fps: number, delay = 0): number =>
  spring({ frame: frame - delay, fps, config: { damping: 200 } });

// Springy pop with slight overshoot (badges, toasts, modals)
export const pop = (frame: number, fps: number, delay = 0): number =>
  spring({ frame: frame - delay, fps, config: { damping: 14, stiffness: 160, mass: 0.8 } });

// Fade + rise style, driven by `enter`
export const fadeUp = (
  frame: number,
  fps: number,
  delay = 0,
  distance = 30,
): React.CSSProperties => {
  const p = enter(frame, fps, delay);
  return {
    opacity: p,
    transform: `translateY(${(1 - p) * distance}px)`,
  };
};

export const fadeOnly = (frame: number, delay = 0, duration = 12): React.CSSProperties => ({
  opacity: interpolate(frame, [delay, delay + duration], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  }),
});

export const easeInOut = Easing.inOut(Easing.cubic);
