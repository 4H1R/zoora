import type { GithubCom4H1RZooraInternalDomainAttendance as Attendance } from "@/api/model"
import type { GithubCom4H1RZooraInternalDomainClassMember as ClassMember } from "@/api/model"
import type { GithubCom4H1RZooraInternalDomainCreateAttendanceDTOStatus as ApiStatus } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { useQueryClient } from "@tanstack/react-query"
import { getCoreRowModel, getSortedRowModel, useReactTable } from "@tanstack/react-table"
import { CheckCircle2Icon, ClockIcon, SaveIcon, ShieldCheckIcon, XCircleIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetClassesIdSessionsSessionIdAttendanceQueryKey,
  usePostClassesIdSessionsSessionIdAttendanceBulk,
} from "@/api/attendance/attendance"
import { DataTable } from "@/components/data-table/data-table"
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

interface RosterRow {
  id: string
  name: string
  userId: string
  status?: Status
  isDirty: boolean
}

interface AttendanceRosterProps {
  classId: string
  classSessionId: string
  members: ClassMember[]
  records: Attendance[]
  canMark: boolean
}

// React Compiler flags useReactTable as an "incompatible library" and silently
// skips memoizing any component that calls it. TanStack Table compares `data` and
// `columns` by reference, so an unmemoized parent rebuilds them every render and
// the table re-renders in an infinite loop (the tab freezes). Isolating the table
// in this leaf keeps the data/columns construction in the compiler-memoized parent,
// so this child receives stable references and the loop never starts.
function RosterTable({ data, columns }: { data: RosterRow[]; columns: ColumnDef<RosterRow>[] }) {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  })

  return (
    <div className="ring-foreground/10 overflow-hidden rounded-2xl ring-1">
      <DataTable table={table} />
    </div>
  )
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

  const setOne = (userId: string, status: Status) => {
    if (!canMark || !userId) return
    setPending((p) => ({ ...p, [userId]: status }))
  }

  const save = () => {
    if (changed.length === 0) return
    bulk.mutate({
      id: classId,
      sessionId: classSessionId,
      data: { entries: changed.map(([user_id, status]) => ({ user_id, status: status as ApiStatus })) },
    })
  }

  const rows: RosterRow[] = members.map((m) => {
    const userId = m.user_id ?? ""
    return {
      id: m.id ?? userId,
      name: m.user?.name ?? "—",
      userId,
      status: statusOf(userId),
      isDirty: userId in pending && pending[userId] !== saved[userId],
    }
  })

  const columns: ColumnDef<RosterRow>[] = [
    {
      accessorKey: "name",
      header: t("org.session.attendance.roster.member"),
      cell: ({ row }) => {
        const { name, status, isDirty } = row.original
        return (
          <div className="flex items-center gap-3">
            <Avatar className="size-9 shrink-0">
              <AvatarFallback className={cn("text-xs font-semibold text-white", getEntityColor(name))}>
                {getInitials(name)}
              </AvatarFallback>
            </Avatar>
            <div className="flex min-w-0 flex-col">
              <span className="flex items-center gap-2 truncate text-sm font-medium">
                {name}
                {isDirty && <span className="bg-primary inline-block size-1.5 shrink-0 rounded-full" />}
              </span>
              {!status && (
                <span className="text-muted-foreground text-xs">{t("org.session.attendance.roster.unmarked")}</span>
              )}
            </div>
          </div>
        )
      },
      enableHiding: false,
    },
    {
      id: "status",
      header: () => <div className="text-end">{t("org.session.attendance.roster.status")}</div>,
      enableSorting: false,
      enableHiding: false,
      cell: ({ row }) => {
        const { status, userId } = row.original
        return (
          <div className="flex justify-end">
            <div className="bg-muted/60 flex shrink-0 items-center gap-0.5 rounded-xl p-0.5">
              {SEGMENTS.map((seg) => {
                const Icon = seg.icon
                const isOn = status === seg.value
                return (
                  <button
                    key={seg.value}
                    type="button"
                    disabled={!canMark}
                    title={t(`org.session.attendance.status.${seg.value}`)}
                    onClick={() => setOne(userId, seg.value)}
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
          </div>
        )
      },
    },
  ]

  if (members.length === 0) {
    return (
      <div className="text-muted-foreground rounded-2xl border border-dashed px-6 py-12 text-center text-sm">
        {t("org.session.attendance.roster.emptyMembers")}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {canMark && (
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
            {changed.length > 0 && (
              <span className="text-muted-foreground font-mono text-[11px]">
                {t("org.session.attendance.roster.unsaved", { count: changed.length })}
              </span>
            )}
            <Button size="sm" disabled={changed.length === 0 || bulk.isPending} onClick={save}>
              <SaveIcon className="size-4" />
              {bulk.isPending ? t("org.session.attendance.roster.saving") : t("org.session.attendance.roster.save")}
            </Button>
          </div>
        </div>
      )}

      <RosterTable data={rows} columns={columns} />
    </div>
  )
}
