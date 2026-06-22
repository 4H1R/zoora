import type { Table } from "@tanstack/react-table"

import { ColumnsIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"

interface ColumnsToggleProps<TData> {
  table: Table<TData>
  columnsLabel?: string
  toggleColumnsLabel?: string
}

/** Popover that toggles per-column visibility. Used standalone on pages whose
 * toolbar isn't the full TableFilter, and inside TableFilter itself. */
export function ColumnsToggle<TData>({
  table,
  columnsLabel = "Columns",
  toggleColumnsLabel = "Toggle columns",
}: ColumnsToggleProps<TData>) {
  return (
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
                <Checkbox checked={col.getIsVisible()} onCheckedChange={(checked) => col.toggleVisibility(!!checked)} />
                {typeof col.columnDef.header === "string" ? col.columnDef.header : col.id}
              </label>
            ))}
        </div>
      </PopoverContent>
    </Popover>
  )
}
