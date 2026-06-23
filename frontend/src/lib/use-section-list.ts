import { useEffect, useState } from "react"
import { useDebounce } from "use-debounce"

import { DEFAULT_PAGE_SIZE } from "@/lib/list"

export interface SectionSort {
  id: string
  desc: boolean
}

export interface SectionListParams {
  search?: string
  order_by?: string
  order_dir?: "asc" | "desc"
  page: number
}

interface UseSectionListOptions {
  defaultSort?: SectionSort
  searchDelayMs?: number
}

function getOrderDir(sort: SectionSort | undefined): "asc" | "desc" | undefined {
  if (!sort) return undefined
  return sort.desc ? "desc" : "asc"
}

/**
 * Local list state for a card-grid section: debounced search, sort, an optional
 * status filter, and pagination. Each section owns its own instance, so the five
 * lists on the class-session page never collide. Page resets to 1 whenever the
 * search, sort, or status changes.
 */
export function useSectionList(opts: UseSectionListOptions = {}) {
  const [searchInput, setSearchInput] = useState("")
  const [debouncedSearch] = useDebounce(searchInput, opts.searchDelayMs ?? 300)
  const [sort, setSort] = useState<SectionSort | undefined>(opts.defaultSort)
  const [status, setStatus] = useState<string | undefined>(undefined)
  const [page, setPage] = useState(1)

  const search = debouncedSearch.trim()

  // Any filter change drops the user back to the first page.
  useEffect(() => {
    setPage(1)
  }, [search, sort?.id, sort?.desc, status])

  const params: SectionListParams = {
    search: search || undefined,
    order_by: sort?.id,
    order_dir: getOrderDir(sort),
    page,
  }

  const isFiltered = search !== "" || status !== undefined

  return {
    params,
    page,
    setPage,
    searchInput,
    setSearchInput,
    sort,
    setSort,
    status,
    setStatus,
    isFiltered,
  }
}

export { DEFAULT_PAGE_SIZE }
