import type { SortOption } from "@/components/data-table/sort-picker"
import type { NavFn } from "@/lib/data-table"
import type { Table } from "@tanstack/react-table"

import { useNavigate, useSearch } from "@tanstack/react-router"
import { SearchIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useDebounce } from "use-debounce"

import { ColumnsToggle } from "@/components/data-table/columns-toggle"
import { SortPicker } from "@/components/data-table/sort-picker"
import { Input } from "@/components/ui/input"
import { paramKeys } from "@/lib/data-table"

interface TableFilterProps<TData> {
  searchPlaceholder?: string
  sortLabel?: string
  sortOptions?: SortOption[]
  table: Table<TData>
  columnsLabel?: string
  toggleColumnsLabel?: string
  /** Show the column show/hide toggle. Only meaningful in table view — pages
   * with a grid/table switcher should pass `useViewMode().isTable`. Default true. */
  showColumnsToggle?: boolean
  /** Namespaces the URL params this toolbar reads/writes; pair with the same
   * prefix on useAdminTable + DataTablePagination. Omit for single-table pages. */
  prefix?: string
  children?: React.ReactNode
}

export function TableFilter<TData>({
  searchPlaceholder,
  sortLabel = "Sort",
  sortOptions: sortOptionsProp,
  table,
  columnsLabel = "Columns",
  toggleColumnsLabel = "Toggle columns",
  showColumnsToggle = true,
  prefix,
  children,
}: TableFilterProps<TData>) {
  const k = paramKeys(prefix)
  const sp = useSearch({ strict: false }) as Record<string, unknown>
  const urlSearch = sp[k.search] as string | undefined
  const order_by = sp[k.order_by] as string | undefined
  const order_dir = sp[k.order_dir] as "asc" | "desc" | undefined
  const navigate = useNavigate() as unknown as NavFn

  const sortOptions: SortOption[] =
    sortOptionsProp ??
    table
      .getAllColumns()
      .filter((col) => col.getCanSort())
      .map((col) => ({
        id: col.id,
        label: typeof col.columnDef.header === "string" ? col.columnDef.header : col.id,
      }))

  const [localSearch, setLocalSearch] = useState(urlSearch ?? "")
  const [debouncedSearch] = useDebounce(localSearch, 300)
  const isFirstRender = useRef(true)

  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false
      return
    }
    navigate({ search: (prev) => ({ ...prev, [k.search]: debouncedSearch || undefined, [k.page]: 1 }) })
  }, [debouncedSearch])

  const sortValue = order_by ? { id: order_by, desc: order_dir === "desc" } : undefined

  return (
    <div className="flex flex-wrap items-center gap-1.5">
      <div className="relative me-1 w-full lg:w-auto">
        <SearchIcon className="text-muted-foreground absolute start-3 top-1/2 size-3.5 -translate-y-1/2" />
        <Input
          value={localSearch}
          onChange={(e) => setLocalSearch(e.target.value)}
          placeholder={searchPlaceholder}
          className="h-9 w-full ps-9 text-sm lg:w-56"
        />
      </div>
      {children}
      <div className="hidden flex-1 lg:flex" />
      <SortPicker
        label={sortLabel}
        options={sortOptions}
        value={sortValue}
        onChange={(v) =>
          navigate({
            search: (prev) => ({
              ...prev,
              [k.order_by]: v?.id,
              [k.order_dir]: v ? (v.desc ? "desc" : "asc") : undefined,
              [k.page]: 1,
            }),
          })
        }
      />
      {showColumnsToggle && (
        <ColumnsToggle table={table} columnsLabel={columnsLabel} toggleColumnsLabel={toggleColumnsLabel} />
      )}
    </div>
  )
}
