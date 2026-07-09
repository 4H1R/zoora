import imageCompression from "browser-image-compression"

// Images at or below this size are already small enough that recompression
// would burn CPU for little gain — send them as-is.
export const COMPRESS_THRESHOLD_BYTES = 512 * 1024

// GIF (animation) and SVG (vector/markup) must never be rasterised through the
// image-compression pipeline — doing so would drop frames or destroy the
// vector, and neither benefits from JPEG-style compression. We treat them as
// non-images here so both the compressor and the blurhash encoder skip them.
const NON_RECOMPRESSIBLE_TYPES = new Set(["image/gif", "image/svg+xml"])

/**
 * True for raster image types we can safely recompress and blurhash.
 * Explicitly excludes GIF and SVG (see NON_RECOMPRESSIBLE_TYPES).
 */
export function isImage(file: Pick<File, "type">): boolean {
  const type = (file.type || "").toLowerCase()
  return type.startsWith("image/") && !NON_RECOMPRESSIBLE_TYPES.has(type)
}

/**
 * Only compressible images larger than the threshold are worth compressing.
 */
export function shouldCompress(file: Pick<File, "type" | "size">): boolean {
  return isImage(file) && file.size > COMPRESS_THRESHOLD_BYTES
}

/**
 * Compress an image via browser-image-compression. Returns the ORIGINAL file
 * unchanged when it shouldn't be compressed or when compression fails, so the
 * caller can always proceed to upload.
 */
export async function compressImage(file: File): Promise<File> {
  if (!shouldCompress(file)) return file
  try {
    return await imageCompression(file, {
      maxSizeMB: 1,
      maxWidthOrHeight: 1920,
      useWebWorker: true,
    })
  } catch {
    return file
  }
}
