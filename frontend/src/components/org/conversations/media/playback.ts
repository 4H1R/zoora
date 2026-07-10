import { create } from "zustand"

/** Telegram-style speed cycle: tap the pill to advance, wraps back to 1×. */
export const PLAYBACK_RATES = [1, 1.5, 2] as const

interface MediaSettingsState {
  /** Global playback speed shared by every audio/video player this session. */
  rate: number
  cycleRate: () => void
}

export const useMediaSettings = create<MediaSettingsState>((set) => ({
  rate: 1,
  cycleRate: () =>
    set((s) => {
      const index = PLAYBACK_RATES.indexOf(s.rate as (typeof PLAYBACK_RATES)[number])
      return { rate: PLAYBACK_RATES[(index + 1) % PLAYBACK_RATES.length] }
    }),
}))

/*
 * Exclusive playback (Telegram behavior): starting any audio/video pauses
 * whatever else was playing. Players register their pause callback when they
 * start and release it when they stop.
 */
let activePause: (() => void) | null = null

export function claimPlayback(pause: () => void): void {
  if (activePause && activePause !== pause) activePause()
  activePause = pause
}

export function releasePlayback(pause: () => void): void {
  if (activePause === pause) activePause = null
}

/** `m:ss` (or `h:mm:ss` past an hour) for player time labels. */
export function formatMediaTime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) return "0:00"
  const total = Math.floor(seconds)
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  const mm = h > 0 ? String(m).padStart(2, "0") : String(m)
  return `${h > 0 ? `${h}:` : ""}${mm}:${String(s).padStart(2, "0")}`
}
