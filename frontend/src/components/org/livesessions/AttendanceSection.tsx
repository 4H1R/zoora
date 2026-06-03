import type { GithubCom4H1RZooraInternalDomainAttendance as Attendance } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import {
  CheckCircle2Icon,
  ClockIcon,
  PencilIcon,
  ShieldCheckIcon,
  SparklesIcon,
  Trash2Icon,
  UserCheckIcon,
  XCircleIcon,
} from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetClassesIdSessionsSessionIdAttendanceQueryKey,
  useDeleteAttendanceAttendanceId,
  useGetClassesIdSessionsSessionIdAttendance,
  usePostClassesIdSessionsSessionIdAttendanceAutoMark,
} from "@/api/attendance/attendance"
import { useGetLiveRooms } from "@/api/live-sessions/live-sessions"
import { Eyebrow } from "@/components/eyebrow"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Skeleton } from "@/components/ui/skeleton"
import { getEntityColor, getInitials } from "@/lib/data-table"
import { cn } from "@/lib/utils"

import { AttendanceEditDialog } from "./AttendanceEditDialog"
import { useAttendancePermissions } from "./use-attendance-permissions"

const STATUS_META: Record<string, { style: string; icon: typeof CheckCircle2Icon }> = {
  present: { style: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400", icon: CheckCircle2Icon },
  absent: { style: "bg-destructive/10 text-destructive", icon: XCircleIcon },
  late: { style: "bg-amber-500/10 text-amber-600 dark:text-amber-400", icon: ClockIcon },
  excused: { style: "bg-primary/10 text-primary", icon: ShieldCheckIcon },
}

interface AutoMarkControlProps {
  classId: string
  classSessionId: string
}

function AutoMarkControl({ classId, classSessionId }: AutoMarkControlProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [roomId, setRoomId] = useState("")
  const [minutes, setMinutes] = useState(5)

  const roomsQuery = useGetLiveRooms({ class_session_id: classSessionId })
  const rooms = (roomsQuery.data?.status === 200 && roomsQuery.data.data.data?.items) || []

  const autoMark = usePostClassesIdSessionsSessionIdAttendanceAutoMark({
    mutation: {
      onSuccess: (result) => {
        const data = (result.status === 200 && result.data.data) || undefined
        toast.success(
          t("org.session.attendance.autoMark.success", {
            marked: data?.marked ?? 0,
            skipped: data?.skipped ?? 0,
          })
        )
        queryClient.invalidateQueries({
          queryKey: getGetClassesIdSessionsSessionIdAttendanceQueryKey(classId, classSessionId),
        })
      },
    },
  })

  const run = () => {
    if (!roomId) return
    autoMark.mutate({
      id: classId,
      sessionId: classSessionId,
      data: { source: "live_room", room_id: roomId, min_duration_seconds: Math.max(0, minutes) * 60 },
    })
  }

  return (
    <div className="bg-card ring-foreground/10 relative isolate flex flex-col gap-4 overflow-hidden rounded-2xl p-5 ring-1">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_right,var(--color-primary)/8%,transparent_60%)]"
      />
      <div className="flex flex-col gap-1.5">
        <Eyebrow className="inline-flex items-center gap-2">
          <SparklesIcon className="size-3.5" />
          {t("org.session.attendance.autoMark.eyebrow")}
        </Eyebrow>
        <p className="text-muted-foreground text-sm leading-relaxed">{t("org.session.attendance.autoMark.hint")}</p>
      </div>

      <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
        <div className="flex flex-1 flex-col gap-1.5">
          <FieldLabelText>{t("org.session.attendance.autoMark.room")}</FieldLabelText>
          <Select value={roomId} onValueChange={(v) => setRoomId(v ?? "")}>
            <SelectTrigger>
              <SelectValue placeholder={t("org.session.attendance.autoMark.roomPlaceholder")} />
            </SelectTrigger>
            <SelectContent>
              {rooms.map((r) => (
                <SelectItem key={r.id} value={r.id ?? ""}>
                  {r.livekit_room_name ?? r.id?.slice(0, 8).toUpperCase() ?? "—"}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="flex flex-col gap-1.5 sm:w-40">
          <FieldLabelText>{t("org.session.attendance.autoMark.minMinutes")}</FieldLabelText>
          <Input type="number" min={0} value={minutes} onChange={(e) => setMinutes(Number(e.target.value))} />
        </div>
        <Button disabled={!roomId || autoMark.isPending} onClick={run}>
          <UserCheckIcon className="size-4" />
          {t("org.session.attendance.autoMark.run")}
        </Button>
      </div>
    </div>
  )
}

function FieldLabelText({ children }: { children: React.ReactNode }) {
  return <span className="text-muted-foreground font-mono text-[10px] tracking-[0.25em] uppercase">{children}</span>
}

interface AttendanceRowProps {
  attendance: Attendance
  canEdit: boolean
  canDelete: boolean
  onEdit: (a: Attendance) => void
  onDelete: (a: Attendance) => void
}

function AttendanceRow({ attendance, canEdit, canDelete, onEdit, onDelete }: AttendanceRowProps) {
  const { t } = useTranslation()
  const name = attendance.user?.name ?? "—"
  const status = attendance.status ?? "absent"
  const meta = STATUS_META[status] ?? STATUS_META.absent
  const StatusIcon = meta.icon

  return (
    <div className="group/row bg-card ring-foreground/10 hover:ring-foreground/25 flex items-center gap-3 rounded-2xl px-4 py-3 ring-1 transition-all">
      <Avatar className="size-9 shrink-0">
        <AvatarFallback className={cn("text-xs font-semibold text-white", getEntityColor(name))}>
          {getInitials(name)}
        </AvatarFallback>
      </Avatar>

      <div className="flex min-w-0 flex-1 flex-col">
        <span className="truncate text-sm font-medium">{name}</span>
        {attendance.remarks ? (
          <span className="text-muted-foreground truncate text-xs">{attendance.remarks}</span>
        ) : null}
      </div>

      {attendance.is_auto_marked ? (
        <span className="text-muted-foreground hidden items-center gap-1 font-mono text-[10px] tracking-[0.2em] uppercase sm:inline-flex">
          <SparklesIcon className="size-3" />
          {t("org.session.attendance.auto")}
        </span>
      ) : (
        <span className="text-muted-foreground hidden items-center gap-1 font-mono text-[10px] tracking-[0.2em] uppercase sm:inline-flex">
          {t("org.session.attendance.manual")}
        </span>
      )}

      <span
        className={cn(
          "inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 font-mono text-[10px] tracking-[0.2em] uppercase",
          meta.style
        )}
      >
        <StatusIcon className="size-3" />
        {t(`org.session.attendance.status.${status}`)}
      </span>

      <div className="flex items-center gap-0.5 opacity-100 transition-opacity sm:opacity-0 sm:group-hover/row:opacity-100">
        {canEdit ? (
          <Button variant="ghost" size="icon-xs" onClick={() => onEdit(attendance)}>
            <PencilIcon />
          </Button>
        ) : null}
        {canDelete ? (
          <Button
            variant="ghost"
            size="icon-xs"
            className="text-destructive hover:bg-destructive/10 hover:text-destructive"
            onClick={() => onDelete(attendance)}
          >
            <Trash2Icon />
          </Button>
        ) : null}
      </div>
    </div>
  )
}

interface AttendanceSectionProps {
  classId: string
  classSessionId: string
}

export function AttendanceSection({ classId, classSessionId }: AttendanceSectionProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { canView, canCreate, canEdit, canDelete } = useAttendancePermissions()
  const [editing, setEditing] = useState<Attendance | null>(null)
  const [deleting, setDeleting] = useState<Attendance | null>(null)

  const query = useGetClassesIdSessionsSessionIdAttendance(
    classId,
    classSessionId,
    { order_by: "status", order_dir: "asc" },
    { query: { enabled: canView && !!classId } }
  )
  const records = (query.data?.status === 200 && query.data.data.data?.items) || []

  const deleteMutation = useDeleteAttendanceAttendanceId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.attendance.form.deleteSuccess"))
        queryClient.invalidateQueries({
          queryKey: getGetClassesIdSessionsSessionIdAttendanceQueryKey(classId, classSessionId),
        })
        setDeleting(null)
      },
    },
  })

  if (!canView) return null

  return (
    <section id="attendance" className="flex scroll-mt-20 flex-col gap-5">
      <div className="flex flex-col gap-1.5">
        <Eyebrow>{t("org.session.attendance.eyebrow")}</Eyebrow>
        <h2 className="text-2xl font-semibold tracking-tight">{t("org.session.attendance.title")}</h2>
      </div>

      {canCreate ? <AutoMarkControl classId={classId} classSessionId={classSessionId} /> : null}

      {query.isPending ? (
        <div className="flex flex-col gap-2">
          <Skeleton className="h-16 w-full rounded-2xl" />
          <Skeleton className="h-16 w-full rounded-2xl" />
          <Skeleton className="h-16 w-full rounded-2xl" />
        </div>
      ) : records.length === 0 ? (
        <div className="bg-card ring-foreground/10 flex flex-col items-center gap-3 rounded-2xl px-6 py-16 text-center ring-1">
          <UserCheckIcon className="text-muted-foreground size-8" />
          <h3 className="text-foreground text-lg font-semibold tracking-tight">
            {t("org.session.attendance.emptyTitle")}
          </h3>
          <p className="text-muted-foreground max-w-md text-sm leading-relaxed">
            {canCreate ? t("org.session.attendance.emptyHint") : t("org.session.attendance.emptyHintMember")}
          </p>
        </div>
      ) : (
        <div className="flex flex-col gap-2">
          {records.map((a) => (
            <AttendanceRow
              key={a.id}
              attendance={a}
              canEdit={canEdit}
              canDelete={canDelete}
              onEdit={setEditing}
              onDelete={setDeleting}
            />
          ))}
        </div>
      )}

      <AttendanceEditDialog
        attendance={editing}
        open={!!editing}
        onOpenChange={(open) => {
          if (!open) setEditing(null)
        }}
        classId={classId}
        classSessionId={classSessionId}
      />

      <DeleteConfirmDialog
        open={!!deleting}
        onOpenChange={(open) => {
          if (deleteMutation.isPending) return
          if (!open) setDeleting(null)
        }}
        resourceName={deleting?.user?.name ?? ""}
        onConfirm={() => {
          if (deleting?.id) deleteMutation.mutate({ attendanceId: deleting.id })
        }}
        isLoading={deleteMutation.isPending}
      />
    </section>
  )
}
