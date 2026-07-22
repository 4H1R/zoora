import type {
  GithubCom4H1RZooraInternalDomainAuditEntry as AuditEntry,
  GetAuditParams,
} from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import { ClockIcon, ScrollTextIcon, ShieldAlertIcon } from "lucide-react"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import { useGetAudit } from "@/api/audit/audit"
import {
  GithubCom4H1RZooraInternalDomainAuditAction as AuditActionEnum,
  GithubCom4H1RZooraInternalDomainAuditOutcome as AuditOutcomeEnum,
  GithubCom4H1RZooraInternalDomainAuditTargetType as AuditTargetTypeEnum,
} from "@/api/model"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { UserAvatar } from "@/components/user-avatar"
import { useOrgGuard } from "@/lib/access"
import { useFormatDate } from "@/lib/format-date"
import { orgHead } from "@/lib/org-head"
import { formatRelativeTime } from "@/lib/relative-time"
import { cn } from "@/lib/utils"

const PAGE_SIZE = 20
const ALL = "all"

const searchSchema = z.object({
  page: z.number().int().min(1).optional(),
  action: z.string().optional(),
  target_type: z.string().optional(),
  outcome: z.string().optional(),
})

export const Route = createFileRoute("/_auth/org/settings/audit")({
  head: () => orgHead("org.audit.title"),
  validateSearch: searchSchema,
  component: AuditLogPage,
})

const ACTIONS = Object.values(AuditActionEnum)
const TARGET_TYPES = Object.values(AuditTargetTypeEnum)
const OUTCOMES = Object.values(AuditOutcomeEnum)

// Verb tint — the audit log's one memorable visual cue. Destructive verbs run
// hot, additive verbs run cool, everything else stays neutral.
const ACTION_TINT: Record<string, string> = {
  created: "text-emerald-600 dark:text-emerald-400",
  enabled: "text-emerald-600 dark:text-emerald-400",
  enrolled: "text-emerald-600 dark:text-emerald-400",
  updated: "text-sky-600 dark:text-sky-400",
  graded: "text-violet-600 dark:text-violet-400",
  deleted: "text-rose-600 dark:text-rose-400",
  disabled: "text-rose-600 dark:text-rose-400",
  unenrolled: "text-amber-600 dark:text-amber-400",
}

// The ledger spine — same semantic buckets as the verb tint, projected onto the
// card's leading edge so a column of entries reads at a glance. Denied always
// overrides to rose; unknown verbs stay neutral.
const ACTION_ACCENT: Record<string, string> = {
  created: "border-s-emerald-500",
  enabled: "border-s-emerald-500",
  enrolled: "border-s-emerald-500",
  updated: "border-s-sky-500",
  graded: "border-s-violet-500",
  deleted: "border-s-rose-500",
  disabled: "border-s-rose-500",
  unenrolled: "border-s-amber-500",
}

function metadataEntries(metadata: AuditEntry["metadata"]): [string, unknown][] {
  if (!metadata || typeof metadata !== "object") return []
  return Object.entries(metadata as Record<string, unknown>)
}

function AuditLogPage() {
  const { t, i18n } = useTranslation()
  const formatDate = useFormatDate()
  const allowed = useOrgGuard("audit:view_any")
  const navigate = Route.useNavigate()
  const { page, action, target_type, outcome } = Route.useSearch()

  const currentPage = page ?? 1
  const hasFilters = !!(action || target_type || outcome)

  const params: GetAuditParams = {
    page: currentPage,
    page_size: PAGE_SIZE,
    action: action || undefined,
    target_type: target_type || undefined,
    outcome: outcome || undefined,
  }

  const { data, isLoading } = useGetAudit(params, { query: { enabled: allowed } })
  const paginated = (data?.status === 200 && data.data.data) || undefined
  const entries: AuditEntry[] = paginated?.items ?? []
  const total = paginated?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const setFilter = (key: "action" | "target_type" | "outcome", value: string) =>
    navigate({
      search: (prev) => ({ ...prev, [key]: value === ALL ? undefined : value, page: undefined }),
    })

  const resetFilters = () => navigate({ search: {} })

  const goToPage = (next: number) =>
    navigate({ search: (prev) => ({ ...prev, page: next <= 1 ? undefined : next }) })

  if (!allowed) return null

  return (
    <div className="space-y-8">
      <div className="min-w-0 space-y-1">
        <h1 className="text-2xl font-bold tracking-tight">{t("org.audit.title")}</h1>
        <p className="text-muted-foreground text-sm">{t("org.audit.subtitle")}</p>
      </div>

      {/* Filter bar */}
      <div className="flex flex-wrap items-center gap-2">
        <FilterSelect
          value={action ?? ALL}
          onValueChange={(v) => setFilter("action", v)}
          placeholder={t("org.audit.filters.action")}
          allLabel={t("org.audit.filters.allActions")}
          options={ACTIONS.map((a) => ({ value: a, label: t(`org.audit.actions.${a}`) }))}
        />
        <FilterSelect
          value={target_type ?? ALL}
          onValueChange={(v) => setFilter("target_type", v)}
          placeholder={t("org.audit.filters.targetType")}
          allLabel={t("org.audit.filters.allTargets")}
          options={TARGET_TYPES.map((tt) => ({ value: tt, label: t(`org.audit.targets.${tt}`) }))}
        />
        <FilterSelect
          value={outcome ?? ALL}
          onValueChange={(v) => setFilter("outcome", v)}
          placeholder={t("org.audit.filters.outcome")}
          allLabel={t("org.audit.filters.allOutcomes")}
          options={OUTCOMES.map((o) => ({ value: o, label: t(`org.audit.outcomes.${o}`) }))}
        />
        {hasFilters && (
          <Button variant="ghost" size="sm" onClick={resetFilters}>
            {t("common.reset")}
          </Button>
        )}
      </div>

      {/* Ledger */}
      <AuditLedger
        isLoading={isLoading}
        entries={entries}
        hasFilters={hasFilters}
        resetFilters={resetFilters}
        t={t}
        i18n={i18n}
        formatDate={formatDate}
      />

      {/* Pagination */}
      {total > PAGE_SIZE && (
        <div className="flex items-center justify-between">
          <p className="text-muted-foreground text-sm">
            {t("org.audit.pageOf", { page: currentPage, total: totalPages })}
          </p>
          <Pagination className="mx-0 w-fit">
            <PaginationContent>
              <PaginationItem>
                <PaginationPrevious
                  aria-disabled={currentPage <= 1}
                  className={cn(currentPage <= 1 && "pointer-events-none opacity-50")}
                  onClick={() => goToPage(currentPage - 1)}
                >
                  {t("common.pagination.previous")}
                </PaginationPrevious>
              </PaginationItem>
              <PaginationItem>
                <PaginationNext
                  aria-disabled={currentPage >= totalPages}
                  className={cn(currentPage >= totalPages && "pointer-events-none opacity-50")}
                  onClick={() => goToPage(currentPage + 1)}
                >
                  {t("common.pagination.next")}
                </PaginationNext>
              </PaginationItem>
            </PaginationContent>
          </Pagination>
        </div>
      )}
    </div>
  )
}

type AuditLedgerProps = {
  isLoading: boolean
  entries: AuditEntry[]
  hasFilters: boolean
  resetFilters: () => void
  t: ReturnType<typeof useTranslation>["t"]
  i18n: ReturnType<typeof useTranslation>["i18n"]
  formatDate: ReturnType<typeof useFormatDate>
}

function AuditLedger({ isLoading, entries, hasFilters, resetFilters, t, i18n, formatDate }: AuditLedgerProps) {
  if (isLoading) {
    return (
      <div className="space-y-3">
        {[0, 1, 2, 3, 4].map((i) => (
          <Card key={i} className="border-s-border flex-row items-start gap-3 border-s-4 p-4">
            <Skeleton className="mt-0.5 size-6 shrink-0 rounded-full" />
            <div className="flex-1 space-y-2">
              <Skeleton className="h-4 w-2/3 rounded-md" />
              <Skeleton className="h-3 w-24 rounded-md" />
            </div>
          </Card>
        ))}
      </div>
    )
  }

  if (entries.length === 0) {
    return (
      <Card className="flex flex-col items-center justify-center gap-3 border-dashed py-16 text-center">
        <ScrollTextIcon className="text-muted-foreground/50 size-8" />
        <p className="text-muted-foreground">
          {hasFilters ? t("org.audit.emptyFiltered") : t("org.audit.empty")}
        </p>
        {hasFilters && (
          <Button variant="outline" onClick={resetFilters}>
            {t("common.reset")}
          </Button>
        )}
      </Card>
    )
  }

  return (
    <ul className="space-y-3">
      {entries.map((entry, i) => {
        const denied = entry.outcome === "denied"
        const meta = metadataEntries(entry.metadata)
        const actorName = entry.actor_name || t("org.audit.system")
        const accent = denied ? "border-s-rose-500" : ACTION_ACCENT[entry.action ?? ""]
        return (
          <li
            key={entry.id}
            className="animate-in fade-in slide-in-from-bottom-2"
            style={{ animationDelay: `${Math.min(i, 8) * 40}ms`, animationFillMode: "backwards" }}
          >
            <Card
              className={cn(
                "gap-2.5 border-s-4 p-4 transition-colors",
                accent ?? "border-s-border",
                "hover:bg-muted/40"
              )}
            >
              <div className="flex items-start gap-3">
                {/* Actor anchor */}
                <UserAvatar name={actorName} size="sm" className="mt-0.5 shrink-0" />

                <div className="min-w-0 flex-1">
                  {/* The sentence + timestamp pinned to the far edge */}
                  <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
                    <span className="font-medium">{actorName}</span>

                    <span className={cn("font-semibold", ACTION_TINT[entry.action ?? ""])}>
                      {t(`org.audit.actions.${entry.action}`)}
                    </span>

                    <span className="bg-muted text-muted-foreground rounded-md px-1.5 py-0.5 font-mono text-xs">
                      {t(`org.audit.targets.${entry.target_type}`)}
                    </span>
                    {Boolean(entry.target_label) && (
                      <span className="min-w-0 truncate font-medium">{entry.target_label}</span>
                    )}

                    {denied && (
                      <Badge variant="destructive" className="gap-1">
                        <ShieldAlertIcon className="size-3" />
                        {t("org.audit.outcomes.denied")}
                      </Badge>
                    )}

                    <span
                      className="text-muted-foreground ms-auto inline-flex shrink-0 items-center gap-1 text-xs tabular-nums"
                      title={formatDate(entry.created_at, "datetime-long")}
                    >
                      <ClockIcon className="size-3.5" />
                      {formatRelativeTime(entry.created_at, i18n.language)}
                    </span>
                  </div>

                  {/* Byline: username + details toggle */}
                  {(Boolean(entry.actor_username) || meta.length > 0) && (
                    <div className="text-muted-foreground mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs">
                      {Boolean(entry.actor_username) && <span>@{entry.actor_username}</span>}
                      {Boolean(entry.actor_username) && meta.length > 0 && (
                        <span aria-hidden className="text-muted-foreground/40">
                          ·
                        </span>
                      )}
                      {meta.length > 0 && (
                        <details className="group">
                          <summary className="hover:text-foreground w-fit cursor-pointer list-none underline-offset-2 group-open:underline">
                            {t("org.audit.showDetails")}
                          </summary>
                          <dl className="mt-2 grid grid-cols-[auto_1fr] gap-x-3 gap-y-1">
                            {meta.map(([k, v]) => (
                              <div key={k} className="col-span-2 grid grid-cols-subgrid">
                                <dt className="text-muted-foreground font-mono">{k}</dt>
                                <dd className="text-foreground min-w-0 truncate font-mono">
                                  {typeof v === "object" ? JSON.stringify(v) : String(v)}
                                </dd>
                              </div>
                            ))}
                          </dl>
                        </details>
                      )}
                    </div>
                  )}
                </div>
              </div>
            </Card>
          </li>
        )
      })}
    </ul>
  )
}

type FilterSelectProps = {
  value: string
  onValueChange: (value: string) => void
  placeholder: string
  allLabel: string
  options: { value: string; label: string }[]
}

function FilterSelect({ value, onValueChange, placeholder, allLabel, options }: FilterSelectProps) {
  const items = [{ value: ALL, label: allLabel }, ...options]
  return (
    <Select value={value} onValueChange={(v) => onValueChange(v ?? ALL)} items={items}>
      <SelectTrigger className="w-auto min-w-40" aria-label={placeholder}>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        {items.map((it) => (
          <SelectItem key={it.value} value={it.value}>
            {it.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
