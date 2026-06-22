import { LayoutGrid, List } from "lucide-react"
import { useTranslation } from "react-i18next"

import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group"
import { cn } from "@/lib/utils"

export type ViewMode = "grid" | "table"

interface ViewModeToggleProps {
  value: ViewMode
  onChange: (mode: ViewMode) => void
  className?: string
}

// Shared grid/table switcher. Drop into any list page that wants both a card
// grid and a DataTable view; the page owns the `viewMode` state and renders
// the matching content.
export function ViewModeToggle({ value, onChange, className }: ViewModeToggleProps) {
  const { t } = useTranslation()

  return (
    <ToggleGroup
      value={[value]}
      onValueChange={(values) => {
        const next = values.find((v) => v !== value)
        if (next) onChange(next as ViewMode)
      }}
      className={cn("border-border rounded-lg border", className)}
    >
      <ToggleGroupItem value="grid" aria-label={t("common.gridView")} className="px-2.5">
        <LayoutGrid className="size-4" />
      </ToggleGroupItem>
      <ToggleGroupItem value="table" aria-label={t("common.tableView")} className="px-2.5">
        <List className="size-4" />
      </ToggleGroupItem>
    </ToggleGroup>
  )
}
