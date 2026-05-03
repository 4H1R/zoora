import type { Table } from "@tanstack/react-table"

import { flexRender } from "@tanstack/react-table"
import { ArrowDownIcon, ArrowUpIcon } from "lucide-react"
import { type ReactNode } from "react"

import { Skeleton } from "@/components/ui/skeleton"
import { TableBody, TableCell, TableHead, TableHeader, Table as TableRoot, TableRow } from "@/components/ui/table"

function SortIcon({ direction }: { direction: false | "asc" | "desc" }) {
  if (direction === "asc") return <ArrowUpIcon className="size-3" />
  if (direction === "desc") return <ArrowDownIcon className="size-3" />
  return <ArrowDownIcon className="size-3 opacity-30" />
}

export interface DataTableProps<TData> {
  table: Table<TData>
  isLoading?: boolean
  emptyIcon?: ReactNode
  emptyTitle?: string
  emptyHint?: string
  skeletonRows?: number
}

export function DataTable<TData>({
  table,
  isLoading,
  emptyIcon,
  emptyTitle,
  emptyHint,
  skeletonRows = 5,
}: DataTableProps<TData>) {
  const colCount = table.getVisibleLeafColumns().length

  return (
    <TableRoot>
      <TableHeader>
        {table.getHeaderGroups().map((headerGroup) => (
          <TableRow key={headerGroup.id} className="bg-muted/40">
            {headerGroup.headers.map((header) => (
              <TableHead
                key={header.id}
                className="text-muted-foreground text-[11px] font-medium tracking-wider uppercase"
              >
                {header.isPlaceholder ? null : header.column.getCanSort() ? (
                  <button
                    type="button"
                    className="hover:text-foreground flex items-center gap-1 transition-colors"
                    onClick={header.column.getToggleSortingHandler()}
                  >
                    {flexRender(header.column.columnDef.header, header.getContext())}
                    <SortIcon direction={header.column.getIsSorted()} />
                  </button>
                ) : (
                  flexRender(header.column.columnDef.header, header.getContext())
                )}
              </TableHead>
            ))}
          </TableRow>
        ))}
      </TableHeader>
      <TableBody>
        {isLoading ? (
          Array.from({ length: skeletonRows }).map((_, i) => (
            <TableRow key={i}>
              {Array.from({ length: colCount }).map((_, j) => (
                <TableCell key={j}>
                  <Skeleton className="h-5 w-full" />
                </TableCell>
              ))}
            </TableRow>
          ))
        ) : table.getRowModel().rows.length === 0 ? (
          <TableRow>
            <TableCell colSpan={colCount} className="h-40 text-center">
              <div className="text-muted-foreground flex flex-col items-center gap-2">
                {emptyIcon}
                {emptyTitle && <p className="text-sm font-medium">{emptyTitle}</p>}
                {emptyHint && <p className="text-xs">{emptyHint}</p>}
              </div>
            </TableCell>
          </TableRow>
        ) : (
          table.getRowModel().rows.map((row) => (
            <TableRow key={row.id} className="group transition-colors">
              {row.getVisibleCells().map((cell) => (
                <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
              ))}
            </TableRow>
          ))
        )}
      </TableBody>
    </TableRoot>
  )
}
