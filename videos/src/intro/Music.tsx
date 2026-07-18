import { Audio } from '@remotion/media';
import { Sequence, interpolate, staticFile } from 'remotion';
import { INTRO_DURATION } from './script';

/**
 * The generated track is ~30.77s (923 frames); the video is longer.
 *
 * Measured structure (RMS envelope + self-similarity, don't trust the
 * generator's own description): intro 0–4s, drop at ~4.2s, steady body
 * ~5–22s, breakdown dip 22.8–24.1s, finale + resolve 24.8–30.77s.
 *
 * Arrangement: the body loops via beat-aligned jumps (each jump replays
 * near-identical material, hidden by equal-power crossfades), then after the
 * LAST jump the track plays straight through to its natural end — breakdown,
 * finale and final chord untouched, no splice. That requires
 * INTRO_DURATION === 923 + sum(loop lengths); the outro scene is padded in
 * script.ts to make the equation hold.
 */
const AUDIO_LEN = 923; // 30.77s
const XF = 30; // crossfade at each loop jump

// Beat-aligned self-similar jump pairs (audio frames): play up to B, jump back to A
const P342 = { A: 265, B: 607 }; // 8.82s ← 20.25s, 20 beats
const P273 = { A: 315, B: 588 }; // 10.50s ← 19.60s, 16 beats
const JUMPS = [P342, P273, P342, P273];

type Seg = { from: number; dur: number; startFrom: number; fadeIn: number; fadeOut: number };

const SEGMENTS: Seg[] = [];
let video = 0; // video frame where the current pass starts
let audio = 0; // audio frame the current pass starts from
for (const p of JUMPS) {
  const joint = video + (p.B - audio);
  SEGMENTS.push({ from: video, dur: p.B - audio + XF, startFrom: audio, fadeIn: video ? XF : 0, fadeOut: XF });
  video = joint;
  audio = p.A;
}
// Final pass: from the last jump target straight to the end of the track
SEGMENTS.push({ from: video, dur: AUDIO_LEN - audio, startFrom: audio, fadeIn: XF, fadeOut: 0 });

const TOTAL = video + AUDIO_LEN - audio;
if (TOTAL !== INTRO_DURATION) {
  throw new Error(
    `Music arrangement is ${TOTAL} frames but the video is ${INTRO_DURATION}. ` +
      `Adjust scene durations in script.ts or the JUMPS list so they match.`,
  );
}

export const IntroMusic: React.FC = () => (
  <>
    {SEGMENTS.map((s, i) => (
      <Sequence key={i} from={s.from} durationInFrames={s.dur}>
        <Audio
          src={staticFile('intro-music.mp3')}
          trimBefore={s.startFrom}
          trimAfter={s.startFrom + s.dur}
          volume={(f) => {
            // Equal-power crossfades — linear fades dip ~3dB at the joint center
            const rIn = s.fadeIn
              ? interpolate(f, [0, s.fadeIn], [0, 1], { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' })
              : 1;
            const rOut = s.fadeOut
              ? interpolate(f, [s.dur - s.fadeOut, s.dur], [1, 0], { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' })
              : 1;
            return Math.sin((rIn * Math.PI) / 2) * Math.sin((rOut * Math.PI) / 2);
          }}
        />
      </Sequence>
    ))}
  </>
);
