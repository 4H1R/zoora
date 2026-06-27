import type { GithubCom4H1RZooraInternalDomainLiveRecording as LiveRecording } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { CircleIcon, DownloadIcon, FilmIcon, Trash2Icon, VideoIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetLiveRoomsIdRecordingsQueryKey,
  useDeleteLiveRoomsIdRecordingsRecordingId,
  useGetLiveRoomsIdRecordings,
  usePostLiveRoomsIdRecordings,
} from "@/api/live-sessions/live-sessions"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { formatSessionDate } from "@/lib/session-status"
import { cn } from "@/lib/utils"

const STATUS_STYLES: Record<string, string> = {
  started: "bg-amber-500/10 text-amber-600 dark:text-amber-400",
  completed: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400",
  failed: "bg-destructive/10 text-destructive",
}

function formatSize(bytes: number | undefined): string {
  const b = bytes ?? 0
  if (b <= 0) return "—"
  const units = ["B", "KB", "MB", "GB"]
  let v = b
  let i = 0
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

interface LiveRoomRecordingsProps {
  roomId: string
  isActive: boolean
  canManage: boolean
}

export function LiveRoomRecordings({ roomId, isActive, canManage }: LiveRoomRecordingsProps) {
  const { t, i18n } = useTranslation()
  const formatDate = (iso?: string) => formatSessionDate(iso, i18n.language, "short")
  const queryClient = useQueryClient()
  const [deleting, setDeleting] = useState<LiveRecording | null>(null)

  const query = useGetLiveRoomsIdRecordings(roomId, undefined, { query: { enabled: !!roomId } })
  const recordings = (query.data?.status === 200 && query.data.data.data?.items) || []

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getGetLiveRoomsIdRecordingsQueryKey(roomId) })

  const startMutation = usePostLiveRoomsIdRecordings({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.liveRooms.recordings.startSuccess"))
        invalidate()
      },
    },
  })

  const deleteMutation = useDeleteLiveRoomsIdRecordingsRecordingId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.liveRooms.recordings.deleteSuccess"))
        invalidate()
        setDeleting(null)
      },
    },
  })

  const recording = recordings.some((r) => r.status === "started")

  return (
    <div className="flex flex-col gap-3">
      {canManage && isActive && (
        <div className="flex justify-end">
          <Button
            size="sm"
            variant="outline"
            disabled={startMutation.isPending || recording}
            onClick={() => startMutation.mutate({ id: roomId })}
          >
            <VideoIcon className="size-3.5" />
            {recording ? t("org.session.liveRooms.recordings.recording") : t("org.session.liveRooms.recordings.start")}
          </Button>
        </div>
      )}

      {query.isPending ? (
        <Skeleton className="h-12 w-full rounded-xl" />
      ) : recordings.length === 0 ? (
        <div className="text-muted-foreground border-foreground/10 flex flex-col items-center gap-2 rounded-xl border border-dashed px-4 py-8 text-center">
          <FilmIcon className="size-5 opacity-60" />
          <span className="font-mono text-[10px] tracking-[0.2em] uppercase">
            {t("org.session.liveRooms.recordings.empty")}
          </span>
        </div>
      ) : (
        <ul className="flex flex-col gap-1.5">
          {recordings.map((r) => {
            const status = r.status ?? "started"
            return (
              <li key={r.id} className="bg-muted/40 flex items-center gap-3 rounded-xl px-3 py-2.5">
                <span
                  className={cn(
                    "inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 font-mono text-[10px] tracking-[0.2em] uppercase",
                    STATUS_STYLES[status] ?? "bg-muted text-muted-foreground"
                  )}
                >
                  {status === "started" && <CircleIcon className="size-2 animate-pulse fill-current" />}
                  {t(`org.session.liveRooms.recordings.status.${status}`)}
                </span>
                <div className="flex min-w-0 flex-1 flex-col">
                  <span className="truncate font-mono text-xs tabular-nums">{formatDate(r.started_at)}</span>
                  <span className="text-muted-foreground font-mono text-[10px]">{formatSize(r.size)}</span>
                </div>
                {Boolean(r.file_url) && (
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    render={<a href={r.file_url} target="_blank" rel="noreferrer" />}
                  >
                    <DownloadIcon />
                  </Button>
                )}
                {canManage && (
                  <Button
                    variant="ghost"
                    size="icon-xs"
                    className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                    onClick={() => setDeleting(r)}
                  >
                    <Trash2Icon />
                  </Button>
                )}
              </li>
            )
          })}
        </ul>
      )}

      <DeleteConfirmDialog
        open={!!deleting}
        onOpenChange={(open) => {
          if (deleteMutation.isPending) return
          if (!open) setDeleting(null)
        }}
        resourceName={deleting ? formatDate(deleting.started_at) : ""}
        onConfirm={() => {
          if (deleting?.id) deleteMutation.mutate({ id: roomId, recordingId: deleting.id })
        }}
        isLoading={deleteMutation.isPending}
      />
    </div>
  )
}
