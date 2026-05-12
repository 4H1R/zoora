import type { GithubCom4H1RZooraInternalDomainLiveRoom as LiveRoom } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { EllipsisVerticalIcon, SquareIcon, Trash2Icon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetAdminLiveRoomsQueryKey,
  useDeleteAdminLiveRoomsId,
  usePostAdminLiveRoomsIdEnd,
} from "@/api/admin-livesessions/admin-livesessions"
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

interface LiveRoomActionsProps {
  room: LiveRoom
  onEnded?: () => void
  onDeleted?: () => void
}

export function LiveRoomActions({ room, onEnded, onDeleted }: LiveRoomActionsProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [deleteOpen, setDeleteOpen] = useState(false)

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminLiveRoomsQueryKey() })
  }

  const endMutation = usePostAdminLiveRoomsIdEnd({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.liveRooms.actions.endSuccess"))
        invalidate()
        onEnded?.()
      },
    },
  })

  const deleteMutation = useDeleteAdminLiveRoomsId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.liveRooms.actions.deleteSuccess"))
        invalidate()
        onDeleted?.()
        setDeleteOpen(false)
      },
    },
  })

  const isActive = room.status === "active"
  const id = room.id

  const handleEnd = () => {
    if (id) endMutation.mutate({ id })
  }

  const handleConfirmDelete = () => {
    if (id) deleteMutation.mutate({ id })
  }

  return (
    <div className="flex items-center justify-end gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
      {isActive && (
        <Button
          variant="ghost"
          size="icon-xs"
          onClick={handleEnd}
          disabled={endMutation.isPending}
          aria-label={t("admin.liveRooms.actions.end")}
        >
          <SquareIcon />
        </Button>
      )}
      <Button
        variant="ghost"
        size="icon-xs"
        className="text-destructive hover:bg-destructive/10 hover:text-destructive"
        onClick={() => setDeleteOpen(true)}
        aria-label={t("admin.liveRooms.actions.delete")}
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
          {isActive && (
            <DropdownMenuGroup>
              <DropdownMenuItem onClick={handleEnd}>
                <SquareIcon data-icon="inline-start" />
                {t("admin.liveRooms.actions.end")}
              </DropdownMenuItem>
            </DropdownMenuGroup>
          )}
          {isActive && <DropdownMenuSeparator />}
          <DropdownMenuGroup>
            <DropdownMenuItem variant="destructive" onClick={() => setDeleteOpen(true)}>
              <Trash2Icon data-icon="inline-start" />
              {t("admin.liveRooms.actions.delete")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>

      <DeleteConfirmDialog
        open={deleteOpen}
        onOpenChange={(open) => {
          if (!deleteMutation.isPending) setDeleteOpen(open)
        }}
        resourceName={room.class_session?.name ?? room.livekit_room_name ?? ""}
        onConfirm={handleConfirmDelete}
        isLoading={deleteMutation.isPending}
      />
    </div>
  )
}
