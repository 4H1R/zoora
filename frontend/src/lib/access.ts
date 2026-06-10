import type {
  GithubCom4H1RZooraInternalDomainPermissionName,
  GithubCom4H1RZooraInternalDomainUser as User,
} from "@/api/model"
import type { UserContext } from "react-access-engine"

import { useNavigate, useParams } from "@tanstack/react-router"
import { useEffect } from "react"
import { defineAccess, useAccess } from "react-access-engine"

export type AppPermission = GithubCom4H1RZooraInternalDomainPermissionName

export function buildAccess(user: User) {
  const isAdmin = !!user.is_admin

  const userRole = user.role
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

  const permSet = new Set<AppPermission>(userPermissions)
  const has = (perm: AppPermission): boolean => isAdmin || permSet.has(perm)
  const hasAny = (perms: AppPermission[]): boolean => perms.some(has)

  return { config, user: accessUser, has, hasAny, isAdmin }
}

// userHasAny is a hook-free permission check over a fetched /users/me object.
// Use it on routes rendered OUTSIDE the org <AccessProvider> (e.g. /live/$liveId),
// where the useAccess* hooks throw. Inside the provider, prefer useCanAny.
export function userHasAny(user: User | undefined, perms: AppPermission[]): boolean {
  if (!user) return false
  return buildAccess(user).hasAny(perms)
}

export function useCanSelfOr(basePerm: AppPermission, anyPerm: AppPermission, targetId: string | undefined) {
  const { can, user } = useAccess()
  if (can(anyPerm)) return true
  if (can(basePerm) && targetId === user.id) return true
  return false
}

export function useCanAny(perms: AppPermission[]): boolean {
  const { can } = useAccess()
  return perms.some((p) => can(p))
}

export function useOrgGuard(perms: AppPermission | AppPermission[]): boolean {
  const list = Array.isArray(perms) ? perms : [perms]
  const allowed = useCanAny(list)
  const navigate = useNavigate()
  const params = useParams({ strict: false }) as { orgId?: string }
  const orgId = params.orgId

  useEffect(() => {
    if (!allowed && orgId) {
      navigate({ to: "/org/$orgId/dashboard", params: { orgId } })
    }
  }, [allowed, orgId, navigate])

  return allowed
}
