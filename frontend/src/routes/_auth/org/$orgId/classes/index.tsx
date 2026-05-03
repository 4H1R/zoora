import { createFileRoute } from "@tanstack/react-router"

import { useGetClasses } from "@/api/classes/classes"
import { ClassCard, ClassCardSkeleton } from "@/components/class-card"

export const Route = createFileRoute("/_auth/org/$orgId/classes/")({
  component: RouteComponent,
})

function RouteComponent() {
  const { orgId } = Route.useParams()
  const { data, isPending } = useGetClasses(undefined)

  const classesData = (data?.status === 200 && data.data.data) || undefined
  const classes = classesData?.items ?? []

  if (isPending) {
    return (
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }, (_, i) => (
          <ClassCardSkeleton key={i} />
        ))}
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {classes.map((cls) => (
        <ClassCard key={cls.id} cls={cls} orgId={orgId} />
      ))}
    </div>
  )
}
