import { useTranslation } from "react-i18next"

import { useBreadcrumbTrail } from "@/components/layout/breadcrumb-context"
import { BreadcrumbTrailView } from "@/components/layout/breadcrumb-trail"

// Mobile-only breadcrumb. Mounted once in the org layout. Renders the same pushed
// trail as the desktop navbar, but only when the page is "deep" — total visible
// crumbs (incl. the Org prefix) >= 3, i.e. the page pushed >= 2 crumbs. Plain
// section pages (no trail) stay hidden on mobile, matching today's behavior.
export function MobileBreadcrumb({ className }: { className?: string }) {
  const { t } = useTranslation()
  const trail = useBreadcrumbTrail()

  if (!trail || trail.length < 2) return null

  return (
    <div className={className}>
      <BreadcrumbTrailView prefixLabel={t("org.panel")} crumbs={trail} />
    </div>
  )
}
