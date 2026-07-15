import { useConnectionState } from "@livekit/components-react"
import { ConnectionState } from "livekit-client"
import { Loader2, WifiOff } from "lucide-react"
import { useEffect, useRef } from "react"
import { useTranslation } from "react-i18next"

// Synthesized via WebAudio so we ship no binary asset.
// Reconnecting -> descending warning tone; reconnected -> ascending chime.
function playTone(kind: "warning" | "success") {
  try {
    const Ctx =
      window.AudioContext ?? (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext
    if (!Ctx) return
    const ctx = new Ctx()
    const now = ctx.currentTime
    const freqs = kind === "warning" ? [660, 440] : [523, 784]
    freqs.forEach((freq, i) => {
      const osc = ctx.createOscillator()
      const gain = ctx.createGain()
      const start = now + i * 0.18
      osc.type = "sine"
      osc.frequency.value = freq
      gain.gain.setValueAtTime(0.0001, start)
      gain.gain.exponentialRampToValueAtTime(0.2, start + 0.02)
      gain.gain.exponentialRampToValueAtTime(0.0001, start + 0.16)
      osc.connect(gain).connect(ctx.destination)
      osc.start(start)
      osc.stop(start + 0.18)
    })
    // Release the context once the last tone finishes.
    window.setTimeout(() => void ctx.close(), 600)
  } catch {
    // Audio is best-effort; ignore failures (autoplay policy, no device, etc.)
  }
}

export function ReconnectOverlay() {
  const { t } = useTranslation()
  const state = useConnectionState()
  const wasReconnecting = useRef(false)

  const reconnecting = state === ConnectionState.Reconnecting || state === ConnectionState.SignalReconnecting

  useEffect(() => {
    if (reconnecting && !wasReconnecting.current) {
      wasReconnecting.current = true
      playTone("warning")
    } else if (!reconnecting && wasReconnecting.current) {
      wasReconnecting.current = false
      if (state === ConnectionState.Connected) playTone("success")
    }
  }, [reconnecting, state])

  if (!reconnecting) return null

  return (
    <div
      role="alert"
      aria-live="assertive"
      className="bg-background/85 absolute inset-0 z-50 flex flex-col items-center justify-center gap-4 backdrop-blur-sm"
    >
      <div className="relative flex size-16 items-center justify-center rounded-full bg-amber-500/15 text-amber-400">
        <WifiOff className="size-7" />
        <Loader2 className="absolute size-16 animate-spin text-amber-400/40" />
      </div>
      <div className="text-center">
        <p className="text-foreground text-base font-semibold">{t("liveRoom.reconnecting")}</p>
        <p className="text-muted-foreground mt-1 text-sm">{t("liveRoom.reconnectHint")}</p>
      </div>
    </div>
  )
}
