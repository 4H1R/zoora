import type { getConversationsIdMessagesResponse } from "@/api/conversations/conversations"

import { describe, expect, it } from "vitest"

import { unwrapMessagesPage } from "./use-messages"

// Fake the orval union response without dragging in the whole generated shape.
// The endpoint returns messages newest-first (DESC) under `data.data`.
function ok(descItems: unknown[]): getConversationsIdMessagesResponse {
  return {
    status: 200,
    data: { data: descItems },
    headers: new Headers(),
  } as unknown as getConversationsIdMessagesResponse
}

describe("unwrapMessagesPage", () => {
  it("reverses the DESC page into an ASCENDING array", () => {
    const res = ok([{ id: "c" }, { id: "b" }, { id: "a" }])
    expect(unwrapMessagesPage(res).map((m) => m.id)).toEqual(["a", "b", "c"])
  })

  it("returns [] when data is absent", () => {
    const res = { status: 200, data: {}, headers: new Headers() } as unknown as getConversationsIdMessagesResponse
    expect(unwrapMessagesPage(res)).toEqual([])
  })

  it("does not mutate the source array", () => {
    const src = [{ id: "b" }, { id: "a" }]
    unwrapMessagesPage(ok(src))
    expect(src.map((m: { id: string }) => m.id)).toEqual(["b", "a"])
  })

  it("throws on a non-200 status so React Query surfaces the error", () => {
    const res = { status: 403, data: {}, headers: new Headers() } as unknown as getConversationsIdMessagesResponse
    expect(() => unwrapMessagesPage(res)).toThrow()
  })
})
