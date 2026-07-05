import { useConnectionState, useParticipants } from "@livekit/components-react"
import { ConnectionState } from "livekit-client"
import { MonitorPlay, Users } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"

import { NetStatList, QUALITY_COLOR, SignalBars, qualityLabel } from "./connection-quality"
import type { NetStats } from "./presence"
import { useConnectionStats } from "./use-connection-stats"

function formatElapsed(seconds: number): string {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  const pad = (n: number) => n.toString().padStart(2, "0")
  return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${pad(m)}:${pad(s)}`
}

function ConnectionInfo({ stats }: { stats: NetStats }) {
  const { t } = useTranslation()
  const { quality } = stats

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
            {qualityLabel(quality, t)}
          </span>
        </div>
        <NetStatList net={stats} showUplink />
      </PopoverContent>
    </Popover>
  )
}

export function RoomHeader({
  sessionName,
  className,
  onOpenPeople,
}: {
  sessionName: string
  className?: string
  onOpenPeople?: () => void
}) {
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

        <button
          type="button"
          onClick={onOpenPeople}
          aria-label={t("liveRoom.controls.people")}
          title={t("liveRoom.controls.people")}
          className="inline-flex items-center gap-1.5 rounded-full bg-muted px-2.5 py-1 text-xs font-medium text-foreground transition-colors hover:bg-muted/70"
        >
          <Users className="size-3.5 text-muted-foreground" />
          <span className="font-mono" dir="ltr">
            {participants.length}
          </span>
        </button>

        <span className="hidden rounded-full bg-muted px-2.5 py-1 font-mono text-xs text-muted-foreground sm:inline-block" dir="ltr">
          {formatElapsed(elapsed)}
        </span>
      </div>
    </header>
  )
}
