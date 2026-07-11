import { Link } from "@tanstack/react-router"
import { BuildingIcon, SettingsIcon, ShieldCheckIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetOrganizationsId } from "@/api/organizations/organizations"
import { useGetUsersMe } from "@/api/users/users"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { useRoleName } from "@/lib/permissions"

function initials(name?: string) {
  if (!name) return "—"
  const words = name.trim().split(/\s+/).filter(Boolean)
  if (words.length === 0) return "—"
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase()
  return (words[0][0] + words[words.length - 1][0]).toUpperCase()
}

/** Compact identity summary for the dashboard. Doubles as a signpost to the
 * account page — non-technical users discover their settings from here. */
export function ProfileCard() {
  const { t } = useTranslation()
  const roleName = useRoleName()

  const { data: meData, isPending } = useGetUsersMe()
  const me = (meData?.status === 200 && meData.data.data) || undefined

  const orgId = me?.organization_id ?? ""
  const { data: orgData } = useGetOrganizationsId(orgId, { query: { enabled: !!orgId } })
  const orgName = (orgData?.status === 200 && orgData.data.data?.name) || undefined

  if (isPending) {
    return (
      <Card className="flex-row items-center gap-4 p-4 sm:p-5">
        <Skeleton className="size-14 shrink-0 rounded-2xl" />
        <div className="flex min-w-0 flex-1 flex-col gap-2">
          <Skeleton className="h-5 w-40" />
          <Skeleton className="h-4 w-28" />
        </div>
        <Skeleton className="h-9 w-32 rounded-lg" />
      </Card>
    )
  }

  const role = me?.is_admin ? t("admin.roleAdmin") : me?.role?.name ? roleName(me.role.name) : t("admin.roleMember")

  return (
    <Card className="ring-foreground/5 relative flex-row items-center gap-4 overflow-hidden p-4 ring-1 sm:p-5">
      <div
        aria-hidden
        className="from-primary/8 pointer-events-none absolute inset-0 bg-gradient-to-br via-transparent to-transparent"
      />
      <div className="from-primary text-primary-foreground ring-foreground/10 relative grid size-14 shrink-0 place-items-center rounded-2xl bg-gradient-to-br to-indigo-700 text-lg font-semibold tracking-tight shadow-sm ring-1 select-none">
        {initials(me?.name)}
      </div>

      <div className="relative min-w-0 flex-1">
        <div className="flex flex-wrap items-center gap-2">
          <p className="text-foreground truncate text-base font-semibold tracking-tight">{me?.name || "—"}</p>
          <span className="border-primary/25 bg-primary/10 text-primary inline-flex h-5 items-center gap-1 rounded-full border px-2 text-[11px] font-medium">
            <ShieldCheckIcon className="size-3" />
            {role}
          </span>
        </div>
        <div className="text-muted-foreground mt-0.5 flex flex-wrap items-center gap-x-3 gap-y-0.5 text-xs">
          <span className="truncate">@{me?.username}</span>
          {orgName && (
            <span className="inline-flex items-center gap-1">
              <BuildingIcon className="size-3" />
              {orgName}
            </span>
          )}
        </div>
      </div>

      <Button variant="outline" size="sm" render={<Link to="/org/account" />} className="relative shrink-0 gap-1.5">
        <SettingsIcon className="size-4" />
        <span className="hidden sm:inline">{t("account.menuAction")}</span>
      </Button>
    </Card>
  )
}
