import type { GithubCom4H1RZooraInternalDomainLiveRoom } from "@/api/model"
import type { CtaMode } from "@/lib/live-room-status"

import { Link } from "@tanstack/react-router"
import { ChevronRight, UserIcon, VideoIcon } from "lucide-react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { getGradient } from "@/components/class-card"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { UserAvatar } from "@/components/user-avatar"
import { badgeLabelKey, badgeStatus, ctaMode } from "@/lib/live-room-status"
import { formatCountdown, useNow } from "@/lib/session-status"
import { cn } from "@/lib/utils"

interface LiveRoomCardProps {
  room: GithubCom4H1RZooraInternalDomainLiveRoom
}

export function LiveRoomCard({ room }: LiveRoomCardProps) {
  const { t } = useTranslation()
  const { can } = useAccess()
  const now = useNow(1000)

  const cls = room.class_session?.class
  const teacherName = cls?.user?.name ?? ""
  const className = cls?.name ?? ""
  const gradient = getGradient(room.id ?? "")
  const status = room.status
  const canManage = can("live_sessions:manage") || can("live_sessions:manage_any")
  const mode = ctaMode(status, canManage)

  const timeLine = (() => {
    if (status === "active") return t("onlineClassesPage.startedLabel")
    if (status === "finished") return t("onlineClassesPage.endedLabel")
    if (room.scheduled_start_time) {
      return t("onlineClassesPage.startsIn", { time: formatCountdown(room.scheduled_start_time, now) })
    }
    return t("onlineClassesPage.notScheduled")
  })()

  const isLive = status === "active"

  return (
    <div
      className={cn(
        "group/card bg-card flex flex-col overflow-hidden rounded-xl ring-1 transition-all",
        isLive ? "ring-2 ring-[#dc2626]/60" : "ring-foreground/10 hover:ring-foreground/30"
      )}
    >
      <div className={cn("relative flex h-28 flex-col justify-end gap-2 bg-gradient-to-br p-3.5", gradient)}>
        <div className="absolute end-3 top-3">
          <StatusBadge status={badgeStatus(status)}>{t(badgeLabelKey(status))}</StatusBadge>
        </div>
        <p className="line-clamp-2 text-sm leading-snug font-semibold text-white drop-shadow-sm">{room.name}</p>
      </div>

      <div className="flex flex-1 flex-col gap-3 p-3.5">
        {className && (
          <p className="text-muted-foreground inline-flex items-center gap-1.5 text-xs">
            <VideoIcon className="size-3.5" />
            {className}
          </p>
        )}

        <div className="flex items-center gap-2">
          {teacherName ? (
            <UserAvatar name={teacherName} size="md" />
          ) : (
            <UserIcon className="text-muted-foreground size-4" />
          )}
          <span className="text-foreground text-xs font-medium">
            {teacherName || t("onlineClassesPage.unknownHost")}
          </span>
        </div>

        <p className="text-muted-foreground text-xs tabular-nums">{timeLine}</p>

        <div className="mt-auto flex items-center justify-end border-t pt-3">
          <LiveRoomCardCta mode={mode} liveId={room.id ?? ""} />
        </div>
      </div>
    </div>
  )
}

function LiveRoomCardCta({ mode, liveId }: { mode: CtaMode; liveId: string }) {
  const { t } = useTranslation()

  if (mode === "join") {
    return (
      <Button size="xs" render={<Link to="/live/$liveId" params={{ liveId }} />}>
        {t("onlineClassesPage.join")}
        <ChevronRight className="size-3.5 rtl:rotate-180" />
      </Button>
    )
  }

  if (mode === "start") {
    return (
      <Button size="xs" variant="outline" render={<Link to="/live/$liveId" params={{ liveId }} />}>
        {t("onlineClassesPage.start")}
        <ChevronRight className="size-3.5 rtl:rotate-180" />
      </Button>
    )
  }

  if (mode === "waiting") {
    return (
      <Button size="xs" variant="outline" disabled>
        {t("onlineClassesPage.notStartedYet")}
      </Button>
    )
  }

  return (
    <Button size="xs" variant="ghost" disabled>
      {t("onlineClassesPage.status.finished")}
    </Button>
  )
}

export function LiveRoomCardSkeleton() {
  return (
    <div className="ring-border bg-card overflow-hidden rounded-xl ring-1">
      <Skeleton className="h-28 rounded-none" />
      <div className="flex flex-col gap-3 p-3.5">
        <Skeleton className="h-3 w-24" />
        <div className="flex items-center gap-2">
          <Skeleton className="size-6 rounded-full" />
          <Skeleton className="h-3 w-20" />
        </div>
        <Skeleton className="h-3 w-28" />
        <div className="flex items-center justify-end border-t pt-3">
          <Skeleton className="h-6 w-20 rounded-lg" />
        </div>
      </div>
    </div>
  )
}
