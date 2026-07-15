import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

import { describe, expect, it } from "vitest"

import { conversationTitle } from "./conversation-title"

const SELF = "self-id"

describe("conversationTitle", () => {
  it("titles a DM after the OTHER member", () => {
    const conv: Conversation = {
      type: "direct",
      name: "",
      members: [
        { user_id: SELF, user: { id: SELF, name: "Me" } },
        { user_id: "partner", user: { id: "partner", name: "Sara" } },
      ],
    }
    expect(conversationTitle(conv, SELF)).toBe("Sara")
  })

  it("falls back to the stored name when the partner row is missing", () => {
    const conv: Conversation = { type: "direct", name: "", members: [] }
    expect(conversationTitle(conv, SELF)).toBe("")
  })

  it("uses the stored name for groups/channels", () => {
    const conv: Conversation = {
      type: "group",
      name: "Engineering",
      members: [{ user_id: SELF, user: { id: SELF, name: "Me" } }],
    }
    expect(conversationTitle(conv, SELF)).toBe("Engineering")
  })
})
