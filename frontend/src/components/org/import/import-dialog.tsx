import type { GithubCom4H1RZooraInternalDomainImportJob as ImportJob } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { DownloadIcon, FileSpreadsheetIcon, TriangleAlertIcon, UploadIcon } from "lucide-react"
import { useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { getGetImportsLatestQueryKey, useGetImportsLatest, usePostImports } from "@/api/imports/imports"
import { usePostMediaPresign } from "@/api/media/media"
import { AUTH_TOKEN_KEY } from "@/api/mutator/custom-instance"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Progress } from "@/components/ui/progress"
import { Spinner } from "@/components/ui/spinner"
import { clientEnv } from "@/config/env"

type ImportType = "users" | "classes" | "class_members"

interface ImportDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  type: ImportType
}

const TEMPLATE_PATHS: Record<ImportType, string> = {
  users: "/templates/users-import-template.xlsx",
  classes: "/templates/classes-import-template.xlsx",
  class_members: "/templates/class-members-import-template.xlsx",
}

const TITLE_KEYS: Record<ImportType, string> = {
  users: "org.import.usersTitle",
  classes: "org.import.classesTitle",
  class_members: "org.import.classMembersTitle",
}

const COLUMNS_KEYS: Record<ImportType, string> = {
  users: "org.import.columnsUsers",
  classes: "org.import.columnsClasses",
  class_members: "org.import.columnsClassMembers",
}
const MAX_SIZE = 10 * 1024 * 1024

const isRunning = (job?: ImportJob | null) => job?.status === "pending" || job?.status === "processing"

export function ImportDialog({ open, onOpenChange, type }: ImportDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const inputRef = useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = useState(false)
  const [downloading, setDownloading] = useState(false)
  // job started (or found running) while this dialog instance is open —
  // switches the view away from the upload phase
  const [activeJobId, setActiveJobId] = useState<string | null>(null)

  const presign = usePostMediaPresign()
  const createImport = usePostImports({
    mutation: {
      onSuccess: (res) => {
        if (res.status === 201 && res.data.data?.id) setActiveJobId(res.data.data.id)
        queryClient.invalidateQueries({ queryKey: getGetImportsLatestQueryKey({ type }) })
      },
      onError: () => toast.error(t("org.import.uploadError")),
    },
  })

  const { data: latestRes } = useGetImportsLatest(
    { type },
    {
      query: {
        enabled: open,
        refetchInterval: (query) => {
          const res = query.state.data
          const job = res?.status === 200 ? res.data.data : undefined
          return isRunning(job) ? 2000 : false
        },
      },
    }
  )
  const job = latestRes?.status === 200 ? latestRes.data.data : undefined

  // resume: reopening while a job runs jumps straight to the progress phase
  const showJob = job && (isRunning(job) || job.id === activeJobId)

  const pickFile = () => inputRef.current?.click()

  const handleFile = async (file: File) => {
    if (uploading || createImport.isPending) return
    if (!file.name.toLowerCase().endsWith(".xlsx")) {
      toast.error(t("org.import.invalidType"))
      return
    }
    if (file.size > MAX_SIZE) {
      toast.error(t("org.import.fileTooLarge"))
      return
    }
    setUploading(true)
    try {
      const res = await presign.mutateAsync({
        data: {
          model_type: "import",
          model_id: crypto.randomUUID(),
          collection_name: "file",
          file_name: file.name,
          mime_type: file.type || "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
          size: file.size,
        },
      })
      const uploadURL = res.status === 201 ? res.data.data?.upload_url : undefined
      const mediaId = res.status === 201 ? res.data.data?.media?.id : undefined
      if (!uploadURL || !mediaId) throw new Error("presign failed")
      const put = await fetch(uploadURL, {
        method: "PUT",
        body: file,
        headers: { "Content-Type": file.type || "application/octet-stream" },
      })
      if (!put.ok) throw new Error(`upload failed: ${put.status}`)
      createImport.mutate({ data: { type, media_id: mediaId } })
    } catch (err) {
      console.error(err)
      toast.error(t("org.import.uploadError"))
    } finally {
      setUploading(false)
    }
  }

  const handleDownload = async () => {
    if (!job?.id) return
    setDownloading(true)
    try {
      // same base URL + bearer header as src/api/mutator/custom-instance.ts
      const res = await fetch(`${clientEnv.VITE_API_URL}/imports/${job.id}/result`, {
        headers: { Authorization: `Bearer ${localStorage.getItem(AUTH_TOKEN_KEY) ?? ""}` },
      })
      if (res.status === 404) {
        toast.error(t("org.import.resultExpired"))
        return
      }
      if (!res.ok) throw new Error(`download failed: ${res.status}`)
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = `import-result-${job.id}.xlsx`
      a.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      console.error(err)
      toast.error(t("org.import.downloadError"))
    } finally {
      setDownloading(false)
    }
  }

  const total = job?.total_rows ?? 0
  const processed = job?.processed_rows ?? 0
  const percent = total > 0 ? Math.round((processed / total) * 100) : 0

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t(TITLE_KEYS[type])}</DialogTitle>
          <DialogDescription>{t(COLUMNS_KEYS[type])}</DialogDescription>
        </DialogHeader>

        {!showJob ? (
          <div className="flex flex-col gap-4">
            <button
              type="button"
              onClick={pickFile}
              onDragOver={(e) => e.preventDefault()}
              onDrop={(e) => {
                e.preventDefault()
                const file = e.dataTransfer.files?.[0]
                if (file) void handleFile(file)
              }}
              className="text-muted-foreground hover:bg-accent flex flex-col items-center gap-2 rounded-lg border border-dashed p-8"
              disabled={uploading || createImport.isPending}
            >
              {uploading || createImport.isPending ? <Spinner /> : <UploadIcon className="size-6" />}
              <span className="text-sm">{t("org.import.dropHint")}</span>
            </button>
            <input
              ref={inputRef}
              type="file"
              accept=".xlsx"
              className="hidden"
              disabled={uploading || createImport.isPending}
              onChange={(e) => {
                const file = e.target.files?.[0]
                e.target.value = ""
                if (file) void handleFile(file)
              }}
            />
            <a
              href={TEMPLATE_PATHS[type]}
              download
              className="text-primary inline-flex items-center gap-2 text-sm hover:underline"
            >
              <FileSpreadsheetIcon className="size-4" />
              {t("org.import.template")}
            </a>
          </div>
        ) : isRunning(job) ? (
          <div className="flex flex-col gap-3">
            <div className="text-muted-foreground flex items-center gap-2 text-sm">
              <Spinner />
              {t("org.import.processing")}
            </div>
            <Progress value={percent} />
            <p className="text-muted-foreground text-sm">{t("org.import.progress", { processed, total })}</p>
            <CountRow job={job} />
          </div>
        ) : (
          <div className="flex flex-col gap-3">
            <p className="text-sm font-medium">
              {job?.status === "failed"
                ? `${t("org.import.failedJob")}${job.error ? `: ${job.error}` : ""}`
                : t("org.import.done")}
            </p>
            <CountRow job={job} />
            {job?.status === "completed" && (
              <>
                <div className="flex items-start gap-2 rounded-md bg-amber-500/10 p-3 text-sm text-amber-600 dark:text-amber-400">
                  <TriangleAlertIcon className="mt-0.5 size-4 shrink-0" />
                  {t("org.import.resultWarning")}
                </div>
                <Button onClick={handleDownload} disabled={downloading}>
                  {downloading ? <Spinner /> : <DownloadIcon data-icon="inline-start" />}
                  {t("org.import.downloadResult")}
                </Button>
              </>
            )}
            <DialogFooter>
              <Button variant="outline" onClick={() => setActiveJobId(null)}>
                {t("org.import.newImport")}
              </Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}

function CountRow({ job }: { job?: ImportJob | null }) {
  const { t } = useTranslation()
  if (!job) return null
  return (
    <div className="flex gap-4 text-sm">
      <span className="text-emerald-600 dark:text-emerald-400">
        {t("org.import.created")}: {job.created_count ?? 0}
      </span>
      <span className="text-muted-foreground">
        {t("org.import.skipped")}: {job.skipped_count ?? 0}
      </span>
      <span className="text-destructive">
        {t("org.import.failed")}: {job.failed_count ?? 0}
      </span>
    </div>
  )
}
