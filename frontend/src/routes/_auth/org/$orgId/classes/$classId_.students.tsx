import type { GithubCom4H1RZooraInternalDomainClassMember as ClassMember } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon, CalendarClockIcon, PlusIcon, UserMinusIcon, UsersIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useAccess } from "react-access-engine"
import { toast } from "sonner"

import {
  getGetClassesIdMembersQueryKey,
  useDeleteClassesIdMembersUserId,
  useGetClassesId,
  useGetClassesIdMembers,
} from "@/api/classes/classes"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { EnrollMemberModal } from "@/components/org/classes/EnrollMemberModal"
import { useClassPermissions } from "@/components/org/classes/use-class-permissions"
import { Eyebrow } from "@/components/eyebrow"
import { Button } from "@/components/ui/button"
import { EmptyState } from "@/components/ui/empty-state"
import { Skeleton } from "@/components/ui/skeleton"
import { UserAvatar } from "@/components/user-avatar"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { formatSessionDate } from "@/lib/session-status"

export const Route = createFileRoute("/_auth/org/$orgId/classes/$classId_/students")({
  head: () => orgHead("org.class.students.title"),
  component: RouteComponent,
})

function StudentCard({
  member,
  index,
  onRemove,
}: {
  member: ClassMember
  index: number
  onRemove?: (member: ClassMember) => void
}) {
  const { t, i18n } = useTranslation()
  const name = member.user?.name ?? t("org.class.students.unknownName")
  const username = member.user?.username ?? ""
  const tileNumber = String(index + 1).padStart(2, "0")
  const joinedStr = member.created_at
    ? formatSessionDate(member.created_at, i18n.language, "short")
    : ""

  return (
    <div className="group/student bg-card text-card-foreground ring-foreground/10 hover:ring-foreground/30 relative isolate flex items-center gap-4 overflow-hidden rounded-2xl p-4 ring-1 transition-all hover:-translate-y-0.5 hover:shadow-lg">
      <UserAvatar name={name} size="lg" />
      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <span className="truncate text-sm font-semibold tracking-tight">{name}</span>
        {username ? (
          <span className="text-muted-foreground truncate font-mono text-xs">@{username}</span>
        ) : null}
        {joinedStr ? (
          <span className="text-muted-foreground mt-1 inline-flex items-center gap-1.5 text-xs">
            <CalendarClockIcon className="size-3" />
            {t("org.class.students.joinedAt", { date: joinedStr })}
          </span>
        ) : null}
      </div>
      {onRemove ? (
        <button
          type="button"
          onClick={() => onRemove(member)}
          aria-label={t("org.class.students.removeAction")}
          title={t("org.class.students.removeAction")}
          className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive focus-visible:ring-ring inline-flex size-8 shrink-0 items-center justify-center rounded-full opacity-0 transition-all focus-visible:opacity-100 focus-visible:ring-2 focus-visible:outline-none group-hover/student:opacity-100"
        >
          <UserMinusIcon className="size-4" />
        </button>
      ) : (
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">/{tileNumber}</span>
      )}
    </div>
  )
}

function StudentCardSkeleton() {
  return (
    <div className="bg-card ring-foreground/10 flex items-center gap-4 rounded-2xl p-4 ring-1">
      <Skeleton className="size-9 rounded-full" />
      <div className="flex flex-1 flex-col gap-2">
        <Skeleton className="h-4 w-32" />
        <Skeleton className="h-3 w-20" />
      </div>
    </div>
  )
}

function RouteComponent() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { orgId, classId } = Route.useParams()
  const { canView } = useClassPermissions()
  const allowed = useOrgGuard(["classes:view", "classes:view_any"])
  const { can, user: accessUser } = useAccess()

  const [enrollOpen, setEnrollOpen] = useState(false)
  const [removeTarget, setRemoveTarget] = useState<ClassMember | null>(null)

  const removeMutation = useDeleteClassesIdMembersUserId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.class.removeMember.success"))
        queryClient.invalidateQueries({ queryKey: getGetClassesIdMembersQueryKey(classId) })
        setRemoveTarget(null)
      },
      onError: (err) => {
        const status = (err as { status?: number })?.status
        if (status === 403) {
          toast.error(t("org.class.removeMember.errorForbidden"))
        } else {
          toast.error(t("org.class.removeMember.errorGeneric"))
        }
      },
    },
  })

  const handleConfirmRemove = () => {
    if (!removeTarget?.user_id) return
    removeMutation.mutate({ id: classId, userId: removeTarget.user_id })
  }

  const { data: classData } = useGetClassesId(classId, { query: { enabled: canView } })
  const cls = (classData?.status === 200 && classData.data.data) || undefined

  // Roster gating mirrors backend canManageClass:
  // admin OR classes:update_any OR caller is class owner.
  const canViewRoster =
    !!cls && (can("classes:update_any") || (!!cls.user_id && cls.user_id === accessUser.id))

  const { data: membersData, isPending: membersPending } = useGetClassesIdMembers(
    classId,
    undefined,
    { query: { enabled: canView && canViewRoster } }
  )
  const membersResult = (membersData?.status === 200 && membersData.data.data) || undefined
  const members = membersResult?.items ?? []
  const studentsTotal = membersResult?.total ?? members.length

  if (!allowed) return null

  const shortId = (cls?.id ?? "").slice(0, 8).toUpperCase()

  return (
    <div className="relative isolate flex flex-col gap-8 pb-16">
      <div className="flex items-center justify-between pt-6">
        <Link
          to="/org/$orgId/classes/$classId"
          params={{ orgId, classId }}
          className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
        >
          <ArrowLeftIcon className="size-3.5" />
          {t("org.class.students.backToClass")}
        </Link>
        <span className="text-muted-foreground font-mono text-xs tracking-[0.25em]">
          № {shortId || "—"}
        </span>
      </div>

      <header className="flex flex-col gap-3">
        <Eyebrow>{cls?.name ?? t("org.class.students.eyebrow")}</Eyebrow>
        <div className="flex items-end justify-between gap-4">
          <div className="flex flex-col gap-3">
            <h1 className="max-w-4xl text-3xl leading-tight font-semibold tracking-tight text-balance md:text-4xl">
              {t("org.class.students.title")}
              <span className="text-muted-foreground ms-2 font-mono text-base font-normal tabular-nums">
                {studentsTotal}
              </span>
            </h1>
            <p className="text-muted-foreground max-w-2xl text-sm leading-relaxed">
              {t("org.class.students.subtitle")}
            </p>
          </div>
          {canViewRoster ? (
            <Button variant="outline" onClick={() => setEnrollOpen(true)}>
              <PlusIcon className="size-4" />
              {t("org.class.students.addMember")}
            </Button>
          ) : null}
        </div>
      </header>

      {!canViewRoster ? (
        <EmptyState
          icon={UsersIcon}
          title={t("org.class.students.noAccessTitle")}
          description={t("org.class.students.noAccessHint")}
        />
      ) : membersPending ? (
        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          <StudentCardSkeleton />
          <StudentCardSkeleton />
          <StudentCardSkeleton />
        </div>
      ) : members.length === 0 ? (
        <EmptyState
          icon={UsersIcon}
          title={t("org.class.students.emptyTitle")}
          description={t("org.class.students.emptyHint")}
        >
          <Button onClick={() => setEnrollOpen(true)}>
            <PlusIcon className="size-4" />
            {t("org.class.students.addMember")}
          </Button>
        </EmptyState>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
          {members.map((m, i) => (
            <StudentCard key={m.id} member={m} index={i} onRemove={setRemoveTarget} />
          ))}
        </div>
      )}

      {canViewRoster ? (
        <>
          <EnrollMemberModal open={enrollOpen} onOpenChange={setEnrollOpen} classId={classId} />
          <DeleteConfirmDialog
            open={!!removeTarget}
            onOpenChange={(open) => !open && setRemoveTarget(null)}
            resourceName={removeTarget?.user?.name ?? t("org.class.students.unknownName")}
            onConfirm={handleConfirmRemove}
            isLoading={removeMutation.isPending}
          />
        </>
      ) : null}
    </div>
  )
}
