import { useGetRoles } from "@/api/roles/roles"
import { useRoleName } from "@/lib/permissions"

export function useRolesMap() {
  const query = useGetRoles()
  const roleName = useRoleName()

  const rolesMap: Record<string, string> = {}
  const items = query.data?.status === 200 && query.data?.data?.data

  if (items) {
    for (const role of items) {
      if (role.id && role.name) rolesMap[role.id] = roleName(role.name)
    }
  }

  return { ...query, rolesMap }
}
