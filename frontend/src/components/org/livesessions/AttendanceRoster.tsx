import type { GithubCom4H1RZooraInternalDomainAttendance as Attendance } from "@/api/model"
import type { GithubCom4H1RZooraInternalDomainClassMember as ClassMember } from "@/api/model"
import type { GithubCom4H1RZooraInternalDomainCreateAttendanceDTOStatus as ApiStatus } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { CheckCircle2Icon, ClockIcon, SaveIcon, ShieldCheckIcon, XCircleIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetClassesIdSessionsSessionIdAttendanceQueryKey,
  usePostClassesIdSessionsSessionIdAttendanceBulk,
} from "@/api/attendance/attendance"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { getEntityColor, getInitials } from "@/lib/data-table"
import { cn } from "@/lib/utils"

type Status = "present" | "late" | "excused" | "absent"

const SEGMENTS: { value: Status; icon: typeof CheckCircle2Icon; active: string }[] = [
  { value: "present", icon: CheckCircle2Icon, active: "bg-emerald-500 text-white" },
  { value: "late", icon: ClockIcon, active: "bg-amber-500 text-white" },
  { value: "excused", icon: ShieldCheckIcon, active: "bg-primary text-primary-foreground" },
  { value: "absent", icon: XCircleIcon, active: "bg-destructive text-white" },
]

interface AttendanceRosterProps {
  classId: string
  classSessionId: string
  members: ClassMember[]
  records: Attendance[]
  canMark: boolean
}

export function AttendanceRoster({ classId, classSessionId, members, records, canMark }: AttendanceRosterProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const saved: Record<string, Status> = {}
  for (const r of records) {
    if (r.user_id && r.status) saved[r.user_id] = r.status as Status
  }

  const [pending, setPending] = useState<Record<string, Status>>({})

  const statusOf = (userId: string): Status | undefined => pending[userId] ?? saved[userId]
  const changed = Object.entries(pending).filter(([uid, st]) => saved[uid] !== st)

  const bulk = usePostClassesIdSessionsSessionIdAttendanceBulk({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.attendance.roster.saveSuccess"))
        setPending({})
        queryClient.invalidateQueries({
          queryKey: getGetClassesIdSessionsSessionIdAttendanceQueryKey(classId, classSessionId),
        })
      },
      onError: () => {
        toast.error(t("org.session.attendance.roster.saveError"))
      },
    },
  })

  const setAll = (status: Status) => {
    const next: Record<string, Status> = {}
    for (const m of members) {
      if (m.user_id) next[m.user_id] = status
    }
    setPending(next)
  }

  const save = () => {
    if (changed.length === 0) return
    bulk.mutate({
      id: classId,
      sessionId: classSessionId,
      data: { entries: changed.map(([user_id, status]) => ({ user_id, status: status as ApiStatus })) },
    })
  }

  if (members.length === 0) {
    return (
      <div className="text-muted-foreground rounded-2xl border border-dashed px-6 py-12 text-center text-sm">
        {t("org.session.attendance.roster.emptyMembers")}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {canMark ? (
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={() => setAll("present")}>
              {t("org.session.attendance.roster.markAllPresent")}
            </Button>
            <Button variant="outline" size="sm" onClick={() => setAll("absent")}>
              {t("org.session.attendance.roster.markAllAbsent")}
            </Button>
          </div>
          <div className="flex items-center gap-3">
            {changed.length > 0 ? (
              <span className="text-muted-foreground font-mono text-[11px]">
                {t("org.session.attendance.roster.unsaved", { count: changed.length })}
              </span>
            ) : null}
            <Button size="sm" disabled={changed.length === 0 || bulk.isPending} onClick={save}>
              <SaveIcon className="size-4" />
              {bulk.isPending ? t("org.session.attendance.roster.saving") : t("org.session.attendance.roster.save")}
            </Button>
          </div>
        </div>
      ) : null}

      <ul className="flex flex-col gap-2">
        {members.map((m) => {
          const name = m.user?.name ?? "—"
          const userId = m.user_id ?? ""
          const current = statusOf(userId)
          const isDirty = userId in pending && pending[userId] !== saved[userId]
          return (
            <li
              key={m.id}
              className={cn(
                "bg-card ring-foreground/10 flex items-center gap-3 rounded-2xl px-4 py-3 ring-1 transition-all",
                isDirty && "ring-primary/40"
              )}
            >
              <Avatar className="size-9 shrink-0">
                <AvatarFallback className={cn("text-xs font-semibold text-white", getEntityColor(name))}>
                  {getInitials(name)}
                </AvatarFallback>
              </Avatar>
              <div className="flex min-w-0 flex-1 flex-col">
                <span className="truncate text-sm font-medium">{name}</span>
                {!current ? (
                  <span className="text-muted-foreground text-xs">{t("org.session.attendance.roster.unmarked")}</span>
                ) : null}
              </div>

              <div className="bg-muted/60 flex shrink-0 items-center gap-0.5 rounded-xl p-0.5">
                {SEGMENTS.map((seg) => {
                  const Icon = seg.icon
                  const isOn = current === seg.value
                  return (
                    <button
                      key={seg.value}
                      type="button"
                      disabled={!canMark}
                      title={t(`org.session.attendance.status.${seg.value}`)}
                      onClick={() => canMark && userId && setPending((p) => ({ ...p, [userId]: seg.value }))}
                      className={cn(
                        "inline-flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-[11px] font-medium transition-colors",
                        isOn ? seg.active : "text-muted-foreground hover:text-foreground",
                        !canMark && "cursor-default"
                      )}
                    >
                      <Icon className="size-3.5" />
                      <span className="hidden sm:inline">{t(`org.session.attendance.status.${seg.value}`)}</span>
                    </button>
                  )
                })}
              </div>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
