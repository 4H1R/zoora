/**
 * Resolve the configured VITE_WS_URL into an absolute WebSocket URL.
 *
 * Zoora is multi-tenant (each org on its own subdomain) and the frontend is
 * served from the same origin as the API, so — like VITE_API_URL=/api/v1 — the
 * WS endpoint should be a host-relative path and the scheme/host derived at
 * runtime from window.location. Baking a fixed host (e.g. ws://localhost:8080)
 * into the build breaks every non-dev origin: the browser dials the baked host
 * instead of the current tenant's, and the token fails to validate there.
 *
 * An absolute ws:// or wss:// value is passed through unchanged so an explicit
 * override (or the test suite) still works.
 */
export function resolveWsUrl(configured: string): string {
  if (/^wss?:\/\//i.test(configured)) return configured

  const scheme = window.location.protocol === "https:" ? "wss:" : "ws:"
  const path = configured.startsWith("/") ? configured : `/${configured}`
  return `${scheme}//${window.location.host}${path}`
}
