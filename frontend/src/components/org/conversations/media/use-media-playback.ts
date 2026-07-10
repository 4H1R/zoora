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
 * element has sought past the end (a long-standing Chromium quirk). When that
 * happens the hook silently seeks far ahead and back once, which forces the
 * real duration to materialize — voice notes are small, so the extra range
 * fetch is negligible.
 */
export function useMediaPlayback(ref: RefObject<HTMLMediaElement | null>): MediaPlayback {
  const [playing, setPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const rate = useMediaSettings((s) => s.rate)

  // Live speed change while playing.
  useEffect(() => {
    const el = ref.current
    if (el) el.playbackRate = rate
  }, [rate, ref])

  useEffect(() => {
    const el = ref.current
    if (!el) return

    const pause = () => el.pause()

    const syncDuration = () => {
      if (Number.isFinite(el.duration) && el.duration > 0) setDuration(el.duration)
    }
    const fixInfiniteDuration = () => {
      if (Number.isFinite(el.duration)) {
        syncDuration()
        return
      }
      const restore = () => {
        el.removeEventListener("timeupdate", restore)
        el.currentTime = 0
        syncDuration()
      }
      el.addEventListener("timeupdate", restore)
      el.currentTime = 1e7
    }

    const onTime = () => {
      // Skip the transient jump made by the infinite-duration fix above.
      if (!Number.isFinite(el.duration)) return
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

    el.addEventListener("loadedmetadata", fixInfiniteDuration)
    el.addEventListener("durationchange", syncDuration)
    el.addEventListener("timeupdate", onTime)
    el.addEventListener("play", onPlay)
    el.addEventListener("pause", onStop)
    el.addEventListener("ended", onStop)
    // Metadata may already be in by the time the effect runs.
    if (el.readyState >= 1) fixInfiniteDuration()

    return () => {
      el.removeEventListener("loadedmetadata", fixInfiniteDuration)
      el.removeEventListener("durationchange", syncDuration)
      el.removeEventListener("timeupdate", onTime)
      el.removeEventListener("play", onPlay)
      el.removeEventListener("pause", onStop)
      el.removeEventListener("ended", onStop)
      releasePlayback(pause)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

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
