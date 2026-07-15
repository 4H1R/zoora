import type { GithubCom4H1RZooraInternalDomainClass as Class } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { Link } from "@tanstack/react-router"
import { CalendarClockIcon, EllipsisVerticalIcon, PencilIcon, Trash2Icon, UserPlusIcon, UsersIcon } from "lucide-react"
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

interface ClassRowActionsProps {
  cls: Class
  onEdit: (cls: Class) => void
  onDelete: (cls: Class) => void
}

function ClassRowActions({ cls, onEdit, onDelete }: ClassRowActionsProps) {
  const { t } = useTranslation()

  return (
    <div className="flex items-center justify-end gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
      {cls.id && (
        <Link
          to="/admin/classes/$classId/members"
          params={{ classId: cls.id }}
          aria-label={t("admin.classes.actions.manageMembers")}
        >
          <Button variant="ghost" size="icon-xs">
            <UserPlusIcon />
          </Button>
        </Link>
      )}
      {cls.id && (
        <Link
          to="/admin/classes/$classId/sessions"
          params={{ classId: cls.id }}
          aria-label={t("admin.classes.actions.viewSessions")}
        >
          <Button variant="ghost" size="icon-xs">
            <CalendarClockIcon />
          </Button>
        </Link>
      )}
      <Button variant="ghost" size="icon-xs" onClick={() => onEdit(cls)}>
        <PencilIcon />
      </Button>
      <Button
        variant="ghost"
        size="icon-xs"
        className="text-destructive hover:bg-destructive/10 hover:text-destructive"
        onClick={() => onDelete(cls)}
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
            {cls.id && (
              <DropdownMenuItem render={<Link to="/admin/classes/$classId/members" params={{ classId: cls.id }} />}>
                <UserPlusIcon data-icon="inline-start" />
                {t("admin.classes.actions.manageMembers")}
              </DropdownMenuItem>
            )}
            {cls.id && (
              <DropdownMenuItem render={<Link to="/admin/classes/$classId/sessions" params={{ classId: cls.id }} />}>
                <CalendarClockIcon data-icon="inline-start" />
                {t("admin.classes.actions.viewSessions")}
              </DropdownMenuItem>
            )}
            <DropdownMenuItem onClick={() => onEdit(cls)}>
              <PencilIcon data-icon="inline-start" />
              {t("admin.classes.actions.edit")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuGroup>
            <DropdownMenuItem variant="destructive" onClick={() => onDelete(cls)}>
              <Trash2Icon data-icon="inline-start" />
              {t("admin.classes.actions.delete")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}

interface UseClassColumnsOptions {
  onEdit: (cls: Class) => void
  onDelete: (cls: Class) => void
}

export function useClassColumns({ onEdit, onDelete }: UseClassColumnsOptions): ColumnDef<Class>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()
  const orgColumn = useOrgColumn<Class>(t("admin.classes.organization"))

  return [
    {
      accessorKey: "name",
      header: t("admin.classes.name"),
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
            {row.original.description && (
              <div className="text-muted-foreground truncate text-xs">{row.original.description}</div>
            )}
          </div>
        </div>
      ),
      enableSorting: true,
      enableHiding: false,
    },
    {
      accessorKey: "user",
      header: t("admin.classes.teacher"),
      cell: ({ row }) => (
        <span className="text-sm">{row.original.user?.name ?? <span className="text-muted-foreground">—</span>}</span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    orgColumn,
    {
      accessorKey: "total_users",
      header: t("admin.classes.capacity"),
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <UsersIcon className="text-muted-foreground size-3.5" />
          <span className="text-sm tabular-nums">
            {row.original.total_users === 0 ? t("admin.classes.unlimited") : (row.original.total_users ?? "—")}
          </span>
        </div>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "created_at",
      header: t("admin.classes.createdAt"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>,
      enableSorting: true,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => <ClassRowActions cls={row.original} onEdit={onEdit} onDelete={onDelete} />,
      enableSorting: false,
      enableHiding: false,
    },
  ]
}
