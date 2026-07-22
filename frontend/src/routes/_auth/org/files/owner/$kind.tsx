import type {
  GetFilesOwnersKindFilesOrderBy,
  GithubCom4H1RZooraInternalDomainMedia as Media,
  GithubCom4H1RZooraInternalDomainOwnerFile as OwnerFile,
} from "@/api/model"
import type { CellContext, ColumnDef } from "@tanstack/react-table"
import type { TFunction } from "i18next"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon, DownloadIcon, FileIcon, Link2Icon, LockIcon, Trash2Icon, VideoIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetFilesOwnersKindFilesQueryKey,
  getGetFilesOwnersQueryKey,
  getMediaIdDownloadUrl,
  useDeleteMediaId,
  useGetFilesOwnersKindFiles,
} from "@/api/media/media"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { ShareDialog } from "@/components/org/files/share-dialog"
import { formatBytes, ownerStyle } from "@/components/org/files/utils"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Spinner } from "@/components/ui/spinner"
import { useCanAny, useOrgGuard } from "@/lib/access"
import { adminSearchSchema, useAdminTable, useFormatDate } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"
import { cn } from "@/lib/utils"

const OWNER_KINDS = ["class", "question_bank", "conversation", "shared", "other"] as const
type OwnerKind = (typeof OWNER_KINDS)[number]

const ownerSearchSchema = adminSearchSchema.extend({
  owner_id: z.string().optional(),
  name: z.string().optional(),
})

export const Route = createFileRoute("/_auth/org/files/owner/$kind")({
  head: () => orgHead("org.nav.files"),
  validateSearch: ownerSearchSchema,
  component: OwnerPage,
})

interface OwnerFileRowActionsProps {
  file: OwnerFile
  canDelete: boolean
  onShare: (file: OwnerFile) => void
  onDelete: (file: OwnerFile) => void
}

function OwnerFileRowActions({ file, canDelete, onShare, onDelete }: OwnerFileRowActionsProps) {
  const { t } = useTranslation()
  const [downloading, setDownloading] = useState(false)

  // Recordings are read-only in this view — no per-recording delete/download
  // endpoint exists yet (managed in class recordings).
  if (file.source !== "media") {
    return (
      <div className="flex items-center justify-end">
        <span
          className="text-muted-foreground inline-flex items-center gap-1 text-[11px]"
          title={t("filesPage.owner.recordingReadOnly")}
        >
          <LockIcon className="size-3" />
        </span>
      </div>
    )
  }

  const handleDownload = async () => {
    if (!file.id) return
    setDownloading(true)
    try {
      const res = await getMediaIdDownloadUrl(file.id)
      const url = res.status === 200 ? res.data.data?.url : undefined
      if (!url) throw new Error("download url failed")
      window.open(url, "_blank", "noopener")
    } catch {
      toast.error(t("filesPage.upload.failed"))
    } finally {
      setDownloading(false)
    }
  }

  return (
    <div className="flex items-center justify-end gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100">
      <Button
        variant="ghost"
        size="icon-xs"
        title={t("filesPage.actions.download")}
        disabled={downloading}
        onClick={handleDownload}
      >
        {downloading ? <Spinner className="size-3.5" /> : <DownloadIcon />}
      </Button>
      <Button variant="ghost" size="icon-xs" title={t("filesPage.actions.share")} onClick={() => onShare(file)}>
        <Link2Icon />
      </Button>
      {canDelete && (
        <Button
          variant="ghost"
          size="icon-xs"
          className="text-destructive hover:bg-destructive/10 hover:text-destructive"
          title={t("filesPage.actions.delete")}
          onClick={() => onDelete(file)}
        >
          <Trash2Icon />
        </Button>
      )}
    </div>
  )
}

function OwnerFileNameCell({ file, tint }: { file: OwnerFile; tint: string }) {
  const isRecording = file.source !== "media"
  return (
    <div className="flex min-w-0 items-center gap-3">
      <span className={cn("flex size-9 shrink-0 items-center justify-center rounded-lg", tint)}>
        {isRecording ? <VideoIcon className="size-4" /> : <FileIcon className="size-4" />}
      </span>
      <div className="min-w-0">
        <p className="truncate text-sm font-medium" dir="auto">
          {file.name}
        </p>
        <p className="text-muted-foreground truncate font-mono text-[11px]" dir="ltr">
          {file.mime_type || "—"}
        </p>
      </div>
    </div>
  )
}

function OwnerFileKindCell({ row }: CellContext<OwnerFile, unknown>) {
  const { t } = useTranslation()
  return row.original.source !== "media" ? (
    <Badge variant="secondary" className="gap-1 text-[11px]">
      <VideoIcon className="size-3" />
      {t("filesPage.owner.recordingBadge")}
    </Badge>
  ) : (
    <span className="text-muted-foreground text-xs">
      {t(`filesPage.folders.${row.original.model_type}`, { defaultValue: row.original.model_type })}
    </span>
  )
}

function OwnerFileSizeCell({ row }: CellContext<OwnerFile, unknown>) {
  return <span className="text-muted-foreground text-xs tabular-nums">{formatBytes(row.original.size ?? 0)}</span>
}

function OwnerFileCreatedAtCell({ row }: CellContext<OwnerFile, unknown>) {
  const formatDate = useFormatDate()
  return <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>
}

interface OwnerFileColumnsDeps {
  t: TFunction
  tint: string
  canDelete: boolean
  onShare: (file: OwnerFile) => void
  onDelete: (file: OwnerFile) => void
}

function buildOwnerFileColumns({
  t,
  tint,
  canDelete,
  onShare,
  onDelete,
}: OwnerFileColumnsDeps): ColumnDef<OwnerFile>[] {
  return [
    {
      accessorKey: "name",
      header: t("filesPage.owner.fileColumns.name"),
      cell: ({ row }) => <OwnerFileNameCell file={row.original} tint={tint} />,
      enableSorting: true,
      enableHiding: false,
    },
    {
      id: "kind",
      header: t("filesPage.owner.fileColumns.kind"),
      cell: OwnerFileKindCell,
      enableSorting: false,
      enableHiding: true,
    },
    {
      accessorKey: "size",
      header: t("filesPage.owner.fileColumns.size"),
      cell: OwnerFileSizeCell,
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: "created_at",
      header: t("filesPage.owner.fileColumns.createdAt"),
      cell: OwnerFileCreatedAtCell,
      enableSorting: true,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <OwnerFileRowActions file={row.original} canDelete={canDelete} onShare={onShare} onDelete={onDelete} />
      ),
      enableSorting: false,
      enableHiding: false,
    },
  ]
}

function OwnerPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { kind } = Route.useParams()
  const { owner_id, name, search, order_by, order_dir, page, page_size } = Route.useSearch()
  const allowed = useOrgGuard(["media:view_any"])
  const canDelete = useCanAny(["media:delete_any"])

  const ownerKind = (OWNER_KINDS as readonly string[]).includes(kind) ? (kind as OwnerKind) : "other"
  const { icon: OwnerGlyph, tint } = ownerStyle(ownerKind)
  const kindLabel = t(`filesPage.owner.kinds.${ownerKind}`, { defaultValue: ownerKind })
  const title = name || kindLabel

  const currentPage = page ?? 1
  const pageSize = page_size ?? 20
  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const listParams = {
    owner_id: owner_id || undefined,
    search: search || undefined,
    order_by: (order_by as GetFilesOwnersKindFilesOrderBy) || undefined,
    order_dir: order_dir || undefined,
    page: currentPage,
    page_size: pageSize,
  }
  const { data, isLoading } = useGetFilesOwnersKindFiles(ownerKind, listParams, { query: { enabled: allowed } })
  const filesData = (data?.status === 200 && data.data.data) || undefined
  const files = (filesData?.items as OwnerFile[] | undefined) ?? []
  const total = filesData?.total ?? 0

  const [shareTarget, setShareTarget] = useState<OwnerFile | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<OwnerFile | null>(null)

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetFilesOwnersKindFilesQueryKey(ownerKind, listParams) })
    queryClient.invalidateQueries({ queryKey: getGetFilesOwnersQueryKey() })
  }

  const deleteMutation = useDeleteMediaId({
    mutation: {
      onSuccess: () => {
        toast.success(t("filesPage.actions.delete"))
        invalidate()
        setDeleteTarget(null)
      },
    },
  })

  const columns = buildOwnerFileColumns({
    t,
    tint,
    canDelete,
    onShare: setShareTarget,
    onDelete: setDeleteTarget,
  })

  const table = useAdminTable({ data: files, columns, rowCount: total, sorting })

  if (!allowed) return null

  // ShareDialog wants a Media; owner files that are media carry the media id.
  const shareMedia: Media | null = shareTarget
    ? { id: shareTarget.id, name: shareTarget.name, file_name: shareTarget.name }
    : null

  return (
    <div className="flex flex-col gap-6">
      <Link
        to="/org/files"
        search={{ view: "owner" }}
        className="text-muted-foreground hover:text-foreground inline-flex w-fit items-center gap-1.5 text-xs font-medium transition-colors"
      >
        <ArrowLeftIcon className="size-3.5 rtl:rotate-180" />
        {t("filesPage.title")}
      </Link>

      <div className="-mt-3 flex min-w-0 items-center gap-3">
        <span className={cn("flex size-11 shrink-0 items-center justify-center rounded-xl", tint)}>
          <OwnerGlyph className="size-5" />
        </span>
        <div className="min-w-0">
          <h1 className="truncate text-2xl font-bold tracking-tight" dir="auto">
            {title}
          </h1>
          <p className="text-muted-foreground text-xs">
            {kindLabel} · {t("filesPage.fileCount", { count: total })}
          </p>
        </div>
      </div>

      <TableFilter
        table={table}
        searchPlaceholder={t("filesPage.searchPlaceholder")}
        sortLabel={t("common.toolbar.sort")}
        columnsLabel={t("common.toolbar.columns")}
        toggleColumnsLabel={t("common.toolbar.toggleColumns")}
      />
      <Card className="gap-0 overflow-hidden p-0">
        <div className="overflow-x-auto">
          <DataTable
            table={table}
            isLoading={isLoading}
            emptyIcon={<OwnerGlyph className="size-8 opacity-40" />}
            emptyTitle={t("filesPage.empty")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>

      <ShareDialog media={shareMedia} onOpenChange={(open) => !open && setShareTarget(null)} />

      <DeleteConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open: boolean) => {
          if (!open) setDeleteTarget(null)
        }}
        resourceName={t("filesPage.deleteConfirm.description", {
          name: deleteTarget?.name || "",
          folder: title,
        })}
        onConfirm={() => {
          if (deleteTarget?.id) deleteMutation.mutate({ id: deleteTarget.id })
        }}
        isLoading={deleteMutation.isPending}
      />
    </div>
  )
}
