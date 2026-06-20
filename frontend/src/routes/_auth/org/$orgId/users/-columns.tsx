import type { GithubCom4H1RZooraInternalDomainUser as User } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { BanIcon, CircleCheckIcon, EllipsisVerticalIcon, PencilIcon, Trash2Icon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { useCanSelfOr } from "@/lib/access"
import { getEntityColor, getInitials, useFormatDate } from "@/lib/data-table"
import { cn } from "@/lib/utils"

interface UserRowActionsProps {
  user: User
  onEdit: (user: User) => void
  onDelete: (user: User) => void
  onDisable: (user: User) => void
  onEnable: (user: User) => void
}

function UserRowActions({ user, onEdit, onDelete, onDisable, onEnable }: UserRowActionsProps) {
  const { t } = useTranslation()
  const canEdit = useCanSelfOr("users:update", "users:update_any", user.id)
  const canDelete = useCanSelfOr("users:delete", "users:delete_any", user.id)
  const canDisable = useCanSelfOr("users:disable", "users:disable_any", user.id)
  const isDisabled = !!user.disabled_at

  if (!canEdit && !canDelete && !canDisable) return null

  return (
    <div className="flex items-center justify-end gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
      {canEdit && (
        <Button variant="ghost" size="icon-xs" onClick={() => onEdit(user)}>
          <PencilIcon />
        </Button>
      )}
      {canDelete && (
        <Button
          variant="ghost"
          size="icon-xs"
          className="text-destructive hover:bg-destructive/10 hover:text-destructive"
          onClick={() => onDelete(user)}
        >
          <Trash2Icon />
        </Button>
      )}
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button variant="ghost" size="icon-xs">
              <EllipsisVerticalIcon />
            </Button>
          }
        />
        <DropdownMenuContent align="end" className="min-w-44">
          {canEdit && (
            <DropdownMenuGroup>
              <DropdownMenuItem onClick={() => onEdit(user)}>
                <PencilIcon data-icon="inline-start" />
                {t("org.users.actions.edit")}
              </DropdownMenuItem>
            </DropdownMenuGroup>
          )}
          {canDisable && (canEdit || canDelete) && <DropdownMenuSeparator />}
          {canDisable && (
            <DropdownMenuGroup>
              {isDisabled ? (
                <DropdownMenuItem onClick={() => onEnable(user)}>
                  <CircleCheckIcon data-icon="inline-start" />
                  {t("org.users.actions.enable")}
                </DropdownMenuItem>
              ) : (
                <DropdownMenuItem onClick={() => onDisable(user)}>
                  <BanIcon data-icon="inline-start" />
                  {t("org.users.actions.disable")}
                </DropdownMenuItem>
              )}
            </DropdownMenuGroup>
          )}
          {canEdit && canDelete && <DropdownMenuSeparator />}
          {canDelete && (
            <DropdownMenuGroup>
              <DropdownMenuItem variant="destructive" onClick={() => onDelete(user)}>
                <Trash2Icon data-icon="inline-start" />
                {t("org.users.actions.delete")}
              </DropdownMenuItem>
            </DropdownMenuGroup>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}

interface UseUserColumnsOptions {
  onEdit: (user: User) => void
  onDelete: (user: User) => void
  onDisable: (user: User) => void
  onEnable: (user: User) => void
  rolesMap: Record<string, string>
}

export function useUserColumns({
  onEdit,
  onDelete,
  onDisable,
  onEnable,
  rolesMap,
}: UseUserColumnsOptions): ColumnDef<User>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  return [
    {
      accessorKey: "name",
      header: t("org.users.name"),
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
            <div className="flex items-center gap-2">
              <div className="truncate text-sm font-medium">{row.original.name}</div>
              {row.original.disabled_at && <Badge variant="secondary">{t("org.users.disabled")}</Badge>}
            </div>
          </div>
        </div>
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: "username",
      header: t("org.users.username"),
      cell: ({ row }) => <span className="font-mono text-xs">{row.original.username}</span>,
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "role",
      header: t("org.users.role"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs">{rolesMap[row.original.role_id ?? ""] ?? "—"}</span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "created_at",
      header: t("org.users.createdAt"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>,
      enableSorting: false,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <UserRowActions
          user={row.original}
          onEdit={onEdit}
          onDelete={onDelete}
          onDisable={onDisable}
          onEnable={onEnable}
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
  ]
}
