/**
 * Waveform helpers for voice-message bubbles: real peaks decoded from the audio
 * bytes when possible, with a deterministic synthetic fallback (CORS-blocked
 * fetches, codecs the browser can't decode) so the bubble always looks like a
 * voice note rather than degrading to a bare progress line.
 */

/** Bar count for voice waveforms — sized to read well at bubble width. */
export const VOICE_BUCKETS = 36

/**
 * In-app voice recordings are named `voice-<ts>.<ext>`; the prefix is how a
 * received audio attachment is recognized as a voice note (waveform layout)
 * versus music (title + seek line).
 */
const VOICE_FILE_PREFIX = "voice-"

export function isVoiceName(name: string): boolean {
  return name.toLowerCase().startsWith(VOICE_FILE_PREFIX)
}

export function voiceFileName(ext: string): string {
  return `${VOICE_FILE_PREFIX}${Date.now()}.${ext}`
}

export interface DecodedAudio {
  /** Normalized RMS peaks (0.12..1) for the waveform. */
  peaks: number[]
  /** Exact clip length in seconds, straight from the decoded buffer. */
  duration: number
}

/**
 * Decode audio bytes into `buckets` normalized RMS peaks (each 0.12..1 so even
 * silence renders a visible bar) plus the exact duration. Throws when the
 * browser can't decode the container/codec — callers fall back to
 * `syntheticPeaks`.
 *
 * The decoded duration is the reliable source for a voice note's length:
 * MediaRecorder webm/ogg reports `Infinity` on the `<audio>` element until it
 * is played through, whereas the fully-decoded buffer always knows its length.
 */
export async function extractPeaks(blob: Blob, buckets = VOICE_BUCKETS): Promise<DecodedAudio> {
  const bytes = await blob.arrayBuffer()
  const ctx = new AudioContext()
  try {
    const decoded = await ctx.decodeAudioData(bytes)
    const data = decoded.getChannelData(0)
    const bucketSize = Math.max(1, Math.floor(data.length / buckets))
    const peaks: number[] = []
    for (let i = 0; i < buckets; i++) {
      const start = i * bucketSize
      const end = Math.min(start + bucketSize, data.length)
      // Sample a stride within the bucket — full-resolution RMS is overkill
      // for a 36-bar sparkline.
      const stride = Math.max(1, Math.floor((end - start) / 64))
      let sum = 0
      let n = 0
      for (let j = start; j < end; j += stride) {
        sum += data[j] * data[j]
        n++
      }
      peaks.push(Math.sqrt(sum / Math.max(1, n)))
    }
    const max = Math.max(...peaks, 0.001)
    return { peaks: peaks.map((p) => Math.max(0.12, p / max)), duration: decoded.duration }
  } finally {
    void ctx.close()
  }
}

/**
 * Deterministic pseudo-waveform seeded by an id — a stable, plausible shape
 * for placeholders and decode failures. FNV-1a over the seed drives per-bar
 * noise, shaped by a slow sine so it reads as speech-like envelopes.
 */
export function syntheticPeaks(seed: string, buckets = VOICE_BUCKETS): number[] {
  let h = 2166136261
  const peaks: number[] = []
  for (let i = 0; i < buckets; i++) {
    h ^= seed.charCodeAt(i % Math.max(1, seed.length)) + i
    h = Math.imul(h, 16777619)
    const noise = ((h >>> 0) % 1000) / 1000
    const envelope = 0.55 + 0.45 * Math.sin(i / 2.8 + (h % 7))
    peaks.push(Math.max(0.12, Math.min(1, (0.3 + 0.7 * noise) * envelope)))
  }
  return peaks
}
