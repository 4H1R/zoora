import { useHotkeys } from "@tanstack/react-hotkeys"
import { ChevronLeft, ChevronRight, Loader2 } from "lucide-react"
import pdfWorkerUrl from "pdfjs-dist/build/pdf.worker.min.mjs?url"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { Document, Page, pdfjs } from "react-pdf"

import "react-pdf/dist/Page/AnnotationLayer.css"
import "react-pdf/dist/Page/TextLayer.css"

import { cn } from "@/lib/utils"

import { clampZoom, ZoomControls } from "./zoom-controls"

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
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState("")
  const scrollRef = useRef<HTMLDivElement | null>(null)
  const wheelLock = useRef(false)
  // Local per-viewer zoom. Multiplies the rendered page width so the page truly
  // overflows the scroll container — the existing overflow-auto then pans it, no
  // CSS transform needed. Wheel stays reserved for host page-nav; zoom is +/- only.
  const [zoom, setZoom] = useState(1)
  const applyZoom = (z: number) => setZoom(clampZoom(z))

  // Measure the scroll viewport width so the page renders at full width. Attach
  // the observer ONCE (an inline ref-callback gets a new identity every render,
  // so React would detach/reattach it — and spin up a fresh ResizeObserver —
  // on every render). The `scrollbar-gutter: stable` on the container keeps this
  // width constant whether or not the scrollbar is showing; without it, a tall
  // page's scrollbar shrinks the content width → page re-renders narrower →
  // scrollbar disappears → width grows → re-renders forever.
  useEffect(() => {
    const el = scrollRef.current
    if (!el) return
    const apply = (w: number) => setContainerWidth((prev) => (prev === w ? prev : w))
    const observer = new ResizeObserver(([entry]) => apply(entry.contentRect.width))
    observer.observe(el)
    apply(el.clientWidth)
    return () => observer.disconnect()
  }, [])

  // Fit the page to the viewport width, but cap it: on a wide monitor a full-bleed
  // page renders huge and forces horizontal scroll, which reads as broken. Capping
  // the base width keeps the slide at a comfortable reading size, centered on the
  // black stage (the inner wrapper's justify-center handles the centering). Zoom
  // still multiplies past the cap for anyone who wants to lean in. Portrait / tall
  // pages overflow vertically for a real scrollbar; INSET leaves room beside it.
  const INSET = 16
  const MAX_STAGE_WIDTH = 1280
  const fitWidth = containerWidth
    ? Math.max(1, Math.min(containerWidth - INSET, MAX_STAGE_WIDTH) * zoom)
    : undefined

  const clamp = (n: number) => Math.max(1, Math.min(n, numPages || 1))

  // Host: click the current-page number to jump directly to a page. Opens a tiny
  // inline input seeded with the current page; Enter commits, Escape/blur cancels.
  const openEditor = () => {
    if (!isHost) return
    setDraft(String(page))
    setEditing(true)
  }
  const commitDraft = () => {
    const n = parseInt(draft, 10)
    if (Number.isFinite(n)) onPageChange(clamp(n))
    setEditing(false)
  }

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
  // edge; Left/Right always flip. Ignored while typing in a field (e.g. chat) or
  // when a focused control needs the key itself. preventDefault/stopPropagation
  // are left off so the guard can bail without swallowing the key — we
  // preventDefault by hand only once we commit to navigating. The callbacks are
  // synced each render, so they always read the latest page/navigate.
  const navByKey = (e: KeyboardEvent, forward: boolean, flipOnly: boolean) => {
    const target = e.target as HTMLElement | null
    // Don't hijack Space/arrows away from an interactive element: a focused
    // button/select/link/menu item needs those keys for its own activation or
    // option navigation (Space "clicks" a button, arrows move select options).
    if (
      target?.isContentEditable ||
      target?.closest(
        "input, textarea, select, button, a[href], [role='button'], [role='menuitem'], [role='menuitemradio'], [role='option'], [role='tab'], [role='slider'], [contenteditable='true']"
      )
    )
      return
    e.preventDefault()
    const el = scrollRef.current
    if (el && !flipOnly) {
      const canScroll = el.scrollHeight > el.clientHeight + 1
      const atTop = el.scrollTop <= 0
      const atBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 1
      const step = el.clientHeight * 0.9
      if (canScroll && forward && !atBottom) return void (el.scrollTop += step)
      if (canScroll && !forward && !atTop) return void (el.scrollTop -= step)
    }
    navigate(forward ? page + 1 : page - 1)
  }
  useHotkeys(
    [
      { hotkey: "PageDown", callback: (e) => navByKey(e, true, false) },
      { hotkey: "ArrowDown", callback: (e) => navByKey(e, true, false) },
      { hotkey: "ArrowRight", callback: (e) => navByKey(e, true, true) },
      { hotkey: "Space", callback: (e) => navByKey(e, true, false) },
      { hotkey: "PageUp", callback: (e) => navByKey(e, false, false) },
      { hotkey: "ArrowUp", callback: (e) => navByKey(e, false, false) },
      { hotkey: "ArrowLeft", callback: (e) => navByKey(e, false, true) },
    ],
    { enabled: isHost, preventDefault: false, stopPropagation: false, ignoreInputs: false }
  )

  return (
    <div className="relative flex h-full w-full flex-col items-center justify-center overflow-hidden rounded-2xl bg-black">
      {/* PDF viewer — the page fills the viewport width. Tall pages overflow
          vertically so the scroll container shows a scrollbar and the wheel
          scrolls. min-h-full on the inner wrapper keeps short pages centered
          while still allowing a taller page to scroll from its top edge (a plain
          flex-center container would clip the overflowing top, making it
          unreachable). */}
      <div ref={scrollRef} onWheel={onWheel} className="min-h-0 w-full flex-1 [scrollbar-gutter:stable] overflow-auto">
        {/* w-max + min-w-full mirrors the min-h-full trick for the horizontal
            axis: when a zoomed page is wider than the viewport the wrapper grows
            to the page width so justify-center can't clip the left edge out of
            scroll reach; when it fits, min-w-full keeps it centered. */}
        <div className="flex min-h-full w-max min-w-full items-center justify-center py-2">
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
          <div className="pointer-events-auto flex items-center gap-0.5 rounded-full border border-white/10 bg-black/60 p-1 shadow-lg ring-1 shadow-black/50 ring-white/5 backdrop-blur-md">
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

            <div className="flex items-center gap-1.5 px-3 text-sm tabular-nums select-none">
              {editing ? (
                <input
                  type="number"
                  min={1}
                  max={numPages}
                  autoFocus
                  value={draft}
                  onChange={(e) => setDraft(e.target.value)}
                  onBlur={() => setEditing(false)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") commitDraft()
                    else if (e.key === "Escape") setEditing(false)
                  }}
                  aria-label={t("liveRoom.stage.goToPage")}
                  className={cn(
                    "w-10 rounded-md bg-white/10 text-center font-semibold text-white outline-none",
                    "ring-1 ring-white/20 focus:ring-white/40",
                    "[appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
                  )}
                />
              ) : isHost ? (
                <button
                  type="button"
                  onClick={openEditor}
                  className="animate-in fade-in zoom-in-95 rounded-md px-1 font-semibold text-white duration-200 hover:bg-white/10"
                  aria-label={t("liveRoom.stage.goToPage")}
                >
                  {page}
                </button>
              ) : (
                <span key={page} className="animate-in fade-in zoom-in-95 font-semibold text-white duration-200">
                  {page}
                </span>
              )}
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

      {!error && !loading && <ZoomControls zoom={zoom} onZoom={applyZoom} onReset={() => setZoom(1)} />}
    </div>
  )
}
