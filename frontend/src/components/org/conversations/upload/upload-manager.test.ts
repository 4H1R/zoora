import { describe, expect, it } from "vitest"

import {
  buildPresignPayload,
  capFiles,
  MAX_MEDIA_PER_MESSAGE,
  MEDIA_COLLECTION_ATTACHMENTS,
  MEDIA_MODEL_CONVERSATION,
} from "./upload-manager"

describe("buildPresignPayload", () => {
  it("builds a conversation media presign body from a file", () => {
    const file = { name: "diagram.png", type: "image/png", size: 4096 }
    expect(buildPresignPayload(file, "conv-123")).toEqual({
      model_type: MEDIA_MODEL_CONVERSATION,
      model_id: "conv-123",
      collection_name: MEDIA_COLLECTION_ATTACHMENTS,
      file_name: "diagram.png",
      mime_type: "image/png",
      size: 4096,
    })
  })

  it("falls back to octet-stream when the file has no type", () => {
    const file = { name: "blob.bin", type: "", size: 10 }
    expect(buildPresignPayload(file, "c1").mime_type).toBe("application/octet-stream")
  })

  it("pins the backend model_type/collection constants", () => {
    expect(MEDIA_MODEL_CONVERSATION).toBe("conversation")
    expect(MEDIA_COLLECTION_ATTACHMENTS).toBe("attachments")
  })
})

describe("capFiles", () => {
  const many = (n: number) => Array.from({ length: n }, (_, i) => i)

  it("returns the list unchanged when under the cap", () => {
    expect(capFiles(many(5))).toEqual(many(5))
  })

  it("keeps exactly the cap when at the limit", () => {
    expect(capFiles(many(MAX_MEDIA_PER_MESSAGE))).toHaveLength(MAX_MEDIA_PER_MESSAGE)
  })

  it("drops files beyond the cap, keeping the first N", () => {
    const capped = capFiles(many(25))
    expect(capped).toHaveLength(MAX_MEDIA_PER_MESSAGE)
    expect(capped[0]).toBe(0)
    expect(capped.at(-1)).toBe(MAX_MEDIA_PER_MESSAGE - 1)
  })

  it("honors a custom max", () => {
    expect(capFiles(many(10), 3)).toEqual([0, 1, 2])
  })

  it("returns empty for a non-positive max", () => {
    expect(capFiles(many(10), 0)).toEqual([])
    expect(capFiles(many(10), -5)).toEqual([])
  })
})
