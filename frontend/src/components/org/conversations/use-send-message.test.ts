import { describe, expect, it } from "vitest"

import type { SendMessageInput } from "./use-send-message"
import { resolveRetryInput } from "./use-send-message"

// A failed plain-text bubble never carries `mentions` (the optimistic
// ChatMessage model deliberately omits them), so a retry can only recover them
// from the stashed original send input.
const failedBubble = { content: "hey @alice", reply_to_message_id: "m-parent" }

describe("resolveRetryInput", () => {
  it("re-POSTs WITH the mentions from the stashed send input", () => {
    const stashed: SendMessageInput = {
      content: "hey @alice",
      replyToMessageId: "m-parent",
      mentions: ["u-alice"],
    }
    expect(resolveRetryInput(stashed, failedBubble)).toEqual({
      content: "hey @alice",
      replyToMessageId: "m-parent",
      mentions: ["u-alice"],
    })
  })

  it("falls back to the cached bubble (no mentions) when the stash is gone", () => {
    expect(resolveRetryInput(undefined, failedBubble)).toEqual({
      content: "hey @alice",
      replyToMessageId: "m-parent",
      mentions: undefined,
    })
  })

  it("recovers the reply target from the stash", () => {
    const stashed: SendMessageInput = { content: "hi", replyToMessageId: "m-99" }
    expect(resolveRetryInput(stashed, { content: "hi" }).replyToMessageId).toBe("m-99")
  })
})
