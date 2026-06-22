import type {
  GetAdminClasses200Data,
  GetAdminLiveRooms200Data,
  GetAdminOrganizations200Data,
  GetAdminPolls200Data,
  GetAdminQuestionBanks200Data,
  GetAdminQuizzes200Data,
  GetAdminUsers200Data,
  GithubCom4H1RZooraInternalDomainOrganizationStats,
  GithubCom4H1RZooraInternalDomainOrganization as Organization,
  GithubCom4H1RZooraInternalDomainUser as User,
} from "@/api/model"

import { createFileRoute, Link } from "@tanstack/react-router"
import {
  ActivityIcon,
  BarChart3Icon,
  Building2Icon,
  ClipboardListIcon,
  GraduationCapIcon,
  LibraryIcon,
  UsersIcon,
  VideoIcon,
} from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetAdminClasses } from "@/api/admin-classes/admin-classes"
import { useGetAdminLiveRooms } from "@/api/admin-livesessions/admin-livesessions"
import { useGetAdminOrganizations, useGetAdminOrganizationsStats } from "@/api/admin-organizations/admin-organizations"
import { useGetAdminPolls } from "@/api/admin-polls/admin-polls"
import { useGetAdminQuestionBanks } from "@/api/admin-questionbanks/admin-questionbanks"
import { useGetAdminQuizzes } from "@/api/admin-quizzes/admin-quizzes"
import { useGetAdminUsers } from "@/api/admin-users/admin-users"
import { useGetUsersMe } from "@/api/users/users"
import { StatCards } from "@/components/data-table/stat-cards"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { adminHead } from "@/lib/admin-head"
import { useFormatDate } from "@/lib/format-date"
import { useRoleName } from "@/lib/permissions"

export const Route = createFileRoute("/_admin/admin/dashboard")({
  head: () => adminHead("admin.dashboard.title"),
  component: RouteComponent,
})

function RecentUsersCard({ users, loading }: { users: User[]; loading: boolean }) {
  const { t } = useTranslation()
  const roleName = useRoleName()
  const formatDate = useFormatDate()

  return (
    <Card className="gap-0 overflow-hidden p-0">
      <CardHeader className="flex flex-row items-center justify-between border-b px-4 py-3">
        <CardTitle className="text-sm font-medium">{t("admin.dashboard.recent.usersTitle")}</CardTitle>
        <Link to="/admin/users" className="text-muted-foreground hover:text-foreground text-xs transition-colors">
          {t("admin.dashboard.recent.viewAll")}
        </Link>
      </CardHeader>
      <CardContent className="p-0">
        {loading ? (
          <div className="divide-y">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex items-center justify-between px-4 py-3">
                <div className="flex flex-col gap-1">
                  <Skeleton className="h-4 w-32" />
                  <Skeleton className="h-3 w-24" />
                </div>
                <Skeleton className="h-5 w-14" />
              </div>
            ))}
          </div>
        ) : users.length === 0 ? (
          <p className="text-muted-foreground px-4 py-6 text-center text-sm">{t("admin.dashboard.recent.empty")}</p>
        ) : (
          <div className="divide-y">
            {users.map((user) => (
              <div key={user.id} className="flex items-center justify-between px-4 py-3">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{user.name || "—"}</p>
                  <p className="text-muted-foreground truncate text-xs">{user.username || "—"}</p>
                </div>
                <div className="ms-3 flex shrink-0 flex-col items-end gap-1">
                  <Badge variant="outline" className="text-xs">
                    {user.is_admin ? t("admin.roleAdmin") : (user.role?.name ? roleName(user.role.name) : t("admin.roleMember"))}
                  </Badge>
                  <span className="text-muted-foreground text-[11px]">{formatDate(user.created_at)}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function RecentOrgsCard({ orgs, loading }: { orgs: Organization[]; loading: boolean }) {
  const { t } = useTranslation()
  const formatDate = useFormatDate()

  const statusVariant = (status?: string): "default" | "secondary" | "destructive" | "outline" => {
    if (status === "active") return "default"
    if (status === "trial") return "secondary"
    if (status === "suspended") return "destructive"
    return "outline"
  }

  return (
    <Card className="gap-0 overflow-hidden p-0">
      <CardHeader className="flex flex-row items-center justify-between border-b px-4 py-3">
        <CardTitle className="text-sm font-medium">{t("admin.dashboard.recent.orgsTitle")}</CardTitle>
        <Link
          to="/admin/organizations"
          className="text-muted-foreground hover:text-foreground text-xs transition-colors"
        >
          {t("admin.dashboard.recent.viewAll")}
        </Link>
      </CardHeader>
      <CardContent className="p-0">
        {loading ? (
          <div className="divide-y">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex items-center justify-between px-4 py-3">
                <Skeleton className="h-4 w-36" />
                <div className="ms-3 flex shrink-0 flex-col items-end gap-1">
                  <Skeleton className="h-5 w-16" />
                  <Skeleton className="h-3 w-20" />
                </div>
              </div>
            ))}
          </div>
        ) : orgs.length === 0 ? (
          <p className="text-muted-foreground px-4 py-6 text-center text-sm">{t("admin.dashboard.recent.empty")}</p>
        ) : (
          <div className="divide-y">
            {orgs.map((org) => (
              <div key={org.id} className="flex items-center justify-between px-4 py-3">
                <p className="min-w-0 truncate text-sm font-medium">{org.name || "—"}</p>
                <div className="ms-3 flex shrink-0 flex-col items-end gap-1">
                  <Badge variant={statusVariant(org.status)} className="text-xs capitalize">
                    {org.status ? t(`admin.orgs.statusLabels.${org.status}`) : "—"}
                  </Badge>
                  <span className="text-muted-foreground text-[11px]">{formatDate(org.created_at)}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

function RouteComponent() {
  const { t } = useTranslation()

  const { data: statsData, isLoading: statsLoading } = useGetAdminOrganizationsStats()
  const { data: classesData, isLoading: classesLoading } = useGetAdminClasses()
  const { data: liveData, isLoading: liveLoading } = useGetAdminLiveRooms()
  const { data: quizzesData, isLoading: quizzesLoading } = useGetAdminQuizzes()
  const { data: pollsData, isLoading: pollsLoading } = useGetAdminPolls()
  const { data: qbData, isLoading: qbLoading } = useGetAdminQuestionBanks()

  const { data: recentUsersData, isLoading: usersLoading } = useGetAdminUsers({
    order_by: "created_at",
    order_dir: "desc",
    page_size: 5,
  })
  const { data: recentOrgsData, isLoading: orgsLoading } = useGetAdminOrganizations({
    order_by: "created_at",
    order_dir: "desc",
    page_size: 5,
  })

  const orgStats = statsData?.data?.data as GithubCom4H1RZooraInternalDomainOrganizationStats | undefined
  const classesTotal = (classesData?.data?.data as GetAdminClasses200Data | undefined)?.total
  const liveTotal = (liveData?.data?.data as GetAdminLiveRooms200Data | undefined)?.total
  const quizzesTotal = (quizzesData?.data?.data as GetAdminQuizzes200Data | undefined)?.total
  const pollsTotal = (pollsData?.data?.data as GetAdminPolls200Data | undefined)?.total
  const qbTotal = (qbData?.data?.data as GetAdminQuestionBanks200Data | undefined)?.total

  const recentUsers = ((recentUsersData?.data?.data as GetAdminUsers200Data | undefined)?.items ?? []) as User[]
  const recentOrgs = ((recentOrgsData?.data?.data as GetAdminOrganizations200Data | undefined)?.items ??
    []) as Organization[]

  const statCards = [
    {
      icon: <Building2Icon />,
      label: t("admin.dashboard.stats.organizations"),
      value: orgStats?.total_organizations,
      loading: statsLoading,
    },
    {
      icon: <ActivityIcon />,
      label: t("admin.dashboard.stats.activeOrgs"),
      value: orgStats?.active_count,
      loading: statsLoading,
    },
    {
      icon: <UsersIcon />,
      label: t("admin.dashboard.stats.users"),
      value: orgStats?.total_users,
      loading: statsLoading,
    },
    {
      icon: <GraduationCapIcon />,
      label: t("admin.dashboard.stats.classes"),
      value: classesTotal,
      loading: classesLoading,
    },
    {
      icon: <VideoIcon />,
      label: t("admin.dashboard.stats.liveSessions"),
      value: liveTotal,
      loading: liveLoading,
    },
    {
      icon: <ClipboardListIcon />,
      label: t("admin.dashboard.stats.quizzes"),
      value: quizzesTotal,
      loading: quizzesLoading,
    },
    {
      icon: <BarChart3Icon />,
      label: t("admin.dashboard.stats.polls"),
      value: pollsTotal,
      loading: pollsLoading,
    },
    {
      icon: <LibraryIcon />,
      label: t("admin.dashboard.stats.questionBanks"),
      value: qbTotal,
      loading: qbLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={t("admin.dashboard.welcome")} />
      <StatCards stats={statCards} className="lg:grid-cols-4" />
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <RecentUsersCard users={recentUsers} loading={usersLoading} />
        <RecentOrgsCard orgs={recentOrgs} loading={orgsLoading} />
      </div>
    </div>
  )
}
