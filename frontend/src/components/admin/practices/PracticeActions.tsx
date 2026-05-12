import type { GithubCom4H1RZooraInternalDomainPracticeRoom as PracticeRoom } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { EllipsisVerticalIcon, PencilIcon, Trash2Icon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetAdminPracticesQueryKey,
  useDeleteAdminPracticesId,
} from "@/api/admin-practices/admin-practices"
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

interface PracticeActionsProps {
  practice: PracticeRoom
  onEdit: (practice: PracticeRoom) => void
}

export function PracticeActions({ practice, onEdit }: PracticeActionsProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [deleteOpen, setDeleteOpen] = useState(false)

  const deleteMutation = useDeleteAdminPracticesId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.practices.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetAdminPracticesQueryKey() })
        setDeleteOpen(false)
      },
    },
  })

  const handleConfirmDelete = () => {
    if (practice.id) deleteMutation.mutate({ id: practice.id })
  }

  return (
    <div className="flex items-center justify-end gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
      <Button variant="ghost" size="icon-xs" onClick={() => onEdit(practice)}>
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
            <DropdownMenuItem onClick={() => onEdit(practice)}>
              <PencilIcon data-icon="inline-start" />
              {t("admin.practices.actions.edit")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuGroup>
            <DropdownMenuItem variant="destructive" onClick={() => setDeleteOpen(true)}>
              <Trash2Icon data-icon="inline-start" />
              {t("admin.practices.actions.delete")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>

      <DeleteConfirmDialog
        open={deleteOpen}
        onOpenChange={(open) => {
          if (!deleteMutation.isPending) setDeleteOpen(open)
        }}
        resourceName={practice.title ?? ""}
        onConfirm={handleConfirmDelete}
        isLoading={deleteMutation.isPending}
      />
    </div>
  )
}
