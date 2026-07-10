import type { ChatMessage } from "./lib/messages"
import type { InfiniteData } from "@tanstack/react-query"

import { useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { useDeleteConversationsMessagesMessageId } from "@/api/conversations/conversations"
import { useChatUi } from "@/stores/chat-ui"

import { removeMessage } from "./lib/optimistic"
import { chatKeys } from "./lib/query-keys"
import { usePinActions } from "./use-pins"
import { useToggleReaction } from "./use-toggle-reaction"

type MessagesCache = InfiniteData<ChatMessage[]>

/**
 * Shared per-message action handlers, so the hover row and the right-click
 * context menu stay in lock-step. Reply/Edit are pure UI intents routed through
 * the chat-ui store; React/Pin go through their hooks; Delete optimistically
 * drops the bubble and re-adds it on error. `delete` here fires immediately —
 * callers own the confirmation UI.
 */
export function useMessageActions(message: ChatMessage, convId: string) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const setReplyTo = useChatUi((s) => s.setReplyTo)
  const setEditing = useChatUi((s) => s.setEditing)
  const deleteMutation = useDeleteConversationsMessagesMessageId()
  const { toggle } = useToggleReaction(convId)
  const { pin, unpin } = usePinActions(convId)

  const messageId = message.id ?? ""
  const key = chatKeys.messages(convId)

  return {
    reply: () => setReplyTo(messageId),
    edit: () => setEditing(messageId),
    react: (emoji: string) => toggle(messageId, emoji),
    togglePin: () => (message.is_pinned ? unpin(messageId) : pin(messageId)),
    remove: () => {
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
            toast.error(t("conversations.actions.deleteError"))
          },
        }
      )
    },
  }
}
