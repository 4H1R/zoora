import type { GithubCom4H1RZooraInternalDomainLead as Lead } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import {
  Building2Icon,
  CheckIcon,
  EllipsisVerticalIcon,
  PhoneIcon,
  Trash2Icon,
  UserPlusIcon,
} from "lucide-react"
import { useTranslation } from "react-i18next"

import { GithubCom4H1RZooraInternalDomainLeadStatus as LeadStatus } from "@/api/model"
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
import { planTier } from "@/lib/plan"
import { cn } from "@/lib/utils"

const STATUS_VARIANT: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  [LeadStatus.LeadStatusNew]: "default",
  [LeadStatus.LeadStatusContacted]: "secondary",
  [LeadStatus.LeadStatusConverted]: "outline",
  [LeadStatus.LeadStatusRejected]: "destructive",
}

function LeadStatusBadge({ status }: { status?: string }) {
  const { t } = useTranslation()
  if (!status) return <span className="text-muted-foreground">—</span>
  return (
    <Badge variant={STATUS_VARIANT[status] ?? "secondary"} className="text-[11px] capitalize">
      {t(`admin.leads.statusLabels.${status}`, { defaultValue: status })}
    </Badge>
  )
}

function LeadPlanBadge({ plan }: { plan?: string }) {
  const { t } = useTranslation()
  if (!plan) return <span className="text-muted-foreground">—</span>
  const tier = planTier(plan)
  return (
    <Badge variant="outline" className="text-[11px]">
      {t(`plans.tiers.${tier}`, { defaultValue: plan })}
    </Badge>
  )
}

interface LeadRowActionsProps {
  lead: Lead
  onConvert: (lead: Lead) => void
  onSetStatus: (lead: Lead, status: string) => void
  onDelete: (lead: Lead) => void
}

function LeadRowActions({ lead, onConvert, onSetStatus, onDelete }: LeadRowActionsProps) {
  const { t } = useTranslation()
  const isConverted = lead.status === LeadStatus.LeadStatusConverted

  return (
    <div className="flex items-center justify-end gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
      {!isConverted ? (
        <Button variant="ghost" size="icon-xs" onClick={() => onConvert(lead)} title={t("admin.leads.actions.convert")}>
          <UserPlusIcon />
        </Button>
      ) : null}
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button variant="ghost" size="icon-xs">
              <EllipsisVerticalIcon />
            </Button>
          }
        />
        <DropdownMenuContent align="end" className="min-w-44">
          {!isConverted ? (
            <>
              <DropdownMenuGroup>
                <DropdownMenuItem onClick={() => onConvert(lead)}>
                  <UserPlusIcon data-icon="inline-start" />
                  {t("admin.leads.actions.convert")}
                </DropdownMenuItem>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />
              <DropdownMenuGroup>
                {lead.status !== LeadStatus.LeadStatusContacted ? (
                  <DropdownMenuItem onClick={() => onSetStatus(lead, LeadStatus.LeadStatusContacted)}>
                    <CheckIcon data-icon="inline-start" />
                    {t("admin.leads.actions.markContacted")}
                  </DropdownMenuItem>
                ) : null}
                {lead.status !== LeadStatus.LeadStatusNew ? (
                  <DropdownMenuItem onClick={() => onSetStatus(lead, LeadStatus.LeadStatusNew)}>
                    <CheckIcon data-icon="inline-start" />
                    {t("admin.leads.actions.markNew")}
                  </DropdownMenuItem>
                ) : null}
                <DropdownMenuItem onClick={() => onSetStatus(lead, LeadStatus.LeadStatusRejected)}>
                  <CheckIcon data-icon="inline-start" />
                  {t("admin.leads.actions.markRejected")}
                </DropdownMenuItem>
              </DropdownMenuGroup>
              <DropdownMenuSeparator />
            </>
          ) : null}
          <DropdownMenuGroup>
            <DropdownMenuItem variant="destructive" onClick={() => onDelete(lead)}>
              <Trash2Icon data-icon="inline-start" />
              {t("admin.leads.actions.delete")}
            </DropdownMenuItem>
          </DropdownMenuGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}

interface UseLeadColumnsOptions {
  onConvert: (lead: Lead) => void
  onSetStatus: (lead: Lead, status: string) => void
  onDelete: (lead: Lead) => void
}

export function useLeadColumns({ onConvert, onSetStatus, onDelete }: UseLeadColumnsOptions): ColumnDef<Lead>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  return [
    {
      accessorKey: "name",
      header: t("admin.leads.name"),
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <div
            className={cn(
              "flex size-9 shrink-0 items-center justify-center rounded-lg text-xs font-semibold text-white",
              getEntityColor(row.original.name ?? "")
            )}
          >
            {getInitials(row.original.name ?? "")}
          </div>
          <div className="min-w-0">
            <div className="truncate text-sm font-medium">{row.original.name}</div>
            <div className="text-muted-foreground flex items-center gap-1 text-xs" dir="ltr">
              <PhoneIcon className="size-3" />
              {row.original.phone}
            </div>
          </div>
        </div>
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: "org_name",
      header: t("admin.leads.orgName"),
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <Building2Icon className="text-muted-foreground size-3.5" />
          <span className="truncate text-sm">{row.original.org_name}</span>
        </div>
      ),
      enableSorting: false,
    },
    {
      accessorKey: "plan",
      header: t("admin.leads.plan"),
      cell: ({ row }) => <LeadPlanBadge plan={row.original.plan} />,
      enableSorting: false,
    },
    {
      accessorKey: "status",
      header: t("admin.leads.status"),
      cell: ({ row }) => <LeadStatusBadge status={row.original.status} />,
      enableSorting: false,
    },
    {
      accessorKey: "created_at",
      header: t("admin.leads.createdAt"),
      cell: ({ row }) => <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>,
      enableSorting: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <LeadRowActions lead={row.original} onConvert={onConvert} onSetStatus={onSetStatus} onDelete={onDelete} />
      ),
      enableSorting: false,
      enableHiding: false,
    },
  ]
}
