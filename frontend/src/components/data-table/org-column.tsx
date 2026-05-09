import type { GithubCom4H1RZooraInternalDomainOrganization as Organization } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { useGetAdminOrganizations } from "@/api/admin-organizations/admin-organizations"
import { getInitials } from "@/components/user-avatar"
import { getEntityColor } from "@/lib/data-table"
import { cn } from "@/lib/utils"

export function useOrgColumn<T extends { organization_id?: string }>(header: string): ColumnDef<T> {
  const { data: orgsData } = useGetAdminOrganizations()
  const orgItems = (orgsData?.data?.data as { items?: Organization[] } | undefined)?.items ?? []
  const orgById = new Map(orgItems.map((o) => [o.id, o.name]))

  return {
    accessorKey: "organization_id",
    header,
    cell: ({ row }) => {
      const orgId = (row.original as { organization_id?: string }).organization_id ?? ""
      const orgName = orgById.get(orgId)
      if (orgName)
        return (
          <div className="flex items-center gap-2">
            <div
              className={cn(
                "flex size-6 shrink-0 items-center justify-center rounded-md text-[10px] font-semibold text-white",
                getEntityColor(orgName)
              )}
            >
              {getInitials(orgName)}
            </div>
            <span className="text-sm">{orgName}</span>
          </div>
        )
      if (orgId) return <span className="text-muted-foreground font-mono text-xs">{orgId.slice(0, 8)}…</span>
      return <span className="text-muted-foreground text-xs">—</span>
    },
    enableSorting: false,
    enableHiding: true,
  }
}
