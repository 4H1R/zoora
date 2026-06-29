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
  const [containerWidth, setContainerWidth] = useState<number | undefined>(undefined)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  // Measure container width when the div mounts so the PDF page fills it.
  const measureRef = (el: HTMLDivElement | null) => {
    if (!el) return
    const observer = new ResizeObserver(([entry]) => {
      setContainerWidth(entry.contentRect.width)
    })
    observer.observe(el)
    setContainerWidth(el.clientWidth)
    return () => observer.disconnect()
  }

  const clamp = (n: number) => Math.max(1, Math.min(n, numPages || 1))

  return (
    <div className="relative flex h-full w-full flex-col items-center justify-center overflow-hidden rounded-2xl bg-black">
      {/* PDF viewer */}
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
              width={containerWidth}
              loading={null}
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
