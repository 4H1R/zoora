import { useGetRoles } from "@/api/roles/roles"

export function useRolesMap() {
  const query = useGetRoles()

  const rolesMap: Record<string, string> = {}
  const items = query.data?.status === 200 && query.data?.data?.data

  if (items) {
    for (const role of items) {
      if (role.id && role.name) rolesMap[role.id] = role.name
    }
  }

  return { ...query, rolesMap }
}
