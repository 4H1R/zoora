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

interface SectionPaginationProps {
  page: number
  pageSize: number
  total: number
  onPageChange: (page: number) => void
}

/** Page navigation for a card-grid section. Renders nothing when there is only
 * a single page of results. */
export function SectionPagination({ page, pageSize, total, onPageChange }: SectionPaginationProps) {
  const { t } = useTranslation()
  const size = pageSize > 0 ? pageSize : total
  const totalPages = size > 0 ? Math.ceil(total / size) : 0
  if (totalPages <= 1) return null

  const from = (page - 1) * size + 1
  const to = Math.min(page * size, total)
  const pages = getPageWindow(page, totalPages)

  return (
    <div className="flex flex-col items-center justify-between gap-3 pt-1 sm:flex-row">
      <span className="text-muted-foreground text-xs tabular-nums">
        {t("org.session.controls.showing", { from, to, total })}
      </span>
      <Pagination className="mx-0 w-auto justify-end">
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious
              aria-label={t("org.session.controls.prev")}
              aria-disabled={page <= 1}
              className={page <= 1 ? "pointer-events-none opacity-50" : ""}
              onClick={page > 1 ? () => onPageChange(page - 1) : undefined}
            />
          </PaginationItem>
          {pages.map((p, i) =>
            p === "…" ? (
              <PaginationItem key={`ellipsis-${i}`}>
                <PaginationEllipsis />
              </PaginationItem>
            ) : (
              <PaginationItem key={p}>
                <PaginationLink isActive={p === page} onClick={() => onPageChange(p)}>
                  {p}
                </PaginationLink>
              </PaginationItem>
            )
          )}
          <PaginationItem>
            <PaginationNext
              aria-label={t("org.session.controls.next")}
              aria-disabled={page >= totalPages}
              className={page >= totalPages ? "pointer-events-none opacity-50" : ""}
              onClick={page < totalPages ? () => onPageChange(page + 1) : undefined}
            />
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    </div>
  )
}
