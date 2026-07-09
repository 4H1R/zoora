import type { MentionCandidate } from "./lib/mentions"
import type { ChatMessage } from "./lib/messages"
import type { InfiniteData } from "@tanstack/react-query"

import { useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"

import { useJumpToMessage } from "./jump-context"
import { formatTimeOfDay } from "./lib/chat-time"
import { removeMessage } from "./lib/optimistic"
import { chatKeys } from "./lib/query-keys"
import { MessageActions } from "./message-actions"
import { MessageContent } from "./message-content"
import { ReactionBar } from "./reaction-bar"
import { ReadReceipt } from "./read-receipt"
import { useSendMessage } from "./use-send-message"

type MessagesCache = InfiniteData<ChatMessage[]>

interface MessageBubbleProps {
  message: ChatMessage
  convId: string
  /** Conversation members, for @mention highlighting inside the content. */
  members: MentionCandidate[]
  /** Author is the signed-in user — aligns to the end side with accent color. */
  isOwn: boolean
  /** Conversation type — DM gets a tick receipt, group a "read by N" caption. */
  conversationType?: string
  /** This is the user's newest confirmed message — gates the group read receipt. */
  isLatestOwn?: boolean
  /** Transient jump flash (5.3): ring the bubble, fading out via transition. */
  isHighlighted?: boolean
}

/**
 * A single message bubble: an optional reply-preview strip, the text content
 * (mention-highlighted, whitespace preserved), a corner timestamp + edited
 * marker, and a hover action row. Optimistic bubbles carry a `_status`: while
 * `sending` the bubble dims; when `failed` it swaps the action row for an inline
 * retry / discard affordance. Reactions arrive in a later phase via the slot
 * below.
 */
export function MessageBubble({
  message,
  convId,
  members,
  isOwn,
  conversationType,
  isLatestOwn = false,
  isHighlighted = false,
}: MessageBubbleProps) {
  const { t, i18n } = useTranslation()
  const queryClient = useQueryClient()
  const jumpToMessage = useJumpToMessage()
  const { retry } = useSendMessage(convId)

  const status = message._status
  const time = formatTimeOfDay(message.created_at, i18n.language)
  const messageId = message.id ?? ""
  const key = chatKeys.messages(convId)

  // The referenced message for the reply preview, read live from the cache.
  // Non-reactive (mirrors the composer): the target is usually already loaded.
  const replyToId = message.reply_to_message_id
  const repliedTo = replyToId
    ? queryClient
        .getQueryData<MessagesCache>(key)
        ?.pages.flat()
        .find((m) => m.id === replyToId)
    : undefined

  // Drop a failed optimistic bubble that never reached the server — purely local.
  function discardFailed() {
    queryClient.setQueryData<MessagesCache>(key, (old) => removeMessage(old, messageId))
  }

  return (
    <div className={cn("group/message flex flex-col", isOwn ? "items-end" : "items-start")}>
      {/* Reply preview: accent start-border strip, click jumps to the target. */}
      {replyToId && (
        <button
          type="button"
          onClick={() => jumpToMessage(replyToId)}
          className={cn(
            "border-primary/60 bg-muted/60 hover:bg-muted mb-0.5 flex max-w-[min(85%,42rem)] flex-col items-start gap-0.5 rounded-lg border-s-2 px-2.5 py-1 text-start transition",
            isOwn ? "me-1" : "ms-1"
          )}
        >
          <span className="text-primary text-xs font-semibold">
            {repliedTo?.sender?.name ?? t("conversations.thread.replyTo")}
          </span>
          <span className="text-muted-foreground line-clamp-1 text-xs">
            {repliedTo?.content ?? t("conversations.thread.replyUnavailable")}
          </span>
        </button>
      )}

      <div className={cn("flex items-center gap-1", isOwn ? "flex-row-reverse" : "flex-row")}>
        <div
          className={cn(
            "relative w-fit max-w-[min(85%,42rem)] rounded-2xl px-3 py-2 text-sm leading-relaxed shadow-sm transition duration-500",
            isOwn ? "bg-primary text-primary-foreground rounded-ee-md" : "bg-muted text-foreground rounded-es-md",
            status === "sending" && "opacity-60",
            isHighlighted && "ring-primary ring-offset-background ring-2 ring-offset-2"
          )}
        >
          <MessageContent content={message.content ?? ""} members={members} isOwn={isOwn} />

          <div
            className={cn(
              "mt-1 flex items-center justify-end gap-1 text-[10px] leading-none tabular-nums",
              isOwn ? "text-primary-foreground/70" : "text-muted-foreground"
            )}
          >
            {message.is_edited && <span className="italic">{t("conversations.thread.edited")}</span>}
            <time dateTime={message.created_at}>{time}</time>
            {/* Read receipt: own + confirmed only (tick for DMs, "read by N" for groups). */}
            {isOwn && !status && (
              <ReadReceipt
                convId={convId}
                messageId={messageId}
                conversationType={conversationType}
                isLatestOwn={isLatestOwn}
              />
            )}
          </div>
        </div>

        {/* Actions only on confirmed messages — not while sending/failed. */}
        {!status && (
          <MessageActions
            message={message}
            isOwn={isOwn}
            convId={convId}
            className="opacity-0 transition group-focus-within/message:opacity-100 group-hover/message:opacity-100"
          />
        )}
      </div>

      {/* Failed send: inline error with retry / discard. */}
      {status === "failed" && (
        <div className={cn("mt-0.5 flex items-center gap-2 px-1 text-xs", isOwn ? "flex-row-reverse" : "flex-row")}>
          <span className="text-destructive">{t("conversations.thread.failed")}</span>
          <button type="button" className="text-primary font-medium hover:underline" onClick={() => retry(messageId)}>
            {t("conversations.actions.retry")}
          </button>
          <button
            type="button"
            className="text-muted-foreground hover:text-foreground hover:underline"
            onClick={discardFailed}
          >
            {t("common.delete")}
          </button>
        </div>
      )}

      {/* Reactions row: pills aligned to the bubble side, non-empty only. */}
      {!status && <ReactionBar message={message} convId={convId} />}
    </div>
  )
}
