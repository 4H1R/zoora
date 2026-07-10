import { describe, expect, it } from "vitest"

import { detectMention, insertMention, resolveMentions, splitMentions } from "./mentions"

const members = [
  { id: "u1", name: "Ali Alizadeh", username: "ali_r" },
  { id: "u2", name: "Sara K", username: "sara.k" },
]

describe("detectMention", () => {
  it("detects an in-progress username token", () => {
    expect(detectMention("hey @ali")).toEqual({ token: "ali", atIndex: 4 })
  })
  it("detects an empty token right after @", () => {
    expect(detectMention("hey @")).toEqual({ token: "", atIndex: 4 })
  })
  it("returns null when not in a token", () => {
    expect(detectMention("hey @ali ")).toBeNull()
  })
})

describe("insertMention", () => {
  it("inserts @username plus trailing space", () => {
    const q = { token: "al", atIndex: 4 }
    expect(insertMention("hey @al", q, 7, "ali_r")).toEqual({ value: "hey @ali_r ", caret: 11 })
  })
})

describe("resolveMentions", () => {
  it("maps @username tokens to member ids, unique + ordered", () => {
    expect(resolveMentions("@ali_r ping @sara.k and @ali_r", members)).toEqual(["u1", "u2"])
  })
  it("ignores non-member and sub-3-char tokens", () => {
    expect(resolveMentions("@ghost @ab", members)).toEqual([])
  })
})

describe("splitMentions", () => {
  it("splits text into plain + mention segments", () => {
    expect(splitMentions("hi @ali_r!")).toEqual([
      { text: "hi ", isMention: false },
      { text: "@ali_r", isMention: true, username: "ali_r" },
      { text: "!", isMention: false },
    ])
  })
  it("returns a single plain segment when there is no mention", () => {
    expect(splitMentions("no mentions here")).toEqual([{ text: "no mentions here", isMention: false }])
  })
})
