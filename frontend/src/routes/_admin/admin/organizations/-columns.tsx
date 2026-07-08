import type { GithubCom4H1RZooraInternalDomainOrganization as Organization } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { CreditCardIcon, EllipsisVerticalIcon, ExternalLinkIcon, PencilIcon, Trash2Icon, UsersIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { GithubCom4H1RZooraInternalDomainOrganizationStatus as OrgStatus } from "@/api/model"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { getEntityColor, getInitials, useFormatDate } from "@/lib/data-table"
import { planSize, planTier, type PlanTier } from "@/lib/plan"
import { cn } from "@/lib/utils"

const STATUS_VARIANT: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  [OrgStatus.OrganizationStatusActive]: "default",
  [OrgStatus.OrganizationStatusTrial]: "secondary",
  [OrgStatus.OrganizationStatusSuspended]: "destructive",
  [OrgStatus.OrganizationStatusArchived]: "outline",
}

function OrgStatusBadge({ status }: { status?: string }) {
  const { t } = useTranslation()
  if (!status) return <span className="text-muted-foreground">—</span>
  return (
    <Badge variant={STATUS_VARIANT[status] ?? "secondary"} className="text-[11px] capitalize">
      {t(`admin.orgs.statusLabels.${status}`, { defaultValue: status })}
    </Badge>
  )
}

const PLAN_VARIANT: Record<PlanTier, "default" | "secondary" | "outline"> = {
  free: "outline",
  plus: "secondary",
  pro: "default",
  max: "secondary",
}

function OrgPlanBadge({ plan }: { plan?: string }) {
  const { t } = useTranslation()
  if (!plan) return <span className="text-muted-foreground">—</span>
  const tier = planTier(plan)
  return (
    <Badge variant={PLAN_VARIANT[tier] ?? "outline"} className="text-[11px]">
      {t(`plans.tiers.${tier}`, { defaultValue: tier })}
      <span className="tabular-nums">{planSize(plan)}</span>
    </Badge>
  )
}

interface OrgRowActionsProps {
  organization: Organization
  onEdit: (org: Organization) => void
  onChangePlan: (org: Organization) => void
  onDelete: (org: Organization) => void
}

function OrgRowActions({ organization, onEdit, onChangePlan, onDelete }: OrgRowActionsProps) {
  const { t } = useTranslation()

  // Org dashboards are host-scoped (one subdomain per tenant), so open the org's
  // own subdomain rather than a path param. Preserves the current scheme/port.
  const handleGoToOrg = () => {
    const base = import.meta.env.VITE_BASE_DOMAIN ?? "localhost"
    const { protocol, port } = window.location
    const portSuffix = port ? `:${port}` : ""
    const href = `${protocol}//${organization.slug}.${base}${portSuffix}/org/dashboard`
    window.open(href, "_blank", "noopener,noreferrer")
  }

  return (
    <div className="flex items-center justify-end gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
      <Button variant="ghost" size="icon-xs" onClick={() => onEdit(organization)}>
        <PencilIcon />
      </Button>
      <Button
        variant="ghost"
        size="icon-xs"
        className="text-destructive hover:bg-destructive/10 hover:text-destructive"
        onClick={() => onDelete(organization)}
      >
        <Trash2Icon />
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button variant="ghost" size="icon-xs">
              <EllipsisVerticalIcon />
            </Button>
          }
        />
        <DropdownMenuContent align="end" className="min-w-44">
          <DropdownMenuGroup>
            <DropdownMenuItem onClick={handleGoToOrg}>
              <ExternalLinkIcon data-icon="inline-start" />
              {t("admin.orgs.actions.goToOrg")}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => onEdit(organization)}>
              <PencilIcon data-icon="inline-start" />
              {t("admin.orgs.actions.edit")}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => onChangePlan(organization)}>
              <CreditCardIcon data-icon="inline-start" />
              {t("admin.orgs.actions.changePlan")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
          <DropdownMenuSeparator />
          <DropdownMenuGroup>
            <DropdownMenuItem variant="destructive" onClick={() => onDelete(organization)}>
              <Trash2Icon data-icon="inline-start" />
              {t("admin.orgs.actions.delete")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}

interface UseOrgColumnsOptions {
  onEdit: (org: Organization) => void
  onChangePlan: (org: Organization) => void
  onDelete: (org: Organization) => void
}

export function useOrgColumns({ onEdit, onChangePlan, onDelete }: UseOrgColumnsOptions): ColumnDef<Organization>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  return [
    {
      accessorKey: "name",
      header: t("admin.orgs.name"),
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <div
            className={cn(
              "flex size-9 shrink-0 items-center justify-center rounded-lg text-xs font-semibold text-white",
              getEntityColor(row.original.name)
            )}
          >
            {getInitials(row.original.name)}
          </div>
          <div className="min-w-0">
            <div className="truncate text-sm font-medium">{row.original.name}</div>
          </div>
        </div>
      ),
      enableSorting: true,
      enableHiding: false,
    },
    {
      accessorKey: "status",
      header: t("admin.orgs.status"),
      cell: ({ row }) => <OrgStatusBadge status={row.original.status} />,
      enableSorting: false,
    },
    {
      accessorKey: "plan",
      header: t("admin.orgs.plan.plan"),
      cell: ({ row }) => <OrgPlanBadge plan={row.original.plan} />,
      enableSorting: false,
    },
    {
      accessorKey: "total_users",
      header: t("admin.orgs.totalUsers"),
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <UsersIcon className="text-muted-foreground size-3.5" />
          <span className="text-sm tabular-nums">{(row.original.total_users ?? 0).toLocaleString()}</span>
        </div>
      ),
      enableSorting: false,
    },
    {
      accessorKey: "created_at",
      header: t("admin.orgs.createdAt"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>,
      enableSorting: true,
    },
    {
      accessorKey: "updated_at",
      header: t("admin.orgs.updatedAt"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{formatDate(row.original.updated_at)}</span>,
      enableSorting: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <OrgRowActions organization={row.original} onEdit={onEdit} onChangePlan={onChangePlan} onDelete={onDelete} />
      ),
      enableSorting: false,
      enableHiding: false,
    },
  ]
}
