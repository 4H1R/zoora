import type { GithubCom4H1RZooraInternalDomainPracticeRoomView as PracticeRoomView } from "@/api/model"
import type { ColumnDef } from "@tanstack/react-table"

import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { getEntityColor, getInitials, useFormatDate } from "@/lib/data-table"
import { cn } from "@/lib/utils"

function completionPct(view: PracticeRoomView): number {
  const members = view.stats?.member_count ?? 0
  const submitted = view.stats?.submitted_count ?? 0
  if (members <= 0) return 0
  return Math.min(100, Math.round((submitted / members) * 100))
}

export function usePracticeHubColumns({
  onViewSubmissions,
}: {
  onViewSubmissions: (view: PracticeRoomView) => void
}): ColumnDef<PracticeRoomView>[] {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  return [
    {
      accessorKey: "title",
      header: t("org.practices.title"),
      cell: ({ row }) => (
        <div className="flex items-center gap-3">
          <div
            className={cn(
              "flex size-9 shrink-0 items-center justify-center rounded-lg text-xs font-semibold text-white",
              getEntityColor(row.original.title ?? "")
            )}
          >
            {getInitials(row.original.title ?? "")}
          </div>
          <span className="truncate text-sm font-medium">{row.original.title}</span>
        </div>
      ),
      enableHiding: false,
    },
    {
      id: "class",
      header: t("org.practices.manager.class"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs">{row.original.class?.name ?? "—"}</span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      id: "teacher",
      header: t("org.practices.manager.teacher"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs">{row.original.user?.name ?? "—"}</span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      id: "submissions",
      header: t("org.practices.manager.submitted"),
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {t("org.practices.manager.submissions", {
            submitted: row.original.stats?.submitted_count ?? 0,
            members: row.original.stats?.member_count ?? 0,
          })}
        </span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      id: "graded",
      header: t("org.practices.manager.graded"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm tabular-nums">
          {row.original.stats?.graded_count ?? 0}
        </span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      id: "completion",
      header: t("org.practices.manager.completion"),
      cell: ({ row }) => {
        const pct = completionPct(row.original)
        return (
          <div className="flex w-28 items-center gap-2">
            <div className="bg-muted h-1.5 flex-1 overflow-hidden rounded-full">
              <div className="bg-primary h-full rounded-full" style={{ width: `${pct}%` }} />
            </div>
            <span className="text-muted-foreground w-9 text-end text-xs tabular-nums">{pct}%</span>
          </div>
        )
      },
      enableSorting: false,
      enableHiding: true,
    },
    {
      id: "due",
      header: t("org.practices.due"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-xs">{formatDate(row.original.end_time)}</span>
      ),
      enableSorting: false,
      enableHiding: true,
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <div className="flex justify-end">
          <Button size="sm" variant="outline" onClick={() => onViewSubmissions(row.original)}>
            {t("org.practices.actions.viewSubmissions")}
          </Button>
        </div>
      ),
      enableSorting: false,
      enableHiding: false,
    },
  ]
}
