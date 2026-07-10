import { QueryClient } from "@tanstack/react-query"
import { describe, expect, it } from "vitest"

import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

import type { ChatMessage } from "./lib/messages"
import { chatKeys } from "./lib/query-keys"
import {
  appendMessageToInfinite,
  bumpConversationInList,
  createChatEventHandler,
  type MessagesInfinite,
} from "./use-chat-ws"

function msg(id: string, extra: Partial<ChatMessage> = {}): ChatMessage {
  return { id, conversation_id: "c1", sender_id: "u2", content: id, created_at: "2026-07-09T10:00:00Z", ...extra }
}

function infinite(...pages: ChatMessage[][]): MessagesInfinite {
  return { pages, pageParams: pages.map(() => null) }
}

describe("appendMessageToInfinite", () => {
  it("no-ops when the cache is absent", () => {
    expect(appendMessageToInfinite(undefined, msg("a"))).toBeUndefined()
  })

  it("no-ops when there are no pages", () => {
    const old = infinite()
    expect(appendMessageToInfinite(old, msg("a"))).toBe(old)
  })

  it("appends to the last page", () => {
    const out = appendMessageToInfinite(infinite([msg("a")], [msg("b")]), msg("c"))
    expect(out?.pages[0].map((m) => m.id)).toEqual(["a"])
    expect(out?.pages[1].map((m) => m.id)).toEqual(["b", "c"])
  })

  it("reconciles an optimistic bubble in place (no duplicate) and clears _status", () => {
    const out = appendMessageToInfinite(infinite([msg("a", { _status: "sending" })]), msg("a"))
    expect(out?.pages[0].map((m) => m.id)).toEqual(["a"])
    expect(out?.pages[0][0]._status).toBeUndefined()
  })
})

describe("bumpConversationInList", () => {
  const list = (): Conversation[] => [
    { id: "c1", unread_count: 0 },
    { id: "c2", unread_count: 3 },
  ]

  it("no-ops when the cache is absent", () => {
    expect(bumpConversationInList(undefined, { convId: "c1", incrementUnread: true })).toBeUndefined()
  })

  it("no-ops when the conversation is not in the list", () => {
    const old = list()
    expect(bumpConversationInList(old, { convId: "zzz", incrementUnread: true })).toBe(old)
  })

  it("moves the conversation to the top, sets last_message and increments unread", () => {
    const out = bumpConversationInList(list(), { convId: "c2", message: msg("m1"), incrementUnread: true })
    expect(out?.map((c) => c.id)).toEqual(["c2", "c1"])
    expect(out?.[0].unread_count).toBe(4)
    expect(out?.[0].last_message?.id).toBe("m1")
  })

  it("does NOT increment unread when incrementUnread is false", () => {
    const out = bumpConversationInList(list(), { convId: "c1", message: msg("m1"), incrementUnread: false })
    expect(out?.[0].id).toBe("c1")
    expect(out?.[0].unread_count).toBe(0)
  })
})

describe("createChatEventHandler new_message", () => {
  // `new_message` arrives ONLY for the joined (focused/open) room: it appends the
  // full message to the thread cache AND bumps the list (unread suppressed, since
  // it's the focused conv or our own send).
  function setup() {
    const qc = new QueryClient()
    qc.setQueryData<MessagesInfinite>(chatKeys.messages("c1"), infinite([msg("m0")]))
    qc.setQueryData<Conversation[]>(chatKeys.conversations(), [{ id: "c1", unread_count: 0 }])
    return qc
  }

  it("appends the full message to the thread and bumps the list for the focused conv (unread suppressed)", () => {
    const qc = setup()
    const handle = createChatEventHandler({
      queryClient: qc,
      getFocusedConvId: () => "c1",
      selfUserId: () => "me",
    })
    handle({ type: "new_message", data: msg("m1", { sender_id: "u2" }) })

    const thread = qc.getQueryData<MessagesInfinite>(chatKeys.messages("c1"))
    expect(thread?.pages[0].map((m) => m.id)).toEqual(["m0", "m1"])
    const conv = qc.getQueryData<Conversation[]>(chatKeys.conversations())
    expect(conv?.[0].unread_count).toBe(0)
    expect(conv?.[0].last_message?.id).toBe("m1")
  })

  it("suppresses the unread bump for the caller's own message", () => {
    const qc = setup()
    const handle = createChatEventHandler({
      queryClient: qc,
      getFocusedConvId: () => null,
      selfUserId: () => "me",
    })
    handle({ type: "new_message", data: msg("m1", { sender_id: "me" }) })
    const thread = qc.getQueryData<MessagesInfinite>(chatKeys.messages("c1"))
    expect(thread?.pages[0].map((m) => m.id)).toEqual(["m0", "m1"])
    expect(qc.getQueryData<Conversation[]>(chatKeys.conversations())?.[0].unread_count).toBe(0)
  })

  it("is idempotent by id on a repeated sighting (no duplicate thread append)", () => {
    const qc = setup()
    const handle = createChatEventHandler({
      queryClient: qc,
      getFocusedConvId: () => "c1",
      selfUserId: () => "me",
    })
    handle({ type: "new_message", data: msg("m1", { sender_id: "u2" }) })
    handle({ type: "new_message", data: msg("m1", { sender_id: "u2" }) })
    const thread = qc.getQueryData<MessagesInfinite>(chatKeys.messages("c1"))
    expect(thread?.pages[0].map((m) => m.id)).toEqual(["m0", "m1"])
  })
})

describe("createChatEventHandler conversation_bump", () => {
  // `conversation_bump` is the per-user sidebar firehose (compact payload). It
  // bumps the LIST only — never the thread cache.
  function setup() {
    const qc = new QueryClient()
    qc.setQueryData<MessagesInfinite>(chatKeys.messages("c2"), infinite([msg("m0", { conversation_id: "c2" })]))
    qc.setQueryData<Conversation[]>(chatKeys.conversations(), [
      { id: "c1", unread_count: 0 },
      { id: "c2", unread_count: 0 },
    ])
    return qc
  }

  const bump = (extra: Record<string, unknown> = {}) => ({
    type: "conversation_bump",
    data: { conversation_id: "c2", id: "m9", sender_id: "u2", content: "hi", created_at: "2026-07-09T10:00:00Z", ...extra },
  })

  it("bumps an unfocused conv from another sender (move-to-top, preview, unread) without touching the thread", () => {
    const qc = setup()
    const handle = createChatEventHandler({
      queryClient: qc,
      getFocusedConvId: () => null,
      selfUserId: () => "me",
    })
    handle(bump())

    const conv = qc.getQueryData<Conversation[]>(chatKeys.conversations())
    expect(conv?.map((c) => c.id)).toEqual(["c2", "c1"])
    expect(conv?.[0].unread_count).toBe(1)
    expect(conv?.[0].last_message?.id).toBe("m9")
    expect(conv?.[0].last_message?.content).toBe("hi")
    // Thread cache is untouched — only the full-payload `new_message` appends.
    const thread = qc.getQueryData<MessagesInfinite>(chatKeys.messages("c2"))
    expect(thread?.pages[0].map((m) => m.id)).toEqual(["m0"])
  })

  it("suppresses the unread bump for the focused conversation", () => {
    const qc = setup()
    const handle = createChatEventHandler({
      queryClient: qc,
      getFocusedConvId: () => "c2",
      selfUserId: () => "me",
    })
    handle(bump())
    const conv = qc.getQueryData<Conversation[]>(chatKeys.conversations())
    expect(conv?.[0].id).toBe("c2")
    expect(conv?.[0].unread_count).toBe(0)
    expect(conv?.[0].last_message?.id).toBe("m9")
  })

  it("suppresses the unread bump for the caller's own message", () => {
    const qc = setup()
    const handle = createChatEventHandler({
      queryClient: qc,
      getFocusedConvId: () => null,
      selfUserId: () => "me",
    })
    handle(bump({ sender_id: "me" }))
    expect(qc.getQueryData<Conversation[]>(chatKeys.conversations())?.find((c) => c.id === "c2")?.unread_count).toBe(0)
  })
})
