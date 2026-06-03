import { CircleDotIcon, LogInIcon, LogOutIcon, TimerIcon, UsersIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetLiveRoomsIdParticipants } from "@/api/live-sessions/live-sessions"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Skeleton } from "@/components/ui/skeleton"
import { getEntityColor, getInitials } from "@/lib/data-table"
import { formatSessionDate } from "@/lib/session-status"
import { cn } from "@/lib/utils"

function formatDuration(seconds: number | undefined): string {
  const s = seconds ?? 0
  if (s <= 0) return "—"
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const sec = s % 60
  if (h > 0) return `${h}h ${m}m`
  if (m > 0) return `${m}m ${sec}s`
  return `${sec}s`
}

export function LiveRoomParticipants({ roomId }: { roomId: string }) {
  const { t, i18n } = useTranslation()
  const formatDate = (iso?: string) => formatSessionDate(iso, i18n.language, "short")

  const query = useGetLiveRoomsIdParticipants(
    roomId,
    { order_by: "joined_at", order_dir: "desc" },
    { query: { enabled: !!roomId } }
  )
  const participants = (query.data?.status === 200 && query.data.data.data?.items) || []

  if (query.isPending) {
    return (
      <div className="flex flex-col gap-2">
        <Skeleton className="h-12 w-full rounded-xl" />
        <Skeleton className="h-12 w-full rounded-xl" />
      </div>
    )
  }

  if (participants.length === 0) {
    return (
      <div className="text-muted-foreground flex flex-col items-center gap-2 rounded-xl border border-dashed py-8 text-center">
        <UsersIcon className="size-5" />
        <span className="text-sm">{t("org.session.liveRooms.participants.empty")}</span>
      </div>
    )
  }

  return (
    <ul className="flex flex-col gap-1.5">
      {participants.map((p) => {
        const name = p.user?.name ?? p.identity ?? "—"
        const active = !p.left_at
        return (
          <li key={p.id} className="bg-muted/40 flex items-center gap-3 rounded-xl px-3 py-2.5">
            <Avatar className="size-8 shrink-0">
              <AvatarFallback className={cn("text-[10px] font-semibold text-white", getEntityColor(name))}>
                {getInitials(name)}
              </AvatarFallback>
            </Avatar>
            <div className="flex min-w-0 flex-1 flex-col">
              <span className="truncate text-sm font-medium">{name}</span>
              <span className="text-muted-foreground inline-flex flex-wrap items-center gap-x-3 gap-y-0.5 font-mono text-[10px]">
                <span className="inline-flex items-center gap-1">
                  <LogInIcon className="size-3" />
                  {formatDate(p.joined_at)}
                </span>
                {p.left_at ? (
                  <span className="inline-flex items-center gap-1">
                    <LogOutIcon className="size-3" />
                    {formatDate(p.left_at)}
                  </span>
                ) : null}
                <span className="inline-flex items-center gap-1">
                  <TimerIcon className="size-3" />
                  {formatDuration(p.total_duration_seconds)}
                </span>
              </span>
            </div>
            <span
              className={cn(
                "inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 font-mono text-[10px] tracking-[0.2em] uppercase",
                active ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400" : "bg-muted text-muted-foreground"
              )}
            >
              {active ? <CircleDotIcon className="size-3 animate-pulse" /> : null}
              {active ? t("org.session.liveRooms.participants.inRoom") : t("org.session.liveRooms.participants.left")}
            </span>
          </li>
        )
      })}
    </ul>
  )
}
