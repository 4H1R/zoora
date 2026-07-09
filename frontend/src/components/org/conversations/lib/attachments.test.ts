import type { ChatMessage, LocalAttachment } from "./messages"
import type { UploadResult } from "../upload/upload-manager"
import type { InfiniteData } from "@tanstack/react-query"

import { describe, expect, it } from "vitest"

import {
  allAttachmentsSucceeded,
  attachmentsOf,
  markAttachmentDone,
  markAttachmentError,
  removeAttachment,
  resetAttachmentUploading,
  resolvedMediaIds,
  updateAttachmentProgress,
} from "./attachments"

type Cache = InfiniteData<ChatMessage[]>

function att(localId: string, over: Partial<LocalAttachment> = {}): LocalAttachment {
  return {
    localId,
    name: `${localId}.png`,
    contentType: "image/png",
    size: 10,
    progress: 0,
    status: "uploading",
    ...over,
  }
}

function cacheWith(msgId: string, atts: LocalAttachment[]): Cache {
  const msg: ChatMessage = { id: msgId, content: "", _status: "sending", _attachments: atts }
  return { pages: [[msg]], pageParams: [{}] }
}

describe("updateAttachmentProgress", () => {
  it("updates only the targeted attachment and clamps to 0..1", () => {
    const cache = cacheWith("m1", [att("a"), att("b")])
    const next = updateAttachmentProgress(cache, "m1", "a", 1.7)
    const atts = next!.pages[0][0]._attachments!
    expect(atts[0].progress).toBe(1)
    expect(atts[1].progress).toBe(0)
  })

  it("returns the same reference when the message is not present", () => {
    const cache = cacheWith("m1", [att("a")])
    expect(updateAttachmentProgress(cache, "nope", "a", 0.5)).toBe(cache)
  })

  it("no-ops on undefined cache", () => {
    expect(updateAttachmentProgress(undefined, "m1", "a", 0.5)).toBeUndefined()
  })
})

describe("markAttachmentDone", () => {
  it("folds in the resolved media id, dims, blurhash and pins progress to 1", () => {
    const cache = cacheWith("m1", [att("a")])
    const result: UploadResult = {
      mediaId: "media-1",
      blurhash: "LKO2",
      width: 800,
      height: 600,
      name: "a.png",
      contentType: "image/png",
      size: 10,
    }
    const atts = markAttachmentDone(cache, "m1", "a", result)!.pages[0][0]._attachments!
    expect(atts[0]).toMatchObject({
      status: "done",
      progress: 1,
      mediaId: "media-1",
      blurhash: "LKO2",
      width: 800,
      height: 600,
    })
  })
})

describe("markAttachmentError / reset / remove", () => {
  it("flips one attachment to error", () => {
    const cache = cacheWith("m1", [att("a"), att("b")])
    const atts = markAttachmentError(cache, "m1", "b")!.pages[0][0]._attachments!
    expect(atts[1].status).toBe("error")
    expect(atts[0].status).toBe("uploading")
  })

  it("resets an errored attachment back to uploading with zero progress", () => {
    const cache = cacheWith("m1", [att("a", { status: "error", progress: 0.4 })])
    const atts = resetAttachmentUploading(cache, "m1", "a")!.pages[0][0]._attachments!
    expect(atts[0]).toMatchObject({ status: "uploading", progress: 0 })
  })

  it("removes a single attachment", () => {
    const cache = cacheWith("m1", [att("a"), att("b")])
    const atts = removeAttachment(cache, "m1", "a")!.pages[0][0]._attachments!
    expect(atts.map((a) => a.localId)).toEqual(["b"])
  })
})

describe("allAttachmentsSucceeded", () => {
  it("is false for an empty list", () => {
    expect(allAttachmentsSucceeded([])).toBe(false)
  })

  it("is false while any attachment is still uploading or errored", () => {
    expect(allAttachmentsSucceeded([att("a", { status: "done", mediaId: "x" }), att("b")])).toBe(false)
    expect(
      allAttachmentsSucceeded([
        att("a", { status: "done", mediaId: "x" }),
        att("b", { status: "error" }),
      ])
    ).toBe(false)
  })

  it("is false when a done attachment has no media id", () => {
    expect(allAttachmentsSucceeded([att("a", { status: "done" })])).toBe(false)
  })

  it("is true when every attachment is done with a media id", () => {
    expect(
      allAttachmentsSucceeded([
        att("a", { status: "done", mediaId: "x" }),
        att("b", { status: "done", mediaId: "y" }),
      ])
    ).toBe(true)
  })
})

describe("resolvedMediaIds", () => {
  it("collects done media ids in order, skipping unresolved ones", () => {
    const ids = resolvedMediaIds([
      att("a", { status: "done", mediaId: "x" }),
      att("b", { status: "error" }),
      att("c", { status: "done", mediaId: "z" }),
      att("d", { status: "uploading" }),
    ])
    expect(ids).toEqual(["x", "z"])
  })
})

describe("attachmentsOf", () => {
  it("reads the attachments off the cached message", () => {
    const cache = cacheWith("m1", [att("a")])
    expect(attachmentsOf(cache, "m1").map((a) => a.localId)).toEqual(["a"])
  })

  it("returns an empty array when the message or cache is missing", () => {
    expect(attachmentsOf(undefined, "m1")).toEqual([])
    expect(attachmentsOf(cacheWith("m1", [att("a")]), "other")).toEqual([])
  })
})
