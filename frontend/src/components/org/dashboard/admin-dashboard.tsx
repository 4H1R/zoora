import type {
  GetClasses200Data,
  GetQuizzes200Data,
  GetUsers200Data,
  GithubCom4H1RZooraInternalDomainClass as Class,
  GithubCom4H1RZooraInternalDomainQuiz as Quiz,
} from "@/api/model"

import { Link, useParams } from "@tanstack/react-router"
import {
  ActivityIcon,
  ArrowRightIcon,
  ClipboardListIcon,
  GraduationCapIcon,
  PlusIcon,
  UsersIcon,
} from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetClasses } from "@/api/classes/classes"
import { useGetOrganizationsId } from "@/api/organizations/organizations"
import { useGetQuizzes } from "@/api/quizzes/quizzes"
import { useGetUsers, useGetUsersMe } from "@/api/users/users"
import { StatCards, type StatItem } from "@/components/data-table/stat-cards"
import { Eyebrow } from "@/components/eyebrow"
import { useDashboardPermissions } from "@/components/org/dashboard/use-dashboard-permissions"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { formatSessionDate } from "@/lib/session-status"
import { cn } from "@/lib/utils"

type ActivityItem = {
  id: string
  kind: "class" | "quiz"
  label: string
  at?: string
}

const DAY_MS = 86_400_000

function startOfDay(d: Date) {
  const x = new Date(d)
  x.setHours(0, 0, 0, 0)
  return x
}

export function AdminDashboard() {
  const { t, i18n } = useTranslation()
  const { orgId } = useParams({ from: "/_auth/org/$orgId/dashboard" })
  const { canViewClasses, canViewQuizzes, canViewUsers, canCreateClass } = useDashboardPermissions()

  const { data: meData } = useGetUsersMe()
  const me = (meData?.status === 200 && meData.data.data) || undefined

  const { data: orgData } = useGetOrganizationsId(orgId)
  const org = (orgData?.status === 200 && orgData.data.data) || undefined

  const { data: classesData, isPending: classesLoading } = useGetClasses(
    { order_by: "created_at", order_dir: "desc" },
    { query: { enabled: canViewClasses } }
  )
  const { data: quizzesData, isPending: quizzesLoading } = useGetQuizzes(
    { order_by: "created_at", order_dir: "desc" },
    { query: { enabled: canViewQuizzes } }
  )
  const { data: usersData, isPending: usersLoading } = useGetUsers(
    { order_by: "created_at", order_dir: "desc" },
    { query: { enabled: canViewUsers } }
  )

  const classesResult = (classesData?.status === 200 && (classesData.data.data as GetClasses200Data)) || undefined
  const quizzesResult = (quizzesData?.status === 200 && (quizzesData.data.data as GetQuizzes200Data)) || undefined
  const usersResult = (usersData?.status === 200 && (usersData.data.data as GetUsers200Data)) || undefined

  const classes = (classesResult?.items ?? []) as Class[]
  const quizzes = (quizzesResult?.items ?? []) as Quiz[]

  const firstName = (me?.name ?? "").trim().split(/\s+/)[0] || me?.username || ""

  const stats: StatItem[] = []
  if (canViewClasses) {
    stats.push({
      icon: <GraduationCapIcon />,
      label: t("org.dashboard.stats.classes"),
      value: classesResult?.total,
      loading: classesLoading,
    })
  }
  if (canViewUsers) {
    stats.push({
      icon: <UsersIcon />,
      label: t("org.dashboard.stats.members"),
      value: usersResult?.total,
      loading: usersLoading,
    })
  }
  if (canViewQuizzes) {
    stats.push({
      icon: <ClipboardListIcon />,
      label: t("org.dashboard.stats.quizzes"),
      value: quizzesResult?.total,
      loading: quizzesLoading,
    })
  }

  // Merge recent classes + quizzes into a single activity feed (real data only).
  const activity: ActivityItem[] = [
    ...(canViewClasses
      ? classes.map((c) => ({ id: c.id!, kind: "class" as const, label: c.name ?? "—", at: c.created_at }))
      : []),
    ...(canViewQuizzes
      ? quizzes.map((q) => ({ id: q.id!, kind: "quiz" as const, label: q.title ?? "—", at: q.created_at }))
      : []),
  ]
    .filter((a) => a.at)
    .sort((a, b) => new Date(b.at!).getTime() - new Date(a.at!).getTime())
    .slice(0, 6)

  // Weekly activity: count created items per day over the trailing 7 days.
  const today = startOfDay(new Date())
  const weekBuckets = Array.from({ length: 7 }, (_, i) => {
    const day = new Date(today.getTime() - (6 - i) * DAY_MS)
    return { day, count: 0 }
  })
  const allRecent: ActivityItem[] = [
    ...(canViewClasses ? classes.map((c) => ({ id: c.id!, kind: "class" as const, label: c.name ?? "", at: c.created_at })) : []),
    ...(canViewQuizzes ? quizzes.map((q) => ({ id: q.id!, kind: "quiz" as const, label: q.title ?? "", at: q.created_at })) : []),
  ]
  for (const a of allRecent) {
    if (!a.at) continue
    const ts = startOfDay(new Date(a.at)).getTime()
    const idx = weekBuckets.findIndex((b) => b.day.getTime() === ts)
    if (idx >= 0) weekBuckets[idx].count += 1
  }
  const weekMax = Math.max(1, ...weekBuckets.map((b) => b.count))
  const weekTotal = weekBuckets.reduce((sum, b) => sum + b.count, 0)

  const showActivityPanels = canViewClasses || canViewQuizzes
  const recentClassesLoading = classesLoading
  const recentClasses = classes.slice(0, 5)

  return (
    <div className="relative isolate flex flex-col gap-6">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-x-0 -top-6 -z-10 h-48 bg-[radial-gradient(ellipse_at_top,var(--color-primary)/8%,transparent_60%)]"
      />

      {/* Hero */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="flex flex-col gap-1.5">
          <Eyebrow className="text-primary">{t("org.dashboard.overview")}</Eyebrow>
          <h1 className="text-2xl font-bold tracking-tight sm:text-3xl">
            {firstName ? t("org.dashboard.welcomeName", { name: firstName }) : t("org.dashboard.welcome")}
          </h1>
          {org?.name ? <p className="text-muted-foreground text-sm">{org.name}</p> : null}
        </div>

        {canCreateClass ? (
          <div className="flex shrink-0 items-center gap-2">
            <Link to="/org/$orgId/classes" params={{ orgId }}>
              <Button size="sm">
                <PlusIcon data-icon="inline-start" />
                {t("org.dashboard.actions.newClass")}
              </Button>
            </Link>
          </div>
        ) : null}
      </div>

      {stats.length > 0 ? <StatCards stats={stats} className="lg:grid-cols-3" /> : null}

      {showActivityPanels ? (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-[1.5fr_1fr]">
          {/* Recent classes */}
          {canViewClasses ? (
            <Card className="gap-0 overflow-hidden p-0">
              <div className="flex items-center justify-between border-b px-4 py-3">
                <Eyebrow>{t("org.dashboard.recentClasses.title")}</Eyebrow>
                <Link
                  to="/org/$orgId/classes"
                  params={{ orgId }}
                  className="text-primary inline-flex items-center gap-1 text-xs font-medium transition-opacity hover:opacity-80"
                >
                  {t("org.dashboard.viewAll")}
                  <ArrowRightIcon className="size-3" />
                </Link>
              </div>
              {recentClassesLoading ? (
                <div className="divide-y">
                  {Array.from({ length: 4 }).map((_, i) => (
                    <div key={i} className="flex items-center gap-3 px-4 py-3">
                      <Skeleton className="size-9 rounded-lg" />
                      <div className="flex flex-1 flex-col gap-1.5">
                        <Skeleton className="h-4 w-40" />
                        <Skeleton className="h-3 w-24" />
                      </div>
                    </div>
                  ))}
                </div>
              ) : recentClasses.length === 0 ? (
                <p className="text-muted-foreground px-4 py-8 text-center text-sm">
                  {t("org.dashboard.recentClasses.empty")}
                </p>
              ) : (
                <div className="divide-y">
                  {recentClasses.map((c) => (
                    <Link
                      key={c.id}
                      to="/org/$orgId/classes/$classId"
                      params={{ orgId, classId: c.id! }}
                      className="hover:bg-muted/50 group flex items-center gap-3 px-4 py-3 transition-colors"
                    >
                      <div className="bg-primary/10 text-primary flex size-9 shrink-0 items-center justify-center rounded-lg">
                        <GraduationCapIcon className="size-4" />
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-medium">{c.name || "—"}</p>
                        <p className="text-muted-foreground truncate text-xs">
                          {c.user?.name || t("org.dashboard.recentClasses.unassigned")}
                          {typeof c.total_users === "number"
                            ? ` · ${t("org.dashboard.recentClasses.members", { count: c.total_users })}`
                            : ""}
                        </p>
                      </div>
                      <ArrowRightIcon className="text-muted-foreground group-hover:text-foreground size-4 shrink-0 transition-colors" />
                    </Link>
                  ))}
                </div>
              )}
            </Card>
          ) : null}

          {/* Recent activity */}
          <Card className="gap-0 overflow-hidden p-0">
            <div className="flex items-center justify-between border-b px-4 py-3">
              <Eyebrow>{t("org.dashboard.activity.title")}</Eyebrow>
              <ActivityIcon className="text-muted-foreground size-4" />
            </div>
            <div className="p-4">
              {classesLoading || quizzesLoading ? (
                <div className="flex flex-col gap-4">
                  {Array.from({ length: 3 }).map((_, i) => (
                    <div key={i} className="flex gap-3">
                      <Skeleton className="size-4 rounded-full" />
                      <div className="flex flex-1 flex-col gap-1.5">
                        <Skeleton className="h-3.5 w-44" />
                        <Skeleton className="h-3 w-16" />
                      </div>
                    </div>
                  ))}
                </div>
              ) : activity.length === 0 ? (
                <p className="text-muted-foreground py-6 text-center text-sm">
                  {t("org.dashboard.activity.empty")}
                </p>
              ) : (
                <ul className="flex flex-col gap-3.5">
                  {activity.map((a) => (
                    <li key={`${a.kind}-${a.id}`} className="flex gap-3">
                      <span
                        className={cn(
                          "mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-md [&>svg]:size-3.5",
                          a.kind === "class"
                            ? "bg-primary/10 text-primary"
                            : "bg-muted text-muted-foreground"
                        )}
                      >
                        {a.kind === "class" ? <GraduationCapIcon /> : <ClipboardListIcon />}
                      </span>
                      <div className="min-w-0 flex-1 text-sm leading-snug">
                        <p className="truncate">
                          {t(a.kind === "class" ? "org.dashboard.activity.classCreated" : "org.dashboard.activity.quizCreated")}{" "}
                          <span className="font-medium">{a.label}</span>
                        </p>
                        <p className="text-muted-foreground mt-0.5 text-xs">
                          {formatSessionDate(a.at, i18n.language, "short")}
                        </p>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </Card>
        </div>
      ) : null}

      {/* Weekly activity chart */}
      {showActivityPanels ? (
        <Card className="gap-0 overflow-hidden p-0">
          <div className="flex items-center justify-between border-b px-4 py-3">
            <Eyebrow>{t("org.dashboard.weekly.title")}</Eyebrow>
            <span className="text-muted-foreground text-xs">
              {t("org.dashboard.weekly.total", { count: weekTotal })}
            </span>
          </div>
          <div className="flex items-end gap-2 px-4 pt-6 pb-4" style={{ height: 140 }}>
            {weekBuckets.map((b, i) => {
              const ratio = b.count / weekMax
              const heightPct = b.count === 0 ? 6 : Math.max(12, Math.round(ratio * 100))
              const isToday = i === weekBuckets.length - 1
              const weekday = b.day.toLocaleDateString(i18n.language, { weekday: "short" })
              return (
                <div key={i} className="flex flex-1 flex-col items-center justify-end gap-2">
                  <span className="text-muted-foreground text-xs tabular-nums">{b.count > 0 ? b.count : ""}</span>
                  <div
                    className={cn(
                      "w-full max-w-7 rounded-t-md transition-all",
                      b.count === 0 ? "bg-muted" : isToday ? "bg-primary" : "bg-primary/40"
                    )}
                    style={{ height: `${heightPct}%` }}
                  />
                  <span className={cn("text-xs", isToday ? "text-foreground font-medium" : "text-muted-foreground")}>
                    {weekday}
                  </span>
                </div>
              )
            })}
          </div>
        </Card>
      ) : null}

      {/* Bare member fallback */}
      {stats.length === 0 && !showActivityPanels ? (
        <Card className="flex flex-col items-center gap-2 px-6 py-12 text-center">
          <div className="bg-primary/10 text-primary mb-1 flex size-12 items-center justify-center rounded-xl">
            <GraduationCapIcon className="size-6" />
          </div>
          <p className="text-sm font-medium">{t("org.dashboard.memberEmpty.title")}</p>
          <p className="text-muted-foreground max-w-sm text-sm">{t("org.dashboard.memberEmpty.hint")}</p>
        </Card>
      ) : null}
    </div>
  )
}
