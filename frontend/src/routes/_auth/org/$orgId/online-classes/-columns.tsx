import type { GithubCom4H1RZooraInternalDomainLiveRoom } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { Link } from "@tanstack/react-router"
import { ChevronRight, UserIcon, VideoIcon } from "lucide-react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { getGradient } from "@/components/class-card"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { UserAvatar } from "@/components/user-avatar"
import { badgeLabelKey, badgeStatus, ctaMode } from "@/lib/live-room-status"
import { cn } from "@/lib/utils"

export function useLiveRoomColumns(): ColumnDef<GithubCom4H1RZooraInternalDomainLiveRoom>[] {
  const { t } = useTranslation()
  const { can } = useAccess()
  const canManage = can("live_sessions:manage") || can("live_sessions:manage_any")

  return [
    {
      accessorKey: "name",
      header: t("onlineClassesPage.table.room"),
      cell: ({ row }) => {
        const room = row.original
        const gradient = getGradient(room.id ?? "")
        return (
          <div className="flex items-center gap-3">
            <div
              className={cn(
                "flex size-9 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br text-white",
                gradient
              )}
            >
              <VideoIcon className="size-4" />
            </div>
            <span className="truncate text-sm font-medium">{room.name}</span>
          </div>
        )
      },
      enableSorting: false,
    },
    {
      accessorKey: "class",
      header: t("onlineClassesPage.table.class"),
      cell: ({ row }) => {
        const className = row.original.class_session?.class?.name ?? ""
        return className ? (
          <span className="text-sm">{className}</span>
        ) : (
          <span className="text-muted-foreground">—</span>
        )
      },
      enableSorting: false,
    },
    {
      accessorKey: "host",
      header: t("onlineClassesPage.table.host"),
      cell: ({ row }) => {
        const teacherName = row.original.class_session?.class?.user?.name ?? ""
        return (
          <div className="flex items-center gap-2">
            {teacherName ? (
              <UserAvatar name={teacherName} size="md" />
            ) : (
              <UserIcon className="text-muted-foreground size-4" />
            )}
            <span className="text-sm">{teacherName || t("onlineClassesPage.unknownHost")}</span>
          </div>
        )
      },
      enableSorting: false,
    },
    {
      accessorKey: "status",
      header: t("onlineClassesPage.table.status"),
      cell: ({ row }) => {
        const status = row.original.status
        return <StatusBadge status={badgeStatus(status)}>{t(badgeLabelKey(status))}</StatusBadge>
      },
      enableSorting: false,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => {
        const room = row.original
        const mode = ctaMode(room.status, canManage)
        const liveId = room.id ?? ""

        if (mode === "join") {
          return (
            <div className="text-end">
              <Button size="xs" render={<Link to="/live/$liveId" params={{ liveId }} />}>
                {t("onlineClassesPage.join")}
                <ChevronRight className="size-3.5 rtl:rotate-180" />
              </Button>
            </div>
          )
        }
        if (mode === "start") {
          return (
            <div className="text-end">
              <Button size="xs" variant="outline" render={<Link to="/live/$liveId" params={{ liveId }} />}>
                {t("onlineClassesPage.start")}
                <ChevronRight className="size-3.5 rtl:rotate-180" />
              </Button>
            </div>
          )
        }
        return null
      },
      enableSorting: false,
    },
  ]
}
