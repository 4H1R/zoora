import type { GithubCom4H1RZooraInternalDomainClassSession as Session } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { DoorOpenIcon, EllipsisVerticalIcon, PencilIcon, Trash2Icon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetClassesIdSessionsQueryKey,
  useDeleteClassesSessionsSessionId,
} from "@/api/classes/classes"
import { SessionRoomsDialog } from "@/components/admin/sessions/SessionRoomsDialog"
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

interface SessionActionsProps {
  session: Session
  classId: string
  onEdit: (session: Session) => void
}

export function SessionActions({ session, classId, onEdit }: SessionActionsProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [roomsOpen, setRoomsOpen] = useState(false)

  const deleteMutation = useDeleteClassesSessionsSessionId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.sessions.form.deleteSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetClassesIdSessionsQueryKey(classId) })
        setDeleteOpen(false)
      },
    },
  })

  const handleConfirmDelete = () => {
    if (session.id) deleteMutation.mutate({ sessionId: session.id })
  }

  return (
    <div className="flex items-center justify-end gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
      <Button
        variant="ghost"
        size="sm"
        onClick={() => setRoomsOpen(true)}
        title={t("admin.sessions.actions.manageRooms")}
      >
        <DoorOpenIcon data-icon="inline-start" />
        <span className="hidden md:inline">{t("admin.sessions.actions.manageRooms")}</span>
      </Button>
      <Button variant="ghost" size="icon-xs" onClick={() => onEdit(session)}>
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
            <DropdownMenuItem onClick={() => setRoomsOpen(true)}>
              <DoorOpenIcon data-icon="inline-start" />
              {t("admin.sessions.actions.manageRooms")}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => onEdit(session)}>
              <PencilIcon data-icon="inline-start" />
              {t("admin.sessions.actions.edit")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuGroup>
            <DropdownMenuItem variant="destructive" onClick={() => setDeleteOpen(true)}>
              <Trash2Icon data-icon="inline-start" />
              {t("admin.sessions.actions.delete")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>

      <SessionRoomsDialog
        open={roomsOpen}
        onOpenChange={setRoomsOpen}
        session={session}
        classId={classId}
      />

      <DeleteConfirmDialog
        open={deleteOpen}
        onOpenChange={(open) => {
          if (!deleteMutation.isPending) setDeleteOpen(open)
        }}
        resourceName={session.name ?? ""}
        onConfirm={handleConfirmDelete}
        isLoading={deleteMutation.isPending}
      />
    </div>
  )
}
