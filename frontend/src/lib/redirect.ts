// Guards the post-login return URL against open-redirect abuse: only a
// root-relative path is allowed. Protocol-relative ("//evil.com") and absolute
// URLs ("https://evil.com") are rejected so the value can only point back into
// this app.
export function safeRedirectPath(value: unknown): string | undefined {
  if (typeof value !== "string") return undefined
  if (!value.startsWith("/") || value.startsWith("//")) return undefined
  // Bouncing back to /login would loop; treat it as no redirect.
  if (value === "/login" || value.startsWith("/login?")) return undefined
  return value
}
