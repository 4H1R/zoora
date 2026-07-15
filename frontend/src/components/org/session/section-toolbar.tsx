import type { SortOption } from "@/components/data-table/sort-picker"
import type { SectionSort } from "@/lib/use-section-list"
import type { ReactNode } from "react"

import { SearchIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { SortPicker } from "@/components/data-table/sort-picker"
import { Input } from "@/components/ui/input"

interface SectionToolbarProps {
  /** Omit search props to hide the search input (e.g. sections whose API has no search). */
  searchValue?: string
  onSearchChange?: (value: string) => void
  searchPlaceholder?: string
  sortOptions: SortOption[]
  sort?: SectionSort
  onSortChange: (value: SectionSort | undefined) => void
  /** Extra filter controls (e.g. a status Select), rendered before the sort picker. */
  children?: ReactNode
}

/** Search + sort + optional filter row for a card-grid section. */
export function SectionToolbar({
  searchValue,
  onSearchChange,
  searchPlaceholder,
  sortOptions,
  sort,
  onSortChange,
  children,
}: SectionToolbarProps) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
      {onSearchChange && (
        <div className="relative w-full sm:max-w-xs">
          <SearchIcon className="text-muted-foreground pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2" />
          <Input
            value={searchValue ?? ""}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder={searchPlaceholder ?? t("org.session.controls.searchPlaceholder")}
            className="ps-9"
          />
        </div>
      )}
      <div className="flex items-center gap-2">
        {children}
        <SortPicker options={sortOptions} value={sort} onChange={onSortChange} label={t("org.session.controls.sort")} />
      </div>
    </div>
  )
}
