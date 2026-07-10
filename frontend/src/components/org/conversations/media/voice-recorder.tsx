import type { VoiceRecorder } from "./use-voice-recorder"

import { SendHorizontalIcon, SquareIcon, Trash2Icon } from "lucide-react"
import { motion } from "motion/react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

import { PlayPauseButton, WaveformBars } from "./controls"
import { formatMediaTime } from "./playback"
import { useMediaPlayback } from "./use-media-playback"
import { extractPeaks, syntheticPeaks, VOICE_BUCKETS } from "./waveform"

// Bars visible in the live strip — newest sample enters at the end.
const LIVE_BARS = 40

/**
 * The composer's recording strip — swaps in for the whole input row while a
 * voice message is being captured (Telegram flow):
 * - recording: trash · pulsing dot + timer · live waveform · stop · send
 * - preview:   trash · play + static waveform + duration      · send
 * Send works from either state (recording sends the take immediately).
 */
export function VoiceRecorderStrip({ recorder, onSend }: { recorder: VoiceRecorder; onSend: () => void }) {
  const { t } = useTranslation()
  const recording = recorder.status === "recording"

  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.15, ease: "easeOut" }}
      className="flex min-h-9 items-center gap-2 px-1"
    >
      <Button
        type="button"
        variant="ghost"
        size="icon-sm"
        className="text-muted-foreground hover:text-destructive shrink-0"
        aria-label={t("conversations.voice.discard")}
        onClick={recorder.cancel}
      >
        <Trash2Icon />
      </Button>

      {recording ? (
        <>
          <span className="text-foreground flex shrink-0 items-center gap-2 text-sm tabular-nums">
            <span className="relative flex size-2">
              <span className="bg-destructive absolute inline-flex size-full animate-ping rounded-full opacity-75" />
              <span className="bg-destructive relative inline-flex size-2 rounded-full" />
            </span>
            {formatMediaTime(recorder.elapsed)}
          </span>

          <LiveBars levels={recorder.levels} />

          <Button
            type="button"
            variant="secondary"
            size="icon-sm"
            className="shrink-0 rounded-full"
            aria-label={t("conversations.voice.stop")}
            onClick={() => void recorder.stop()}
          >
            <SquareIcon className="size-3 fill-current" />
          </Button>
        </>
      ) : (
        recorder.blob && <VoicePreview blob={recorder.blob} duration={recorder.duration} />
      )}

      <Button
        type="button"
        size="icon-sm"
        className="shrink-0 rounded-full"
        aria-label={t("conversations.voice.send")}
        onClick={onSend}
      >
        <SendHorizontalIcon className="rtl:rotate-180" />
      </Button>
    </motion.div>
  )
}

/** Live loudness bars — a fixed window that fills from the end as samples land. */
function LiveBars({ levels }: { levels: number[] }) {
  const recent = levels.slice(-LIVE_BARS)
  const bars = [...Array<number>(Math.max(0, LIVE_BARS - recent.length)).fill(0), ...recent]

  return (
    <div className="flex h-8 min-w-0 flex-1 items-center gap-0.5" aria-hidden>
      {bars.map((level, i) => (
        <span
          key={i}
          style={{ height: `${Math.round(10 + level * 90)}%` }}
          className={cn(
            "min-h-1 flex-1 rounded-full transition-[height] duration-100",
            level > 0 ? "bg-primary" : "bg-primary/20"
          )}
        />
      ))}
    </div>
  )
}

/** Playable preview of the finished take before sending. */
function VoicePreview({ blob, duration }: { blob: Blob; duration: number }) {
  const audioRef = useRef<HTMLAudioElement>(null)
  const playback = useMediaPlayback(audioRef)

  const [url, setUrl] = useState<string | null>(null)
  const [peaks, setPeaks] = useState<number[] | null>(null)

  useEffect(() => {
    const objectUrl = URL.createObjectURL(blob)
    setUrl(objectUrl)
    let alive = true
    extractPeaks(blob, VOICE_BUCKETS)
      .then((p) => alive && setPeaks(p))
      .catch(() => alive && setPeaks(syntheticPeaks("preview")))
    return () => {
      alive = false
      URL.revokeObjectURL(objectUrl)
    }
  }, [blob])

  const total = playback.duration > 0 ? playback.duration : duration
  const fraction = total > 0 ? Math.min(1, playback.currentTime / total) : 0

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.97 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.15 }}
      className="bg-muted/60 flex min-w-0 flex-1 items-center gap-2 rounded-full px-1 py-0.5"
    >
      {url && <audio ref={audioRef} src={url} preload="metadata" className="hidden" />}
      <PlayPauseButton playing={playback.playing} onToggle={playback.toggle} tone="accent" size="sm" disabled={!url} />
      <WaveformBars
        peaks={peaks ?? syntheticPeaks("preview")}
        pending={!peaks}
        progress={fraction}
        tone="accent"
        onSeek={playback.seekTo}
        className="h-6"
      />
      <span className="text-muted-foreground pe-1.5 text-xs tabular-nums">
        {formatMediaTime(playback.playing || playback.currentTime > 0 ? playback.currentTime : total)}
      </span>
    </motion.div>
  )
}
