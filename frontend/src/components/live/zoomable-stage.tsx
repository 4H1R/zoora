import { type PointerEvent as ReactPointerEvent, type ReactNode, useRef, useState } from "react"

import { cn } from "@/lib/utils"

import { clampZoom, ZOOM_STEP, ZoomControls } from "./zoom-controls"

// Wraps the main stage video and lets any viewer zoom it in/out with the +/-
// controls, then drag to pan once zoomed past 1×. Zoom is purely local (CSS
// transform on the client) — it never touches the published track, so each
// participant frames the content for themselves.
export function ZoomableStage({ children, className }: { children: ReactNode; className?: string }) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [zoom, setZoom] = useState(1)
  const [offset, setOffset] = useState({ x: 0, y: 0 })
  const [dragging, setDragging] = useState(false)
  const drag = useRef<{ pointerId: number; startX: number; startY: number; baseX: number; baseY: number } | null>(null)

  // Panning is only meaningful while zoomed; clamp the offset so the scaled
  // frame never pulls its own edge into view (no empty gutters).
  const clamp = (next: { x: number; y: number }, z: number) => {
    const el = containerRef.current
    if (!el) return { x: 0, y: 0 }
    const maxX = (el.clientWidth * (z - 1)) / 2
    const maxY = (el.clientHeight * (z - 1)) / 2
    return {
      x: Math.max(-maxX, Math.min(maxX, next.x)),
      y: Math.max(-maxY, Math.min(maxY, next.y)),
    }
  }

  const applyZoom = (z: number) => {
    const next = clampZoom(z)
    setZoom(next)
    setOffset((o) => clamp(o, next))
  }

  const reset = () => {
    setZoom(1)
    setOffset({ x: 0, y: 0 })
  }

  const zoomed = zoom > 1

  const onPointerDown = (e: ReactPointerEvent<HTMLDivElement>) => {
    if (!zoomed) return
    e.currentTarget.setPointerCapture(e.pointerId)
    drag.current = { pointerId: e.pointerId, startX: e.clientX, startY: e.clientY, baseX: offset.x, baseY: offset.y }
    setDragging(true)
  }

  const onPointerMove = (e: ReactPointerEvent<HTMLDivElement>) => {
    const d = drag.current
    if (!d || d.pointerId !== e.pointerId) return
    setOffset(clamp({ x: d.baseX + (e.clientX - d.startX), y: d.baseY + (e.clientY - d.startY) }, zoom))
  }

  const endDrag = (e: ReactPointerEvent<HTMLDivElement>) => {
    if (drag.current?.pointerId !== e.pointerId) return
    drag.current = null
    setDragging(false)
  }

  // Wheel zoom keeps the natural "scroll to zoom" feel over the stage.
  const onWheel = (e: React.WheelEvent<HTMLDivElement>) => {
    if (e.deltaY === 0) return
    applyZoom(zoom + (e.deltaY < 0 ? ZOOM_STEP : -ZOOM_STEP))
  }

  return (
    <div
      ref={containerRef}
      className={cn("relative h-full w-full overflow-hidden rounded-2xl", className)}
      onWheel={onWheel}
    >
      <div
        className={cn(
          "size-full touch-none will-change-transform",
          // Animate zoom steps, but not the pan — a live drag must track the
          // pointer 1:1, not lag behind a 150ms tween.
          dragging ? "" : "transition-transform duration-150 ease-out",
          zoomed ? (dragging ? "cursor-grabbing" : "cursor-grab") : "cursor-default"
        )}
        style={{ transform: `translate(${offset.x}px, ${offset.y}px) scale(${zoom})` }}
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={endDrag}
        onPointerCancel={endDrag}
      >
        {children}
      </div>

      <ZoomControls zoom={zoom} onZoom={applyZoom} onReset={reset} />
    </div>
  )
}
