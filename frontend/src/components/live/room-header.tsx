import { useConnectionState, useParticipants } from "@livekit/components-react"
import { ConnectionQuality, ConnectionState } from "livekit-client"
import { MonitorPlay, Users } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"

import { useConnectionStats, type ConnectionStats } from "./use-connection-stats"

function formatElapsed(seconds: number): string {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  const pad = (n: number) => n.toString().padStart(2, "0")
  return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${pad(m)}:${pad(s)}`
}

const QUALITY_BARS: Record<ConnectionQuality, number> = {
  [ConnectionQuality.Excellent]: 3,
  [ConnectionQuality.Good]: 2,
  [ConnectionQuality.Poor]: 1,
  [ConnectionQuality.Lost]: 0,
  [ConnectionQuality.Unknown]: 0,
}

const QUALITY_COLOR: Record<ConnectionQuality, string> = {
  [ConnectionQuality.Excellent]: "text-emerald-400",
  [ConnectionQuality.Good]: "text-emerald-400",
  [ConnectionQuality.Poor]: "text-amber-400",
  [ConnectionQuality.Lost]: "text-red-400",
  [ConnectionQuality.Unknown]: "text-muted-foreground",
}

function SignalBars({ quality }: { quality: ConnectionQuality }) {
  const filled = QUALITY_BARS[quality] ?? 0
  const color = QUALITY_COLOR[quality] ?? "text-muted-foreground"
  const heights = ["h-1.5", "h-2.5", "h-3.5"]
  return (
    <span className="flex items-end gap-0.5" dir="ltr">
      {heights.map((h, i) => (
        <span
          key={h}
          className={cn(
            "w-1 rounded-sm",
            h,
            i < filled ? cn(color, "bg-current") : "bg-current text-muted-foreground/25"
          )}
        />
      ))}
    </span>
  )
}

function StatRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-4 text-xs">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-mono text-foreground" dir="ltr">
        {value}
      </span>
    </div>
  )
}

function ConnectionInfo({ stats }: { stats: ConnectionStats }) {
  const { t } = useTranslation()
  const { quality, rtt, jitter, packetLoss, downKbps } = stats

  const qualityLabel =
    quality === ConnectionQuality.Excellent || quality === ConnectionQuality.Good
      ? t("liveRoom.connection.good")
      : quality === ConnectionQuality.Poor
        ? t("liveRoom.connection.poor")
        : quality === ConnectionQuality.Lost
          ? t("liveRoom.connection.lost")
          : t("liveRoom.connection.unknown")

  const na = t("liveRoom.connection.na")
  const fmtMs = (v: number | null) => (v == null ? na : `${v} ${t("liveRoom.connection.ms")}`)

  return (
    <Popover>
      <PopoverTrigger
        aria-label={t("liveRoom.connection.title")}
        className="inline-flex size-7 items-center justify-center rounded-full bg-muted text-foreground transition-colors hover:bg-muted/70"
      >
        <SignalBars quality={quality} />
      </PopoverTrigger>
      <PopoverContent align="end" className="w-56 space-y-2.5">
        <div className="flex items-center justify-between">
          <span className="text-sm font-semibold text-foreground">{t("liveRoom.connection.title")}</span>
          <span className={cn("text-xs font-medium", QUALITY_COLOR[quality] ?? "text-muted-foreground")}>
            {qualityLabel}
          </span>
        </div>
        <div className="space-y-1.5">
          <StatRow label={t("liveRoom.connection.ping")} value={fmtMs(rtt)} />
          <StatRow label={t("liveRoom.connection.jitter")} value={fmtMs(jitter)} />
          <StatRow
            label={t("liveRoom.connection.packetLoss")}
            value={packetLoss == null ? na : `${packetLoss.toFixed(1)} %`}
          />
          <StatRow
            label={t("liveRoom.connection.bitrate")}
            value={downKbps == null ? na : `${downKbps} ${t("liveRoom.connection.kbps")}`}
          />
        </div>
      </PopoverContent>
    </Popover>
  )
}

export function RoomHeader({ sessionName, className }: { sessionName: string; className?: string }) {
  const { t } = useTranslation()
  const participants = useParticipants()
  const state = useConnectionState()
  const stats = useConnectionStats()
  const [elapsed, setElapsed] = useState(0)

  useEffect(() => {
    const start = Date.now()
    const id = setInterval(() => setElapsed(Math.floor((Date.now() - start) / 1000)), 1000)
    return () => clearInterval(id)
  }, [])

  const connected = state === ConnectionState.Connected
  const reconnecting = state === ConnectionState.Reconnecting || state === ConnectionState.SignalReconnecting

  // Solid bg, no backdrop-blur: in landscape this header floats over the <video>
  // stage, and a backdrop-filter pass over a video paints it black on some GPUs.
  return (
    <header className="flex shrink-0 items-center justify-between gap-3 border-b border-border bg-background/95 px-4 py-2.5 sm:px-5">
      <div className="flex min-w-0 items-center gap-3">
        <div className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-primary/15 text-primary">
          <MonitorPlay className="size-5" />
        </div>
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold text-foreground">{sessionName}</p>
          {className && <p className="truncate text-xs text-muted-foreground">{className}</p>}
        </div>
      </div>

      <div className="flex items-center gap-2">
        {reconnecting ? (
          <span className="inline-flex items-center gap-1.5 rounded-full bg-amber-500/15 px-2.5 py-1 text-[11px] font-medium text-amber-300">
            <span className="size-1.5 animate-pulse rounded-full bg-amber-400" />
            {t("liveRoom.reconnecting")}
          </span>
        ) : (
          <span
            className={cn(
              "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-[11px] font-semibold tracking-wide",
              connected ? "bg-red-500 text-white" : "bg-muted text-muted-foreground"
            )}
          >
            <span className={cn("size-1.5 rounded-full bg-white", connected && "animate-pulse")} />
            {connected ? t("status.live") : t("liveRoom.connecting")}
          </span>
        )}

        {connected && <ConnectionInfo stats={stats} />}

        <span className="inline-flex items-center gap-1.5 rounded-full bg-muted px-2.5 py-1 text-xs font-medium text-foreground">
          <Users className="size-3.5 text-muted-foreground" />
          <span className="font-mono" dir="ltr">
            {participants.length}
          </span>
        </span>

        <span className="hidden rounded-full bg-muted px-2.5 py-1 font-mono text-xs text-muted-foreground sm:inline-block" dir="ltr">
          {formatElapsed(elapsed)}
        </span>
      </div>
    </header>
  )
}
