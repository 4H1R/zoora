import { ParticipantEvent, type Participant } from "livekit-client"
import { Cpu, Globe, Monitor, Smartphone, Tablet, Wifi } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { UserAvatar } from "@/components/user-avatar"
import { cn } from "@/lib/utils"

import { NetStatList, QUALITY_COLOR, SignalBars, qualityLabel } from "../connection-quality"
import { readPresence, type DeviceType } from "../presence"
import type { RoomRole } from "../room-role"

const DEVICE_ICON: Record<DeviceType, typeof Smartphone> = {
  mobile: Smartphone,
  tablet: Tablet,
  desktop: Monitor,
}

function InfoRow({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Smartphone
  label: string
  value: string
}) {
  return (
    <div className="flex items-center gap-3 rounded-lg border border-border px-3 py-2.5">
      <Icon className="size-4 shrink-0 text-muted-foreground" />
      <span className="text-xs text-muted-foreground">{label}</span>
      <span className="ms-auto truncate text-sm font-medium text-foreground" dir="ltr">
        {value}
      </span>
    </div>
  )
}

export function ParticipantInfoDialog({
  participant,
  role,
  onClose,
}: {
  participant: Participant | null
  role: RoomRole
  onClose: () => void
}) {
  // Attributes and connection quality mutate on the live Participant object
  // without changing its identity, so re-render on the relevant events.
  const [, force] = useState(0)
  useEffect(() => {
    if (!participant) return
    const rerender = () => force((n) => n + 1)
    participant.on(ParticipantEvent.AttributesChanged, rerender)
    participant.on(ParticipantEvent.ConnectionQualityChanged, rerender)
    return () => {
      participant.off(ParticipantEvent.AttributesChanged, rerender)
      participant.off(ParticipantEvent.ConnectionQualityChanged, rerender)
    }
  }, [participant])

  return (
    <Dialog open={participant !== null} onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="sm:max-w-md">
        {participant && (
          <ParticipantInfoBody participant={participant} role={role} />
        )}
      </DialogContent>
    </Dialog>
  )
}

function ParticipantInfoBody({ participant, role }: { participant: Participant; role: RoomRole }) {
  const { t } = useTranslation()
  const name = participant.name || participant.identity
  const { device, net } = readPresence(participant.attributes)
  // Live value straight off the participant beats the periodically-published one.
  const quality = participant.connectionQuality
  const unknown = t("liveRoom.people.info.unknown")

  const DeviceIcon = device ? DEVICE_ICON[device.device] : Cpu
  const deviceLabel = device
    ? t(`liveRoom.people.info.deviceType.${device.device}`)
    : unknown

  return (
    <>
      <DialogHeader>
        <DialogTitle className="sr-only">{t("liveRoom.people.info.title")}</DialogTitle>
        <div className="flex items-center gap-3">
          <UserAvatar name={name} size="md" online={true} />
          <div className="min-w-0 text-start">
            <p className="truncate text-base font-semibold text-foreground">{name}</p>
            <p className="text-xs text-muted-foreground">{t(`liveRoom.people.roles.${role}`)}</p>
          </div>
        </div>
      </DialogHeader>

      <div className="space-y-2">
        <InfoRow icon={DeviceIcon} label={t("liveRoom.people.info.device")} value={deviceLabel} />
        <InfoRow icon={Cpu} label={t("liveRoom.people.info.os")} value={device?.os ?? unknown} />
        <InfoRow icon={Globe} label={t("liveRoom.people.info.browser")} value={device?.browser ?? unknown} />
      </div>

      <div className="space-y-2.5 rounded-lg border border-border p-3">
        <div className="flex items-center justify-between">
          <span className="flex items-center gap-2 text-sm font-semibold text-foreground">
            <Wifi className="size-4 text-muted-foreground" />
            {t("liveRoom.people.info.network")}
          </span>
          <span className="flex items-center gap-2">
            <SignalBars quality={quality} />
            <span className={cn("text-xs font-medium", QUALITY_COLOR[quality] ?? "text-muted-foreground")}>
              {qualityLabel(quality, t)}
            </span>
          </span>
        </div>
        <NetStatList net={net} showUplink />
      </div>
    </>
  )
}
