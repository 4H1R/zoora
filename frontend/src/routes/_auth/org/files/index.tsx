import { createFileRoute, Link } from "@tanstack/react-router"
import { FolderOpenIcon, UploadCloudIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetFilesFolders } from "@/api/media/media"
import { SHARED_FOLDER, folderStyle, formatBytes } from "@/components/org/files/utils"
import { PageHeader } from "@/components/page-header"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org/files/")({
  head: () => orgHead("org.nav.files"),
  component: FilesPage,
})

function FilesPage() {
  const { t } = useTranslation()
  const allowed = useOrgGuard(["media:view_any"])

  const { data, isLoading } = useGetFilesFolders({ query: { enabled: allowed } })
  const folders = (data?.status === 200 && data.data.data) || []

  // Shared folder is pinned first — it's the only folder that accepts uploads.
  const sorted = [...folders].sort((a, b) =>
    a.model_type === SHARED_FOLDER ? -1 : b.model_type === SHARED_FOLDER ? 1 : 0
  )

  if (!allowed) return null

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("filesPage.title")} />
      <p className="text-muted-foreground -mt-4 text-sm">{t("filesPage.description")}</p>

      {isLoading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-36 rounded-2xl" />
          ))}
        </div>
      ) : sorted.length === 0 ? (
        <EmptyState icon={FolderOpenIcon} title={t("filesPage.empty")} />
      ) : (
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
                  "hover:border-primary/40 hover:shadow-md hover:-translate-y-0.5",
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
                  <p className="truncate text-sm font-semibold">
                    {t(`filesPage.folders.${type}`, { defaultValue: type })}
                  </p>
                  <p className="text-muted-foreground mt-0.5 text-xs">
                    {t("filesPage.fileCount", { count: folder.file_count ?? 0 })}
                    {(folder.total_size ?? 0) > 0 && <> · {formatBytes(folder.total_size ?? 0)}</>}
                  </p>
                </div>
              </Link>
            )
          })}
        </div>
      )}
    </div>
  )
}
