import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetClassesId } from "@/api/classes/classes"
import { Eyebrow } from "@/components/eyebrow"
import { OrgGradebookView } from "@/components/org/classes/OrgGradebookView"
import { useClassPermissions } from "@/components/org/classes/use-class-permissions"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/classes/$classId_/gradebook")({
  head: () => orgHead("org.class.gradebook.title"),
  component: RouteComponent,
})

function RouteComponent() {
  const { t } = useTranslation()
  const { classId } = Route.useParams()
  const { canView } = useClassPermissions()
  const allowed = useOrgGuard(["classes:view", "classes:view_any"])

  const { data: classData } = useGetClassesId(classId, { query: { enabled: canView } })
  const cls = (classData?.status === 200 && classData.data.data) || undefined

  if (!allowed) return null

  const shortId = (cls?.id ?? "").slice(0, 8).toUpperCase()

  return (
    <div className="relative isolate flex flex-col gap-8 pb-16">
      <div className="flex items-center justify-between pt-6">
        <Link
          to="/org/classes/$classId"
          params={{ classId }}
          className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
        >
          <ArrowLeftIcon className="size-3.5" />
          {t("org.class.gradebook.backToClass")}
        </Link>
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">
          № {shortId || "—"}
        </span>
      </div>

      <header className="flex flex-col gap-3">
        <Eyebrow>{cls?.name ?? t("org.class.gradebook.eyebrow")}</Eyebrow>
        <h1 className="max-w-4xl text-3xl leading-tight font-semibold tracking-tight text-balance md:text-4xl">
          {t("org.class.gradebook.title")}
        </h1>
        <p className="text-muted-foreground max-w-2xl text-sm leading-relaxed">
          {t("org.class.gradebook.subtitle")}
        </p>
      </header>

      <OrgGradebookView classId={classId} cls={cls} />
    </div>
  )
}
