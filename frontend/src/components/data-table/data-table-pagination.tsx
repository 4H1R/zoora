import type { NavFn } from "@/lib/data-table"
import type { Table } from "@tanstack/react-table"

import { useNavigate, useSearch } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { paramKeys } from "@/lib/data-table"

const DEFAULT_PAGE_SIZE_OPTIONS = [8, 20, 50, 100, 200]

function getPageWindow(current: number, total: number): (number | "…")[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)

  const pages: (number | "…")[] = [1]
  const left = Math.max(2, current - 1)
  const right = Math.min(total - 1, current + 1)

  if (left > 2) pages.push("…")
  for (let i = left; i <= right; i++) pages.push(i)
  if (right < total - 1) pages.push("…")

  pages.push(total)
  return pages
}

export function DataTablePagination<TData>({
  table,
  pageSizeOptions = DEFAULT_PAGE_SIZE_OPTIONS,
  prefix,
}: {
  table: Table<TData>
  pageSizeOptions?: number[]
  /** Namespaces the URL params this pager reads/writes (page/page_size); pair
   * with the same prefix on useAdminTable + TableFilter. */
  prefix?: string
}) {
  const { t } = useTranslation()
  const k = paramKeys(prefix)
  const sp = useSearch({ strict: false }) as Record<string, unknown>
  const urlPage = sp[k.page] as number | undefined
  const urlPageSize = sp[k.page_size] as number | undefined
  const navigate = useNavigate() as unknown as NavFn

  const total = table.getRowCount()
  const page = urlPage ?? 1
  const pageSize = urlPageSize ?? 20
  const totalPages = Math.ceil(total / pageSize)
  const from = total === 0 ? 0 : (page - 1) * pageSize + 1
  const to = Math.min(page * pageSize, total)

  if (totalPages <= 0) return null

  const pages = getPageWindow(page, totalPages)

  return (
    <div className="bg-muted/30 border-t px-4 py-3">
      <div className="flex items-center justify-between sm:hidden">
        <PaginationPrevious
          aria-disabled={page <= 1}
          className={page <= 1 ? "pointer-events-none opacity-50" : ""}
          onClick={page > 1 ? () => navigate({ search: (prev) => ({ ...prev, [k.page]: page - 1 }) }) : undefined}
        />
        <Select
          value={String(pageSize)}
          onValueChange={(v) => navigate({ search: (prev) => ({ ...prev, [k.page_size]: Number(v), [k.page]: 1 }) })}
        >
          <SelectTrigger className="h-7 w-16 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {pageSizeOptions.map((s) => (
              <SelectItem key={s} value={String(s)} className="text-xs">
                {s}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <PaginationNext
          aria-disabled={page >= totalPages}
          className={page >= totalPages ? "pointer-events-none opacity-50" : ""}
          onClick={
            page < totalPages ? () => navigate({ search: (prev) => ({ ...prev, [k.page]: page + 1 }) }) : undefined
          }
        />
      </div>

      <div className="hidden items-center justify-between gap-4 sm:flex">
        <span className="text-muted-foreground text-xs tabular-nums">
          {t("admin.pagination.showing", { from, to, total })}
        </span>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2">
            <span className="text-muted-foreground text-xs">{t("admin.pagination.rowsPerPage")}</span>
            <Select
              value={String(pageSize)}
              onValueChange={(v) =>
                navigate({ search: (prev) => ({ ...prev, [k.page_size]: Number(v), [k.page]: 1 }) })
              }
            >
              <SelectTrigger className="h-7 w-16 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {pageSizeOptions.map((s) => (
                  <SelectItem key={s} value={String(s)} className="text-xs">
                    {s}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <Pagination className="mx-0 w-auto justify-start">
            <PaginationContent>
              <PaginationItem>
                <PaginationPrevious
                  aria-disabled={page <= 1}
                  className={page <= 1 ? "pointer-events-none opacity-50" : ""}
                  onClick={
                    page > 1 ? () => navigate({ search: (prev) => ({ ...prev, [k.page]: page - 1 }) }) : undefined
                  }
                />
              </PaginationItem>
              {pages.map((p, i) =>
                p === "…" ? (
                  <PaginationItem key={`ellipsis-${i}`}>
                    <PaginationEllipsis />
                  </PaginationItem>
                ) : (
                  <PaginationItem key={p}>
                    <PaginationLink
                      isActive={p === page}
                      onClick={() => navigate({ search: (prev) => ({ ...prev, [k.page]: p }) })}
                    >
                      {p}
                    </PaginationLink>
                  </PaginationItem>
                )
              )}
              <PaginationItem>
                <PaginationNext
                  aria-disabled={page >= totalPages}
                  className={page >= totalPages ? "pointer-events-none opacity-50" : ""}
                  onClick={
                    page < totalPages
                      ? () => navigate({ search: (prev) => ({ ...prev, [k.page]: page + 1 }) })
                      : undefined
                  }
                />
              </PaginationItem>
            </PaginationContent>
          </Pagination>
        </div>
      </div>
    </div>
  )
}
