import type { GithubCom4H1RZooraInternalDomainOrganization as Organization } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { useGetAdminOrganizations } from "@/api/admin-organizations/admin-organizations"

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
      if (orgName) return <span className="text-sm">{orgName}</span>
      if (orgId) return <span className="text-muted-foreground font-mono text-xs">{orgId.slice(0, 8)}…</span>
      return <span className="text-muted-foreground text-xs">—</span>
    },
    enableSorting: false,
    enableHiding: true,
  }
}
