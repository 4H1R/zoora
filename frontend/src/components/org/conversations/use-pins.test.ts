import { describe, expect, it } from "vitest"

import type { getConversationsIdPinsResponse } from "@/api/conversations/conversations"

import { unwrapPins } from "./use-pins"

// Fake the orval union response without dragging in the whole generated shape.
// The endpoint returns pinned messages newest-pin-first (pinned_at DESC).
function ok(items: unknown[]): getConversationsIdPinsResponse {
  return {
    status: 200,
    data: { data: items },
    headers: new Headers(),
  } as unknown as getConversationsIdPinsResponse
}

describe("unwrapPins", () => {
  it("returns the pinned messages in their given (pinned_at DESC) order", () => {
    const res = ok([{ id: "c" }, { id: "b" }, { id: "a" }])
    expect(unwrapPins(res).map((m) => m.id)).toEqual(["c", "b", "a"])
  })

  it("returns [] when data is absent", () => {
    const res = { status: 200, data: {}, headers: new Headers() } as unknown as getConversationsIdPinsResponse
    expect(unwrapPins(res)).toEqual([])
  })

  it("throws on a non-200 status so React Query surfaces the error", () => {
    const res = { status: 403, data: {}, headers: new Headers() } as unknown as getConversationsIdPinsResponse
    expect(() => unwrapPins(res)).toThrow()
  })
})
