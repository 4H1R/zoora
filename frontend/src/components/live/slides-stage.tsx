import { ChevronLeft, ChevronRight, Loader2 } from "lucide-react"
import { useEffect, useRef, useState } from "react"
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
  const scrollRef = useRef<HTMLDivElement | null>(null)
  const wheelLock = useRef(false)

  // Measure the scroll viewport width so the page renders at full width. Keep
  // the SAME value when unchanged, otherwise the ref-callback reattaching every
  // render would setState → re-render forever.
  const measureRef = (el: HTMLDivElement | null) => {
    scrollRef.current = el
    if (!el) return
    const apply = (w: number) => setContainerWidth((prev) => (prev === w ? prev : w))
    const observer = new ResizeObserver(([entry]) => apply(entry.contentRect.width))
    observer.observe(el)
    apply(el.clientWidth)
    return () => observer.disconnect()
  }

  // Fit the page to the full viewport width. Portrait / tall pages then overflow
  // vertically, giving a real scrollbar + wheel scroll. Landscape slides fit
  // without overflow. INSET leaves breathing room beside the scrollbar.
  const INSET = 16
  const fitWidth = containerWidth ? Math.max(1, containerWidth - INSET) : undefined

  const clamp = (n: number) => Math.max(1, Math.min(n, numPages || 1))

  // Host: wheel scrolls a tall page, then flips to the prev/next page once the
  // scroll hits the top/bottom edge (or immediately when the page fits). A short
  // lock stops wheel momentum from skipping several pages per gesture. Viewers
  // just scroll their own view — page nav is host-driven.
  const navigate = (n: number) => {
    const target = clamp(n)
    if (wheelLock.current || target === page) return
    wheelLock.current = true
    onPageChange(target)
    setTimeout(() => {
      wheelLock.current = false
    }, 400)
  }

  const onWheel = (e: React.WheelEvent<HTMLDivElement>) => {
    if (!isHost) return
    const el = scrollRef.current
    if (!el) return
    const canScroll = el.scrollHeight > el.clientHeight + 1
    const atTop = el.scrollTop <= 0
    const atBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 1
    if (e.deltaY > 0 && (!canScroll || atBottom)) navigate(page + 1)
    else if (e.deltaY < 0 && (!canScroll || atTop)) navigate(page - 1)
  }

  // Host keyboard nav: PageDown/Down/Right/Space go forward, PageUp/Up/Left go
  // back. Up/Down/PageUp/PageDown scroll a tall page first and only flip at the
  // edge; Left/Right always flip. Ignored while typing in a field (e.g. chat).
  useEffect(() => {
    if (!isHost) return
    const FORWARD = ["PageDown", "ArrowDown", "ArrowRight", " ", "Spacebar"]
    const BACKWARD = ["PageUp", "ArrowUp", "ArrowLeft"]
    const onKey = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null
      // Don't hijack Space/arrows away from an interactive element: a focused
      // button/select/link/menu item needs those keys for its own activation or
      // option navigation (Space "clicks" a button, arrows move select options).
      if (
        target?.isContentEditable ||
        target?.closest(
          "input, textarea, select, button, a[href], [role='button'], [role='menuitem'], [role='menuitemradio'], [role='option'], [role='tab'], [role='slider'], [contenteditable='true']",
        )
      )
        return
      const forward = FORWARD.includes(e.key)
      const backward = BACKWARD.includes(e.key)
      if (!forward && !backward) return
      e.preventDefault()
      const el = scrollRef.current
      const flipOnly = e.key === "ArrowLeft" || e.key === "ArrowRight"
      if (el && !flipOnly) {
        const canScroll = el.scrollHeight > el.clientHeight + 1
        const atTop = el.scrollTop <= 0
        const atBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 1
        const step = el.clientHeight * 0.9
        if (canScroll && forward && !atBottom) return void (el.scrollTop += step)
        if (canScroll && backward && !atTop) return void (el.scrollTop -= step)
      }
      navigate(forward ? page + 1 : page - 1)
    }
    window.addEventListener("keydown", onKey)
    return () => window.removeEventListener("keydown", onKey)
  }, [isHost, page, numPages, onPageChange])

  return (
    <div className="relative flex h-full w-full flex-col items-center justify-center overflow-hidden rounded-2xl bg-black">
      {/* PDF viewer — the page fills the viewport width. Tall pages overflow
          vertically so the scroll container shows a scrollbar and the wheel
          scrolls. min-h-full on the inner wrapper keeps short pages centered
          while still allowing a taller page to scroll from its top edge (a plain
          flex-center container would clip the overflowing top, making it
          unreachable). */}
      <div
        ref={measureRef}
        onWheel={onWheel}
        className="min-h-0 w-full flex-1 overflow-auto"
      >
        <div className="flex min-h-full w-full items-center justify-center py-2">
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
              <Page pageNumber={page} width={fitWidth} loading={null} />
            </Document>
          )}
        </div>

        {loading && !error && (
          <div className="absolute inset-0 flex items-center justify-center">
            <div className="flex flex-col items-center gap-2 text-zinc-400">
              <Loader2 className="size-6 animate-spin" />
              <p className="text-sm">{t("liveRoom.stage.slidesLoading")}</p>
            </div>
          </div>
        )}
      </div>

      {/* Page navigation — a single glass segmented pill. Host gets prev/next
          controls flanking a current/total counter; viewers see just the
          counter. Anchored top-center so it never collides with the ControlBar,
          which owns the bottom-center slot. Tabular figures keep the width
          steady; the active page pops on change. */}
      {numPages > 0 && (
        <div className="pointer-events-none absolute inset-x-0 top-0 z-10 flex justify-center pt-4">
          <div className="pointer-events-auto flex items-center gap-0.5 rounded-full border border-white/10 bg-black/60 p-1 shadow-lg shadow-black/50 ring-1 ring-white/5 backdrop-blur-md">
            {isHost && (
              <button
                type="button"
                onClick={() => onPageChange(clamp(page - 1))}
                disabled={page <= 1}
                className={cn(
                  "grid size-8 place-items-center rounded-full text-zinc-300 transition-all",
                  "hover:bg-white/10 hover:text-white active:scale-90",
                  "disabled:pointer-events-none disabled:opacity-30 rtl:rotate-180"
                )}
                aria-label={t("liveRoom.stage.prevPage")}
              >
                <ChevronLeft className="size-5" />
              </button>
            )}

            <div className="flex select-none items-center gap-1.5 px-3 text-sm tabular-nums">
              <span key={page} className="animate-in fade-in zoom-in-95 font-semibold text-white duration-200">
                {page}
              </span>
              <span className="text-zinc-500">/</span>
              <span className="text-zinc-400">{numPages}</span>
            </div>

            {isHost && (
              <button
                type="button"
                onClick={() => onPageChange(clamp(page + 1))}
                disabled={page >= numPages}
                className={cn(
                  "grid size-8 place-items-center rounded-full text-zinc-300 transition-all",
                  "hover:bg-white/10 hover:text-white active:scale-90",
                  "disabled:pointer-events-none disabled:opacity-30 rtl:rotate-180"
                )}
                aria-label={t("liveRoom.stage.nextPage")}
              >
                <ChevronRight className="size-5" />
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
