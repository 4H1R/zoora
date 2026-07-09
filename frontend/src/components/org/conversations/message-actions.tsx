import type { ChatMessage } from "./lib/messages"
import type { InfiniteData } from "@tanstack/react-query"

import { useQueryClient } from "@tanstack/react-query"
import { PencilIcon, PinIcon, PinOffIcon, ReplyIcon, Trash2Icon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useDeleteConversationsMessagesMessageId } from "@/api/conversations/conversations"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { useChatUi } from "@/stores/chat-ui"

import { removeMessage } from "./lib/optimistic"
import { chatKeys } from "./lib/query-keys"
import { ReactionPicker } from "./reaction-picker"
import { usePinActions } from "./use-pins"
import { useToggleReaction } from "./use-toggle-reaction"

type MessagesCache = InfiniteData<ChatMessage[]>

interface MessageActionsProps {
  message: ChatMessage
  /** Whether the signed-in user authored this message — gates edit/delete. */
  isOwn: boolean
  convId: string
  className?: string
}

/**
 * The per-message hover action row. Reply is offered on every message; Edit and
 * Delete only on the signed-in user's own. Delete confirms via an AlertDialog,
 * then optimistically drops the bubble from the thread cache and re-adds it on
 * error. Reply/Edit are pure UI intents routed through the chat-ui store (the
 * composer reacts). Future React/Pin actions mount in the marked slot below.
 */
export function MessageActions({ message, isOwn, convId, className }: MessageActionsProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const setReplyTo = useChatUi((s) => s.setReplyTo)
  const setEditing = useChatUi((s) => s.setEditing)
  const deleteMutation = useDeleteConversationsMessagesMessageId()
  const { toggle } = useToggleReaction(convId)
  const { pin, unpin } = usePinActions(convId)
  const [confirmOpen, setConfirmOpen] = useState(false)

  const messageId = message.id ?? ""
  const key = chatKeys.messages(convId)

  function handleDelete() {
    setConfirmOpen(false)
    // Snapshot for rollback, then optimistically drop the bubble.
    const snapshot = queryClient.getQueryData<MessagesCache>(key)
    queryClient.setQueryData<MessagesCache>(key, (old) => removeMessage(old, messageId))
    deleteMutation.mutate(
      { messageId },
      {
        onError: () => {
          // Re-add the removed bubble, then reconcile against the server.
          if (snapshot) queryClient.setQueryData(key, snapshot)
          queryClient.invalidateQueries({ queryKey: key })
        },
      }
    )
  }

  return (
    <div className={cn("flex items-center gap-0.5", className)}>
      {/* React: quick-row + full picker; opens toward the message. */}
      <ReactionPicker align={isOwn ? "end" : "start"} onSelect={(emoji) => toggle(messageId, emoji)} />

      <Button
        type="button"
        variant="ghost"
        size="icon-sm"
        className="text-muted-foreground hover:text-foreground size-7 rounded-full"
        aria-label={t("conversations.actions.reply")}
        onClick={() => setReplyTo(messageId)}
      >
        <ReplyIcon className="rtl:-scale-x-100" />
      </Button>

      {isOwn && (
        <>
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className="text-muted-foreground hover:text-foreground size-7 rounded-full"
            aria-label={t("conversations.actions.edit")}
            onClick={() => setEditing(messageId)}
          >
            <PencilIcon />
          </Button>

          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className="text-muted-foreground hover:text-destructive size-7 rounded-full"
            aria-label={t("conversations.actions.delete")}
            onClick={() => setConfirmOpen(true)}
          >
            <Trash2Icon />
          </Button>
        </>
      )}

      {/* Pin / unpin — offered on every message; backend enforces who may pin. */}
      <Button
        type="button"
        variant="ghost"
        size="icon-sm"
        className="text-muted-foreground hover:text-foreground size-7 rounded-full"
        aria-label={message.is_pinned ? t("conversations.actions.unpin") : t("conversations.actions.pin")}
        onClick={() => (message.is_pinned ? unpin(messageId) : pin(messageId))}
      >
        {message.is_pinned ? <PinOffIcon /> : <PinIcon />}
      </Button>

      <AlertDialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("conversations.actions.deleteConfirm.title")}</AlertDialogTitle>
            <AlertDialogDescription>{t("conversations.actions.deleteConfirm.description")}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("common.cancel")}</AlertDialogCancel>
            <AlertDialogAction variant="destructive" onClick={handleDelete}>
              {t("common.delete")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
