import type { RefObject } from "react"

import { useEffect, useState } from "react"

import { claimPlayback, releasePlayback, useMediaSettings } from "./playback"

export interface MediaPlayback {
  playing: boolean
  currentTime: number
  /** 0 until the element reports a finite duration. */
  duration: number
  toggle: () => void
  /** Seek to a 0..1 fraction of the duration. */
  seekTo: (fraction: number) => void
}

/**
 * Playback state + controls for an `<audio>`/`<video>` element. Applies the
 * global Telegram-style speed setting (live, even mid-play), and claims the
 * exclusive-playback slot so starting one player pauses every other.
 *
 * MediaRecorder-produced webm/ogg often reports `duration: Infinity` until the
 * element has sought past the end (a long-standing Chromium quirk). Two things
 * counter it: callers that have decoded the clip (voice notes) pass the exact
 * length as `knownDuration`, and as a fallback the hook silently seeks far
 * ahead and back once to force the real duration to materialize — voice notes
 * are small, so the extra range fetch is negligible.
 */
export function useMediaPlayback(
  ref: RefObject<HTMLMediaElement | null>,
  src?: string,
  knownDuration?: number
): MediaPlayback {
  const [playing, setPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [nativeDuration, setNativeDuration] = useState(0)
  const rate = useMediaSettings((s) => s.rate)

  // Prefer the element's own finite duration; fall back to a caller-supplied
  // decoded length while the element still reports Infinity.
  const duration = nativeDuration > 0 ? nativeDuration : knownDuration && knownDuration > 0 ? knownDuration : 0

  // Live speed change while playing.
  useEffect(() => {
    const el = ref.current
    if (el) el.playbackRate = rate
  }, [rate, ref])

  useEffect(() => {
    const el = ref.current
    if (!el) return

    const pause = () => el.pause()

    // True only during the seek probe below, so the huge currentTime it
    // produces never leaks into the progress bar.
    let probing = false

    const syncDuration = () => {
      if (Number.isFinite(el.duration) && el.duration > 0) setNativeDuration(el.duration)
    }
    const probeInfiniteDuration = () => {
      if (Number.isFinite(el.duration)) {
        syncDuration()
        return
      }
      probing = true
      const restore = () => {
        el.removeEventListener("timeupdate", restore)
        probing = false
        el.currentTime = 0
        setCurrentTime(0)
        syncDuration()
      }
      el.addEventListener("timeupdate", restore)
      el.currentTime = 1e7
    }

    const onTime = () => {
      // Ignore the transient jump from the infinite-duration probe. Otherwise
      // report progress even while el.duration is still Infinity — the played
      // fraction is computed against the decoded knownDuration.
      if (probing) return
      setCurrentTime(el.currentTime)
    }
    const onPlay = () => {
      claimPlayback(pause)
      el.playbackRate = useMediaSettings.getState().rate
      setPlaying(true)
    }
    const onStop = () => {
      releasePlayback(pause)
      setPlaying(false)
    }

    el.addEventListener("loadedmetadata", probeInfiniteDuration)
    el.addEventListener("durationchange", syncDuration)
    el.addEventListener("timeupdate", onTime)
    el.addEventListener("play", onPlay)
    el.addEventListener("pause", onStop)
    el.addEventListener("ended", onStop)
    // Metadata may already be in by the time the effect runs.
    if (el.readyState >= 1) probeInfiniteDuration()

    return () => {
      el.removeEventListener("loadedmetadata", probeInfiniteDuration)
      el.removeEventListener("durationchange", syncDuration)
      el.removeEventListener("timeupdate", onTime)
      el.removeEventListener("play", onPlay)
      el.removeEventListener("pause", onStop)
      el.removeEventListener("ended", onStop)
      releasePlayback(pause)
    }
    // Re-bind when the media element (re)mounts. Conditionally-rendered players
    // (e.g. video, which mounts only once its src resolves) attach their element
    // AFTER the initial mount, so keying on `src` re-runs this to catch it — and
    // also re-subscribes on a source swap.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [src])

  function toggle() {
    const el = ref.current
    if (!el) return
    if (el.paused) void el.play()
    else el.pause()
  }

  function seekTo(fraction: number) {
    const el = ref.current
    if (!el || duration <= 0) return
    el.currentTime = Math.min(1, Math.max(0, fraction)) * duration
    setCurrentTime(el.currentTime)
  }

  return { playing, currentTime, duration, toggle, seekTo }
}
