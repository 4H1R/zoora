import { describe, expect, it } from "vitest"

import { countReaders, isReadBy, isReadByAll } from "./read-receipts"

describe("isReadBy", () => {
  it("is false when the pointer is missing or empty", () => {
    expect(isReadBy(undefined, "m5")).toBe(false)
    expect(isReadBy("", "m5")).toBe(false)
  })

  it("is false when the message id is empty", () => {
    expect(isReadBy("m5", "")).toBe(false)
  })

  it("is true when the pointer equals the message id", () => {
    expect(isReadBy("m5", "m5")).toBe(true)
  })

  it("is true when the pointer is lexically past the message id", () => {
    expect(isReadBy("m9", "m5")).toBe(true)
  })

  it("is false when the pointer is behind the message id", () => {
    expect(isReadBy("m4", "m5")).toBe(false)
  })

  it("compares uuidv7-style ids lexically (time order)", () => {
    const older = "0190a000-0000-7000-8000-000000000000"
    const newer = "0190b000-0000-7000-8000-000000000000"
    expect(isReadBy(newer, older)).toBe(true)
    expect(isReadBy(older, newer)).toBe(false)
    expect(isReadBy(older, older)).toBe(true)
  })
})

describe("countReaders", () => {
  it("counts only pointers at or past the message, excluding the author", () => {
    const pointers = { me: "m9", a: "m9", b: "m5", c: "m2" }
    // author=me excluded; a(>=m5) counts, b(=m5) counts, c(<m5) does not.
    expect(countReaders(pointers, "m5", "me")).toBe(2)
  })

  it("returns 0 when nobody else has reached the message", () => {
    expect(countReaders({ me: "m9", a: "m1" }, "m5", "me")).toBe(0)
  })

  it("ignores missing/empty pointers", () => {
    const pointers = { a: undefined, b: "", c: "m9" }
    expect(countReaders(pointers, "m5", "me")).toBe(1)
  })

  it("never counts the author even if their own pointer is ahead", () => {
    expect(countReaders({ me: "m9" }, "m5", "me")).toBe(0)
  })
})

describe("isReadByAll", () => {
  it("is false when there are no other members (solo conversation)", () => {
    expect(isReadByAll({ me: "m9" }, [], "m5")).toBe(false)
  })

  it("is true only when every other member has reached the message", () => {
    const pointers = { a: "m9", b: "m5" }
    expect(isReadByAll(pointers, ["a", "b"], "m5")).toBe(true)
  })

  it("is false when any other member is behind or missing", () => {
    expect(isReadByAll({ a: "m9", b: "m2" }, ["a", "b"], "m5")).toBe(false)
    expect(isReadByAll({ a: "m9" }, ["a", "b"], "m5")).toBe(false)
  })
})
