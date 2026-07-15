import type { GithubCom4H1RZooraInternalDomainQuiz as Quiz } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { EllipsisVerticalIcon, ListChecksIcon, PencilIcon, Trash2Icon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getGetAdminQuizzesQueryKey } from "@/api/admin-quizzes/admin-quizzes"
import { getGetQuizzesQueryKey, useDeleteQuizzesId } from "@/api/quizzes/quizzes"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

interface QuizActionsProps {
  quiz: Quiz
  onEdit: (q: Quiz) => void
  onManageQuestions: (q: Quiz) => void
}

export function QuizActions({ quiz, onEdit, onManageQuestions }: QuizActionsProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [deleteOpen, setDeleteOpen] = useState(false)

  const deleteMutation = useDeleteQuizzesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.quizzes.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetQuizzesQueryKey() })
        queryClient.invalidateQueries({ queryKey: getGetAdminQuizzesQueryKey() })
        setDeleteOpen(false)
      },
    },
  })

  const handleConfirmDelete = () => {
    if (quiz.id) deleteMutation.mutate({ id: quiz.id })
  }

  return (
    <div className="flex items-center justify-end gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
      <Button variant="ghost" size="icon-xs" onClick={() => onManageQuestions(quiz)}>
        <ListChecksIcon />
      </Button>
      <Button variant="ghost" size="icon-xs" onClick={() => onEdit(quiz)}>
        <PencilIcon />
      </Button>
      <Button
        variant="ghost"
        size="icon-xs"
        className="text-destructive hover:bg-destructive/10 hover:text-destructive"
        onClick={() => setDeleteOpen(true)}
      >
        <Trash2Icon />
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button variant="ghost" size="icon-xs">
              <EllipsisVerticalIcon />
            </Button>
          }
        />
        <DropdownMenuContent align="end" className="min-w-44">
          <DropdownMenuGroup>
            <DropdownMenuItem onClick={() => onManageQuestions(quiz)}>
              <ListChecksIcon data-icon="inline-start" />
              {t("admin.quizzes.actions.manageQuestions")}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => onEdit(quiz)}>
              <PencilIcon data-icon="inline-start" />
              {t("admin.quizzes.actions.edit")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuGroup>
            <DropdownMenuItem variant="destructive" onClick={() => setDeleteOpen(true)}>
              <Trash2Icon data-icon="inline-start" />
              {t("admin.quizzes.actions.delete")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>

      <DeleteConfirmDialog
        open={deleteOpen}
        onOpenChange={(open) => {
          if (!deleteMutation.isPending) setDeleteOpen(open)
        }}
        resourceName={quiz.title ?? ""}
        onConfirm={handleConfirmDelete}
        isLoading={deleteMutation.isPending}
      />
    </div>
  )
}
