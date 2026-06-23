import { useRouterState } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { useBreadcrumbTrail, type Crumb } from "@/components/layout/breadcrumb-context"
import { BreadcrumbTrailView } from "@/components/layout/breadcrumb-trail"

export function SidebarBreadcrumb({
  className,
  prefixLabel,
  pathPrefix,
  segmentKeys,
  defaultSegment = "dashboard",
}: {
  className?: string
  prefixLabel: string
  pathPrefix: RegExp
  segmentKeys: Record<string, string>
  defaultSegment?: string
}) {
  const { t } = useTranslation()
  const pathname = useRouterState({ select: (s) => s.location.pathname })
  const trail = useBreadcrumbTrail()

  // Path-based fallback: a single crumb for the top-level section (Org / Classes).
  const segments = pathname.replace(pathPrefix, "").split("/").filter(Boolean)
  const currentSegment = segments[0] ?? defaultSegment
  const currentKey = segmentKeys[currentSegment] ?? currentSegment
  const fallback: Crumb[] = [{ label: t(currentKey) }]

  return <BreadcrumbTrailView className={className} prefixLabel={prefixLabel} crumbs={trail ?? fallback} />
}
