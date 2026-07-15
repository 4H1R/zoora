import type { NetStats } from "./presence"
import type { TFunction } from "i18next"

import { ConnectionQuality } from "livekit-client"
import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"

export const QUALITY_BARS: Record<ConnectionQuality, number> = {
  [ConnectionQuality.Excellent]: 3,
  [ConnectionQuality.Good]: 2,
  [ConnectionQuality.Poor]: 1,
  [ConnectionQuality.Lost]: 0,
  [ConnectionQuality.Unknown]: 0,
}

export const QUALITY_COLOR: Record<ConnectionQuality, string> = {
  [ConnectionQuality.Excellent]: "text-emerald-400",
  [ConnectionQuality.Good]: "text-emerald-400",
  [ConnectionQuality.Poor]: "text-amber-400",
  [ConnectionQuality.Lost]: "text-red-400",
  [ConnectionQuality.Unknown]: "text-muted-foreground",
}

export function qualityLabel(quality: ConnectionQuality, t: TFunction): string {
  if (quality === ConnectionQuality.Excellent || quality === ConnectionQuality.Good)
    return t("liveRoom.connection.good")
  if (quality === ConnectionQuality.Poor) return t("liveRoom.connection.poor")
  if (quality === ConnectionQuality.Lost) return t("liveRoom.connection.lost")
  return t("liveRoom.connection.unknown")
}

export function SignalBars({ quality }: { quality: ConnectionQuality }) {
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
            i < filled ? cn(color, "bg-current") : "text-muted-foreground/25 bg-current"
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
      <span className="text-foreground font-mono" dir="ltr">
        {value}
      </span>
    </div>
  )
}

type NetFields = Pick<NetStats, "rtt" | "jitter" | "packetLoss" | "downKbps" | "upKbps">

/** Ping / jitter / loss / bitrate rows shared by the header popover and the
 * participant-info dialog. `net` may be null (nothing measured/published yet). */
export function NetStatList({ net, showUplink = false }: { net: NetFields | null; showUplink?: boolean }) {
  const { t } = useTranslation()
  const na = t("liveRoom.connection.na")
  const fmtMs = (v: number | null | undefined) => (v == null ? na : `${v} ${t("liveRoom.connection.ms")}`)
  const fmtKbps = (v: number | null | undefined) => (v == null ? na : `${v} ${t("liveRoom.connection.kbps")}`)
  return (
    <div className="space-y-1.5">
      <StatRow label={t("liveRoom.connection.ping")} value={fmtMs(net?.rtt)} />
      <StatRow label={t("liveRoom.connection.jitter")} value={fmtMs(net?.jitter)} />
      <StatRow
        label={t("liveRoom.connection.packetLoss")}
        value={net?.packetLoss == null ? na : `${net.packetLoss.toFixed(1)} %`}
      />
      <StatRow label={t("liveRoom.connection.downlink")} value={fmtKbps(net?.downKbps)} />
      {showUplink && <StatRow label={t("liveRoom.connection.uplink")} value={fmtKbps(net?.upKbps)} />}
    </div>
  )
}
