import type { GithubCom4H1RZooraInternalDomainRole as Role } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { EllipsisVerticalIcon, PencilIcon, Trash2Icon } from "lucide-react"
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
import { getEntityColor, getInitials, useFormatDate } from "@/lib/data-table"
import { cn } from "@/lib/utils"

interface RoleRowActionsProps {
  role: Role
  onEdit: (role: Role) => void
  onDelete: (role: Role) => void
}

function RoleRowActions({ role, onEdit, onDelete }: RoleRowActionsProps) {
  const { t } = useTranslation()

  if (role.is_preset) return null

  return (
    <div className="flex items-center justify-end gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
      <Button variant="ghost" size="icon-xs" onClick={() => onEdit(role)}>
        <PencilIcon />
      </Button>
      <Button
        variant="ghost"
        size="icon-xs"
        className="text-destructive hover:bg-destructive/10 hover:text-destructive"
        onClick={() => onDelete(role)}
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
            <DropdownMenuItem onClick={() => onEdit(role)}>
              <PencilIcon data-icon="inline-start" />
              {t("org.roles.actions.edit")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuGroup>
            <DropdownMenuItem variant="destructive" onClick={() => onDelete(role)}>
              <Trash2Icon data-icon="inline-start" />
              {t("org.roles.actions.delete")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}

interface UseRoleColumnsOptions {
  onEdit: (role: Role) => void
  onDelete: (role: Role) => void
}

export function useRoleColumns({ onEdit, onDelete }: UseRoleColumnsOptions): ColumnDef<Role>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  return [
    {
      accessorKey: "name",
      header: t("org.roles.name"),
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
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: "permissions",
      header: t("org.roles.permissions"),
      cell: ({ row }) => {
        const count = row.original.permissions?.length ?? 0
        return (
          <Badge variant="secondary" className="text-[11px]">
            {t("org.roles.permissionsCount", { count })}
          </Badge>
        )
      },
      enableSorting: false,
    },
    {
      accessorKey: "created_at",
      header: t("org.roles.createdAt"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>,
      enableSorting: false,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => <RoleRowActions role={row.original} onEdit={onEdit} onDelete={onDelete} />,
      enableSorting: false,
      enableHiding: false,
    },
  ]
}
