import type { DeviceType } from "../presence"
import type { RoomRole } from "../room-role"
import type { Participant } from "livekit-client"

import { ParticipantEvent } from "livekit-client"
import { Cpu, Globe, Monitor, Smartphone, Tablet, Wifi } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { UserAvatar } from "@/components/user-avatar"
import { cn } from "@/lib/utils"

import { NetStatList, QUALITY_COLOR, qualityLabel, SignalBars } from "../connection-quality"
import { readPresence } from "../presence"

const DEVICE_ICON: Record<DeviceType, typeof Smartphone> = {
  mobile: Smartphone,
  tablet: Tablet,
  desktop: Monitor,
}

function InfoRow({ icon: Icon, label, value }: { icon: typeof Smartphone; label: string; value: string }) {
  return (
    <div className="border-border flex items-center gap-3 rounded-lg border px-3 py-2.5">
      <Icon className="text-muted-foreground size-4 shrink-0" />
      <span className="text-muted-foreground text-xs">{label}</span>
      <span className="text-foreground ms-auto truncate text-sm font-medium" dir="ltr">
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
        {participant && <ParticipantInfoBody participant={participant} role={role} />}
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
  const deviceLabel = device ? t(`liveRoom.people.info.deviceType.${device.device}`) : unknown

  return (
    <>
      <DialogHeader>
        <DialogTitle className="sr-only">{t("liveRoom.people.info.title")}</DialogTitle>
        <div className="flex items-center gap-3">
          <UserAvatar name={name} size="md" online={true} />
          <div className="min-w-0 text-start">
            <p className="text-foreground truncate text-base font-semibold">{name}</p>
            <p className="text-muted-foreground text-xs">{t(`liveRoom.people.roles.${role}`)}</p>
          </div>
        </div>
      </DialogHeader>

      <div className="space-y-2">
        <InfoRow icon={DeviceIcon} label={t("liveRoom.people.info.device")} value={deviceLabel} />
        <InfoRow icon={Cpu} label={t("liveRoom.people.info.os")} value={device?.os ?? unknown} />
        <InfoRow icon={Globe} label={t("liveRoom.people.info.browser")} value={device?.browser ?? unknown} />
      </div>

      <div className="border-border space-y-2.5 rounded-lg border p-3">
        <div className="flex items-center justify-between">
          <span className="text-foreground flex items-center gap-2 text-sm font-semibold">
            <Wifi className="text-muted-foreground size-4" />
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
