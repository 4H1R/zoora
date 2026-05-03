import type { SortOption } from "@/components/data-table/sort-picker"
import type { NavFn } from "@/lib/data-table"
import type { Table } from "@tanstack/react-table"

import { useNavigate, useSearch } from "@tanstack/react-router"
import { ColumnsIcon, SearchIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useDebounce } from "use-debounce"

import { SortPicker } from "@/components/data-table/sort-picker"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"

interface TableFilterProps<TData> {
  searchPlaceholder?: string
  sortLabel?: string
  sortOptions?: SortOption[]
  table: Table<TData>
  columnsLabel?: string
  toggleColumnsLabel?: string
  children?: React.ReactNode
}

export function TableFilter<TData>({
  searchPlaceholder,
  sortLabel = "Sort",
  sortOptions: sortOptionsProp,
  table,
  columnsLabel = "Columns",
  toggleColumnsLabel = "Toggle columns",
  children,
}: TableFilterProps<TData>) {
  const {
    search: urlSearch,
    order_by,
    order_dir,
  } = useSearch({ strict: false }) as {
    search?: string
    order_by?: string
    order_dir?: "asc" | "desc"
  }
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
    navigate({ search: (prev) => ({ ...prev, search: debouncedSearch || undefined, page: 1 }) })
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
              order_by: v?.id,
              order_dir: v ? (v.desc ? "desc" : "asc") : undefined,
              page: 1,
            }),
          })
        }
      />
      <Popover>
        <PopoverTrigger
          render={
            <Button variant="outline" size="sm" className="h-9 gap-1.5 px-2.5 text-xs font-medium">
              <ColumnsIcon className="size-3.5" />
              {columnsLabel}
            </Button>
          }
        />
        <PopoverContent align="end" className="w-44 p-1.5">
          <p className="text-muted-foreground px-1.5 py-1 text-[11px] font-medium tracking-wider uppercase">
            {toggleColumnsLabel}
          </p>
          <div className="mt-0.5 flex flex-col">
            {table
              .getAllColumns()
              .filter((col) => col.getCanHide())
              .map((col) => (
                <label
                  key={col.id}
                  className="hover:bg-accent flex cursor-pointer items-center gap-2 rounded-md px-1.5 py-1.5 text-sm"
                >
                  <Checkbox
                    checked={col.getIsVisible()}
                    onCheckedChange={(checked) => col.toggleVisibility(!!checked)}
                  />
                  {typeof col.columnDef.header === "string" ? col.columnDef.header : col.id}
                </label>
              ))}
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}
