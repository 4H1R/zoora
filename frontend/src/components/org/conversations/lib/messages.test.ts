import { describe, expect, it } from "vitest"

import { dedupSortMessages, deriveCursors, groupMessages, reconcileOptimistic } from "./messages"
import type { ChatMessage } from "./messages"

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
