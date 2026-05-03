import type { GithubCom4H1RZooraInternalDomainUser as User } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { CheckIcon, EllipsisVerticalIcon, MinusIcon, PencilIcon, Trash2Icon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useOrgColumn } from "@/components/data-table/org-column"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { getEntityColor, getInitials, useFormatDate } from "@/lib/data-table"
import { cn } from "@/lib/utils"

interface UserRowActionsProps {
  user: User
  onEdit: (user: User) => void
  onDelete: (user: User) => void
}

function UserRowActions({ user, onEdit, onDelete }: UserRowActionsProps) {
  const { t } = useTranslation()

  return (
    <div className="flex items-center justify-end gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
      <Button variant="ghost" size="icon-xs" onClick={() => onEdit(user)}>
        <PencilIcon />
      </Button>
      <Button
        variant="ghost"
        size="icon-xs"
        className="text-destructive hover:bg-destructive/10 hover:text-destructive"
        onClick={() => onDelete(user)}
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
            <DropdownMenuItem onClick={() => onEdit(user)}>
              <PencilIcon data-icon="inline-start" />
              {t("admin.users.actions.edit")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuGroup>
            <DropdownMenuItem variant="destructive" onClick={() => onDelete(user)}>
              <Trash2Icon data-icon="inline-start" />
              {t("admin.users.actions.delete")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}

interface UseUserColumnsOptions {
  onEdit: (user: User) => void
  onDelete: (user: User) => void
}

export function useUserColumns({ onEdit, onDelete }: UseUserColumnsOptions): ColumnDef<User>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()
  const orgColumn = useOrgColumn<User>(t("admin.users.organization"))

  return [
    {
      accessorKey: "name",
      header: t("admin.users.name"),
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <div
            className={cn(
              "flex size-9 shrink-0 items-center justify-center rounded-lg text-xs font-semibold text-white",
              getEntityColor(row.original.name ?? "")
            )}
          >
            {getInitials(row.original.name ?? "")}
          </div>
          <div className="min-w-0">
            <div className="truncate text-sm font-medium">{row.original.name}</div>
          </div>
        </div>
      ),
      enableSorting: true,
      enableHiding: false,
    },
    {
      accessorKey: "username",
      header: t("admin.users.username"),
      cell: ({ row }) => <span className="font-mono text-xs">{row.original.username}</span>,
      enableSorting: true,
      enableHiding: true,
    },
    orgColumn,
    {
      accessorKey: "is_admin",
      header: t("admin.users.isAdmin"),
      cell: ({ row }) =>
        row.original.is_admin ? (
          <CheckIcon className="text-primary size-4" />
        ) : (
          <MinusIcon className="text-muted-foreground size-4" />
        ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "created_at",
      header: t("admin.users.createdAt"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>,
      enableSorting: true,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => <UserRowActions user={row.original} onEdit={onEdit} onDelete={onDelete} />,
      enableSorting: false,
      enableHiding: false,
    },
  ]
}
