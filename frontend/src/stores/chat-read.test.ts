import { beforeEach, describe, expect, it } from "vitest"

import { useChatRead } from "./chat-read"

beforeEach(() => {
  useChatRead.setState({ byConv: {} })
})

describe("chat-read store", () => {
  it("seeds read pointers for a conversation", () => {
    useChatRead.getState().seed("c1", { a: "m3", b: "m5" })
    expect(useChatRead.getState().byConv.c1).toEqual({ a: "m3", b: "m5" })
  })

  it("advances a pointer only when strictly newer", () => {
    const { seed, applyRead } = useChatRead.getState()
    seed("c1", { a: "m5" })
    applyRead("c1", "a", "m3") // older — ignored
    expect(useChatRead.getState().byConv.c1.a).toBe("m5")
    applyRead("c1", "a", "m9") // newer — applied
    expect(useChatRead.getState().byConv.c1.a).toBe("m9")
  })

  it("seed never regresses a newer live pointer", () => {
    const { seed, applyRead } = useChatRead.getState()
    applyRead("c1", "a", "m9")
    seed("c1", { a: "m4" }) // stale server snapshot — ignored
    expect(useChatRead.getState().byConv.c1.a).toBe("m9")
  })

  it("keeps a stable reference when nothing changes", () => {
    const { seed } = useChatRead.getState()
    seed("c1", { a: "m5" })
    const before = useChatRead.getState().byConv
    seed("c1", { a: "m5" }) // no-op
    expect(useChatRead.getState().byConv).toBe(before)
  })

  it("keeps conversations independent", () => {
    const { applyRead } = useChatRead.getState()
    applyRead("c1", "a", "m5")
    applyRead("c2", "a", "m9")
    expect(useChatRead.getState().byConv.c1.a).toBe("m5")
    expect(useChatRead.getState().byConv.c2.a).toBe("m9")
  })
})
