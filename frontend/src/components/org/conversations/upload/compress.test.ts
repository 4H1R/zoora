import { describe, expect, it } from "vitest"

import { COMPRESS_THRESHOLD_BYTES, isImage, shouldCompress } from "./compress"

// Fake File-like objects so we can exercise the pure predicates without a real
// File/Blob (jsdom's File is fine too, but this keeps the tests focused on the
// mime/size branches only).
const fake = (type: string, size = 0) => ({ type, size })

describe("isImage", () => {
  it("accepts common raster image mime types", () => {
    expect(isImage(fake("image/png"))).toBe(true)
    expect(isImage(fake("image/jpeg"))).toBe(true)
    expect(isImage(fake("image/webp"))).toBe(true)
    expect(isImage(fake("IMAGE/PNG"))).toBe(true) // case-insensitive
  })

  it("rejects GIF and SVG (must not be recompressed)", () => {
    expect(isImage(fake("image/gif"))).toBe(false)
    expect(isImage(fake("image/svg+xml"))).toBe(false)
  })

  it("rejects non-image types and empty/unknown types", () => {
    expect(isImage(fake("application/pdf"))).toBe(false)
    expect(isImage(fake("video/mp4"))).toBe(false)
    expect(isImage(fake(""))).toBe(false)
  })
})

describe("shouldCompress", () => {
  it("compresses images strictly larger than the threshold", () => {
    expect(shouldCompress(fake("image/png", COMPRESS_THRESHOLD_BYTES + 1))).toBe(true)
    expect(shouldCompress(fake("image/jpeg", 5 * 1024 * 1024))).toBe(true)
  })

  it("skips images at or below the threshold", () => {
    expect(shouldCompress(fake("image/png", COMPRESS_THRESHOLD_BYTES))).toBe(false)
    expect(shouldCompress(fake("image/png", 1024))).toBe(false)
  })

  it("skips GIF/SVG and non-images regardless of size", () => {
    expect(shouldCompress(fake("image/gif", 10 * 1024 * 1024))).toBe(false)
    expect(shouldCompress(fake("image/svg+xml", 10 * 1024 * 1024))).toBe(false)
    expect(shouldCompress(fake("application/pdf", 10 * 1024 * 1024))).toBe(false)
  })
})
