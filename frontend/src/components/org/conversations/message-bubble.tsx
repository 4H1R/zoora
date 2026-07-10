import type { ChatMessage } from "./lib/messages"
import type { InfiniteData } from "@tanstack/react-query"
import type { ReactNode } from "react"

import { useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"

import { AttachmentBubble } from "./attachment-bubble"
import { useJumpToMessage } from "./jump-context"
import { formatTimeOfDay } from "./lib/chat-time"
import { mediaIdStrings } from "./lib/messages"
import { removeMessage } from "./lib/optimistic"
import { chatKeys } from "./lib/query-keys"
import { MessageContent } from "./message-content"
import { MessageContextMenu } from "./message-context-menu"
import { MessageStatus } from "./message-status"
import { ReactionBar } from "./reaction-bar"
import { useSendAttachments } from "./use-send-attachments"
import { useSendMessage } from "./use-send-message"

type MessagesCache = InfiniteData<ChatMessage[]>

interface MessageBubbleProps {
  message: ChatMessage
  convId: string
  /** Author is the signed-in user — aligns to the end side with accent color. */
  isOwn: boolean
  /** Conversation type — drives the read (double-tick) rule per bubble. */
  conversationType?: string
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
  isOwn,
  conversationType,
  isHighlighted = false,
}: MessageBubbleProps) {
  const { t, i18n } = useTranslation()
  const queryClient = useQueryClient()
  const jumpToMessage = useJumpToMessage()
  const { retry } = useSendMessage(convId)
  const { retry: retryAttachments, discard: discardAttachments } = useSendAttachments(convId)

  const status = message._status
  const time = formatTimeOfDay(message.created_at, i18n.language)
  const messageId = message.id ?? ""
  const key = chatKeys.messages(convId)

  // An attachment bubble carries client-only previews OR confirmed media ids;
  // failed retries/discards route through the attachment pipeline.
  const isAttachmentMsg = (message._attachments?.length ?? 0) > 0
  const hasMedia = isAttachmentMsg || mediaIdStrings(message).length > 0
  const hasContent = !!message.content?.trim()

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
  // Attachment bubbles also abort any in-flight uploads and revoke blob URLs.
  function discardFailed() {
    if (isAttachmentMsg) {
      discardAttachments(messageId)
      return
    }
    queryClient.setQueryData<MessagesCache>(key, (old) => removeMessage(old, messageId))
  }

  function retryFailed() {
    if (isAttachmentMsg) retryAttachments(messageId)
    else retry(messageId)
  }

  // Menu (tap on touch / right-click on desktop) only on confirmed messages — a
  // sending/failed bubble has no server id to act on, so it renders bare.
  const maybeWithMenu = (bubble: ReactNode) =>
    status ? (
      bubble
    ) : (
      <MessageContextMenu message={message} isOwn={isOwn} convId={convId}>
        {bubble}
      </MessageContextMenu>
    )

  return (
    <div className={cn("group/message flex flex-col", isOwn ? "items-start" : "items-end")}>
      {/* Reply preview: accent start-border strip, click jumps to the target. */}
      {replyToId && (
        <button
          type="button"
          onClick={() => jumpToMessage(replyToId)}
          className={cn(
            "border-primary/60 bg-muted/60 hover:bg-muted mb-0.5 flex max-w-[50%] flex-col items-start gap-0.5 rounded-lg border-s-2 px-2.5 py-1 text-start transition",
            isOwn ? "ms-1" : "me-1"
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

      <div className={cn("flex max-w-[50%] items-center gap-1", isOwn ? "flex-row" : "flex-row-reverse")}>
        {maybeWithMenu(
          <div
            className={cn(
              "relative w-fit rounded-2xl text-sm leading-relaxed shadow-sm transition duration-500",
              hasMedia && !hasContent ? "p-1.5" : "px-3 py-2",
              isOwn ? "bg-primary text-primary-foreground rounded-es-md" : "bg-muted text-foreground rounded-ee-md",
              status === "sending" && "opacity-60",
              isHighlighted && "ring-primary ring-offset-background ring-2 ring-offset-2"
            )}
          >
            {hasMedia && <AttachmentBubble message={message} convId={convId} isOwn={isOwn} />}

            {hasContent && <MessageContent content={message.content ?? ""} isOwn={isOwn} />}

            <div
              className={cn(
                "mt-1.5 flex items-center justify-end gap-1 ps-2 text-[10px] leading-none tabular-nums",
                isOwn ? "text-primary-foreground/70" : "text-muted-foreground"
              )}
            >
              {message.is_edited && <span className="italic">{t("conversations.thread.edited")}</span>}
              <time dateTime={message.created_at}>{time}</time>
              {/* Delivery status on own bubbles: clock (sending) → tick (sent) →
                  double tick (read). A failed send shows its own retry row below. */}
              {isOwn && status !== "failed" && (
                <MessageStatus
                  convId={convId}
                  messageId={messageId}
                  conversationType={conversationType}
                  status={status}
                />
              )}
            </div>
          </div>
        )}
      </div>

      {/* Failed send: inline error with retry / discard. */}
      {status === "failed" && (
        <div className={cn("mt-0.5 flex items-center gap-2 px-1 text-xs", isOwn ? "flex-row" : "flex-row-reverse")}>
          <span className="text-destructive">{t("conversations.thread.failed")}</span>
          <button type="button" className="text-primary font-medium hover:underline" onClick={retryFailed}>
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
