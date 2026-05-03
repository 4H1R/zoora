import { useGetRoles } from "@/api/roles/roles"

export function useRolesMap(organizationId?: string) {
  const query = useGetRoles({ organization_id: organizationId })

  const rolesMap: Record<string, string> = {}
  const items = query.data?.status === 200 && query.data?.data?.data

  if (items) {
    for (const role of items) {
      if (role.id && role.name) rolesMap[role.id] = role.name
    }
  }

  return { ...query, rolesMap }
}
