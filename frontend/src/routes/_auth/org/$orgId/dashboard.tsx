import { createFileRoute, Link } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

export const Route = createFileRoute("/_auth/org/$orgId/dashboard")({
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
