import { Skeleton } from "@/components/ui/skeleton"

/**
 * Loading placeholder for the conversation list. Mirrors {@link ConversationItem}
 * exactly — size-10 avatar, a name/time top row and a preview line — so the swap
 * to real data doesn't shift the layout.
 */
export function ConversationListSkeleton({ rows = 8 }: { rows?: number }) {
  return (
    <div className="flex flex-col gap-0.5 px-2 pt-1">
      {Array.from({ length: rows }, (_, i) => (
        <div key={i} className="flex items-center gap-3 rounded-xl px-2.5 py-2.5">
          <Skeleton className="size-10 shrink-0 rounded-full" />
          <div className="flex min-w-0 flex-1 flex-col gap-1.5">
            <div className="flex items-center justify-between gap-2">
              <Skeleton className="h-3.5 w-28" />
              <Skeleton className="h-3 w-8 shrink-0" />
            </div>
            <Skeleton className="h-3 w-40 max-w-full" />
          </div>
        </div>
      ))}
    </div>
  )
}
