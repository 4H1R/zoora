import { createFileRoute, Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/$orgId/dashboard")({
  head: () => orgHead("org.nav.dashboard"),
  component: RouteComponent,
})

function RouteComponent() {
  const { orgId } = Route.useParams()
  const { t } = useTranslation()

  return (
    <div>
      <Link to="/org/$orgId/classes" params={{ orgId }}>
        <button type="button">{t("nav.classes")}</button>
      </Link>
    </div>
  )
}
