import { describe, expect, it } from "vitest"

import type { MentionCandidate } from "./mentions"
import { detectMention, insertMention, resolveMentions } from "./mentions"

const members: MentionCandidate[] = [
  { id: "u-ali", name: "Ali" },
  { id: "u-ali-a", name: "Ali Alizadeh" },
  { id: "u-sara", name: "Sara" },
]

describe("detectMention", () => {
  it("detects an @token at the caret at the start of text", () => {
    expect(detectMention("@al")).toEqual({ token: "al", atIndex: 0 })
  })

  it("detects an @token after whitespace", () => {
    expect(detectMention("hi @sar")).toEqual({ token: "sar", atIndex: 3 })
  })

  it("detects a bare @ with an empty token", () => {
    expect(detectMention("hey @")).toEqual({ token: "", atIndex: 4 })
  })

  it("returns null when @ is glued to a preceding word (email-like)", () => {
    expect(detectMention("mail@ali")).toBeNull()
  })

  it("returns null when there is no @ before the caret", () => {
    expect(detectMention("just text")).toBeNull()
  })

  it("returns null when a space follows the token (mention already finished)", () => {
    expect(detectMention("@Ali ")).toBeNull()
  })
})

describe("insertMention", () => {
  it("replaces the in-progress token span with `@<Name> `", () => {
    const value = "hi @al"
    const query = detectMention(value)!
    const out = insertMention(value, query, value.length, "Ali Alizadeh")
    expect(out.value).toBe("hi @Ali Alizadeh ")
    expect(out.caret).toBe(out.value.length)
  })

  it("keeps trailing text after the caret intact", () => {
    const value = "@al rest"
    // Caret sits right after "@al" (index 3), before " rest".
    const query = detectMention(value.slice(0, 3))!
    const out = insertMention(value, query, 3, "Ali")
    expect(out.value).toBe("@Ali  rest")
    expect(out.caret).toBe("@Ali ".length)
  })
})

describe("resolveMentions", () => {
  it("maps inserted display names to their ids", () => {
    const ids = resolveMentions("hey @Ali Alizadeh and @Sara", members)
    expect(ids.sort()).toEqual(["u-ali-a", "u-sara"].sort())
  })

  it("excludes an id whose text was later deleted", () => {
    // Only Sara survives in the final content.
    const ids = resolveMentions("thanks @Sara", members)
    expect(ids).toEqual(["u-sara"])
  })

  it("resolves longest-name-first so a prefix name does not steal the span", () => {
    // "@Ali Alizadeh" must resolve to the full name, NOT to the shorter "@Ali".
    const ids = resolveMentions("ping @Ali Alizadeh", members)
    expect(ids).toEqual(["u-ali-a"])
  })

  it("still resolves the short prefix name on its own", () => {
    const ids = resolveMentions("ping @Ali here", members)
    expect(ids).toEqual(["u-ali"])
  })

  it("resolves both when both distinct names appear", () => {
    const ids = resolveMentions("@Ali Alizadeh and also @Ali", members)
    expect(ids.sort()).toEqual(["u-ali", "u-ali-a"].sort())
  })

  it("returns an empty array when there are no mentions", () => {
    expect(resolveMentions("plain message", members)).toEqual([])
  })

  it("returns each id at most once", () => {
    const ids = resolveMentions("@Sara @Sara", members)
    expect(ids).toEqual(["u-sara"])
  })
})
