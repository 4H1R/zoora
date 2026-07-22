import type { GithubCom4H1RZooraInternalDomainMediaOwner as MediaOwner } from "@/api/model"
import type { CellContext, ColumnDef } from "@tanstack/react-table"

import { createFileRoute, Link } from "@tanstack/react-router"
import { ChevronRightIcon, FolderOpenIcon, UploadCloudIcon } from "lucide-react"
import { useTranslation } from "react-i18next"
import { z } from "zod"

import { useGetFilesFolders, useGetFilesOwners } from "@/api/media/media"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { folderStyle, formatBytes, ownerStyle, SHARED_FOLDER } from "@/components/org/files/utils"
import { PageHeader } from "@/components/page-header"
import { Card } from "@/components/ui/card"
import { EmptyState } from "@/components/ui/empty-state"
import { Progress } from "@/components/ui/progress"
import { Skeleton } from "@/components/ui/skeleton"
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import { useOrgGuard } from "@/lib/access"
import { useAdminTable } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"
import { cn } from "@/lib/utils"

const filesSearchSchema = z.object({
  view: z.enum(["type", "owner"]).optional(),
  page: z.number().int().positive().optional().default(1),
  page_size: z.number().int().positive().optional().default(20),
})

export const Route = createFileRoute("/_auth/org/files/")({
  head: () => orgHead("org.nav.files"),
  validateSearch: filesSearchSchema,
  component: FilesPage,
})

function FilesPage() {
  const { t } = useTranslation()
  const allowed = useOrgGuard(["media:view_any"])
  const { view = "type" } = Route.useSearch()
  const navigate = Route.useNavigate()

  if (!allowed) return null

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("filesPage.title")} />
      <div className="-mt-4 flex flex-wrap items-center justify-between gap-3">
        <p className="text-muted-foreground text-sm">
          {view === "owner" ? t("filesPage.owner.description") : t("filesPage.description")}
        </p>
        <ToggleGroup
          value={[view]}
          onValueChange={(next) => {
            const picked = next[0]
            if (picked === "type" || picked === "owner") {
              navigate({ search: (prev) => ({ ...prev, view: picked, page: 1 }) })
            }
          }}
          variant="outline"
          size="sm"
        >
          <ToggleGroupItem value="type">{t("filesPage.tabs.byType")}</ToggleGroupItem>
          <ToggleGroupItem value="owner">{t("filesPage.tabs.byOwner")}</ToggleGroupItem>
        </ToggleGroup>
      </div>

      {view === "owner" ? <ByOwnerView /> : <ByTypeView />}
    </div>
  )
}

function ByTypeView() {
  const { t } = useTranslation()
  const { data, isLoading } = useGetFilesFolders()
  const folders = (data?.status === 200 && data.data.data) || []

  // Shared folder is pinned first — it's the only folder that accepts uploads.
  const sorted = [...folders].sort((a, b) =>
    a.model_type === SHARED_FOLDER ? -1 : b.model_type === SHARED_FOLDER ? 1 : 0
  )

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-36 rounded-2xl" />
        ))}
      </div>
    )
  }

  if (sorted.length === 0) {
    return <EmptyState icon={FolderOpenIcon} title={t("filesPage.empty")} />
  }

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {sorted.map((folder) => {
        const type = folder.model_type ?? ""
        const isShared = type === SHARED_FOLDER
        const { icon: Icon, tint } = folderStyle(type)
        return (
          <Link
            key={type}
            to="/org/files/$folder"
            params={{ folder: type }}
            className={cn(
              "group bg-card relative flex flex-col gap-4 overflow-hidden rounded-2xl border p-5 transition-all",
              "hover:border-primary/40 hover:-translate-y-0.5 hover:shadow-md",
              isShared && "border-primary/30"
            )}
          >
            {/* Soft top wash echoes the folder tint without shouting. */}
            <div
              aria-hidden
              className={cn(
                "pointer-events-none absolute inset-x-0 top-0 h-16 opacity-60",
                "bg-[radial-gradient(ellipse_80%_100%_at_50%_0%,var(--color-primary)/6%,transparent)]"
              )}
            />
            <div className="flex items-start justify-between">
              <span
                className={cn(
                  "flex size-11 items-center justify-center rounded-xl transition-transform group-hover:scale-105",
                  tint
                )}
              >
                <Icon className="size-5" />
              </span>
              {isShared && (
                <span className="bg-primary/10 text-primary inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[11px] font-medium">
                  <UploadCloudIcon className="size-3" />
                  {t("filesPage.actions.upload")}
                </span>
              )}
            </div>
            <div className="min-w-0">
              <p className="truncate text-sm font-semibold">{t(`filesPage.folders.${type}`, { defaultValue: type })}</p>
              <p className="text-muted-foreground mt-0.5 text-xs">
                {t("filesPage.fileCount", { count: folder.file_count ?? 0 })}
                {(folder.total_size ?? 0) > 0 && <> · {formatBytes(folder.total_size ?? 0)}</>}
              </p>
            </div>
          </Link>
        )
      })}
    </div>
  )
}

// ownerName resolves the display label: server sends the entity name, but the
// nameless shared/other buckets are labelled client-side.
function ownerName(t: (k: string) => string, o: MediaOwner) {
  if (o.name) return o.name
  if (o.owner_kind === "shared") return t("filesPage.owner.sharedName")
  if (o.owner_kind === "other") return t("filesPage.owner.otherName")
  return "—"
}

function QuotaHeader({ used, limit, unlimited }: { used: number; limit: number; unlimited: boolean }) {
  const { t } = useTranslation()
  const pct = !unlimited && limit > 0 ? Math.min(100, Math.round((used / limit) * 100)) : 0
  return (
    <Card className="gap-2 p-4">
      <div className="flex items-center justify-between text-sm">
        <span className="text-muted-foreground">
          {unlimited
            ? t("filesPage.owner.quota.usedUnlimited", { used: formatBytes(used) })
            : t("filesPage.owner.quota.used", { used: formatBytes(used), limit: formatBytes(limit) })}
        </span>
        {unlimited ? (
          <span className="text-muted-foreground text-xs">{t("filesPage.owner.quota.unlimited")}</span>
        ) : (
          <span className="text-xs font-medium tabular-nums">{pct}%</span>
        )}
      </div>
      {!unlimited && <Progress value={pct} className="h-2" />}
    </Card>
  )
}

function OwnerNameCell({ row }: CellContext<MediaOwner, unknown>) {
  const { t } = useTranslation()
  const kind = row.original.owner_kind ?? "other"
  const { icon: Icon, tint } = ownerStyle(kind)
  return (
    <div className="flex min-w-0 items-center gap-3">
      <span className={cn("flex size-9 shrink-0 items-center justify-center rounded-lg", tint)}>
        <Icon className="size-4" />
      </span>
      <div className="min-w-0">
        <p className="truncate text-sm font-medium" dir="auto">
          {ownerName(t, row.original)}
        </p>
        <p className="text-muted-foreground text-xs">{t(`filesPage.owner.kinds.${kind}`, { defaultValue: kind })}</p>
      </div>
    </div>
  )
}

function OwnerFilesCell({ row }: CellContext<MediaOwner, unknown>) {
  return <span className="text-muted-foreground text-xs tabular-nums">{row.original.file_count ?? 0}</span>
}

function OwnerSizeCell({ row }: CellContext<MediaOwner, unknown>) {
  return <span className="text-sm font-medium tabular-nums">{formatBytes(row.original.total_size ?? 0)}</span>
}

function OwnerChevronCell() {
  return <ChevronRightIcon className="text-muted-foreground size-4 rtl:rotate-180" />
}

function ByOwnerView() {
  const { t } = useTranslation()
  const navigate = Route.useNavigate()
  const { page, page_size } = Route.useSearch()
  const currentPage = page ?? 1
  const pageSize = page_size ?? 20

  const { data, isLoading } = useGetFilesOwners({ page: currentPage, page_size: pageSize })
  const payload = (data?.status === 200 && data.data.data) || undefined
  const owners = (payload?.owners as MediaOwner[] | undefined) ?? []
  const total = payload?.total ?? 0
  const quota = payload?.quota

  const columns: ColumnDef<MediaOwner>[] = [
    {
      accessorKey: "name",
      header: t("filesPage.owner.columns.name"),
      cell: OwnerNameCell,
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: "file_count",
      header: t("filesPage.owner.columns.files"),
      cell: OwnerFilesCell,
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "total_size",
      header: t("filesPage.owner.columns.size"),
      cell: OwnerSizeCell,
      enableSorting: false,
      enableHiding: false,
    },
    {
      id: "chevron",
      header: "",
      cell: OwnerChevronCell,
      enableSorting: false,
      enableHiding: false,
    },
  ]

  const table = useAdminTable({ data: owners, columns, rowCount: total, sorting: [] })

  return (
    <div className="flex flex-col gap-4">
      {quota && (
        <QuotaHeader used={quota.used_bytes ?? 0} limit={quota.limit_bytes ?? 0} unlimited={quota.unlimited ?? false} />
      )}
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<FolderOpenIcon className="size-8 opacity-40" />}
            emptyTitle={t("filesPage.owner.empty")}
            onRowClick={(row) =>
              navigate({
                to: "/org/files/owner/$kind",
                params: { kind: row.owner_kind ?? "other" },
                search: {
                  ...(row.owner_id ? { owner_id: row.owner_id } : {}),
                  name: ownerName(t, row),
                },
              })
            }
          />
        </div>
        <DataTablePagination table={table} />
      </Card>
    </div>
  )
}
