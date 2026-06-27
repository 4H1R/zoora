import type {
  GithubCom4H1RZooraInternalDomainQuestion as Question,
  GithubCom4H1RZooraInternalDomainQuestionBank as Bank,
} from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { HelpCircleIcon, PencilIcon, PlusIcon, Trash2Icon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetQuestionBanksIdQuestionsQueryKey,
  useDeleteQuestionBanksQuestionsQuestionId,
  useGetQuestionBanksIdQuestions,
} from "@/api/question-banks/question-banks"
import { QuestionCreateModal } from "@/components/admin/questions/QuestionCreateModal"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"

import { useBankPermissions } from "./use-bank-permissions"

interface QuestionBankQuestionsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  bank: Bank | null
}

export function QuestionBankQuestionsDialog({
  open,
  onOpenChange,
  bank,
}: QuestionBankQuestionsDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { canEdit, canDelete } = useBankPermissions()

  const bankId = bank?.id ?? ""

  const { data, isLoading } = useGetQuestionBanksIdQuestions(
    bankId,
    {},
    { query: { enabled: open && !!bankId } }
  )
  const questions: Question[] = (data?.status === 200 && data.data.data?.items) || []

  const [formOpen, setFormOpen] = useState(false)
  const [editingQuestion, setEditingQuestion] = useState<Question | null>(null)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deletingQuestion, setDeletingQuestion] = useState<Question | null>(null)

  const invalidate = () => {
    if (bankId) {
      queryClient.invalidateQueries({
        queryKey: getGetQuestionBanksIdQuestionsQueryKey(bankId),
      })
    }
  }

  const deleteMutation = useDeleteQuestionBanksQuestionsQuestionId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.questionBanks.questions.deleteSuccess"))
        invalidate()
        setDeleteOpen(false)
        setDeletingQuestion(null)
      },
    },
  })

  const openCreate = () => {
    setEditingQuestion(null)
    setFormOpen(true)
  }

  const openEdit = (q: Question) => {
    setEditingQuestion(q)
    setFormOpen(true)
  }

  const handleFormOpenChange = (next: boolean) => {
    setFormOpen(next)
    if (!next) {
      setEditingQuestion(null)
      invalidate()
    }
  }

  return (
    <>
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>
              {bank?.name ?? t("org.session.questionBanks.questions.title")}
              <span className="text-muted-foreground ms-2 text-sm font-normal">
                · {t("org.session.questionBanks.questions.title")}
              </span>
            </DialogTitle>
            <DialogDescription>
              {bank?.description || t("org.session.questionBanks.questions.description")}
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
              <div className="flex items-center gap-2">
                <span className="text-muted-foreground">
                  {t("org.session.questionBanks.questions.count")}
                </span>
                <Badge variant="secondary">{questions.length}</Badge>
              </div>
              {canEdit && (
                <Button size="sm" onClick={openCreate}>
                  <PlusIcon className="size-4" />
                  {t("org.session.questionBanks.questions.add")}
                </Button>
              )}
            </div>

            {!isLoading && questions.length === 0 ? (
              <EmptyState
                icon={HelpCircleIcon}
                title={t("org.session.questionBanks.questions.emptyTitle")}
                description={t("org.session.questionBanks.questions.emptyHint")}
              />
            ) : (
              <ul className="divide-border max-h-[28rem] divide-y overflow-y-auto rounded-md border">
                {isLoading ? (
                  <>
                    <li className="px-3 py-3"><Skeleton className="h-5 w-3/5" /></li>
                    <li className="px-3 py-3"><Skeleton className="h-5 w-2/5" /></li>
                    <li className="px-3 py-3"><Skeleton className="h-5 w-1/2" /></li>
                  </>
                ) : (
                  questions.map((q) => (
                  <li key={q.id} className="group/qrow flex items-start gap-3 px-3 py-3">
                    <div className="min-w-0 flex-1">
                      <div className="line-clamp-2 text-sm leading-snug">{q.text}</div>
                      <div className="mt-1 flex items-center gap-2">
                        <Badge variant="outline" className="text-[10px] uppercase">
                          {t(`admin.questions.types.${q.type ?? "descriptive"}`)}
                        </Badge>
                        <span className="text-muted-foreground font-mono text-[10px]">
                          {(q.options?.length ?? 0)} {t("org.session.questionBanks.questions.options")}
                        </span>
                      </div>
                    </div>
                    <div className="flex items-center gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover/qrow:opacity-100">
                      {canEdit && (
                        <Button
                          variant="ghost"
                          size="icon-xs"
                          onClick={() => openEdit(q)}
                          title={t("common.edit")}
                        >
                          <PencilIcon />
                        </Button>
                      )}
                      {canDelete && (
                        <Button
                          variant="ghost"
                          size="icon-xs"
                          className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                          title={t("common.delete")}
                          onClick={() => {
                            setDeletingQuestion(q)
                            setDeleteOpen(true)
                          }}
                        >
                          <Trash2Icon />
                        </Button>
                      )}
                    </div>
                  </li>
                ))
                )}
              </ul>
            )}
          </div>
        </DialogContent>
      </Dialog>

      <QuestionCreateModal
        open={formOpen}
        onOpenChange={handleFormOpenChange}
        question={editingQuestion}
        defaultBankId={bankId || undefined}
      />

      <DeleteConfirmDialog
        open={deleteOpen}
        onOpenChange={(next) => {
          if (deleteMutation.isPending) return
          setDeleteOpen(next)
          if (!next) setDeletingQuestion(null)
        }}
        resourceName={deletingQuestion?.text?.slice(0, 60) ?? ""}
        onConfirm={() => {
          if (deletingQuestion?.id) {
            deleteMutation.mutate({ questionId: deletingQuestion.id })
          }
        }}
        isLoading={deleteMutation.isPending}
      />
    </>
  )
}
