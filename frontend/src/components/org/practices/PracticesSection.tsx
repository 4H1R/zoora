import type { GithubCom4H1RZooraInternalDomainPracticeRoom as PracticeRoom } from "@/api/model"
import type { SortOption } from "@/components/data-table/sort-picker"

import { useQueryClient } from "@tanstack/react-query"
import {
  CalendarClockIcon,
  ClockIcon,
  DumbbellIcon,
  PencilIcon,
  PlusIcon,
  SendIcon,
  Trash2Icon,
  TrophyIcon,
} from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetPracticesQueryKey,
  useDeletePracticesId,
  useGetPractices,
} from "@/api/practices/practices"
import { MediaAttachmentList } from "@/components/media/MediaAttachmentList"
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
import { formatScore } from "@/lib/score"
import { formatSessionDate } from "@/lib/session-status"
import { useSectionList } from "@/lib/use-section-list"

import { PracticeFormDialog } from "./PracticeFormDialog"
import { PracticeSubmitDialog } from "./PracticeSubmitDialog"
import { usePracticePermissions } from "./use-practice-permissions"

interface PracticeCardProps {
  practice: PracticeRoom
  index: number
  canSubmit: boolean
  onEdit: (p: PracticeRoom) => void
  onDelete: (p: PracticeRoom) => void
  onSubmit: (p: PracticeRoom) => void
}

function PracticeCard({ practice, index, canSubmit, onEdit, onDelete, onSubmit }: PracticeCardProps) {
  const { t, i18n } = useTranslation()
  const canEdit = useCanSelfOr("practices:update", "practices:update_any", practice.user_id)
  const canDelete = useCanSelfOr("practices:delete", "practices:delete_any", practice.user_id)
  const tileNumber = String(index + 1).padStart(2, "0")
  const createdStr = formatSessionDate(practice.created_at, i18n.language, "short")
  const startStr = formatSessionDate(practice.start_time, i18n.language, "short")
  const endStr = formatSessionDate(practice.end_time, i18n.language, "short")

  return (
    <div className="group/practice bg-card text-card-foreground ring-foreground/10 hover:ring-foreground/30 relative isolate flex flex-col gap-5 overflow-hidden rounded-2xl p-5 ring-1 transition-all">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/8%,transparent_60%)] opacity-0 transition-opacity group-hover/practice:opacity-100"
      />
      <div className="flex items-start justify-between gap-3">
        <div className="bg-muted text-foreground flex size-10 items-center justify-center rounded-xl">
          <DumbbellIcon className="size-5" />
        </div>
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">/{tileNumber}</span>
      </div>

      <div className="flex flex-col gap-2">
        <Eyebrow>{t("org.session.practices.cardEyebrow")}</Eyebrow>
        <h3 className="line-clamp-2 text-xl leading-snug font-semibold tracking-tight text-balance">
          {practice.title ?? "—"}
        </h3>
        {practice.content && (
          <p className="text-muted-foreground line-clamp-2 text-sm leading-relaxed">{practice.content}</p>
        )}
        {practice.attachments && practice.attachments.length > 0 && (
          <MediaAttachmentList mediaIds={practice.attachments} className="pt-1" />
        )}
      </div>

      <div className="border-foreground/10 grid grid-cols-2 gap-3 border-t border-dashed pt-3">
        <div className="flex flex-col gap-1">
          <Eyebrow className="text-[10px]">{t("org.session.practices.window")}</Eyebrow>
          <span className="inline-flex items-center gap-1.5 font-mono text-xs tabular-nums">
            <ClockIcon className="size-3.5" />
            {startStr} → {endStr}
          </span>
        </div>
        <div className="flex flex-col gap-1">
          <Eyebrow className="text-[10px]">{t("org.session.practices.maxScore")}</Eyebrow>
          <span className="inline-flex items-center gap-1.5 font-mono text-sm tabular-nums">
            <TrophyIcon className="size-3.5" />
            {formatScore(practice.max_score ?? 0)}
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
            <div className="flex items-center gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover/practice:opacity-100">
              {canEdit && (
                <Button
                  variant="ghost"
                  size="icon-xs"
                  title={t("org.session.practices.actions.edit")}
                  onClick={() => onEdit(practice)}
                >
                  <PencilIcon />
                </Button>
              )}
              {canDelete && (
                <Button
                  variant="ghost"
                  size="icon-xs"
                  className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                  title={t("org.session.practices.actions.delete")}
                  onClick={() => onDelete(practice)}
                >
                  <Trash2Icon />
                </Button>
              )}
            </div>
          )}
          {canSubmit && (
            <Button size="sm" onClick={() => onSubmit(practice)}>
              <SendIcon className="size-3.5" />
              {t("org.session.practices.actions.submit")}
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}

function PracticeCardSkeleton() {
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

interface PracticesSectionProps {
  classSessionId: string
}

export function PracticesSection({ classSessionId }: PracticesSectionProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { canView, canCreate, canSubmit } = usePracticePermissions()

  const list = useSectionList()
  const sortOptions: SortOption[] = [
    { id: "created_at", label: t("org.session.controls.sortFields.created_at") },
    { id: "title", label: t("org.session.controls.sortFields.title") },
    { id: "start_time", label: t("org.session.controls.sortFields.start_time") },
    { id: "end_time", label: t("org.session.controls.sortFields.end_time") },
  ]

  const practicesQuery = useGetPractices(
    { class_session_id: classSessionId, ...list.params },
    { query: { enabled: canView } }
  )
  const practicesData = (practicesQuery.data?.status === 200 && practicesQuery.data.data.data) || undefined
  const practices = practicesData?.items ?? []
  const total = practicesData?.total ?? 0
  const pageSize = practicesData?.page_size ?? DEFAULT_PAGE_SIZE

  const [formOpen, setFormOpen] = useState(false)
  const [editingPractice, setEditingPractice] = useState<PracticeRoom | null>(null)
  const [submitOpen, setSubmitOpen] = useState(false)
  const [submitPractice, setSubmitPractice] = useState<PracticeRoom | null>(null)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deletingPractice, setDeletingPractice] = useState<PracticeRoom | null>(null)

  const deleteMutation = useDeletePracticesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.practices.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetPracticesQueryKey() })
        setDeleteOpen(false)
        setDeletingPractice(null)
      },
    },
  })

  const openCreate = () => {
    setEditingPractice(null)
    setFormOpen(true)
  }

  if (!canView) return null

  return (
    <section id="practices" className="flex flex-col gap-5 scroll-mt-20">
      <div className="flex items-end justify-between gap-4">
        <div className="flex flex-col gap-1.5">
          <h2 className="text-2xl font-semibold tracking-tight">{t("org.session.practices.title")}</h2>
        </div>
        {canCreate && (
          <Button onClick={openCreate}>
            <PlusIcon className="size-4" />
            {t("org.session.practices.newPractice")}
          </Button>
        )}
      </div>

      {(practices.length > 0 || list.isFiltered) && (
        <SectionToolbar
          searchValue={list.searchInput}
          onSearchChange={list.setSearchInput}
          sortOptions={sortOptions}
          sort={list.sort}
          onSortChange={list.setSort}
        />
      )}

      {practicesQuery.isPending ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <PracticeCardSkeleton />
          <PracticeCardSkeleton />
          <PracticeCardSkeleton />
        </div>
      ) : practices.length === 0 ? (
        list.isFiltered ? (
          <SectionNoResults />
        ) : (
          <EmptyState
            icon={DumbbellIcon}
            title={t("org.session.practices.emptyTitle")}
            description={
              canCreate
                ? t("org.session.practices.emptyHint")
                : t("org.session.practices.emptyHintMember")
            }
          >
            {canCreate && (
              <Button onClick={openCreate}>
                <PlusIcon className="size-4" />
                {t("org.session.practices.newPractice")}
              </Button>
            )}
          </EmptyState>
        )
      ) : (
        <>
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {practices.map((p, i) => (
              <PracticeCard
                key={p.id}
                practice={p}
                index={(list.page - 1) * pageSize + i}
                canSubmit={canSubmit}
                onEdit={(practice) => {
                  setEditingPractice(practice)
                  setFormOpen(true)
                }}
                onDelete={(practice) => {
                  setDeletingPractice(practice)
                  setDeleteOpen(true)
                }}
                onSubmit={(practice) => {
                  setSubmitPractice(practice)
                  setSubmitOpen(true)
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

      <PracticeFormDialog
        open={formOpen}
        onOpenChange={(open) => {
          setFormOpen(open)
          if (!open) setEditingPractice(null)
        }}
        practice={editingPractice}
        classSessionId={classSessionId}
      />

      <PracticeSubmitDialog
        open={submitOpen}
        onOpenChange={(open) => {
          setSubmitOpen(open)
          if (!open) setSubmitPractice(null)
        }}
        practice={submitPractice}
      />

      <DeleteConfirmDialog
        open={deleteOpen}
        onOpenChange={(open) => {
          if (deleteMutation.isPending) return
          setDeleteOpen(open)
          if (!open) setDeletingPractice(null)
        }}
        resourceName={deletingPractice?.title ?? ""}
        onConfirm={() => {
          if (deletingPractice?.id) deleteMutation.mutate({ id: deletingPractice.id })
        }}
        isLoading={deleteMutation.isPending}
      />
    </section>
  )
}
