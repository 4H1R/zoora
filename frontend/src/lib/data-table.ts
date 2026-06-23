import type { ColumnDef, OnChangeFn, SortingState, Table, VisibilityState } from "@tanstack/react-table"

import { useNavigate } from "@tanstack/react-router"
import {
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table"
import { useState } from "react"
import { z } from "zod"

export type SearchUpdater = (prev: Record<string, unknown>) => Record<string, unknown>
export type NavFn = (opts: { search: SearchUpdater }) => void

export interface ParamKeys {
  search: string
  status: string
  order_by: string
  order_dir: string
  page: string
  page_size: string
}

// Maps the canonical table params to their URL keys. With no prefix the keys
// are the bare names (search, order_by, …) so existing single-table pages are
// untouched. A prefix namespaces them (sessions_search, students_order_by, …)
// so multiple server-driven tables can share one route without colliding.
export function paramKeys(prefix?: string): ParamKeys {
  const p = prefix ? `${prefix}_` : ""
  return {
    search: `${p}search`,
    status: `${p}status`,
    order_by: `${p}order_by`,
    order_dir: `${p}order_dir`,
    page: `${p}page`,
    page_size: `${p}page_size`,
  }
}

export { getInitials } from "@/components/user-avatar"

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

// Source of truth lives in ./format-date. Re-exported here for call-site convenience.
export { useFormatDate } from "./format-date"

export const adminSearchSchema = z.object({
  search: z.string().optional(),
  status: z.string().optional(),
  order_by: z.string().optional(),
  order_dir: z.enum(["asc", "desc"]).optional(),
  page: z.number().int().positive().optional().default(1),
  page_size: z.number().int().positive().optional().default(20),
})

function getOrderDir(first: SortingState[number] | undefined): "asc" | "desc" | undefined {
  if (!first) return undefined
  return first.desc ? "desc" : "asc"
}

export function createSortingHandler(
  navigate: NavFn,
  sorting: SortingState,
  prefix?: string
): OnChangeFn<SortingState> {
  const k = paramKeys(prefix)
  return (updater) => {
    const next = typeof updater === "function" ? updater(sorting) : updater
    const first = next[0]
    navigate({
      search: (prev) => ({
        ...prev,
        [k.order_by]: first?.id,
        [k.order_dir]: getOrderDir(first),
        [k.page]: 1,
      }),
    })
  }
}

interface UseAdminTableOptions<TData> {
  data: TData[]
  columns: ColumnDef<TData>[]
  rowCount: number
  sorting: SortingState
  /** Namespaces the URL params this table writes (order_by/order_dir/page).
   * Omit for single-table pages; set when several tables share one route. */
  prefix?: string
}

export function useAdminTable<TData>({
  data,
  columns,
  rowCount,
  sorting,
  prefix,
}: UseAdminTableOptions<TData>): Table<TData> {
  const navigate = useNavigate() as unknown as NavFn
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})

  return useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    rowCount,
    manualSorting: true,
    manualPagination: true,
    onSortingChange: createSortingHandler(navigate, sorting, prefix),
    onColumnVisibilityChange: (updater) => {
      setColumnVisibility((prev) => (typeof updater === "function" ? updater(prev) : updater))
    },
    state: { sorting, columnVisibility },
  })
}

interface UseClientTableOptions<TData> {
  data: TData[]
  columns: ColumnDef<TData>[]
  sorting: SortingState
  /** URL search term (`search` param). Filters client-side via the global filter. */
  globalFilter?: string
  /** 1-based page from the URL `page` param. */
  page?: number
  /** Page size from the URL `page_size` param. */
  pageSize?: number
}

// Client-side sibling of useAdminTable: same URL-driven TableFilter /
// DataTablePagination wiring, but the data lives entirely in the browser.
// Use for "me"-style endpoints that return a full (small) list with no
// server pagination/search/sort. Filtering, sorting and paging all happen
// over the in-memory array; URL stays the source of truth.
export function useClientTable<TData>({
  data,
  columns,
  sorting,
  globalFilter,
  page,
  pageSize = 20,
}: UseClientTableOptions<TData>): Table<TData> {
  const navigate = useNavigate() as unknown as NavFn
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})

  return useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    globalFilterFn: "includesString",
    onSortingChange: createSortingHandler(navigate, sorting),
    onColumnVisibilityChange: (updater) => {
      setColumnVisibility((prev) => (typeof updater === "function" ? updater(prev) : updater))
    },
    state: {
      sorting,
      globalFilter: globalFilter ?? "",
      columnVisibility,
      pagination: { pageIndex: Math.max(0, (page ?? 1) - 1), pageSize },
    },
  })
}
