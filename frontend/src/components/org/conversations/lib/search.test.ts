import type { GithubCom4H1RZooraInternalDomainConversationMessage as ConversationMessage } from "@/api/model"

import { describe, expect, it } from "vitest"

import { matchesKey, nextMatchIndex } from "./search"

describe("nextMatchIndex", () => {
  it("advances forward within range", () => {
    expect(nextMatchIndex(0, 3, 1)).toBe(1)
    expect(nextMatchIndex(1, 3, 1)).toBe(2)
  })

  it("wraps forward from the last item to the first", () => {
    expect(nextMatchIndex(2, 3, 1)).toBe(0)
  })

  it("goes backward within range", () => {
    expect(nextMatchIndex(2, 3, -1)).toBe(1)
  })

  it("wraps backward from the first item to the last", () => {
    expect(nextMatchIndex(0, 3, -1)).toBe(2)
  })

  it("returns -1 when there are no matches", () => {
    expect(nextMatchIndex(0, 0, 1)).toBe(-1)
    expect(nextMatchIndex(0, 0, -1)).toBe(-1)
  })

  it("resolves an unset cursor of -1 forward to the first item", () => {
    expect(nextMatchIndex(-1, 4, 1)).toBe(0)
  })

  it("stays on the only item in a single-match list", () => {
    expect(nextMatchIndex(0, 1, 1)).toBe(0)
    expect(nextMatchIndex(0, 1, -1)).toBe(0)
  })
})

describe("matchesKey", () => {
  it("joins message ids in order", () => {
    const msgs: ConversationMessage[] = [{ id: "m1" }, { id: "m2" }, { id: "m3" }]
    expect(matchesKey(msgs)).toBe("m1,m2,m3")
  })

  it("changes when the set of matches changes", () => {
    const a: ConversationMessage[] = [{ id: "m1" }, { id: "m2" }]
    const b: ConversationMessage[] = [{ id: "m2" }, { id: "m3" }]
    expect(matchesKey(a)).not.toBe(matchesKey(b))
  })

  it("is empty for no matches", () => {
    expect(matchesKey([])).toBe("")
  })
})
