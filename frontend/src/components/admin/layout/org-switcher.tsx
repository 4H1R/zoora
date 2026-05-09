import type { GithubCom4H1RZooraInternalDomainOrganization } from "@/api/model"

import { CheckIcon, ChevronDownIcon, PlusIcon, SearchIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useDebounce } from "use-debounce"

import { useGetAdminOrganizations } from "@/api/admin-organizations/admin-organizations"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { SidebarMenu, SidebarMenuButton, SidebarMenuItem, useSidebar } from "@/components/ui/sidebar"
import { getInitials } from "@/components/user-avatar"
import { getEntityColor } from "@/lib/data-table"
import { cn } from "@/lib/utils"

type Organization = GithubCom4H1RZooraInternalDomainOrganization

const RECENT_ORGS_KEY = "admin:recent-orgs"
const MAX_RECENT = 3
const ALL_MARK_CLASS = "bg-emerald-700"

function getRecentOrgIds(): string[] {
  try {
    const raw = localStorage.getItem(RECENT_ORGS_KEY)
    return raw ? JSON.parse(raw) : []
  } catch {
    return []
  }
}

function pushRecentOrgId(id: string) {
  const ids = getRecentOrgIds().filter((x) => x !== id)
  ids.unshift(id)
  localStorage.setItem(RECENT_ORGS_KEY, JSON.stringify(ids.slice(0, MAX_RECENT)))
}

function formatCount(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1).replace(/\.0$/, "")}k`
  return String(n)
}

export function OrgSwitcher({
  selected,
  onSelect,
  onCreateOrg,
}: {
  selected?: Organization | null
  onSelect: (org: Organization | null) => void
  onCreateOrg?: () => void
}) {
  const { t } = useTranslation()
  const { isMobile } = useSidebar()
  const [search, setSearch] = useState("")
  const [debouncedSearch] = useDebounce(search, 300)

  const { data } = useGetAdminOrganizations({ search: debouncedSearch || undefined })
  const orgsData = (data?.status === 200 && data.data.data) || undefined
  const organizations = orgsData?.items ?? []
  const totalOrgs = orgsData?.total ?? organizations.length

  const recentIds = getRecentOrgIds()
  const recentOrgs = recentIds.map((id) => organizations.find((o) => o.id === id)).filter((o): o is Organization => !!o)
  const recentIdSet = new Set(recentIds)
  const remainingOrgs = organizations.filter((o) => !recentIdSet.has(o.id!))

  const handleSelect = (org: Organization | null) => {
    if (org?.id) pushRecentOrgId(org.id)
    onSelect(org)
    setSearch("")
  }

  const displayName = selected?.name ?? t("admin.orgs.switcher.all")
  const initials = getInitials(selected?.name)
  const markColor = selected ? getEntityColor(selected.name) : ALL_MARK_CLASS

  const orgItem = (org: Organization) => (
    <DropdownMenuItem key={org.id} className="gap-2.5 rounded-md px-2.5 py-1.5" onClick={() => handleSelect(org)}>
      <div
        className={cn(
          "flex size-5 shrink-0 items-center justify-center rounded-[5px] text-[10px] font-semibold !text-white",
          getEntityColor(org.name)
        )}
      >
        {getInitials(org.name)}
      </div>
      <span className="flex-1 truncate text-[13px]">{org.name}</span>
      <span className="text-muted-foreground text-[11px]">{formatCount(org.total_users ?? 0)}</span>
      <CheckIcon
        className={cn("text-primary size-3.5 shrink-0 stroke-2", selected?.id === org.id ? "opacity-100" : "opacity-0")}
      />
    </DropdownMenuItem>
  )

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <SidebarMenuButton
                size="lg"
                className="border-border hover:border-border-strong hover:bg-sidebar-accent data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground border transition-colors group-data-[collapsible=icon]:justify-center"
              />
            }
          >
            <div
              className={cn(
                "flex size-5.5 shrink-0 items-center justify-center rounded-[5px] text-[10px] font-semibold !text-white",
                markColor
              )}
            >
              {initials}
            </div>
            <div className="grid flex-1 text-start leading-tight">
              <span className="text-muted-foreground text-[10px] font-medium tracking-wider uppercase">
                {t("admin.orgs.switcher.label")}
              </span>
              <span className="truncate text-sm font-medium">{displayName}</span>
            </div>
            <ChevronDownIcon className="text-muted-foreground ms-auto size-3.5" />
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="min-w-64 rounded-lg"
            side={isMobile ? "bottom" : "right"}
            align="start"
            sideOffset={4}
          >
            <div className="flex items-center gap-2 border-b px-3 py-2.5">
              <SearchIcon className="text-muted-foreground size-3.5 shrink-0" />
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                onKeyDown={(e) => e.stopPropagation()}
                placeholder={t("admin.orgs.switcher.search")}
                className="h-auto border-0 bg-transparent px-0 py-0 text-[13px] shadow-none focus-visible:ring-0"
              />
            </div>

            {/* Scope */}
            <DropdownMenuGroup>
              <DropdownMenuLabel>{t("admin.orgs.switcher.scope")}</DropdownMenuLabel>
              <DropdownMenuItem className="gap-2.5 rounded-md px-2.5 py-1.5" onClick={() => handleSelect(null)}>
                <div
                  className={cn(
                    "flex size-5 shrink-0 items-center justify-center rounded-[5px] text-xs font-semibold !text-white",
                    ALL_MARK_CLASS
                  )}
                >
                  EC
                </div>
                <span className="flex-1 text-[13px]">{t("admin.orgs.switcher.all")}</span>
                <span className="text-muted-foreground text-[11px]">{totalOrgs}</span>
                <CheckIcon
                  className={cn("text-primary size-3.5 shrink-0 stroke-2", !selected ? "opacity-100" : "opacity-0")}
                />
              </DropdownMenuItem>
            </DropdownMenuGroup>

            {/* Recent */}
            {recentOrgs.length > 0 && !search && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuGroup>
                  <DropdownMenuLabel>{t("admin.orgs.switcher.recent")}</DropdownMenuLabel>
                  {recentOrgs.map(orgItem)}
                </DropdownMenuGroup>
              </>
            )}

            {/* All organizations */}
            {!search && remainingOrgs.length > 0 && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuGroup>
                  <DropdownMenuLabel>{t("admin.orgs.switcher.allOrgs")}</DropdownMenuLabel>
                  {remainingOrgs.slice(0, 8).map(orgItem)}
                </DropdownMenuGroup>
              </>
            )}

            {/* When searching, show all results without recent split */}
            {search && organizations.length > 0 && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuGroup>
                  <DropdownMenuLabel>{t("admin.organizations")}</DropdownMenuLabel>
                  {organizations.slice(0, 8).map(orgItem)}
                </DropdownMenuGroup>
              </>
            )}

            {/* Create organization footer */}
            {onCreateOrg && (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuItem className="gap-2.5 rounded-md px-2.5 py-1.5" onClick={onCreateOrg}>
                  <PlusIcon className="text-muted-foreground size-3.5" />
                  <span className="flex-1 text-[13px]">{t("admin.orgs.switcher.create")}</span>
                  <kbd className="text-muted-foreground text-[11px]">⌘N</kbd>
                </DropdownMenuItem>
              </>
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
