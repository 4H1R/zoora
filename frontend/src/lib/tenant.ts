// Tenant resolution from the request host. The org boundary is the subdomain
// (`<slug>.<base>`); `admin.<base>` is the platform-admin scope. The backend
// re-resolves and enforces this on every request — these helpers only drive
// client-side gating (which login screen to show, where to navigate).

const BASE_DOMAIN = import.meta.env.VITE_BASE_DOMAIN ?? "localhost"
const ADMIN_SUBDOMAIN = import.meta.env.VITE_ADMIN_SUBDOMAIN ?? "admin"

// currentSlug returns the left-most host label, or "" for the apex. The
// canonical `www` host mirrors the apex (it serves the landing page), so it
// also resolves to "".
export function currentSlug(base = BASE_DOMAIN): string {
  const host = window.location.hostname
  if (host === base) return ""
  const suffix = `.${base}`
  if (!host.endsWith(suffix)) return ""
  const slug = host.slice(0, -suffix.length).split(".")[0]
  return slug === "www" ? "" : slug
}

export function isAdminHost(): boolean {
  return currentSlug() === ADMIN_SUBDOMAIN
}
