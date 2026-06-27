import type { GithubCom4H1RZooraInternalDomainLiveRoom as LiveRoom } from "@/api/model"
import type { SortOption } from "@/components/data-table/sort-picker"

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
import { SectionNoResults } from "@/components/org/session/section-no-results"
import { SectionPagination } from "@/components/org/session/section-pagination"
import { SectionToolbar } from "@/components/org/session/section-toolbar"
import { Eyebrow } from "@/components/eyebrow"
import { Button } from "@/components/ui/button"
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible"
import { EmptyState } from "@/components/ui/empty-state"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { DEFAULT_PAGE_SIZE } from "@/lib/list"
import { formatSessionDate } from "@/lib/session-status"
import { useSectionList } from "@/lib/use-section-list"
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

      <div className="flex items-center justify-between gap-2">
        <h3 className="min-w-0 flex-1 truncate text-sm font-medium tracking-tight">
          {room.name?.trim() || room.id?.slice(0, 8).toUpperCase() || "—"}
        </h3>
        <span
          className={cn(
            "inline-flex shrink-0 items-center gap-1.5 rounded-full px-2.5 py-1 font-mono text-[10px] tracking-[0.25em] uppercase",
            STATUS_STYLES[status] ?? "bg-muted text-muted-foreground"
          )}
        >
          {isActive && (
            <span className="relative flex size-1.5">
              <span className="bg-destructive absolute inline-flex h-full w-full animate-ping rounded-full opacity-75" />
              <span className="bg-destructive relative inline-flex size-1.5 rounded-full" />
            </span>
          )}
          {t(`org.session.liveRooms.status.${status}`)}
        </span>
      </div>

      <div className="border-foreground/10 grid grid-cols-2 gap-3 border-t border-dashed pt-3">
        <div className="flex flex-col gap-1">
          <Eyebrow className="text-[10px]">{t("org.session.liveRooms.startedAt")}</Eyebrow>
          <span className="font-mono text-xs tabular-nums">
            {formatDate(room.actual_start_time ?? room.scheduled_start_time)}
          </span>
        </div>
        <div className="flex flex-col gap-1">
          <Eyebrow className="text-[10px]">{t("org.session.liveRooms.endedAt")}</Eyebrow>
          <span className="font-mono text-xs tabular-nums">{formatDate(room.actual_end_time)}</span>
        </div>
      </div>

      <div className="mt-auto flex items-center gap-2">
        {isActive && canJoin && (
          <Button size="sm" className="flex-1" onClick={enter}>
            <RadioIcon className="size-3.5" />
            {t("org.session.liveRooms.actions.join")}
          </Button>
        )}
        {!isActive && !isFinished && canManage && (
          <Button
            size="sm"
            className="flex-1"
            disabled={startMutation.isPending}
            onClick={() => room.id && startMutation.mutate({ id: room.id })}
          >
            <RadioIcon className="size-3.5" />
            {t("org.session.liveRooms.actions.start")}
          </Button>
        )}
        {isActive && canManage && (
          <Button
            size="sm"
            variant="outline"
            disabled={endMutation.isPending}
            onClick={() => room.id && endMutation.mutate({ id: room.id })}
          >
            <SquareIcon className="size-3.5" />
            {t("org.session.liveRooms.actions.end")}
          </Button>
        )}
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
          {open && <LiveRoomRecordings roomId={room.id ?? ""} isActive={isActive} canManage={canManage} />}
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

  const list = useSectionList()
  const sortOptions: SortOption[] = [
    { id: "created_at", label: t("org.session.controls.sortFields.created_at") },
    { id: "status", label: t("org.session.controls.sortFields.status") },
    { id: "actual_start_time", label: t("org.session.controls.sortFields.actual_start_time") },
    { id: "actual_end_time", label: t("org.session.controls.sortFields.actual_end_time") },
  ]
  const statusItems = [
    { value: "all", label: t("org.session.controls.status.all") },
    { value: "created", label: t("org.session.controls.status.created") },
    { value: "active", label: t("org.session.controls.status.active") },
    { value: "finished", label: t("org.session.controls.status.finished") },
  ]

  const query = useGetLiveRooms(
    { class_session_id: classSessionId, status: list.status, ...list.params },
    { query: { enabled: canViewAny } }
  )
  const roomsData = (query.data?.status === 200 && query.data.data.data) || undefined
  const rooms = roomsData?.items ?? []
  const total = roomsData?.total ?? 0
  const pageSize = roomsData?.page_size ?? DEFAULT_PAGE_SIZE

  if (!canViewAny) return null

  return (
    <section id="live-rooms" className="flex scroll-mt-20 flex-col gap-5">
      <div className="flex items-end justify-between gap-4">
        <div className="flex flex-col gap-1.5">
          <h2 className="text-2xl font-semibold tracking-tight">{t("org.session.liveRooms.title")}</h2>
        </div>
        {canCreate && (
          <Button onClick={() => setFormOpen(true)}>
            <PlusIcon className="size-4" />
            {t("org.session.liveRooms.newRoom")}
          </Button>
        )}
      </div>

      {(rooms.length > 0 || list.isFiltered) && (
        <SectionToolbar
          sortOptions={sortOptions}
          sort={list.sort}
          onSortChange={list.setSort}
        >
          <Select
            items={statusItems}
            value={list.status ?? "all"}
            onValueChange={(v) => list.setStatus(v && v !== "all" ? v : undefined)}
          >
            <SelectTrigger className="h-8 w-auto gap-1.5 text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {statusItems.map((item) => (
                <SelectItem key={item.value} value={item.value} className="text-xs">
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </SectionToolbar>
      )}

      {query.isPending ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <CardSkeleton />
          <CardSkeleton />
          <CardSkeleton />
        </div>
      ) : rooms.length === 0 ? (
        list.isFiltered ? (
          <SectionNoResults />
        ) : (
          <EmptyState
            icon={VideoIcon}
            title={t("org.session.liveRooms.emptyTitle")}
            description={canCreate ? t("org.session.liveRooms.emptyHint") : t("org.session.liveRooms.emptyHintMember")}
          >
            {canCreate && (
              <Button onClick={() => setFormOpen(true)}>
                <PlusIcon className="size-4" />
                {t("org.session.liveRooms.newRoom")}
              </Button>
            )}
          </EmptyState>
        )
      ) : (
        <>
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {rooms.map((room, i) => (
              <LiveRoomCard
                key={room.id}
                room={room}
                index={(list.page - 1) * pageSize + i}
                canJoin={canJoin}
                canManage={canManage || canCreate}
              />
            ))}
          </div>
          <SectionPagination
            page={list.page}
            pageSize={pageSize}
            total={total}
            onPageChange={list.setPage}
          />
        </>
      )}

      <LiveRoomFormDialog open={formOpen} onOpenChange={setFormOpen} classSessionId={classSessionId} />
    </section>
  )
}
