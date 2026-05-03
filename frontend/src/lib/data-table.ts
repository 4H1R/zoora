import type { ColumnDef, SortingState, Table, VisibilityState } from "@tanstack/react-table"

import { useNavigate } from "@tanstack/react-router"
import { getCoreRowModel, useReactTable } from "@tanstack/react-table"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { z } from "zod"

// ── types ────────────────────────────────────────────────────────────────────

export type SearchUpdater = (prev: Record<string, unknown>) => Record<string, unknown>
export type NavFn = (opts: { search: SearchUpdater }) => void

// ── color + initials ─────────────────────────────────────────────────────────

export const ENTITY_COLORS = [
  "bg-slate-600",
  "bg-emerald-700",
  "bg-stone-500",
  "bg-amber-700",
  "bg-gray-800",
  "bg-sky-700",
  "bg-violet-700",
  "bg-rose-700",
]

export function getEntityColor(name?: string) {
  if (!name) return ENTITY_COLORS[0]
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return ENTITY_COLORS[Math.abs(hash) % ENTITY_COLORS.length]
}

export function getInitials(name?: string) {
  if (!name) return "?"
  return name
    .split(/\s+/)
    .slice(0, 2)
    .map((w) => w[0])
    .join("")
    .toUpperCase()
}

// ── date formatting ──────────────────────────────────────────────────────────

export function useFormatDate() {
  const { i18n } = useTranslation()
  return (dateStr?: string) => {
    if (!dateStr) return "—"
    return new Intl.DateTimeFormat(i18n.language, {
      year: "numeric",
      month: "short",
      day: "numeric",
    }).format(new Date(dateStr))
  }
}

// ── search schema ────────────────────────────────────────────────────────────

export const adminSearchSchema = z.object({
  search: z.string().optional(),
  status: z.string().optional(),
  order_by: z.string().optional(),
  order_dir: z.enum(["asc", "desc"]).optional(),
  page: z.number().int().positive().optional().default(1),
  page_size: z.number().int().positive().optional().default(8),
})

// ── useAdminTable ────────────────────────────────────────────────────────────

interface UseAdminTableOptions<TData> {
  data: TData[]
  columns: ColumnDef<TData>[]
  rowCount: number
  sorting: SortingState
}

export function useAdminTable<TData>({ data, columns, rowCount, sorting }: UseAdminTableOptions<TData>): Table<TData> {
  const navigate = useNavigate() as unknown as NavFn
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})

  return useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    rowCount,
    manualSorting: true,
    manualPagination: true,
    onSortingChange: (updater) => {
      const next = typeof updater === "function" ? updater(sorting) : updater
      const first = next[0]
      navigate({
        search: (prev) => ({
          ...prev,
          order_by: first?.id,
          order_dir: first ? (first.desc ? ("desc" as const) : ("asc" as const)) : undefined,
          page: 1,
        }),
      })
    },
    onColumnVisibilityChange: (updater) => {
      setColumnVisibility((prev) => (typeof updater === "function" ? updater(prev) : updater))
    },
    state: { sorting, columnVisibility },
  })
}
