import type { ChatMessage } from "./messages"

import { describe, expect, it } from "vitest"

import {
  dedupSortMessages,
  deriveCursors,
  findGroupIndex,
  groupMessages,
  newestReadableId,
  nextPageParam,
  prevPageParam,
  reconcileOptimistic,
} from "./messages"

// Build a page of `n` ASCENDING messages whose ids sort after `startCode`.
function page(startCode: number, n: number): ChatMessage[] {
  return Array.from({ length: n }, (_, i) => ({ id: String.fromCharCode(startCode + i) }) as ChatMessage)
}

// Build a message with an id and a created_at derived from a base time + offset seconds.
const BASE = new Date("2026-07-09T10:00:00.000Z").getTime()
function msg(id: string, senderId: string, offsetSeconds = 0, extra: Partial<ChatMessage> = {}): ChatMessage {
  return {
    id,
    sender_id: senderId,
    created_at: new Date(BASE + offsetSeconds * 1000).toISOString(),
    ...extra,
  }
}

describe("dedupSortMessages", () => {
  it("dedups by id, sorts ascending", () => {
    expect(dedupSortMessages([{ id: "b" }, { id: "a" }, { id: "b" }] as any).map((m) => m.id)).toEqual(["a", "b"])
  })

  it("last-write-wins on duplicate id", () => {
    const out = dedupSortMessages([
      { id: "a", content: "first" },
      { id: "a", content: "second" },
    ] as any)
    expect(out).toHaveLength(1)
    expect(out[0].content).toBe("second")
  })

  it("empty stays empty", () => {
    expect(dedupSortMessages([])).toEqual([])
  })
})

describe("deriveCursors", () => {
  it("before=first, after=last", () => {
    expect(deriveCursors([{ id: "a" }, { id: "b" }, { id: "c" }] as any)).toEqual({ before: "a", after: "c" })
  })

  it("empty -> nulls", () => expect(deriveCursors([])).toEqual({ before: null, after: null }))
})

describe("groupMessages", () => {
  it("groups consecutive same-sender within 5min", () => {
    const groups = groupMessages([msg("a", "u1", 0), msg("b", "u1", 60)])
    const messageGroups = groups.filter((g) => g.type === "messages")
    const dayGroups = groups.filter((g) => g.type === "day")
    expect(dayGroups).toHaveLength(1)
    expect(messageGroups).toHaveLength(1)
    expect(messageGroups[0].messages.map((m) => m.id)).toEqual(["a", "b"])
    expect(messageGroups[0].senderId).toBe("u1")
  })

  it("splits on sender change", () => {
    const groups = groupMessages([msg("a", "u1", 0), msg("b", "u2", 60)])
    const messageGroups = groups.filter((g) => g.type === "messages")
    expect(messageGroups).toHaveLength(2)
    expect(messageGroups[0].senderId).toBe("u1")
    expect(messageGroups[1].senderId).toBe("u2")
  })

  it("splits same sender when gap > 5min", () => {
    const groups = groupMessages([msg("a", "u1", 0), msg("b", "u1", 6 * 60)])
    const messageGroups = groups.filter((g) => g.type === "messages")
    expect(messageGroups).toHaveLength(2)
  })

  it("inserts day divider when date changes", () => {
    const day2 = 26 * 60 * 60 // +26h -> next calendar day
    const groups = groupMessages([msg("a", "u1", 0), msg("b", "u1", day2)])
    const dayGroups = groups.filter((g) => g.type === "day")
    expect(dayGroups).toHaveLength(2)
    // day divider precedes its messages
    expect(groups[0].type).toBe("day")
  })

  it("uses first message id as message-group id and day-<isoDate> for divider", () => {
    const groups = groupMessages([msg("a", "u1", 0)])
    expect(groups[0].type).toBe("day")
    expect(groups[0].id).toMatch(/^day-\d{4}-\d{2}-\d{2}$/)
    expect(groups[1].type).toBe("messages")
    expect(groups[1].id).toBe("a")
  })

  it("empty -> no groups", () => {
    expect(groupMessages([])).toEqual([])
  })
})

describe("nextPageParam (NEWER / bottom)", () => {
  const LIMIT = 3

  it("latest-seed never fetches newer (WS appends live)", () => {
    // Full latest page — still no newer page: the latest page IS the newest.
    expect(nextPageParam([page(97, 3)], page(97, 3), LIMIT, false)).toBeUndefined()
  })

  it("around-seed with a FULL page allows a next (after) cursor at the newest id", () => {
    const p = page(97, 3) // a,b,c
    expect(nextPageParam([p], p, LIMIT, true)).toEqual({ after: "c" })
  })

  it("around-seed with a SHORT page is exhausted (undefined)", () => {
    const p = page(97, 2) // a,b (< limit)
    expect(nextPageParam([p], p, LIMIT, true)).toBeUndefined()
  })

  it("around-seed empty page -> undefined", () => {
    expect(nextPageParam([[]], [], LIMIT, true)).toBeUndefined()
  })

  it("around-seed uses the OVERALL newest id across all loaded pages", () => {
    const older = page(97, 3) // a,b,c (prepended top)
    const bottom = page(100, 3) // d,e,f (newest-position page)
    expect(nextPageParam([older, bottom], bottom, LIMIT, true)).toEqual({ after: "f" })
  })
})

describe("prevPageParam (OLDER / top)", () => {
  const LIMIT = 3

  it("full first page allows a previous (before) cursor at the oldest id", () => {
    const p = page(97, 3) // a,b,c
    expect(prevPageParam([p], p, LIMIT)).toEqual({ before: "a" })
  })

  it("short first page is exhausted (undefined)", () => {
    const p = page(97, 2) // a,b (< limit)
    expect(prevPageParam([p], p, LIMIT)).toBeUndefined()
  })

  it("empty first page -> undefined", () => {
    expect(prevPageParam([[]], [], LIMIT)).toBeUndefined()
  })

  it("uses the OVERALL oldest id across all loaded pages", () => {
    const top = page(97, 3) // a,b,c (oldest-position page)
    const bottom = page(100, 3) // d,e,f
    expect(prevPageParam([top, bottom], top, LIMIT)).toEqual({ before: "a" })
  })
})

describe("findGroupIndex", () => {
  it("returns the index of the messages group containing the id (past a leading day divider)", () => {
    const groups = groupMessages([msg("a", "u1", 0), msg("b", "u2", 60)])
    // groups: [day, {a}, {b}] — day divider offsets the message groups by one.
    expect(groups[0].type).toBe("day")
    expect(findGroupIndex(groups, "a")).toBe(1)
    expect(findGroupIndex(groups, "b")).toBe(2)
  })

  it("finds an id nested inside a multi-message group", () => {
    const groups = groupMessages([msg("a", "u1", 0), msg("b", "u1", 60)])
    // groups: [day, {a,b}] — both live in the same group at index 1.
    expect(findGroupIndex(groups, "a")).toBe(1)
    expect(findGroupIndex(groups, "b")).toBe(1)
  })

  it("respects day-divider offsets across days", () => {
    const day2 = 26 * 60 * 60 // +26h -> next calendar day
    const groups = groupMessages([msg("a", "u1", 0), msg("b", "u1", day2)])
    // groups: [day1, {a}, day2, {b}]
    expect(findGroupIndex(groups, "a")).toBe(1)
    expect(findGroupIndex(groups, "b")).toBe(3)
  })

  it("returns -1 when the message is not loaded", () => {
    const groups = groupMessages([msg("a", "u1", 0)])
    expect(findGroupIndex(groups, "zzz")).toBe(-1)
  })

  it("returns -1 for empty groups", () => {
    expect(findGroupIndex([], "a")).toBe(-1)
  })
})

describe("newestReadableId", () => {
  it("returns the newest (last) non-optimistic id", () => {
    expect(newestReadableId([msg("a", "u1"), msg("b", "u1")])).toBe("b")
  })

  it("skips trailing optimistic bubbles and returns the newest confirmed id", () => {
    const messages = [
      msg("a", "u1"),
      msg("b", "u1"),
      msg("c", "u1", 0, { _status: "sending" }),
      msg("d", "u1", 0, { _status: "failed" }),
    ]
    expect(newestReadableId(messages)).toBe("b")
  })

  it("returns null when every message is optimistic", () => {
    const messages = [msg("a", "u1", 0, { _status: "sending" }), msg("b", "u1", 0, { _status: "failed" })]
    expect(newestReadableId(messages)).toBeNull()
  })

  it("returns null for an empty list", () => {
    expect(newestReadableId([])).toBeNull()
  })
})

describe("reconcileOptimistic", () => {
  it("same id replaces optimistic, clears _status", () => {
    const out = reconcileOptimistic([{ id: "x", _status: "sending" }] as any, { id: "x" } as any)
    expect(out.filter((m) => m.id === "x")).toHaveLength(1)
    expect(out[0]._status).toBeUndefined()
  })

  it("new id appends", () => {
    const out = reconcileOptimistic([{ id: "x" }] as any, { id: "y" } as any)
    expect(out.map((m) => m.id)).toEqual(["x", "y"])
  })

  it("returns a new array (immutability)", () => {
    const existing = [{ id: "x" }] as any
    const out = reconcileOptimistic(existing, { id: "y" } as any)
    expect(out).not.toBe(existing)
  })
})
