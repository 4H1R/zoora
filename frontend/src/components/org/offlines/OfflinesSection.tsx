import type { GithubCom4H1RZooraInternalDomainOfflineRoom as OfflineRoom } from "@/api/model"
import type { SortOption } from "@/components/data-table/sort-picker"

import { useQueryClient } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import { CalendarClockIcon, EyeIcon, FilmIcon, PencilIcon, PlusIcon, Trash2Icon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getGetOfflinesQueryKey, useDeleteOfflinesId, useGetOfflines } from "@/api/offlines/offlines"
import { SectionNoResults } from "@/components/org/session/section-no-results"
import { SectionPagination } from "@/components/org/session/section-pagination"
import { SectionToolbar } from "@/components/org/session/section-toolbar"
import { Eyebrow } from "@/components/eyebrow"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"
import { useCanSelfOr } from "@/lib/access"
import { DEFAULT_PAGE_SIZE } from "@/lib/list"
import { formatSessionDate } from "@/lib/session-status"
import { useSectionList } from "@/lib/use-section-list"

import { OfflineFormDialog } from "./OfflineFormDialog"
import { useOfflinePermissions } from "./use-offline-permissions"

interface OfflineCardProps {
  room: OfflineRoom
  index: number
  onEdit: (r: OfflineRoom) => void
  onDelete: (r: OfflineRoom) => void
}

function OfflineCard({ room, index, onEdit, onDelete }: OfflineCardProps) {
  const { t, i18n } = useTranslation()
  const canEdit = useCanSelfOr("offlines:update", "offlines:update_any", room.creator_id)
  const canDelete = useCanSelfOr("offlines:delete", "offlines:delete_any", room.creator_id)
  const tileNumber = String(index + 1).padStart(2, "0")
  const createdStr = formatSessionDate(room.created_at, i18n.language, "short")
  const publishedStr = room.published_at ? formatSessionDate(room.published_at, i18n.language, "short") : null

  return (
    <div className="group/offline bg-card text-card-foreground ring-foreground/10 hover:ring-foreground/30 relative isolate flex flex-col gap-5 overflow-hidden rounded-2xl p-5 ring-1 transition-all">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/8%,transparent_60%)] opacity-0 transition-opacity group-hover/offline:opacity-100"
      />
      <div className="flex items-start justify-between gap-3">
        <div className="bg-muted text-foreground flex size-10 items-center justify-center rounded-xl">
          <FilmIcon className="size-5" />
        </div>
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">/{tileNumber}</span>
      </div>

      <div className="flex flex-col gap-2">
        <Eyebrow>{t("org.session.offlines.cardEyebrow")}</Eyebrow>
        <h3 className="line-clamp-2 text-xl leading-snug font-semibold tracking-tight text-balance">
          {room.title ?? "—"}
        </h3>
        {room.description && (
          <p className="text-muted-foreground line-clamp-2 text-sm leading-relaxed">{room.description}</p>
        )}
      </div>

      <div className="border-foreground/10 grid grid-cols-2 gap-3 border-t border-dashed pt-3">
        <div className="flex flex-col gap-1">
          <Eyebrow className="text-[10px]">{t("org.session.offlines.views")}</Eyebrow>
          <span className="inline-flex items-center gap-1.5 font-mono text-sm tabular-nums">
            <EyeIcon className="size-3.5" />
            {room.view_count ?? 0}
          </span>
        </div>
        <div className="flex flex-col gap-1">
          <Eyebrow className="text-[10px]">{t("org.session.offlines.published")}</Eyebrow>
          <span className="inline-flex items-center gap-1.5 font-mono text-xs tabular-nums">
            <CalendarClockIcon className="size-3.5" />
            {publishedStr ?? t("org.session.offlines.notPublished")}
          </span>
        </div>
      </div>

      <div className="border-foreground/10 mt-auto flex items-center justify-between gap-2 border-t border-dashed pt-3">
        <span className="text-muted-foreground inline-flex items-center gap-2 font-mono text-xs">
          <CalendarClockIcon className="size-3.5" />
          {createdStr}
        </span>
        <div className="flex items-center gap-1.5">
          {(canEdit || canDelete) && (
            <div className="flex items-center gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover/offline:opacity-100">
              {canEdit && (
                <Button variant="ghost" size="icon-xs" title={t("org.session.offlines.actions.edit")} onClick={() => onEdit(room)}>
                  <PencilIcon />
                </Button>
              )}
              {canDelete && (
                <Button
                  variant="ghost"
                  size="icon-xs"
                  className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                  title={t("org.session.offlines.actions.delete")}
                  onClick={() => onDelete(room)}
                >
                  <Trash2Icon />
                </Button>
              )}
            </div>
          )}
          <Button
            size="sm"
            render={<Link to="/org/offlines/$offlineId" params={{ offlineId: room.id ?? "" }} />}
          >
            {t("org.session.offlines.open")}
          </Button>
        </div>
      </div>
    </div>
  )
}

function OfflineCardSkeleton() {
  return (
    <div className="bg-card ring-foreground/10 flex flex-col gap-5 rounded-2xl p-5 ring-1">
      <div className="flex items-center justify-between">
        <Skeleton className="size-10 rounded-xl" />
        <Skeleton className="h-3 w-8" />
      </div>
      <div className="flex flex-col gap-2">
        <Skeleton className="h-3 w-16" />
        <Skeleton className="h-6 w-4/5" />
        <Skeleton className="h-3 w-3/5" />
      </div>
      <div className="grid grid-cols-2 gap-3 border-t border-dashed pt-3">
        <Skeleton className="h-8 w-20" />
        <Skeleton className="h-8 w-20" />
      </div>
    </div>
  )
}

interface OfflinesSectionProps {
  classSessionId: string
}

export function OfflinesSection({ classSessionId }: OfflinesSectionProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { canView, canCreate } = useOfflinePermissions()

  const list = useSectionList()
  const sortOptions: SortOption[] = [
    { id: "created_at", label: t("org.session.controls.sortFields.created_at") },
    { id: "title", label: t("org.session.controls.sortFields.title") },
    { id: "published_at", label: t("org.session.controls.sortFields.published_at") },
    { id: "view_count", label: t("org.session.controls.sortFields.view_count") },
  ]

  const offlinesQuery = useGetOfflines(
    { class_session_id: classSessionId, ...list.params },
    { query: { enabled: canView } }
  )
  const offlinesData = (offlinesQuery.data?.status === 200 && offlinesQuery.data.data.data) || undefined
  const rooms = offlinesData?.items ?? []
  const total = offlinesData?.total ?? 0
  const pageSize = offlinesData?.page_size ?? DEFAULT_PAGE_SIZE

  const [formOpen, setFormOpen] = useState(false)
  const [editingRoom, setEditingRoom] = useState<OfflineRoom | null>(null)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deletingRoom, setDeletingRoom] = useState<OfflineRoom | null>(null)

  const deleteMutation = useDeleteOfflinesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.offlines.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetOfflinesQueryKey() })
        setDeleteOpen(false)
        setDeletingRoom(null)
      },
    },
  })

  const openCreate = () => {
    setEditingRoom(null)
    setFormOpen(true)
  }

  if (!canView) return null

  return (
    <section id="offlines" className="flex flex-col gap-5 scroll-mt-20">
      <div className="flex items-end justify-between gap-4">
        <div className="flex flex-col gap-1.5">
          <h2 className="text-2xl font-semibold tracking-tight">{t("org.session.offlines.title")}</h2>
        </div>
        {canCreate && (
          <Button onClick={openCreate}>
            <PlusIcon className="size-4" />
            {t("org.session.offlines.newOffline")}
          </Button>
        )}
      </div>

      {(rooms.length > 0 || list.isFiltered) && (
        <SectionToolbar
          searchValue={list.searchInput}
          onSearchChange={list.setSearchInput}
          sortOptions={sortOptions}
          sort={list.sort}
          onSortChange={list.setSort}
        />
      )}

      {offlinesQuery.isPending ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <OfflineCardSkeleton />
          <OfflineCardSkeleton />
          <OfflineCardSkeleton />
        </div>
      ) : rooms.length === 0 ? (
        list.isFiltered ? (
          <SectionNoResults />
        ) : (
          <EmptyState
            icon={FilmIcon}
            title={t("org.session.offlines.emptyTitle")}
            description={
              canCreate
                ? t("org.session.offlines.emptyHint")
                : t("org.session.offlines.emptyHintMember")
            }
          >
            {canCreate && (
              <Button onClick={openCreate}>
                <PlusIcon className="size-4" />
                {t("org.session.offlines.newOffline")}
              </Button>
            )}
          </EmptyState>
        )
      ) : (
        <>
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {rooms.map((r, i) => (
              <OfflineCard
                key={r.id}
                room={r}
                index={(list.page - 1) * pageSize + i}
                onEdit={(room) => {
                  setEditingRoom(room)
                  setFormOpen(true)
                }}
                onDelete={(room) => {
                  setDeletingRoom(room)
                  setDeleteOpen(true)
                }}
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

      <OfflineFormDialog
        open={formOpen}
        onOpenChange={(open) => {
          setFormOpen(open)
          if (!open) setEditingRoom(null)
        }}
        room={editingRoom}
        classSessionId={classSessionId}
      />

      <DeleteConfirmDialog
        open={deleteOpen}
        onOpenChange={(open) => {
          if (deleteMutation.isPending) return
          setDeleteOpen(open)
          if (!open) setDeletingRoom(null)
        }}
        resourceName={deletingRoom?.title ?? ""}
        onConfirm={() => {
          if (deletingRoom?.id) deleteMutation.mutate({ id: deletingRoom.id })
        }}
        isLoading={deleteMutation.isPending}
      />
    </section>
  )
}
