// Aparat helpers. Tutorials store only the Aparat *video hash*; the embed URL
// and (author-time) thumbnail are derived from it. Aparat hashes are short
// alphanumeric codes, e.g. the "AbCdE" in aparat.com/v/AbCdE.

const HASH_RE = /^[A-Za-z0-9]+$/

/**
 * Pull the video hash out of whatever the admin pasted — a full watch URL, an
 * embed URL, or a bare hash. Returns "" when nothing usable is found.
 */
export function extractAparatHash(input: string): string {
  const raw = input.trim()
  if (!raw) return ""
  // Bare hash pasted directly.
  if (HASH_RE.test(raw)) return raw
  // /v/<hash> (watch page) or /embed/videohash/<hash> (embed URL).
  const m = raw.match(/\/v\/([A-Za-z0-9]+)/) ?? raw.match(/videohash\/([A-Za-z0-9]+)/)
  return m ? m[1] : ""
}

/** The iframe src for playing a tutorial. */
export function aparatEmbedUrl(hash: string): string {
  return `https://www.aparat.com/video/video/embed/videohash/${hash}/vt/frame`
}

/** The public watch URL (used for oEmbed lookups and "open on Aparat"). */
export function aparatWatchUrl(hash: string): string {
  return `https://www.aparat.com/v/${hash}`
}

// NOTE: oEmbed (title + thumbnail) is resolved server-side via
// getAdminTutorialsAparatOembed — Aparat's oEmbed endpoint sends no CORS
// headers, so a browser fetch is always blocked. Do not re-add a client fetch.
