import type { ChatMessage } from "./messages"
import type { InfiniteData } from "@tanstack/react-query"

import { describe, expect, it } from "vitest"

import { applyReactionCounts, insertOptimistic, markStatus, removeMessage, replaceMessage } from "./optimistic"

type Cache = InfiniteData<ChatMessage[]>

function m(id: string, extra: Partial<ChatMessage> = {}): ChatMessage {
  return { id, ...extra }
}

function cache(pages: ChatMessage[][]): Cache {
  return { pages, pageParams: pages.map(() => ({})) }
}

describe("insertOptimistic", () => {
  it("creates a single-page cache when old is undefined", () => {
    const msg = m("a", { _status: "sending" })
    const out = insertOptimistic(undefined, msg)
    expect(out).toEqual({ pages: [[msg]], pageParams: [{}] })
  })

  it("creates a single page when the cache has no pages", () => {
    const msg = m("a", { _status: "sending" })
    const out = insertOptimistic(cache([]), msg)
    expect(out.pages).toEqual([[msg]])
  })

  it("appends to the LAST page", () => {
    const old = cache([[m("a")], [m("b"), m("c")]])
    const out = insertOptimistic(old, m("d", { _status: "sending" }))
    expect(out.pages[0].map((x) => x.id)).toEqual(["a"])
    expect(out.pages[1].map((x) => x.id)).toEqual(["b", "c", "d"])
  })

  it("replaces in place when the id already exists (idempotent retry)", () => {
    const old = cache([[m("a", { _status: "failed" })]])
    const out = insertOptimistic(old, m("a", { _status: "sending", content: "hi" }))
    expect(out.pages[0]).toHaveLength(1)
    expect(out.pages[0][0].content).toBe("hi")
  })

  it("returns a NEW object (immutability)", () => {
    const old = cache([[m("a")]])
    const out = insertOptimistic(old, m("b"))
    expect(out).not.toBe(old)
    expect(out.pages).not.toBe(old.pages)
    expect(old.pages[0].map((x) => x.id)).toEqual(["a"])
  })
})

describe("replaceMessage", () => {
  it("replaces by id across pages and clears _status", () => {
    const old = cache([[m("a")], [m("b", { _status: "sending" })]])
    const out = replaceMessage(old, m("b", { content: "server" }))!
    expect(out.pages[1][0]._status).toBeUndefined()
    expect(out.pages[1][0].content).toBe("server")
  })

  it("no-ops when the id is absent", () => {
    const old = cache([[m("a")]])
    const out = replaceMessage(old, m("zzz"))
    expect(out).toBe(old)
  })

  it("no-ops safely on undefined", () => {
    expect(replaceMessage(undefined, m("a"))).toBeUndefined()
  })
})

describe("removeMessage", () => {
  it("drops the matching message and keeps the rest", () => {
    const old = cache([[m("a"), m("b")], [m("c")]])
    const out = removeMessage(old, "b")!
    expect(out.pages[0].map((x) => x.id)).toEqual(["a"])
    expect(out.pages[1].map((x) => x.id)).toEqual(["c"])
  })

  it("removes from whichever page holds the id", () => {
    const old = cache([[m("a")], [m("b"), m("c")]])
    const out = removeMessage(old, "c")!
    expect(out.pages[1].map((x) => x.id)).toEqual(["b"])
  })

  it("no-ops when the id is absent", () => {
    const old = cache([[m("a")]])
    expect(removeMessage(old, "zzz")).toBe(old)
  })

  it("no-ops safely on undefined", () => {
    expect(removeMessage(undefined, "a")).toBeUndefined()
  })

  it("returns a NEW object and does not mutate the source", () => {
    const old = cache([[m("a"), m("b")]])
    const out = removeMessage(old, "a")!
    expect(out).not.toBe(old)
    expect(old.pages[0].map((x) => x.id)).toEqual(["a", "b"])
  })
})

describe("applyReactionCounts", () => {
  it("overwrites the matching message's reactions map", () => {
    const old = cache([[m("a", { reactions: { "👍": 1 } })], [m("b")]])
    const out = applyReactionCounts(old, "b", { "❤️": 2 })!
    expect(out.pages[1][0].reactions).toEqual({ "❤️": 2 })
    // Untouched message keeps its own map.
    expect(out.pages[0][0].reactions).toEqual({ "👍": 1 })
  })

  it("replaces (does not merge) the existing map", () => {
    const old = cache([[m("a", { reactions: { "👍": 3, "🎉": 1 } })]])
    const out = applyReactionCounts(old, "a", { "👍": 4 })!
    expect(out.pages[0][0].reactions).toEqual({ "👍": 4 })
  })

  it("finds the message on whichever page holds it", () => {
    const old = cache([[m("a")], [m("b"), m("c")]])
    const out = applyReactionCounts(old, "c", { "😂": 5 })!
    expect(out.pages[1][1].reactions).toEqual({ "😂": 5 })
  })

  it("no-ops when the id is absent", () => {
    const old = cache([[m("a")]])
    expect(applyReactionCounts(old, "zzz", { "👍": 1 })).toBe(old)
  })

  it("no-ops safely on undefined", () => {
    expect(applyReactionCounts(undefined, "a", { "👍": 1 })).toBeUndefined()
  })

  it("returns a NEW object and does not mutate the source", () => {
    const old = cache([[m("a", { reactions: { "👍": 1 } })]])
    const out = applyReactionCounts(old, "a", { "👍": 2 })!
    expect(out).not.toBe(old)
    expect(out.pages).not.toBe(old.pages)
    expect(old.pages[0][0].reactions).toEqual({ "👍": 1 })
  })
})

describe("markStatus", () => {
  it("flips the matching message to failed", () => {
    const old = cache([[m("a", { _status: "sending" })]])
    const out = markStatus(old, "a", "failed")!
    expect(out.pages[0][0]._status).toBe("failed")
  })

  it("flips a failed message back to sending (retry)", () => {
    const old = cache([[m("a", { _status: "failed" })]])
    const out = markStatus(old, "a", "sending")!
    expect(out.pages[0][0]._status).toBe("sending")
  })

  it("no-ops when the id is absent", () => {
    const old = cache([[m("a")]])
    const out = markStatus(old, "zzz", "failed")
    expect(out).toBe(old)
  })

  it("no-ops safely on undefined", () => {
    expect(markStatus(undefined, "a", "failed")).toBeUndefined()
  })
})
