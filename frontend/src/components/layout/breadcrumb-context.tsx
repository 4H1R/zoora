import type { ReactNode } from "react"

import { createContext, useContext, useEffect, useState } from "react"

// A single breadcrumb crumb. `to` present → clickable link; `to` omitted →
// current-page (non-link) per the "omit to = current" convention. `label` null
// or `loading` true → render a skeleton (async entity name not yet resolved).
export type Crumb = {
  label: string | null
  to?: string
  params?: Record<string, string>
  loading?: boolean
}

type BreadcrumbContextValue = {
  crumbs: Crumb[] | null
  setCrumbs: (crumbs: Crumb[] | null) => void
}

const BreadcrumbContext = createContext<BreadcrumbContextValue | null>(null)

export function BreadcrumbProvider({ children }: { children: ReactNode }) {
  const [crumbs, setCrumbs] = useState<Crumb[] | null>(null)
  return <BreadcrumbContext.Provider value={{ crumbs, setCrumbs }}>{children}</BreadcrumbContext.Provider>
}

// Reader. Provider-optional on purpose: the navbar's SidebarBreadcrumb is shared
// with the admin layout (no provider there), so this must return null rather than
// throw when there is no provider — callers fall back to path-based labels.
export function useBreadcrumbTrail(): Crumb[] | null {
  return useContext(BreadcrumbContext)?.crumbs ?? null
}

// Setter hook for detail pages. Sets the trail whenever its content changes and
// clears it to null on unmount, so a trail never leaks onto the next page. The
// clear lives in its own effect (stable deps) so it runs ONLY on unmount — not on
// every content change — avoiding a fallback flicker while names stream in.
export function useBreadcrumb(crumbs: Crumb[]): void {
  const ctx = useContext(BreadcrumbContext)
  if (!ctx) throw new Error("useBreadcrumb must be used within BreadcrumbProvider")
  const { setCrumbs } = ctx
  const serialized = JSON.stringify(crumbs)

  useEffect(() => {
    setCrumbs(crumbs)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [serialized, setCrumbs])

  useEffect(() => {
    return () => setCrumbs(null)
  }, [setCrumbs])
}
