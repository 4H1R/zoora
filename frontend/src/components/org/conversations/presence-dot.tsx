import { cn } from "@/lib/utils"

interface PresenceDotProps {
  /** Online → a live emerald dot; offline/unknown → a muted dot. */
  online?: boolean
  /**
   * Ring the dot in the surrounding surface colour so it reads as a badge sitting
   * on an avatar. Turn off for inline (in-text) use.
   */
  ringed?: boolean
  className?: string
}

/**
 * A small presence indicator. Emerald when the user is online (with a soft glow),
 * muted otherwise. Positioned onto an avatar by the caller via logical props
 * (e.g. `absolute -bottom-0.5 -end-0.5`) so it stays RTL-correct.
 */
export function PresenceDot({ online = false, ringed = true, className }: PresenceDotProps) {
  return (
    <span
      aria-hidden
      className={cn(
        "block size-2.5 rounded-full transition-colors",
        online ? "bg-emerald-500" : "bg-muted-foreground/40",
        ringed && "ring-background ring-2",
        className
      )}
    />
  )
}
