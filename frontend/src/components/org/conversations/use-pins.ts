import type { ChatMessage } from "./lib/messages"
import type { getConversationsIdPinsResponse } from "@/api/conversations/conversations"

import { useQuery, useQueryClient } from "@tanstack/react-query"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getConversationsIdPins,
  usePostConversationsMessagesMessageIdPin,
  usePostConversationsMessagesMessageIdUnpin,
} from "@/api/conversations/conversations"

import { chatKeys } from "./lib/query-keys"

/**
 * Unwrap the orval pins response into a flat `ChatMessage[]`. The endpoint
 * returns the pinned messages ordered newest-pin-first (`pinned_at` DESC) under
 * `data.data`; we surface that order unchanged so the bar can preview `pins[0]`
 * as the most recent. Throws on a non-200 so React Query surfaces the error
 * (the orval fetcher resolves error statuses).
 */
export function unwrapPins(res: getConversationsIdPinsResponse): ChatMessage[] {
  if (res.status !== 200) {
    throw new Error(`Failed to load pins (status ${res.status})`)
  }
  return res.data.data ?? []
}

/**
 * Pinned messages for a conversation, `pinned_at` DESC. Keyed on
 * `chatKeys.pins(convId)` so `usePinActions` (and any future WS reconciler) can
 * invalidate the exact cache. There is no WS echo for pin/unpin, so freshness
 * comes solely from the manual invalidations in `usePinActions`.
 */
export function usePins(convId: string) {
  const query = useQuery({
    queryKey: chatKeys.pins(convId),
    queryFn: async ({ signal }) => unwrapPins(await getConversationsIdPins(convId, { signal })),
    staleTime: Infinity,
    enabled: Boolean(convId),
  })

  return {
    pins: query.data ?? [],
    isLoading: query.isLoading,
  }
}

/**
 * Pin/unpin a message. Both endpoints take no body and flip server-side. Since
 * there is no realtime echo for pinning, `onSuccess` invalidates two caches:
 *  - the pins list (`chatKeys.pins`), so the bar re-fetches;
 *  - the message thread (`chatKeys.messages`), so each bubble's `is_pinned`
 *    flag refreshes and the per-message action toggles its label/icon.
 */
export function usePinActions(convId: string) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const pinMutation = usePostConversationsMessagesMessageIdPin()
  const unpinMutation = usePostConversationsMessagesMessageIdUnpin()

  function invalidate() {
    queryClient.invalidateQueries({ queryKey: chatKeys.pins(convId) })
    queryClient.invalidateQueries({ queryKey: chatKeys.messages(convId) })
  }

  const onError = () => toast.error(t("conversations.actions.pinError"))

  function pin(messageId: string) {
    pinMutation.mutate({ messageId }, { onSuccess: invalidate, onError })
  }

  function unpin(messageId: string) {
    unpinMutation.mutate({ messageId }, { onSuccess: invalidate, onError })
  }

  return {
    pin,
    unpin,
    isPending: pinMutation.isPending || unpinMutation.isPending,
  }
}
