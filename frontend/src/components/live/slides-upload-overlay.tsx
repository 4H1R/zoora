import { FileText, Loader2, X } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Progress } from "@/components/ui/progress"
import { cn } from "@/lib/utils"

// Upload lifecycle for a host-shared PDF. `preparing` = presigning (no bytes on
// the wire yet), `uploading` = PUT in flight with a real byte percentage,
// `processing` = uploaded, resolving the shared download URL. Only `preparing`
// and `uploading` are abortable — once `processing`, the bytes are already up.
export type SlidesUpload = {
  fileName: string
  phase: "preparing" | "uploading" | "processing"
  progress: number // 0–100; only meaningful during `uploading`
}

// A centered glass card over the stage that reports upload progress so the host
// isn't staring at a dead screen wondering whether anything is happening. Matches
// the dark glass pills used elsewhere in the slides stage (black/60 + blur).
export function SlidesUploadOverlay({ upload, onCancel }: { upload: SlidesUpload; onCancel: () => void }) {
  const { t } = useTranslation()
  const { fileName, phase, progress } = upload
  const abortable = phase !== "processing"

  const status =
    phase === "preparing"
      ? t("liveRoom.stage.preparing")
      : phase === "processing"
        ? t("liveRoom.stage.processing")
        : t("liveRoom.stage.uploading")

  return (
    <div className="absolute inset-0 z-30 flex items-center justify-center p-4">
      <div
        className={cn(
          "animate-in fade-in zoom-in-95 w-full max-w-sm rounded-2xl border border-white/10 bg-black/70 p-5",
          "shadow-2xl ring-1 shadow-black/50 ring-white/5 backdrop-blur-md duration-200"
        )}
      >
        <div className="flex items-center gap-3">
          <div className="grid size-10 shrink-0 place-items-center rounded-xl bg-white/10 text-white">
            <FileText className="size-5" />
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium text-white" dir="ltr" title={fileName}>
              {fileName}
            </p>
            <div className="mt-0.5 flex items-center gap-1.5 text-xs text-zinc-400">
              {phase !== "uploading" && <Loader2 className="size-3 animate-spin" />}
              <span>{status}</span>
            </div>
          </div>
          {abortable && (
            <button
              type="button"
              onClick={onCancel}
              aria-label={t("liveRoom.stage.cancelUpload")}
              title={t("liveRoom.stage.cancelUpload")}
              className="grid size-8 shrink-0 place-items-center rounded-full text-zinc-300 transition-all hover:bg-white/10 hover:text-white active:scale-90"
            >
              <X className="size-4" />
            </button>
          )}
        </div>

        <div className="mt-4 flex items-center gap-3">
          <Progress
            value={phase === "processing" ? 100 : phase === "uploading" ? progress : null}
            className="flex-1 [&_[data-slot=progress-indicator]]:bg-white [&_[data-slot=progress-track]]:bg-white/15"
          />
          {phase === "uploading" && (
            <span className="w-9 shrink-0 text-end text-xs font-medium text-zinc-300 tabular-nums">{progress}%</span>
          )}
        </div>
      </div>
    </div>
  )
}
