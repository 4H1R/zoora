import { describe, expect, it } from "vitest"

import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

import { directPartnerId } from "./presence"

function conv(over: Partial<Conversation>): Conversation {
  return { id: "c1", type: "direct", ...over }
}

describe("directPartnerId", () => {
  it("returns the other member's user_id for a DM", () => {
    const c = conv({ members: [{ user_id: "me" }, { user_id: "partner" }] })
    expect(directPartnerId(c, "me")).toBe("partner")
  })

  it("falls back to the nested user.id when user_id is absent", () => {
    const c = conv({ members: [{ user: { id: "me" } }, { user: { id: "partner" } }] })
    expect(directPartnerId(c, "me")).toBe("partner")
  })

  it("returns undefined for non-direct conversations", () => {
    const c = conv({ type: "group", members: [{ user_id: "me" }, { user_id: "partner" }] })
    expect(directPartnerId(c, "me")).toBeUndefined()
  })

  it("returns undefined when members are not loaded", () => {
    expect(directPartnerId(conv({ members: undefined }), "me")).toBeUndefined()
  })
})
