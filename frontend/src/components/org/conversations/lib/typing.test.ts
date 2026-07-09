import { describe, expect, it } from "vitest"

import { activeTypers, markTyping, pruneExpired, typingCopy } from "./typing"

describe("markTyping", () => {
  it("records a fresh expiry TYPING_TTL_MS ahead of now", () => {
    const map = markTyping({}, "u1", 1000)
    expect(map).toEqual({ u1: 1000 + 5000 })
  })

  it("refreshes an existing user's expiry without dropping others", () => {
    const map = markTyping({ u1: 1000, u2: 2000 }, "u1", 5000)
    expect(map).toEqual({ u1: 5000 + 5000, u2: 2000 })
  })
})

describe("pruneExpired", () => {
  it("drops entries whose expiry has passed now", () => {
    const map = { u1: 1000, u2: 5000 }
    expect(pruneExpired(map, 2000)).toEqual({ u2: 5000 })
  })

  it("keeps entries whose expiry is still ahead", () => {
    const map = { u1: 5000, u2: 6000 }
    expect(pruneExpired(map, 1000)).toEqual(map)
  })

  it("returns the SAME reference when nothing expired (setState skip)", () => {
    const map = { u1: 5000 }
    expect(pruneExpired(map, 1000)).toBe(map)
  })
})

describe("activeTypers", () => {
  it("excludes expired entries", () => {
    const map = { u1: 1000, u2: 5000 }
    expect(activeTypers(map, 2000)).toEqual(["u2"])
  })

  it("preserves insertion order (stable ordering across renders)", () => {
    let map = {}
    map = markTyping(map, "b", 0)
    map = markTyping(map, "a", 0)
    map = markTyping(map, "c", 0)
    expect(activeTypers(map, 0)).toEqual(["b", "a", "c"])
  })

  it("returns [] for an empty map", () => {
    expect(activeTypers({}, 0)).toEqual([])
  })

  it("returns [] once every entry has expired", () => {
    const map = { u1: 100, u2: 200 }
    expect(activeTypers(map, 500)).toEqual([])
  })
})

describe("typingCopy", () => {
  it("returns null for no typers", () => {
    expect(typingCopy([])).toBeNull()
  })

  it("returns the singular key + name for one typer", () => {
    expect(typingCopy(["Ann"])).toEqual({ key: "conversations.typing.one", params: { name: "Ann" } })
  })

  it("returns the dual key + both names for two typers", () => {
    expect(typingCopy(["Ann", "Bo"])).toEqual({
      key: "conversations.typing.two",
      params: { name1: "Ann", name2: "Bo" },
    })
  })

  it("returns the 'many' key for three or more typers", () => {
    expect(typingCopy(["Ann", "Bo", "Cid"])).toEqual({ key: "conversations.typing.many" })
  })
})
