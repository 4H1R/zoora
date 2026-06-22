import type { GithubCom4H1RZooraInternalDomainClassMember as ClassMember } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon, UserPlusIcon, UsersIcon, UserXIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetClassesIdMembersQueryKey,
  useDeleteClassesIdMembersUserId,
  useGetClassesId,
  useGetClassesIdMembers,
  usePostClassesIdMembers,
} from "@/api/classes/classes"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { UserSelect } from "@/components/form/user-select"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { EmptyState } from "@/components/ui/empty-state"
import { Spinner } from "@/components/ui/spinner"
import { getEntityColor, getInitials, useFormatDate } from "@/lib/data-table"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_admin/admin/classes/$classId/members")({
  component: ClassMembersPage,
})

function ClassMembersPage() {
  const { t } = useTranslation()
  const { classId } = Route.useParams()
  const queryClient = useQueryClient()
  const formatDate = useFormatDate()

  const [selectedUserId, setSelectedUserId] = useState<string>("")
  const [removeTarget, setRemoveTarget] = useState<ClassMember | null>(null)

  const { data: classData } = useGetClassesId(classId)
  const cls = (classData?.status === 200 && classData.data.data) || undefined

  const { data, isLoading } = useGetClassesIdMembers(classId, {})
  const membersData = (data?.status === 200 && data.data.data) || undefined
  const members = membersData?.items ?? []
  const total = membersData?.total ?? 0
  const capacity = cls?.total_users ?? 0
  const capacityReached = capacity > 0 && total >= capacity

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetClassesIdMembersQueryKey(classId) })
  }

  const enrollMutation = usePostClassesIdMembers({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.classMembers.form.addSuccess"))
        setSelectedUserId("")
        invalidate()
      },
    },
  })

  const removeMutation = useDeleteClassesIdMembersUserId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.classMembers.form.removeSuccess"))
        setRemoveTarget(null)
        invalidate()
      },
    },
  })

  const handleAdd = () => {
    if (!selectedUserId) return
    enrollMutation.mutate({ id: classId, data: { user_id: selectedUserId } })
  }

  const handleConfirmRemove = () => {
    if (removeTarget?.user_id) {
      removeMutation.mutate({ id: classId, userId: removeTarget.user_id })
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={cls?.name ? `${cls.name} · ${t("admin.classMembers.title")}` : t("admin.classMembers.title")}
        actions={
          <Link to="/admin/classes">
            <Button variant="outline" size="sm">
              <ArrowLeftIcon data-icon="inline-start" />
              {t("admin.classMembers.backToClasses")}
            </Button>
          </Link>
        }
      />

      <Card className="relative overflow-hidden border-dashed p-0">
        <div className="bg-muted/40 absolute inset-0 -z-10" aria-hidden />
        <div className="flex flex-col gap-4 p-5 sm:flex-row sm:items-end">
          <div className="flex-1">
            <div className="mb-1 flex items-center gap-2 text-xs font-semibold tracking-wide uppercase">
              <UserPlusIcon className="size-3.5" />
              {t("admin.classMembers.form.addTitle")}
            </div>
            <p className="text-muted-foreground mb-3 text-xs">
              {capacity === 0
                ? t("admin.classMembers.form.unlimitedHint")
                : t("admin.classMembers.form.capacityHint", { used: total, capacity })}
            </p>
            <UserSelect
              value={selectedUserId || undefined}
              onChange={setSelectedUserId}
              placeholder={t("admin.classMembers.form.userPlaceholder")}
              organizationId={cls?.organization_id || undefined}
            />
          </div>
          <Button
            onClick={handleAdd}
            disabled={!selectedUserId || enrollMutation.isPending || capacityReached}
            size="sm"
          >
            {enrollMutation.isPending && <Spinner />}
            <UserPlusIcon data-icon="inline-start" />
            {t("admin.classMembers.form.addButton")}
          </Button>
        </div>
      </Card>

      <div className="flex items-baseline justify-between px-1">
        <h2 className="text-sm font-semibold">
          {t("admin.classMembers.roster")}
          <span className="text-muted-foreground ms-2 tabular-nums">{total}</span>
        </h2>
        {capacity > 0 && (
          <span className="text-muted-foreground text-xs tabular-nums">
            {total} / {capacity}
          </span>
        )}
      </div>

      {isLoading ? (
        <Card className="text-muted-foreground flex items-center justify-center gap-2 p-10 text-sm">
          <Spinner />
          {t("common.loading")}
        </Card>
      ) : members.length === 0 ? (
        <EmptyState
          icon={UsersIcon}
          title={t("admin.classMembers.empty.title")}
          description={t("admin.classMembers.empty.hint")}
        />
      ) : (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {members.map((m) => (
            <MemberCard
              key={m.id}
              member={m}
              joined={formatDate(m.created_at)}
              onRemove={() => setRemoveTarget(m)}
            />
          ))}
        </div>
      )}

      <DeleteConfirmDialog
        open={!!removeTarget}
        onOpenChange={(open) => {
          if (!open) setRemoveTarget(null)
        }}
        resourceName={removeTarget?.user?.name ?? removeTarget?.user?.username ?? ""}
        onConfirm={handleConfirmRemove}
        isLoading={removeMutation.isPending}
      />
    </div>
  )
}

interface MemberCardProps {
  member: ClassMember
  joined: string
  onRemove: () => void
}

function MemberCard({ member, joined, onRemove }: MemberCardProps) {
  const { t } = useTranslation()
  const name = member.user?.name ?? "—"

  return (
    <Card className="group hover:border-foreground/20 relative flex flex-row items-center gap-3 overflow-hidden p-3 transition-colors">
      <div
        className={cn(
          "flex size-11 shrink-0 items-center justify-center rounded-xl text-sm font-semibold text-white",
          getEntityColor(name)
        )}
      >
        {getInitials(name)}
      </div>
      <div className="min-w-0 flex-1">
        <div className="truncate text-sm font-medium">{name}</div>
        {member.user?.username && (
          <div className="text-muted-foreground truncate font-mono text-xs">{member.user.username}</div>
        )}
        <div className="text-muted-foreground/70 mt-0.5 truncate text-[10px] tracking-wide uppercase">
          {t("admin.classMembers.joined", { date: joined })}
        </div>
      </div>
      <Button
        variant="ghost"
        size="icon-xs"
        onClick={onRemove}
        aria-label={t("admin.classMembers.actions.remove")}
        className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive opacity-100 transition-opacity sm:opacity-0 sm:group-hover:opacity-100"
      >
        <UserXIcon />
      </Button>
    </Card>
  )
}
