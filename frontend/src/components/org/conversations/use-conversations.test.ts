import type { getConversationsResponse } from "@/api/conversations/conversations"

import { describe, expect, it } from "vitest"

import { unwrapConversations } from "./use-conversations"

// Minimal helper to fake the orval union response without dragging in the whole
// generated shape.
function ok(items: unknown[]): getConversationsResponse {
  return {
    status: 200,
    data: { data: { items } },
    headers: new Headers(),
  } as unknown as getConversationsResponse
}

describe("unwrapConversations", () => {
  it("flattens the paginated items into a flat array", () => {
    const res = ok([{ id: "a" }, { id: "b" }])
    expect(unwrapConversations(res).map((c) => c.id)).toEqual(["a", "b"])
  })

  it("returns [] when items is absent", () => {
    const res = { status: 200, data: { data: {} }, headers: new Headers() } as unknown as getConversationsResponse
    expect(unwrapConversations(res)).toEqual([])
  })

  it("throws on a non-200 status so React Query surfaces the error", () => {
    const res = { status: 403, data: {}, headers: new Headers() } as unknown as getConversationsResponse
    expect(() => unwrapConversations(res)).toThrow()
  })
})
