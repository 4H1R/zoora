import { useEffect, useRef, useState } from "react"

import { voiceFileName } from "./waveform"

// First supported container wins; Chromium/Firefox take webm+opus, Safari mp4.
const MIME_CANDIDATES = ["audio/webm;codecs=opus", "audio/webm", "audio/mp4", "audio/ogg;codecs=opus"]

const EXT_BY_MIME: Record<string, string> = {
  "audio/webm": "webm",
  "audio/mp4": "m4a",
  "audio/ogg": "ogg",
}

// Live level sampling cadence — ~10 bars/second reads naturally at the
// composer's bar width.
const LEVEL_INTERVAL_MS = 100
const CLOCK_INTERVAL_MS = 200

// Discard blips shorter than this (accidental taps).
const MIN_DURATION_S = 0.4

export type VoiceRecorderStatus = "idle" | "recording" | "preview"

export interface VoiceRecorder {
  status: VoiceRecorderStatus
  /** Seconds since recording started (live). */
  elapsed: number
  /** Live 0..1 loudness samples, appended while recording. */
  levels: number[]
  /** The finished take (preview state only). */
  blob: Blob | null
  /** Final take length in seconds (preview state only). */
  duration: number
  /** Ask for the mic and start recording. False when access is denied. */
  start: () => Promise<boolean>
  /** Stop recording and move to preview. */
  stop: () => Promise<void>
  /** Discard everything and return to idle. */
  cancel: () => void
  /** Finish (stopping first if still recording) and return the voice File. */
  finish: () => Promise<File | null>
}

/**
 * MediaRecorder-based voice-message capture. Alongside the recording it runs
 * an AnalyserNode sampling RMS loudness so the composer can animate a live
 * waveform. All capture resources (mic stream, audio context, timers) are torn
 * down on stop/cancel/unmount — the mic indicator must never linger.
 */
export function useVoiceRecorder(): VoiceRecorder {
  const [status, setStatus] = useState<VoiceRecorderStatus>("idle")
  const [elapsed, setElapsed] = useState(0)
  const [levels, setLevels] = useState<number[]>([])
  const [blob, setBlob] = useState<Blob | null>(null)
  const [duration, setDuration] = useState(0)

  const recorderRef = useRef<MediaRecorder | null>(null)
  const streamRef = useRef<MediaStream | null>(null)
  const audioCtxRef = useRef<AudioContext | null>(null)
  const chunksRef = useRef<Blob[]>([])
  const timersRef = useRef<number[]>([])
  const startedAtRef = useRef(0)
  const mimeRef = useRef("audio/webm")

  // Release everything that keeps the mic hot.
  function stopCapture() {
    timersRef.current.forEach((id) => window.clearInterval(id))
    timersRef.current = []
    streamRef.current?.getTracks().forEach((track) => track.stop())
    streamRef.current = null
    void audioCtxRef.current?.close().catch(() => {})
    audioCtxRef.current = null
  }

  function reset() {
    setStatus("idle")
    setElapsed(0)
    setLevels([])
    setBlob(null)
    setDuration(0)
  }

  async function start(): Promise<boolean> {
    let stream: MediaStream
    try {
      stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    } catch {
      return false
    }

    const mime = MIME_CANDIDATES.find((m) => MediaRecorder.isTypeSupported(m))
    const recorder = new MediaRecorder(stream, mime ? { mimeType: mime } : undefined)
    mimeRef.current = (recorder.mimeType || mime || "audio/webm").split(";")[0]
    chunksRef.current = []
    recorder.ondataavailable = (e) => {
      if (e.data.size > 0) chunksRef.current.push(e.data)
    }
    recorder.start(250)

    // Loudness taps for the live waveform.
    const ctx = new AudioContext()
    const analyser = ctx.createAnalyser()
    analyser.fftSize = 512
    ctx.createMediaStreamSource(stream).connect(analyser)
    const buf = new Uint8Array(analyser.fftSize)
    const levelTimer = window.setInterval(() => {
      analyser.getByteTimeDomainData(buf)
      let sum = 0
      for (let i = 0; i < buf.length; i++) {
        const v = (buf[i] - 128) / 128
        sum += v * v
      }
      const rms = Math.sqrt(sum / buf.length)
      setLevels((prev) => [...prev, Math.min(1, rms * 3.5)])
    }, LEVEL_INTERVAL_MS)
    const clockTimer = window.setInterval(
      () => setElapsed((Date.now() - startedAtRef.current) / 1000),
      CLOCK_INTERVAL_MS
    )

    recorderRef.current = recorder
    streamRef.current = stream
    audioCtxRef.current = ctx
    timersRef.current = [levelTimer, clockTimer]
    startedAtRef.current = Date.now()

    setElapsed(0)
    setLevels([])
    setBlob(null)
    setStatus("recording")
    return true
  }

  // Stop the MediaRecorder and resolve with the assembled take.
  function stopRecorder(): Promise<Blob | null> {
    const recorder = recorderRef.current
    recorderRef.current = null
    const assemble = () =>
      chunksRef.current.length > 0 ? new Blob(chunksRef.current, { type: mimeRef.current }) : null

    if (!recorder || recorder.state === "inactive") return Promise.resolve(assemble())
    return new Promise((resolve) => {
      recorder.onstop = () => resolve(assemble())
      recorder.stop()
    })
  }

  async function stop(): Promise<void> {
    const seconds = (Date.now() - startedAtRef.current) / 1000
    const take = await stopRecorder()
    stopCapture()
    if (!take || seconds < MIN_DURATION_S) {
      reset()
      return
    }
    setDuration(seconds)
    setBlob(take)
    setStatus("preview")
  }

  function cancel() {
    void stopRecorder()
    stopCapture()
    reset()
  }

  async function finish(): Promise<File | null> {
    let take = blob
    let seconds = duration
    if (status === "recording") {
      seconds = (Date.now() - startedAtRef.current) / 1000
      take = await stopRecorder()
      stopCapture()
    }
    reset()
    if (!take || take.size === 0 || seconds < MIN_DURATION_S) return null
    const ext = EXT_BY_MIME[mimeRef.current] ?? "webm"
    return new File([take], voiceFileName(ext), { type: mimeRef.current })
  }

  // Unmount safety net: kill the mic if the composer disappears mid-recording.
  useEffect(() => {
    return () => {
      const recorder = recorderRef.current
      if (recorder && recorder.state !== "inactive") recorder.stop()
      stopCapture()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return { status, elapsed, levels, blob, duration, start, stop, cancel, finish }
}
