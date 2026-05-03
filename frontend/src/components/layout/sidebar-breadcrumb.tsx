import { useRouterState } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"

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

  const segments = pathname.replace(pathPrefix, "").split("/").filter(Boolean)
  const currentSegment = segments[0] ?? defaultSegment
  const currentKey = segmentKeys[currentSegment] ?? currentSegment

  return (
    <Breadcrumb className={className}>
      <BreadcrumbList>
        <BreadcrumbItem>
          <span className="text-muted-foreground text-sm">{prefixLabel}</span>
        </BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem>
          <BreadcrumbPage>{t(currentKey)}</BreadcrumbPage>
        </BreadcrumbItem>
      </BreadcrumbList>
    </Breadcrumb>
  )
}
