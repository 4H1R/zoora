import { interpolate, useCurrentFrame, AbsoluteFill } from 'remotion';
import { colors, font } from './tokens';

export type SubtitleLine = {
  /** Persian narration line — doubles as the voice-over script */
  text: string;
  from: number;
  to: number;
};

export const Subtitles: React.FC<{ lines: SubtitleLine[] }> = ({ lines }) => {
  const frame = useCurrentFrame();
  const line = lines.find((l) => frame >= l.from && frame < l.to);
  if (!line) return null;

  const appear = interpolate(frame, [line.from, line.from + 8], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const exit = interpolate(frame, [line.to - 8, line.to], [1, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });
  const opacity = Math.min(appear, exit);

  return (
    <AbsoluteFill style={{ justifyContent: 'flex-end', alignItems: 'center', pointerEvents: 'none' }}>
      <div
        dir="rtl"
        style={{
          opacity,
          transform: `translateY(${(1 - appear) * 16}px)`,
          marginBottom: 56,
          maxWidth: 1100,
          padding: '14px 32px',
          borderRadius: 16,
          background: 'rgba(9, 9, 11, 0.72)',
          backdropFilter: 'blur(8px)',
          color: colors.white,
          fontFamily: font,
          fontSize: 34,
          fontWeight: 600,
          lineHeight: 1.7,
          textAlign: 'center',
        }}
      >
        {line.text}
      </div>
    </AbsoluteFill>
  );
};
