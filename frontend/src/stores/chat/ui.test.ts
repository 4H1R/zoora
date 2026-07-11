import { beforeEach, describe, expect, it } from "vitest"

import { useChatUi } from "./ui"

describe("useChatUi", () => {
  beforeEach(() => {
    useChatUi.setState({ replyTo: null, editingMessageId: null, scrollToMessageId: null })
  })

  it("defaults to null", () => {
    const s = useChatUi.getState()
    expect(s.replyTo).toBeNull()
    expect(s.editingMessageId).toBeNull()
    expect(s.scrollToMessageId).toBeNull()
  })

  it("setReplyTo / setEditing / requestScrollTo update state", () => {
    useChatUi.getState().setReplyTo("m1")
    useChatUi.getState().setEditing("m2")
    useChatUi.getState().requestScrollTo("m3")
    const s = useChatUi.getState()
    expect(s.replyTo).toBe("m1")
    expect(s.editingMessageId).toBe("m2")
    expect(s.scrollToMessageId).toBe("m3")
  })
})
