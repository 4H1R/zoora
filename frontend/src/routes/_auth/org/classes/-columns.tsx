import type { GithubCom4H1RZooraInternalDomainClass } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { Link } from "@tanstack/react-router"
import { ChevronRight, UserIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { getGradient } from "@/components/class-card"
import { getInitials } from "@/components/user-avatar"
import { Button } from "@/components/ui/button"
import { UserAvatar } from "@/components/user-avatar"
import { cn } from "@/lib/utils"

export function useClassColumns(): ColumnDef<GithubCom4H1RZooraInternalDomainClass>[] {
  const { t } = useTranslation()

  return [
    {
      accessorKey: "name",
      header: t("classesPage.table.class"),
      cell: ({ row }) => {
        const cls = row.original
        const gradient = getGradient(cls.id ?? "")
        return (
          <div className="flex items-center gap-3">
            <div
              className={cn(
                "flex size-9 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br text-sm font-bold text-white",
                gradient
              )}
            >
              {getInitials(cls.name)}
            </div>
            <div className="min-w-0">
              <Link
                to="/org/classes/$classId"
                params={{ classId: cls.id! }}
                className="hover:text-primary truncate text-sm font-medium transition-colors"
              >
                {cls.name}
              </Link>
            </div>
          </div>
        )
      },
    },
    {
      accessorKey: "user",
      header: t("classesPage.table.instructor"),
      cell: ({ row }) => {
        const cls = row.original
        const teacherName = cls.user?.name ?? ""
        if (!cls.user_id) return <span className="text-muted-foreground">—</span>
        return (
          <div className="flex items-center gap-2">
            {teacherName ? (
              <UserAvatar name={teacherName} size="md" />
            ) : (
              <UserIcon className="text-muted-foreground size-4" />
            )}
            <span className="text-sm">{teacherName}</span>
          </div>
        )
      },
      enableSorting: false,
    },
    {
      accessorKey: "total_users",
      header: t("classesPage.table.capacity"),
      cell: ({ row }) => {
        const capacity = row.original.total_users ?? 0
        return capacity > 0 ? (
          <span className="text-sm tabular-nums">{capacity}</span>
        ) : (
          <span className="text-muted-foreground">—</span>
        )
      },
      enableSorting: false,
    },
    {
      id: "actions",
      header: "",
      enableSorting: false,
      enableHiding: false,
      cell: ({ row }) => {
        const cls = row.original
        return (
          <div className="text-end">
            <Button
              variant="outline"
              size="xs"
              render={<Link to="/org/classes/$classId" params={{ classId: cls.id! }} />}
            >
              {t("common.continue")}
              <ChevronRight className="size-3.5 rtl:rotate-180" />
            </Button>
          </div>
        )
      },
    },
  ]
}
