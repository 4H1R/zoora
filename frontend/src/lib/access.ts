import type {
  GithubCom4H1RZooraInternalDomainPermissionName,
  GithubCom4H1RZooraInternalDomainRole as Role,
  GithubCom4H1RZooraInternalDomainUser as User,
} from "@/api/model"
import type { UserContext } from "react-access-engine"

import { defineAccess, useAccess } from "react-access-engine"

export type AppPermission = GithubCom4H1RZooraInternalDomainPermissionName

export function buildAccess(user: User, allRoles: Role[]) {
  const isAdmin = !!user.is_admin

  const userRole = allRoles.find((r) => r.id === user.role_id)
  const roleName = userRole?.name

  const userPermissions: AppPermission[] = (userRole?.permissions ?? [])
    .map((p) => p.name)
    .filter((name): name is AppPermission => !!name)

  const roles: string[] = []
  if (isAdmin) roles.push("admin")
  if (roleName) roles.push(roleName)

  const permissions: Record<string, string[]> = {}
  if (isAdmin) permissions["admin"] = ["*"]
  if (roleName) permissions[roleName] = userPermissions

  const config = defineAccess({
    roles,
    permissions,
  })
  const accessUser: UserContext<string, string> = { id: user.id ?? "", roles }

  return { config, user: accessUser }
}

export function useCanSelfOr(basePerm: AppPermission, anyPerm: AppPermission, targetId: string | undefined) {
  const { can, user } = useAccess()
  if (can(anyPerm)) return true
  if (can(basePerm) && targetId === user.id) return true
  return false
}
