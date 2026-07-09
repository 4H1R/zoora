import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"

import { formatTimeOfDay } from "./lib/chat-time"
import type { ChatMessage } from "./lib/messages"

interface MessageBubbleProps {
  message: ChatMessage
  /** Author is the signed-in user — aligns to the end side with accent color. */
  isOwn: boolean
  /** Transient jump flash (5.3): ring the bubble, fading out via transition. */
  isHighlighted?: boolean
}

/**
 * A single message bubble: text content (whitespace preserved, URLs left plain
 * for now), a corner timestamp, an edited marker, and an optimistic-status
 * affordance. Mention highlighting, attachments, reply previews and reactions
 * arrive in later phases via the clearly-marked slots below — keep this clean.
 */
export function MessageBubble({ message, isOwn, isHighlighted = false }: MessageBubbleProps) {
  const { t, i18n } = useTranslation()
  const status = message._status
  const time = formatTimeOfDay(message.created_at, i18n.language)

  return (
    <div className={cn("flex flex-col", isOwn ? "items-end" : "items-start")}>
      {/* Phase 6: reply preview slot renders here (above the bubble). */}

      <div
        className={cn(
          "relative w-fit max-w-[min(85%,42rem)] rounded-2xl px-3 py-2 text-sm leading-relaxed shadow-sm transition duration-500",
          isOwn
            ? "bg-primary text-primary-foreground rounded-ee-md"
            : "bg-muted text-foreground rounded-es-md",
          status === "sending" && "opacity-60",
          isHighlighted && "ring-primary ring-offset-background ring-2 ring-offset-2"
        )}
      >
        <p className="break-words whitespace-pre-wrap">{message.content}</p>

        <div
          className={cn(
            "mt-1 flex items-center justify-end gap-1 text-[10px] leading-none tabular-nums",
            isOwn ? "text-primary-foreground/70" : "text-muted-foreground"
          )}
        >
          {message.is_edited && (
            <span className="italic">{t("conversations.thread.edited")}</span>
          )}
          <time dateTime={message.created_at}>{time}</time>
          {status === "failed" && (
            <span
              className="bg-destructive size-1.5 rounded-full"
              role="img"
              aria-label={t("conversations.thread.failed")}
            />
          )}
        </div>
      </div>

      {/* Phase 7: reactions row renders here (below the bubble). */}
    </div>
  )
}
