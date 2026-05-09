import type { GithubCom4H1RZooraInternalDomainPermission as Permission } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { useFormatDate } from "@/lib/data-table"
import { usePermissionLabel } from "@/lib/permissions"

export function usePermissionColumns(): ColumnDef<Permission>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()
  const permissionLabel = usePermissionLabel()

  return [
    {
      accessorKey: "name",
      header: t("admin.permissions.name"),
      cell: ({ row }) => (
        <Badge variant="secondary" className="font-mono text-xs font-normal">
          {row.original.name ? permissionLabel(row.original.name) : ""}
        </Badge>
      ),
      enableSorting: true,
      enableHiding: false,
    },
    {
      id: "resource",
      header: t("admin.permissions.resource"),
      accessorFn: (row) => row.name?.split(":")[0] ?? "",
      cell: ({ getValue }) => (
        <span className="text-muted-foreground text-sm">
          {t(`permissions.resources.${getValue<string>()}`, { defaultValue: getValue<string>() })}
        </span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      id: "action",
      header: t("admin.permissions.action"),
      accessorFn: (row) => row.name?.split(":")[1] ?? "",
      cell: ({ getValue }) => (
        <span className="text-muted-foreground text-sm">
          {t(`permissions.actions.${getValue<string>()}`, { defaultValue: getValue<string>() })}
        </span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "created_at",
      header: t("admin.permissions.createdAt"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>,
      enableSorting: true,
    },
    {
      accessorKey: "updated_at",
      header: t("admin.permissions.updatedAt"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{formatDate(row.original.updated_at)}</span>,
      enableSorting: true,
    },
  ]
}
