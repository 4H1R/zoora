import type {
  GithubCom4H1RZooraInternalDomainClassMember as ClassMember,
  GithubCom4H1RZooraInternalDomainClassSession as Session,
} from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { Link } from "@tanstack/react-router"
import { CalendarClockIcon, UserMinusIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { SessionStatusPill } from "@/components/session/status-pill"
import { Button } from "@/components/ui/button"
import { UserAvatar } from "@/components/user-avatar"
import { formatSessionDate, getSessionStatus } from "@/lib/session-status"

// Server-driven columns for the class-detail page. Sortable column `id`s map
// 1:1 to the backend's white-listed order_by tokens (sessions: name,
// start_time, created_at — members: name, created_at), so TanStack's sort
// state round-trips straight through the URL into the API call.

export function useSessionColumns(now: number): ColumnDef<Session>[] {
  const { t, i18n } = useTranslation()

  return [
    {
      id: "status",
      header: t("org.class.sessions.table.status"),
      enableSorting: false,
      enableHiding: false,
      cell: ({ row }) => <SessionStatusPill status={getSessionStatus(row.original.start_time, now)} size="sm" />,
    },
    {
      accessorKey: "name",
      header: t("org.class.sessions.table.name"),
      enableHiding: false,
      cell: ({ row }) => (
        <div className="flex min-w-0 flex-col">
          <span className="truncate font-medium">{row.original.name}</span>
          {row.original.description ? (
            <span className="text-muted-foreground truncate text-xs">{row.original.description}</span>
          ) : null}
        </div>
      ),
    },
    {
      accessorKey: "start_time",
      header: t("org.class.sessions.table.startTime"),
      cell: ({ row }) => (
        <span className="text-muted-foreground inline-flex items-center gap-1.5 font-mono text-xs whitespace-nowrap">
          <CalendarClockIcon className="size-3" />
          {formatSessionDate(row.original.start_time, i18n.language, "short")}
        </span>
      ),
    },
    {
      accessorKey: "created_at",
      header: t("org.class.sessions.table.createdAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground font-mono text-xs whitespace-nowrap">
          {row.original.created_at ? formatSessionDate(row.original.created_at, i18n.language, "short") : "—"}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      enableSorting: false,
      enableHiding: false,
      cell: ({ row }) => (
        <div className="flex justify-end">
          <Button
            variant="ghost"
            size="sm"
            render={
              <Link
                to="/org/classes/classsessions/$classSessionId"
                params={{ classSessionId: row.original.id! }}
              />
            }
          >
            {t("org.class.sessions.open")} →
          </Button>
        </div>
      ),
    },
  ]
}

export function useStudentColumns(onRemove?: (member: ClassMember) => void): ColumnDef<ClassMember>[] {
  const { t, i18n } = useTranslation()

  return [
    {
      accessorKey: "name",
      header: t("org.class.students.table.name"),
      enableHiding: false,
      cell: ({ row }) => {
        const name = row.original.user?.name ?? t("org.class.students.unknownName")
        const username = row.original.user?.username
        return (
          <div className="flex min-w-0 items-center gap-2.5">
            <UserAvatar name={name} size="sm" />
            <div className="flex min-w-0 flex-col">
              <span className="truncate font-medium">{name}</span>
              {username ? <span className="text-muted-foreground truncate font-mono text-xs">@{username}</span> : null}
            </div>
          </div>
        )
      },
    },
    {
      accessorKey: "created_at",
      header: t("org.class.students.table.joined"),
      cell: ({ row }) => (
        <span className="text-muted-foreground inline-flex items-center gap-1.5 font-mono text-xs whitespace-nowrap">
          <CalendarClockIcon className="size-3" />
          {row.original.created_at ? formatSessionDate(row.original.created_at, i18n.language, "short") : "—"}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      enableSorting: false,
      enableHiding: false,
      cell: ({ row }) =>
        onRemove ? (
          <div className="flex justify-end">
            <button
              type="button"
              onClick={() => onRemove(row.original)}
              aria-label={t("org.class.students.removeAction")}
              title={t("org.class.students.removeAction")}
              className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive focus-visible:ring-ring inline-flex size-7 items-center justify-center rounded-full transition-all focus-visible:ring-2 focus-visible:outline-none"
            >
              <UserMinusIcon className="size-3.5" />
            </button>
          </div>
        ) : null,
    },
  ]
}
