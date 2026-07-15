import { decode, encode } from "blurhash"

import { isImage } from "./compress"

// BlurHash component counts (x, y). 4x3 is the conventional default for
// landscape-ish thumbnails: enough detail to read the image, cheap to encode.
export function blurhashComponents(): [number, number] {
  return [4, 3]
}

// Downscale target for encoding — a tiny canvas keeps encode() fast and is all
// BlurHash needs.
const ENCODE_MAX_EDGE = 32

/**
 * Load a File into an HTMLImageElement via an object URL. Rejects on decode
 * failure. Canvas/Image are only partially implemented under jsdom, so callers
 * must treat this as best-effort and guard failures.
 */
function loadImageElement(file: File): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const url = URL.createObjectURL(file)
    const img = new Image()
    img.onload = () => {
      URL.revokeObjectURL(url)
      resolve(img)
    }
    img.onerror = () => {
      URL.revokeObjectURL(url)
      reject(new Error("image decode failed"))
    }
    img.src = url
  })
}

/**
 * Encode a BlurHash placeholder string for an image File. Returns null for
 * non-images (incl. GIF/SVG) and on any failure — the canvas path is defensive
 * because jsdom cannot render pixels, so this is never unit-tested end to end.
 */
export async function encodeBlurhash(file: File): Promise<string | null> {
  if (!isImage(file)) return null
  try {
    const img = await loadImageElement(file)
    const natW = img.naturalWidth || img.width
    const natH = img.naturalHeight || img.height
    if (!natW || !natH) return null

    const scale = Math.min(ENCODE_MAX_EDGE / natW, ENCODE_MAX_EDGE / natH, 1)
    const w = Math.max(1, Math.round(natW * scale))
    const h = Math.max(1, Math.round(natH * scale))

    const canvas = document.createElement("canvas")
    canvas.width = w
    canvas.height = h
    const ctx = canvas.getContext("2d")
    if (!ctx) return null
    ctx.drawImage(img, 0, 0, w, h)
    const { data } = ctx.getImageData(0, 0, w, h)

    const [cx, cy] = blurhashComponents()
    return encode(data, w, h, cx, cy)
  } catch {
    return null
  }
}

/**
 * Decode a BlurHash string into a tiny PNG data URL usable as an <img>/CSS
 * background placeholder. Returns null on any failure (bad hash, no canvas) so
 * callers can fall back to a plain tinted box. The decoded grid is deliberately
 * small — it's scaled up (blurred) by the browser, so 32px is plenty.
 */
export function blurhashToDataUrl(hash: string, width = 32, height = 32): string | null {
  if (!hash) return null
  try {
    const pixels = decode(hash, width, height)
    const canvas = document.createElement("canvas")
    canvas.width = width
    canvas.height = height
    const ctx = canvas.getContext("2d")
    if (!ctx) return null
    const imageData = ctx.createImageData(width, height)
    imageData.data.set(pixels)
    ctx.putImageData(imageData, 0, 0)
    return canvas.toDataURL()
  } catch {
    return null
  }
}

/**
 * Natural pixel dimensions of an image File, or null for non-images / failure.
 */
export async function imageDimensions(file: File): Promise<{ width: number; height: number } | null> {
  if (!isImage(file)) return null
  try {
    const img = await loadImageElement(file)
    const width = img.naturalWidth || img.width
    const height = img.naturalHeight || img.height
    if (!width || !height) return null
    return { width, height }
  } catch {
    return null
  }
}
