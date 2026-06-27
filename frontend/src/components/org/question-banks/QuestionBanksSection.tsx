import type { GithubCom4H1RZooraInternalDomainQuestionBank as Bank } from "@/api/model"
import type { SortOption } from "@/components/data-table/sort-picker"

import { useQueryClient } from "@tanstack/react-query"
import {
  CalendarClockIcon,
  LibraryIcon,
  ListChecksIcon,
  PencilIcon,
  PlusIcon,
  Trash2Icon,
} from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetQuestionBanksQueryKey,
  useDeleteQuestionBanksId,
  useGetQuestionBanks,
} from "@/api/question-banks/question-banks"
import { SectionNoResults } from "@/components/org/session/section-no-results"
import { SectionPagination } from "@/components/org/session/section-pagination"
import { SectionToolbar } from "@/components/org/session/section-toolbar"
import { Eyebrow } from "@/components/eyebrow"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"
import { DEFAULT_PAGE_SIZE } from "@/lib/list"
import { formatSessionDate } from "@/lib/session-status"
import { useSectionList } from "@/lib/use-section-list"

import { QuestionBankFormDialog } from "./QuestionBankFormDialog"
import { QuestionBankQuestionsDialog } from "./QuestionBankQuestionsDialog"
import { useBankPermissions } from "./use-bank-permissions"

interface BankCardProps {
  bank: Bank
  index: number
  canEdit: boolean
  canDelete: boolean
  onEdit: (b: Bank) => void
  onManage: (b: Bank) => void
  onDelete: (b: Bank) => void
}

function BankCard({ bank, index, canEdit, canDelete, onEdit, onManage, onDelete }: BankCardProps) {
  const { t, i18n } = useTranslation()
  const tileNumber = String(index + 1).padStart(2, "0")
  const createdStr = bank.created_at
    ? formatSessionDate(bank.created_at, i18n.language, "short")
    : "—"

  return (
    <div className="group/bank bg-card text-card-foreground ring-foreground/10 hover:ring-foreground/30 relative isolate flex flex-col gap-5 overflow-hidden rounded-2xl p-5 ring-1 transition-all">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/8%,transparent_60%)] opacity-0 transition-opacity group-hover/bank:opacity-100"
      />
      <div className="flex items-start justify-between gap-3">
        <div className="bg-muted text-foreground flex size-10 items-center justify-center rounded-xl">
          <LibraryIcon className="size-5" />
        </div>
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">/{tileNumber}</span>
      </div>

      <div className="flex flex-col gap-2">
        <Eyebrow>{t("org.session.questionBanks.cardEyebrow")}</Eyebrow>
        <h3 className="line-clamp-2 text-xl leading-snug font-semibold tracking-tight text-balance">
          {bank.name ?? "—"}
        </h3>
        {bank.description && (
          <p className="text-muted-foreground line-clamp-2 text-sm leading-relaxed">
            {bank.description}
          </p>
        )}
      </div>

      <div className="border-foreground/10 mt-auto flex items-center justify-between gap-2 border-t border-dashed pt-3">
        <span className="text-muted-foreground inline-flex items-center gap-2 font-mono text-xs">
          <CalendarClockIcon className="size-3.5" />
          {createdStr}
        </span>
        <div className="flex items-center gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover/bank:opacity-100">
          <Button
            variant="ghost"
            size="icon-xs"
            title={t("org.session.questionBanks.actions.manage")}
            onClick={() => onManage(bank)}
          >
            <ListChecksIcon />
          </Button>
          {canEdit && (
            <Button
              variant="ghost"
              size="icon-xs"
              title={t("org.session.questionBanks.actions.edit")}
              onClick={() => onEdit(bank)}
            >
              <PencilIcon />
            </Button>
          )}
          {canDelete && (
            <Button
              variant="ghost"
              size="icon-xs"
              className="text-destructive hover:bg-destructive/10 hover:text-destructive"
              title={t("org.session.questionBanks.actions.delete")}
              onClick={() => onDelete(bank)}
            >
              <Trash2Icon />
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}

function BankCardSkeleton() {
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
      <Skeleton className="h-4 w-1/3" />
    </div>
  )
}

export function QuestionBanksSection() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { canView, canCreate, canEdit, canDelete } = useBankPermissions()

  const list = useSectionList()
  const sortOptions: SortOption[] = [
    { id: "created_at", label: t("org.session.controls.sortFields.created_at") },
    { id: "name", label: t("org.session.controls.sortFields.name") },
  ]

  const banksQuery = useGetQuestionBanks({ ...list.params }, { query: { enabled: canView } })
  const banksData = (banksQuery.data?.status === 200 && banksQuery.data.data.data) || undefined
  const banks: Bank[] = banksData?.items ?? []
  const total = banksData?.total ?? 0
  const pageSize = banksData?.page_size ?? DEFAULT_PAGE_SIZE

  const [formOpen, setFormOpen] = useState(false)
  const [editingBank, setEditingBank] = useState<Bank | null>(null)
  const [questionsOpen, setQuestionsOpen] = useState(false)
  const [managingBank, setManagingBank] = useState<Bank | null>(null)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deletingBank, setDeletingBank] = useState<Bank | null>(null)

  const deleteMutation = useDeleteQuestionBanksId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.questionBanks.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetQuestionBanksQueryKey() })
        setDeleteOpen(false)
        setDeletingBank(null)
      },
    },
  })

  const openCreate = () => {
    setEditingBank(null)
    setFormOpen(true)
  }

  if (!canView) return null

  return (
    <section id="question-banks" className="flex flex-col gap-5 scroll-mt-20">
      <div className="flex items-end justify-between gap-4">
        <div className="flex flex-col gap-1.5">
          <Eyebrow>{t("org.session.questionBanks.eyebrow")}</Eyebrow>
          <h2 className="text-2xl font-semibold tracking-tight">
            {t("org.session.questionBanks.title")}
          </h2>
        </div>
        {canCreate && (
          <Button onClick={openCreate}>
            <PlusIcon className="size-4" />
            {t("org.session.questionBanks.newBank")}
          </Button>
        )}
      </div>

      {(banks.length > 0 || list.isFiltered) && (
        <SectionToolbar
          searchValue={list.searchInput}
          onSearchChange={list.setSearchInput}
          sortOptions={sortOptions}
          sort={list.sort}
          onSortChange={list.setSort}
        />
      )}

      {banksQuery.isPending ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <BankCardSkeleton />
          <BankCardSkeleton />
          <BankCardSkeleton />
        </div>
      ) : banks.length === 0 ? (
        list.isFiltered ? (
          <SectionNoResults />
        ) : (
          <EmptyState
            icon={LibraryIcon}
            title={t("org.session.questionBanks.emptyTitle")}
            description={t("org.session.questionBanks.emptyHint")}
          >
            {canCreate && (
              <Button onClick={openCreate}>
                <PlusIcon className="size-4" />
                {t("org.session.questionBanks.newBank")}
              </Button>
            )}
          </EmptyState>
        )
      ) : (
        <>
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {banks.map((b, i) => (
              <BankCard
                key={b.id}
                bank={b}
                index={(list.page - 1) * pageSize + i}
                canEdit={canEdit}
                canDelete={canDelete}
                onEdit={(bank) => {
                  setEditingBank(bank)
                  setFormOpen(true)
                }}
                onManage={(bank) => {
                  setManagingBank(bank)
                  setQuestionsOpen(true)
                }}
                onDelete={(bank) => {
                  setDeletingBank(bank)
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

      <QuestionBankFormDialog
        open={formOpen}
        onOpenChange={(open) => {
          setFormOpen(open)
          if (!open) setEditingBank(null)
        }}
        bank={editingBank}
      />

      <QuestionBankQuestionsDialog
        open={questionsOpen}
        onOpenChange={(open) => {
          setQuestionsOpen(open)
          if (!open) setManagingBank(null)
        }}
        bank={managingBank}
      />

      <DeleteConfirmDialog
        open={deleteOpen}
        onOpenChange={(open) => {
          if (deleteMutation.isPending) return
          setDeleteOpen(open)
          if (!open) setDeletingBank(null)
        }}
        resourceName={deletingBank?.name ?? ""}
        onConfirm={() => {
          if (deletingBank?.id) deleteMutation.mutate({ id: deletingBank.id })
        }}
        isLoading={deleteMutation.isPending}
      />
    </section>
  )
}
