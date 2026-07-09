import type { MentionCandidate } from "./mentions"

import { describe, expect, it } from "vitest"

import { detectMention, highlightMentions, insertMention, resolveMentions } from "./mentions"

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

describe("highlightMentions", () => {
  it("returns a single plain segment when there are no mentions", () => {
    expect(highlightMentions("just plain text", members)).toEqual([{ text: "just plain text", isMention: false }])
  })

  it("returns an empty array for empty content", () => {
    expect(highlightMentions("", members)).toEqual([])
  })

  it("wraps an inserted name and preserves the plain text around it", () => {
    expect(highlightMentions("hey @Sara!", members)).toEqual([
      { text: "hey ", isMention: false },
      { text: "@Sara", isMention: true, userId: "u-sara" },
      { text: "!", isMention: false },
    ])
  })

  it("preserves leading/trailing whitespace between segments", () => {
    expect(highlightMentions("  @Sara  ", members)).toEqual([
      { text: "  ", isMention: false },
      { text: "@Sara", isMention: true, userId: "u-sara" },
      { text: "  ", isMention: false },
    ])
  })

  it("claims the longest name first so a prefix name does not steal the span", () => {
    // "@Ali Alizadeh" must highlight as the full name, NOT the shorter "@Ali".
    expect(highlightMentions("ping @Ali Alizadeh", members)).toEqual([
      { text: "ping ", isMention: false },
      { text: "@Ali Alizadeh", isMention: true, userId: "u-ali-a" },
    ])
  })

  it("still highlights the short prefix name on its own", () => {
    expect(highlightMentions("ping @Ali here", members)).toEqual([
      { text: "ping ", isMention: false },
      { text: "@Ali", isMention: true, userId: "u-ali" },
      { text: " here", isMention: false },
    ])
  })

  it("highlights both the long and the short name when both appear", () => {
    expect(highlightMentions("@Ali Alizadeh & @Ali", members)).toEqual([
      { text: "@Ali Alizadeh", isMention: true, userId: "u-ali-a" },
      { text: " & ", isMention: false },
      { text: "@Ali", isMention: true, userId: "u-ali" },
    ])
  })

  it("highlights every occurrence of the same name", () => {
    expect(highlightMentions("@Sara @Sara", members)).toEqual([
      { text: "@Sara", isMention: true, userId: "u-sara" },
      { text: " ", isMention: false },
      { text: "@Sara", isMention: true, userId: "u-sara" },
    ])
  })
})
