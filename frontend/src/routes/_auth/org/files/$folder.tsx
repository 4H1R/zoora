import type { GetFilesOrderBy, GithubCom4H1RZooraInternalDomainMedia as Media } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon, DownloadIcon, FileIcon, Link2Icon, Trash2Icon, UploadCloudIcon } from "lucide-react"
import { useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetFilesFoldersQueryKey,
  getGetFilesQueryKey,
  getMediaIdDownloadUrl,
  useDeleteMediaId,
  useGetFiles,
  usePostMediaPresign,
} from "@/api/media/media"
import { useGetUsersMe } from "@/api/users/users"
import { ShareDialog } from "@/components/org/files/share-dialog"
import { MAX_SHARED_UPLOAD_BYTES, uploadSharedFile } from "@/components/org/files/upload-shared"
import { SHARED_FOLDER, folderStyle, formatBytes } from "@/components/org/files/utils"
import { DataTable } from "@/components/data-table/data-table"
import { DataTablePagination } from "@/components/data-table/data-table-pagination"
import { TableFilter } from "@/components/data-table/table-filter"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Spinner } from "@/components/ui/spinner"
import { useCanAny, useOrgGuard } from "@/lib/access"
import { adminSearchSchema, useAdminTable, useFormatDate } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org/files/$folder")({
  head: () => orgHead("org.nav.files"),
  validateSearch: adminSearchSchema,
  component: FolderPage,
})

interface FileRowActionsProps {
  media: Media
  canDelete: boolean
  onShare: (media: Media) => void
  onDelete: (media: Media) => void
}

function FileRowActions({ media, canDelete, onShare, onDelete }: FileRowActionsProps) {
  const { t } = useTranslation()
  const [downloading, setDownloading] = useState(false)

  const handleDownload = async () => {
    if (!media.id) return
    setDownloading(true)
    try {
      const res = await getMediaIdDownloadUrl(media.id)
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
      <Button variant="ghost" size="icon-xs" title={t("filesPage.actions.share")} onClick={() => onShare(media)}>
        <Link2Icon />
      </Button>
      {canDelete && (
        <Button
          variant="ghost"
          size="icon-xs"
          className="text-destructive hover:bg-destructive/10 hover:text-destructive"
          title={t("filesPage.actions.delete")}
          onClick={() => onDelete(media)}
        >
          <Trash2Icon />
        </Button>
      )}
    </div>
  )
}

function FolderPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { folder } = Route.useParams()
  const { search, order_by, order_dir, page, page_size } = Route.useSearch()
  const allowed = useOrgGuard(["media:view_any"])
  const canDelete = useCanAny(["media:delete_any"])
  const canUpload = useCanAny(["media:create"])

  const isShared = folder === SHARED_FOLDER
  const { icon: FolderGlyph, tint } = folderStyle(folder)
  const folderLabel = t(`filesPage.folders.${folder}`, { defaultValue: folder })

  const { data: meResponse } = useGetUsersMe()
  const orgId = (meResponse?.status === 200 && meResponse.data.data?.organization_id) || ""

  const currentPage = page ?? 1
  const pageSize = page_size ?? 20
  const sorting = order_by ? [{ id: order_by, desc: order_dir === "desc" }] : []

  const listParams = {
    model_type: folder,
    search: search || undefined,
    order_by: (order_by as GetFilesOrderBy) || undefined,
    order_dir: order_dir || undefined,
    page: currentPage,
    page_size: pageSize,
  }
  const { data, isLoading } = useGetFiles(listParams, { query: { enabled: allowed } })
  const filesData = (data?.status === 200 && data.data.data) || undefined
  const files = (filesData?.items as Media[] | undefined) ?? []
  const total = filesData?.total ?? 0

  const [shareTarget, setShareTarget] = useState<Media | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Media | null>(null)
  const [uploading, setUploading] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const presign = usePostMediaPresign()

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetFilesQueryKey(listParams) })
    queryClient.invalidateQueries({ queryKey: getGetFilesFoldersQueryKey() })
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

  const handleUploadFiles = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const picked = Array.from(e.target.files ?? [])
    e.target.value = ""
    if (picked.length === 0 || !orgId) return

    const oversize = picked.filter((f) => f.size > MAX_SHARED_UPLOAD_BYTES)
    if (oversize.length > 0) {
      toast.error(t("filesPage.upload.tooLarge"))
      return
    }

    setUploading(true)
    try {
      for (const file of picked) {
        await uploadSharedFile(presign.mutateAsync, orgId, file)
      }
      toast.success(t("filesPage.upload.success"))
      invalidate()
    } catch {
      toast.error(t("filesPage.upload.failed"))
    } finally {
      setUploading(false)
    }
  }

  const formatDate = useFormatDate()
  const columns: ColumnDef<Media>[] = [
    {
      accessorKey: "name",
      header: t("filesPage.columns.name"),
      cell: ({ row }) => (
        <div className="flex min-w-0 items-center gap-3">
          <span className={cn("flex size-9 shrink-0 items-center justify-center rounded-lg", tint)}>
            <FileIcon className="size-4" />
          </span>
          <div className="min-w-0">
            <p className="truncate text-sm font-medium" dir="auto">
              {row.original.name || row.original.file_name}
            </p>
            <p className="text-muted-foreground truncate font-mono text-[11px]" dir="ltr">
              {row.original.mime_type || "—"}
            </p>
          </div>
        </div>
      ),
      enableSorting: true,
      enableHiding: false,
    },
    {
      accessorKey: "size",
      header: t("filesPage.columns.size"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs tabular-nums">{formatBytes(row.original.size ?? 0)}</span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      accessorKey: "created_at",
      header: t("filesPage.columns.createdAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs">{formatDate(row.original.created_at)}</span>
      ),
      enableSorting: true,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <FileRowActions media={row.original} canDelete={canDelete} onShare={setShareTarget} onDelete={setDeleteTarget} />
      ),
      enableSorting: false,
      enableHiding: false,
    },
  ]

  const table = useAdminTable({ data: files, columns, rowCount: total, sorting })

  if (!allowed) return null

  return (
    <div className="flex flex-col gap-6">
      <Link
        to="/org/files"
        className="text-muted-foreground hover:text-foreground inline-flex w-fit items-center gap-1.5 text-xs font-medium transition-colors"
      >
        <ArrowLeftIcon className="size-3.5 rtl:rotate-180" />
        {t("filesPage.title")}
      </Link>

      <div className="-mt-3 flex items-end justify-between gap-6">
        <div className="flex min-w-0 items-center gap-3">
          <span className={cn("flex size-11 shrink-0 items-center justify-center rounded-xl", tint)}>
            <FolderGlyph className="size-5" />
          </span>
          <div className="min-w-0">
            <h1 className="truncate text-2xl font-bold tracking-tight">{folderLabel}</h1>
            <p className="text-muted-foreground text-xs">{t("filesPage.fileCount", { count: total })}</p>
          </div>
        </div>
        {isShared && canUpload && (
          <div className="flex shrink-0 items-center gap-2">
            <input ref={inputRef} type="file" multiple hidden onChange={handleUploadFiles} />
            <Button size="sm" disabled={uploading || !orgId} onClick={() => inputRef.current?.click()}>
              {uploading ? (
                <Spinner className="size-4" data-icon="inline-start" />
              ) : (
                <UploadCloudIcon data-icon="inline-start" />
              )}
              {t("filesPage.actions.upload")}
            </Button>
          </div>
        )}
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
            emptyIcon={<FolderGlyph className="size-8 opacity-40" />}
            emptyTitle={t("filesPage.empty")}
          />
        </div>
        <DataTablePagination table={table} />
      </Card>

      <ShareDialog media={shareTarget} onOpenChange={(open) => !open && setShareTarget(null)} />

      <DeleteConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open: boolean) => {
          if (!open) setDeleteTarget(null)
        }}
        resourceName={t("filesPage.deleteConfirm.description", {
          name: deleteTarget?.name || deleteTarget?.file_name || "",
          folder: folderLabel,
        })}
        onConfirm={() => {
          if (deleteTarget?.id) deleteMutation.mutate({ id: deleteTarget.id })
        }}
        isLoading={deleteMutation.isPending}
      />
    </div>
  )
}
