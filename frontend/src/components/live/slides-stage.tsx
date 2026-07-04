import { ChevronLeft, ChevronRight, Loader2 } from "lucide-react"
import { useState } from "react"
import { Document, Page, pdfjs } from "react-pdf"
import { useTranslation } from "react-i18next"
import pdfWorkerUrl from "pdfjs-dist/build/pdf.worker.min.mjs?url"

import "react-pdf/dist/Page/AnnotationLayer.css"
import "react-pdf/dist/Page/TextLayer.css"

import { cn } from "@/lib/utils"

pdfjs.GlobalWorkerOptions.workerSrc = pdfWorkerUrl

interface SlidesStageProps {
  url: string
  page: number
  numPages: number
  isHost: boolean
  onLoadNumPages: (n: number) => void
  onPageChange: (page: number) => void
}

export function SlidesStage({ url, page, numPages, isHost, onLoadNumPages, onPageChange }: SlidesStageProps) {
  const { t } = useTranslation()
  const [container, setContainer] = useState<{ w: number; h: number } | undefined>(undefined)
  const [aspect, setAspect] = useState<number | undefined>(undefined) // page width / height
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  // Measure container width AND height so we can fit the whole page on screen.
  // Keep the SAME state reference when dimensions are unchanged, otherwise the
  // ref-callback reattaching every render would setState → re-render forever.
  const applySize = (w: number, h: number) =>
    setContainer((prev) => (prev && prev.w === w && prev.h === h ? prev : { w, h }))

  const measureRef = (el: HTMLDivElement | null) => {
    if (!el) return
    const observer = new ResizeObserver(([entry]) => {
      applySize(entry.contentRect.width, entry.contentRect.height)
    })
    observer.observe(el)
    applySize(el.clientWidth, el.clientHeight)
    return () => observer.disconnect()
  }

  // Fit the page inside the container on BOTH axes (contain), preserving aspect
  // ratio, so the entire slide is visible without scrolling. Falls back to
  // full container width until the page's own aspect ratio is known.
  const INSET = 16 // px of breathing room around the page; also avoids a phantom scrollbar
  const fitWidth = (() => {
    if (!container) return undefined
    if (!aspect) return container.w
    return Math.max(1, Math.min(container.w - INSET, (container.h - INSET) * aspect))
  })()

  const clamp = (n: number) => Math.max(1, Math.min(n, numPages || 1))

  return (
    <div className="relative flex h-full w-full flex-col items-center justify-center overflow-hidden rounded-2xl bg-black">
      {/* PDF viewer — the page is sized to fit within both the container width
          and height (contain), so the whole slide is visible without scrolling.
          overflow-auto stays as a safety net for extreme aspect ratios. */}
      <div
        ref={measureRef}
        className="flex min-h-0 w-full flex-1 items-center justify-center overflow-auto"
      >
        {error ? (
          <p className="text-sm text-zinc-400">{t("liveRoom.stage.slidesError")}</p>
        ) : (
          <Document
            file={url}
            loading={null}
            onLoadSuccess={({ numPages: n }) => {
              onLoadNumPages(n)
              setLoading(false)
            }}
            onLoadError={() => {
              setLoading(false)
              setError(true)
            }}
          >
            <Page
              pageNumber={page}
              width={fitWidth}
              loading={null}
              onLoadSuccess={(p) => setAspect(p.originalWidth / p.originalHeight)}
            />
          </Document>
        )}

        {loading && !error && (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="flex flex-col items-center gap-2 text-zinc-400">
              <Loader2 className="size-6 animate-spin" />
              <p className="text-sm">{t("liveRoom.stage.slidesLoading")}</p>
            </div>
          </div>
        )}
      </div>

      {/* Page label + navigation (host only for nav buttons) */}
      <div className="absolute bottom-0 inset-x-0 flex items-center justify-center gap-3 pb-4">
        {isHost && (
          <button
            type="button"
            onClick={() => onPageChange(clamp(page - 1))}
            disabled={page <= 1}
            className={cn(
              "flex size-9 items-center justify-center rounded-xl bg-black/60 text-zinc-100 backdrop-blur-md transition-colors hover:bg-primary/80 disabled:opacity-40",
              "rtl:rotate-180"
            )}
            aria-label={t("liveRoom.stage.prevPage")}
          >
            <ChevronLeft className="size-5" />
          </button>
        )}

        {numPages > 0 && (
          <span className="rounded-lg bg-black/60 px-3 py-1 text-sm text-zinc-200 backdrop-blur-md">
            {t("liveRoom.stage.pageOf", { page, total: numPages })}
          </span>
        )}

        {isHost && (
          <button
            type="button"
            onClick={() => onPageChange(clamp(page + 1))}
            disabled={page >= numPages}
            className={cn(
              "flex size-9 items-center justify-center rounded-xl bg-black/60 text-zinc-100 backdrop-blur-md transition-colors hover:bg-primary/80 disabled:opacity-40",
              "rtl:rotate-180"
            )}
            aria-label={t("liveRoom.stage.nextPage")}
          >
            <ChevronRight className="size-5" />
          </button>
        )}
      </div>
    </div>
  )
}
