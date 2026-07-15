import type { Crumb } from "@/components/layout/breadcrumb-context"

import { Link } from "@tanstack/react-router"
import { Fragment } from "react"

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Skeleton } from "@/components/ui/skeleton"

// Renders the `Org` (prefix) crumb followed by each pushed crumb. Shared by the
// desktop navbar and the mobile bar so both render identically from one trail.
// `Link.to` is a typed route union, not a plain string, so the generic crumb
// path is cast with `as never` (assignable to any target) at this single site.
export function BreadcrumbTrailView({
  prefixLabel,
  crumbs,
  className,
}: {
  prefixLabel: string
  crumbs: Crumb[]
  className?: string
}) {
  return (
    <Breadcrumb className={className}>
      <BreadcrumbList>
        <BreadcrumbItem>
          <span className="text-muted-foreground text-sm">{prefixLabel}</span>
        </BreadcrumbItem>
        {crumbs.map((crumb, index) => (
          <Fragment key={index}>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              {crumb.loading || crumb.label == null ? (
                <Skeleton className="h-4 w-24" />
              ) : crumb.to ? (
                <BreadcrumbLink
                  className="max-w-[22ch] truncate"
                  render={<Link to={crumb.to as never} params={crumb.params as never} />}
                >
                  {crumb.label}
                </BreadcrumbLink>
              ) : (
                <BreadcrumbPage className="max-w-[22ch] truncate">{crumb.label}</BreadcrumbPage>
              )}
            </BreadcrumbItem>
          </Fragment>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  )
}
