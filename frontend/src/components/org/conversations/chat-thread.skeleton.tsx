import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"

// Mirrors the loaded thread density: avatar + stacked bubbles per group,
// alternating start/end alignment with varied widths. Keep in sync with
// <MessageGroup> / <MessageBubble> so the swap-in is seamless.
const ROWS: { own: boolean; widths: string[] }[] = [
  { own: false, widths: ["w-40", "w-56"] },
  { own: true, widths: ["w-32"] },
  { own: false, widths: ["w-64"] },
  { own: true, widths: ["w-48", "w-28"] },
  { own: false, widths: ["w-36", "w-52", "w-24"] },
  { own: true, widths: ["w-44"] },
]

export function ChatThreadSkeleton() {
  return (
    <div className="flex flex-1 flex-col justify-end gap-3 overflow-hidden p-4">
      {ROWS.map((row, i) => (
        <div
          key={i}
          className={cn(
            "flex items-end gap-2",
            row.own ? "flex-row-reverse ps-12" : "flex-row pe-12"
          )}
        >
          {row.own ? (
            <div className="w-8 shrink-0" />
          ) : (
            <Skeleton className="size-8 shrink-0 rounded-full" />
          )}
          <div className={cn("flex flex-col gap-1.5", row.own ? "items-end" : "items-start")}>
            {row.widths.map((w, j) => (
              <Skeleton key={j} className={cn("h-9 rounded-2xl", w)} />
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}
