import type { WsEvent } from "./lib/ws-client"
import type { ReactNode } from "react"

import { useQueryClient } from "@tanstack/react-query"
import { createContext, useContext, useEffect, useRef, useState } from "react"

import { AUTH_TOKEN_KEY } from "@/api/mutator/custom-instance"
import { useGetUsersMe } from "@/api/users/users"
import { clientEnv } from "@/config/env"

import { chatKeys } from "./lib/query-keys"
import { ChatWsClient } from "./lib/ws-client"
import { resolveWsUrl } from "./lib/ws-url"
import { createChatEventHandler } from "./use-chat-ws"

type Status = "online" | "offline"

type ChatWsContextValue = {
  join: (convId: string) => void
  leave: (convId: string) => void
  typing: (convId: string) => void
  /** Subscribe to the RAW WS event stream (Phase 7 typing/presence). Returns an unsubscribe fn. */
  subscribe: (fn: (e: WsEvent) => void) => () => void
  /** Mark which thread is on-screen so the reducer can suppress its unread bumps. */
  setFocusedConvId: (id: string | null) => void
  status: Status
}

const ChatWsContext = createContext<ChatWsContextValue | null>(null)

/**
 * Owns the single chat WebSocket connection for an org that has the chat plan.
 * On mount it wires a `ChatWsClient` whose events (a) drive the React Query
 * cache reducer and (b) fan out to raw subscribers, then `connect()`s. The
 * connect effect has `[]` deps and reads the live reducer/subscribers through
 * refs, so it wires exactly one socket for the provider's lifetime.
 */
export function ChatProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient()
  const { data: meData } = useGetUsersMe()

  const clientRef = useRef<ChatWsClient | null>(null)
  const focusedConvIdRef = useRef<string | null>(null)
  const selfUserIdRef = useRef<string | null>(null)
  const subscribersRef = useRef<Set<(e: WsEvent) => void>>(new Set())
  const handleRef = useRef<(e: WsEvent) => void>(() => {})
  const [status, setStatus] = useState<Status>("offline")

  // Keep the self-user id fresh for the reducer (unread self-suppression).
  selfUserIdRef.current = (meData?.status === 200 && meData.data.data?.id) || null

  // Rebuild the cache reducer every render so it closes over the live
  // queryClient + ref readers, WITHOUT re-running the connect effect below.
  // (Deliberately not useCallback — React Compiler aside, the ref is the churn
  // barrier the socket depends on.)
  handleRef.current = createChatEventHandler({
    queryClient,
    getFocusedConvId: () => focusedConvIdRef.current,
    selfUserId: () => selfUserIdRef.current,
  })

  useEffect(() => {
    const client = new ChatWsClient(
      resolveWsUrl(clientEnv.VITE_WS_URL),
      () => localStorage.getItem(AUTH_TOKEN_KEY),
      (e) => {
        handleRef.current(e)
        for (const fn of subscribersRef.current) fn(e)
      },
      (s) => {
        setStatus(s)
        // NB: the chat socket's liveness must NOT drive React Query's global
        // `onlineManager` — a chat drop would otherwise pause every query and
        // mutation app-wide. We keep `status` purely for chat UI.
        if (s === "online") {
          // Reconnected (or first connect): the list may have drifted while
          // offline — refetch it, and ensure the focused thread is (re)joined.
          queryClient.invalidateQueries({ queryKey: chatKeys.conversations() })
          const focused = focusedConvIdRef.current
          if (focused) client.join(focused)
        }
      }
    )
    clientRef.current = client
    client.connect()
    return () => {
      client.close()
      clientRef.current = null
    }
    // Intentionally empty: the socket is wired once; live handlers come via refs.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const value: ChatWsContextValue = {
    join: (convId) => clientRef.current?.join(convId),
    leave: (convId) => clientRef.current?.leave(convId),
    typing: (convId) => clientRef.current?.typing(convId),
    subscribe: (fn) => {
      subscribersRef.current.add(fn)
      return () => {
        subscribersRef.current.delete(fn)
      }
    },
    setFocusedConvId: (id) => {
      focusedConvIdRef.current = id
    },
    status,
  }

  return <ChatWsContext.Provider value={value}>{children}</ChatWsContext.Provider>
}

/** Access the chat WS surface. Throws outside `<ChatProvider>`. */
export function useChatWs(): ChatWsContextValue {
  const ctx = useContext(ChatWsContext)
  if (!ctx) throw new Error("useChatWs must be used within <ChatProvider>")
  return ctx
}
