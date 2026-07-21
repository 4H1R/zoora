import type { ColumnDef } from "@tanstack/react-table"

import { GraduationCapIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

// One flattened grade cell: a single graded item within a class.
export interface GradeRow {
  classId: string
  className: string
  item: string
  value: string
  maxScore?: number
}

export function useGradeColumns(): ColumnDef<GradeRow>[] {
  const { t } = useTranslation()

  return [
    {
      id: "className",
      accessorFn: (r) => r.className,
      header: t("org.grades.table.class"),
      cell: ({ row }) => (
        <div className="flex items-center gap-2">
          <div className="bg-muted text-muted-foreground flex size-8 shrink-0 items-center justify-center rounded-lg [&>svg]:size-4">
            <GraduationCapIcon />
          </div>
          <span className="truncate text-sm font-medium">{row.original.className || "—"}</span>
        </div>
      ),
    },
    {
      id: "item",
      accessorFn: (r) => r.item,
      header: t("org.grades.table.item"),
      cell: ({ getValue }) => <span className="text-sm">{(getValue() as string) || "—"}</span>,
    },
    {
      id: "value",
      accessorFn: (r) => r.value,
      header: t("org.grades.table.score"),
      cell: ({ getValue, row }) => {
        const v = getValue() as string
        const max = row.original.maxScore
        if (!v || !v.trim()) return <span className="font-medium tabular-nums">—</span>
        return (
          <span dir={max != null ? "ltr" : undefined} className="font-medium tabular-nums">
            {v}
            {max != null && <span className="text-muted-foreground font-normal"> / {max}</span>}
          </span>
        )
      },
    },
  ]
}
