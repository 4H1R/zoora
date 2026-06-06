import type { GithubCom4H1RZooraInternalDomainLiveRoom as LiveRoom } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { ChevronDownIcon, FilmIcon, PlusIcon, RadioIcon, SquareIcon, UsersIcon, VideoIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetLiveRoomsQueryKey,
  useGetLiveRooms,
  useGetLiveRoomsIdParticipants,
  usePostLiveRoomsIdEnd,
  usePostLiveRoomsIdStart,
} from "@/api/live-sessions/live-sessions"
import { Eyebrow } from "@/components/eyebrow"
import { Button } from "@/components/ui/button"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import { Skeleton } from "@/components/ui/skeleton"
import { formatSessionDate } from "@/lib/session-status"
import { cn } from "@/lib/utils"

import { LiveRoomFormDialog } from "./LiveRoomFormDialog"
import { LiveRoomRecordings } from "./LiveRoomRecordings"
import { useLivesessionPermissions } from "./use-livesession-permissions"

const STATUS_STYLES: Record<string, string> = {
  active: "bg-destructive/10 text-destructive",
  created: "bg-primary/10 text-primary",
  finished: "bg-muted text-muted-foreground",
}

function ParticipantCountBadge({ roomId, isActive }: { roomId: string; isActive: boolean }) {
  const { t } = useTranslation()
  const query = useGetLiveRoomsIdParticipants(
    roomId,
    { order_by: "joined_at", order_dir: "desc" },
    { query: { enabled: !!roomId } }
  )
  const items = (query.data?.status === 200 && query.data.data.data?.items) || []
  const active = items.filter((p) => !p.left_at).length
  const total = items.length
  if (query.isPending || total === 0) return null
  return (
    <span className="text-muted-foreground inline-flex items-center gap-1.5 font-mono text-[10px] tracking-[0.2em] uppercase">
      <UsersIcon className="size-3" />
      {isActive ? t("org.session.liveRooms.participantsActive", { active, total }) : t("org.session.liveRooms.participantsTotal", { total })}
    </span>
  )
}

interface LiveRoomCardProps {
  room: LiveRoom
  index: number
  canJoin: boolean
  canManage: boolean
}

function LiveRoomCard({ room, index, canJoin, canManage }: LiveRoomCardProps) {
  const { t, i18n } = useTranslation()
  const formatDate = (iso?: string) => formatSessionDate(iso, i18n.language, "short")
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)

  const status = room.status ?? "created"
  const isActive = status === "active"
  const isFinished = status === "finished"
  const tileNumber = String(index + 1).padStart(2, "0")

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getGetLiveRoomsQueryKey() })

  const startMutation = usePostLiveRoomsIdStart({
    mutation: {
      onSuccess: () => {
        invalidate()
        if (room.id) navigate({ to: "/live/$liveId", params: { liveId: room.id } })
      },
    },
  })

  const endMutation = usePostLiveRoomsIdEnd({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.liveRooms.actions.endSuccess"))
        invalidate()
      },
    },
  })

  const enter = () => {
    if (room.id) navigate({ to: "/live/$liveId", params: { liveId: room.id } })
  }

  return (
    <div className="group/room bg-card text-card-foreground ring-foreground/10 hover:ring-foreground/30 relative isolate flex flex-col gap-4 overflow-hidden rounded-2xl p-5 ring-1 transition-all">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/8%,transparent_60%)] opacity-0 transition-opacity group-hover/room:opacity-100"
      />
      <div className="flex items-start justify-between gap-3">
        <div
          className={cn(
            "flex size-10 items-center justify-center rounded-xl",
            isActive ? "bg-destructive/10 text-destructive" : "bg-muted text-foreground"
          )}
        >
          <VideoIcon className="size-5" />
        </div>
        <div className="flex items-center gap-3">
          <ParticipantCountBadge roomId={room.id ?? ""} isActive={isActive} />
          <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">/{tileNumber}</span>
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <span
          className={cn(
            "inline-flex w-fit items-center gap-1.5 rounded-full px-2.5 py-1 font-mono text-[10px] tracking-[0.25em] uppercase",
            STATUS_STYLES[status] ?? "bg-muted text-muted-foreground"
          )}
        >
          {isActive ? (
            <span className="relative flex size-1.5">
              <span className="bg-destructive absolute inline-flex h-full w-full animate-ping rounded-full opacity-75" />
              <span className="bg-destructive relative inline-flex size-1.5 rounded-full" />
            </span>
          ) : null}
          {t(`org.session.liveRooms.status.${status}`)}
        </span>
        <h3 className="line-clamp-1 text-sm font-medium tracking-tight">
          {room.name?.trim() || room.livekit_room_name || room.id?.slice(0, 8).toUpperCase() || "—"}
        </h3>
      </div>

      <div className="border-foreground/10 grid grid-cols-2 gap-3 border-t border-dashed pt-3">
        <div className="flex flex-col gap-1">
          <Eyebrow className="text-[10px]">{t("org.session.liveRooms.startedAt")}</Eyebrow>
          <span className="font-mono text-xs tabular-nums">{formatDate(room.actual_start_time)}</span>
        </div>
        <div className="flex flex-col gap-1">
          <Eyebrow className="text-[10px]">{t("org.session.liveRooms.endedAt")}</Eyebrow>
          <span className="font-mono text-xs tabular-nums">{formatDate(room.actual_end_time)}</span>
        </div>
      </div>

      <div className="mt-auto flex items-center gap-2">
        {isActive && canJoin ? (
          <Button size="sm" className="flex-1" onClick={enter}>
            <RadioIcon className="size-3.5" />
            {t("org.session.liveRooms.actions.join")}
          </Button>
        ) : null}
        {!isActive && !isFinished && canManage ? (
          <Button
            size="sm"
            className="flex-1"
            disabled={startMutation.isPending}
            onClick={() => room.id && startMutation.mutate({ id: room.id })}
          >
            <RadioIcon className="size-3.5" />
            {t("org.session.liveRooms.actions.start")}
          </Button>
        ) : null}
        {isActive && canManage ? (
          <Button
            size="sm"
            variant="outline"
            disabled={endMutation.isPending}
            onClick={() => room.id && endMutation.mutate({ id: room.id })}
          >
            <SquareIcon className="size-3.5" />
            {t("org.session.liveRooms.actions.end")}
          </Button>
        ) : null}
      </div>

      <Collapsible open={open} onOpenChange={setOpen}>
        <CollapsibleTrigger
          render={
            <Button variant="ghost" size="sm" className="w-full justify-between">
              <span className="inline-flex items-center gap-2">
                <FilmIcon className="size-3.5" />
                {t("org.session.liveRooms.recordingsLabel")}
              </span>
              <ChevronDownIcon className={cn("size-4 transition-transform", open && "rotate-180")} />
            </Button>
          }
        />
        <CollapsibleContent className="flex flex-col gap-4 pt-3">
          {open ? <LiveRoomRecordings roomId={room.id ?? ""} isActive={isActive} canManage={canManage} /> : null}
        </CollapsibleContent>
      </Collapsible>
    </div>
  )
}

function CardSkeleton() {
  return (
    <div className="bg-card ring-foreground/10 flex flex-col gap-4 rounded-2xl p-5 ring-1">
      <div className="flex items-center justify-between">
        <Skeleton className="size-10 rounded-xl" />
        <Skeleton className="h-3 w-8" />
      </div>
      <Skeleton className="h-5 w-24 rounded-full" />
      <div className="grid grid-cols-2 gap-3 border-t border-dashed pt-3">
        <Skeleton className="h-8 w-20" />
        <Skeleton className="h-8 w-20" />
      </div>
      <Skeleton className="h-9 w-full" />
    </div>
  )
}

export function LiveRoomsSection({ classSessionId }: { classSessionId: string }) {
  const { t } = useTranslation()
  const { canView, canJoin, canCreate, canManage } = useLivesessionPermissions()
  const canViewAny = canView || canJoin
  const [formOpen, setFormOpen] = useState(false)

  const query = useGetLiveRooms({ class_session_id: classSessionId }, { query: { enabled: canViewAny } })
  const rooms = (query.data?.status === 200 && query.data.data.data?.items) || []

  if (!canViewAny) return null

  return (
    <section id="live-rooms" className="flex scroll-mt-20 flex-col gap-5">
      <div className="flex items-end justify-between gap-4">
        <div className="flex flex-col gap-1.5">
          <Eyebrow>{t("org.session.liveRooms.eyebrow")}</Eyebrow>
          <h2 className="text-2xl font-semibold tracking-tight">{t("org.session.liveRooms.title")}</h2>
        </div>
        {canCreate ? (
          <Button onClick={() => setFormOpen(true)}>
            <PlusIcon className="size-4" />
            {t("org.session.liveRooms.newRoom")}
          </Button>
        ) : null}
      </div>

      {query.isPending ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <CardSkeleton />
          <CardSkeleton />
          <CardSkeleton />
        </div>
      ) : rooms.length === 0 ? (
        <div className="bg-card ring-foreground/10 flex flex-col items-center gap-3 rounded-2xl px-6 py-16 text-center ring-1">
          <VideoIcon className="text-muted-foreground size-8" />
          <h3 className="text-foreground text-lg font-semibold tracking-tight">
            {t("org.session.liveRooms.emptyTitle")}
          </h3>
          <p className="text-muted-foreground max-w-md text-sm leading-relaxed">
            {canCreate ? t("org.session.liveRooms.emptyHint") : t("org.session.liveRooms.emptyHintMember")}
          </p>
          {canCreate ? (
            <Button className="mt-2" onClick={() => setFormOpen(true)}>
              <PlusIcon className="size-4" />
              {t("org.session.liveRooms.newRoom")}
            </Button>
          ) : null}
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {rooms.map((room, i) => (
            <LiveRoomCard key={room.id} room={room} index={i} canJoin={canJoin} canManage={canManage || canCreate} />
          ))}
        </div>
      )}

      <LiveRoomFormDialog open={formOpen} onOpenChange={setFormOpen} classSessionId={classSessionId} />
    </section>
  )
}
