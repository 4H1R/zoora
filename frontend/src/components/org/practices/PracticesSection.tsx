import type { GithubCom4H1RZooraInternalDomainPracticeRoom as PracticeRoom } from "@/api/model"

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
import { Eyebrow } from "@/components/eyebrow"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useCanSelfOr } from "@/lib/access"
import { formatScore } from "@/lib/score"
import { formatSessionDate } from "@/lib/session-status"

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
    <div className="group/practice bg-card text-card-foreground ring-foreground/10 hover:ring-foreground/30 relative isolate flex flex-col gap-5 overflow-hidden rounded-2xl p-5 ring-1 transition-all hover:-translate-y-0.5 hover:shadow-lg">
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
        {practice.content ? (
          <p className="text-muted-foreground line-clamp-2 text-sm leading-relaxed">{practice.content}</p>
        ) : null}
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
          {(canEdit || canDelete) ? (
            <div className="flex items-center gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover/practice:opacity-100">
              {canEdit ? (
                <Button
                  variant="ghost"
                  size="icon-xs"
                  title={t("org.session.practices.actions.edit")}
                  onClick={() => onEdit(practice)}
                >
                  <PencilIcon />
                </Button>
              ) : null}
              {canDelete ? (
                <Button
                  variant="ghost"
                  size="icon-xs"
                  className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                  title={t("org.session.practices.actions.delete")}
                  onClick={() => onDelete(practice)}
                >
                  <Trash2Icon />
                </Button>
              ) : null}
            </div>
          ) : null}
          {canSubmit ? (
            <Button size="sm" onClick={() => onSubmit(practice)}>
              <SendIcon className="size-3.5" />
              {t("org.session.practices.actions.submit")}
            </Button>
          ) : null}
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

function EmptyState({ canCreate, onCreate }: { canCreate: boolean; onCreate: () => void }) {
  const { t } = useTranslation()
  return (
    <div className="bg-card ring-foreground/10 flex flex-col items-center gap-3 rounded-2xl px-6 py-16 text-center ring-1">
      <DumbbellIcon className="text-muted-foreground size-8" />
      <h3 className="text-foreground text-lg font-semibold tracking-tight">
        {t("org.session.practices.emptyTitle")}
      </h3>
      <p className="text-muted-foreground max-w-md text-sm leading-relaxed">
        {canCreate ? t("org.session.practices.emptyHint") : t("org.session.practices.emptyHintMember")}
      </p>
      {canCreate ? (
        <Button className="mt-2" onClick={onCreate}>
          <PlusIcon className="size-4" />
          {t("org.session.practices.newPractice")}
        </Button>
      ) : null}
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

  const practicesQuery = useGetPractices(
    { class_session_id: classSessionId },
    { query: { enabled: canView } }
  )
  const practices =
    (practicesQuery.data?.status === 200 && practicesQuery.data.data.data?.items) || []

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
          <Eyebrow>{t("org.session.practices.eyebrow")}</Eyebrow>
          <h2 className="text-2xl font-semibold tracking-tight">{t("org.session.practices.title")}</h2>
        </div>
        {canCreate ? (
          <Button onClick={openCreate}>
            <PlusIcon className="size-4" />
            {t("org.session.practices.newPractice")}
          </Button>
        ) : null}
      </div>

      {practicesQuery.isPending ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <PracticeCardSkeleton />
          <PracticeCardSkeleton />
          <PracticeCardSkeleton />
        </div>
      ) : practices.length === 0 ? (
        <EmptyState canCreate={canCreate} onCreate={openCreate} />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {practices.map((p, i) => (
            <PracticeCard
              key={p.id}
              practice={p}
              index={i}
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
