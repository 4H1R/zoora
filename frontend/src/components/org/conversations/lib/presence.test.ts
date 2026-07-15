import type { Presence } from "./presence"
import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

import { describe, expect, it } from "vitest"

import { directPartnerId, pickFreshestStatus } from "./presence"

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

describe("pickFreshestStatus", () => {
  const older: Presence = { online: false, lastSeen: "2024-01-01T00:00:00.000Z" }
  const newer: Presence = { online: true, lastSeen: "2024-01-02T00:00:00.000Z" }

  it("returns live when live is newer", () => {
    expect(pickFreshestStatus(newer, older)).toBe(newer)
  })

  it("returns snapshot when snapshot is newer (self-heals a stale live entry)", () => {
    expect(pickFreshestStatus(older, newer)).toBe(newer)
  })

  it("returns snapshot when only snapshot is present", () => {
    expect(pickFreshestStatus(undefined, newer)).toBe(newer)
  })

  it("returns live when only live is present", () => {
    expect(pickFreshestStatus(newer, undefined)).toBe(newer)
  })

  it("returns undefined when neither is present", () => {
    expect(pickFreshestStatus(undefined, undefined)).toBeUndefined()
  })

  it("prefers live on an exact tie", () => {
    const liveTie: Presence = { online: true, lastSeen: "2024-01-01T00:00:00.000Z" }
    const snapshotTie: Presence = { online: false, lastSeen: "2024-01-01T00:00:00.000Z" }
    expect(pickFreshestStatus(liveTie, snapshotTie)).toBe(liveTie)
  })

  it("prefers snapshot when live has no lastSeen but snapshot has a parseable one", () => {
    const liveNoTs: Presence = { online: true }
    expect(pickFreshestStatus(liveNoTs, newer)).toBe(newer)
  })

  it("prefers live when snapshot has no lastSeen but live has a parseable one", () => {
    const snapshotNoTs: Presence = { online: false }
    expect(pickFreshestStatus(newer, snapshotNoTs)).toBe(newer)
  })

  it("prefers snapshot when live's lastSeen is unparseable", () => {
    const liveBad: Presence = { online: true, lastSeen: "not-a-date" }
    expect(pickFreshestStatus(liveBad, newer)).toBe(newer)
  })

  it("prefers live when snapshot's lastSeen is unparseable", () => {
    const snapshotBad: Presence = { online: false, lastSeen: "not-a-date" }
    expect(pickFreshestStatus(newer, snapshotBad)).toBe(newer)
  })

  it("prefers live when neither lastSeen is parseable", () => {
    const liveNoTs: Presence = { online: true }
    const snapshotNoTs: Presence = { online: false }
    expect(pickFreshestStatus(liveNoTs, snapshotNoTs)).toBe(liveNoTs)
  })
})
