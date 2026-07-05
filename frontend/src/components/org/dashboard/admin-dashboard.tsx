import { GraduationCapIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetUsersMe } from "@/api/users/users"
import { Eyebrow } from "@/components/eyebrow"
import { InstallAppBanner } from "@/components/pwa/install-app-banner"
import { TileGrid } from "@/components/org/dashboard/tile-grid"
import { useDashboardTiles } from "@/components/org/dashboard/use-dashboard-tiles"
import { useGreeting } from "@/components/org/dashboard/use-greeting"
import { EmptyState } from "@/components/ui/empty-state"

export function AdminDashboard() {
  const { t } = useTranslation()
  const tiles = useDashboardTiles()

  const { data: meData } = useGetUsersMe()
  const me = (meData?.status === 200 && meData.data.data) || undefined

  const firstName = (me?.name ?? "").trim().split(/\s+/)[0] || me?.username || ""
  const initial = firstName.charAt(0).toUpperCase()
  const greeting = useGreeting(firstName)

  return (
    <div className="relative isolate flex flex-col gap-6">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-x-0 -top-6 -z-10 h-48 bg-[radial-gradient(ellipse_at_top,var(--color-primary)/8%,transparent_60%)]"
      />

      <div className="flex items-center gap-3.5">
        {initial && (
          <div
            aria-hidden
            className="ring-primary/15 relative size-11 shrink-0 rounded-xl shadow-sm ring-1 ring-inset"
          >
            {/* base gradient + top sheen for depth */}
            <div className="from-primary/25 to-primary/5 absolute inset-0 rounded-xl bg-gradient-to-br" />
            <div className="absolute inset-0 rounded-xl bg-gradient-to-t from-transparent to-white/15 dark:to-white/5" />
            <span className="text-primary absolute inset-0 grid place-items-center text-base font-semibold tracking-tight">
              {initial}
            </span>
            <span className="bg-success ring-background absolute -end-0.5 -bottom-0.5 size-3 rounded-full ring-2" />
          </div>
        )}
        <div className="flex min-w-0 flex-col gap-0.5">
          <Eyebrow className="text-primary">{t("org.dashboard.overview")}</Eyebrow>
          <h1 className="truncate text-xl font-bold tracking-tight text-balance sm:text-2xl">
            {greeting}
          </h1>
        </div>
      </div>

      <InstallAppBanner />

      {/* Tiles gated by permission */}
      {tiles.length > 0 ? (
        <TileGrid tiles={tiles} />
      ) : (
        <EmptyState
          icon={GraduationCapIcon}
          title={t("org.dashboard.memberEmpty.title")}
          description={t("org.dashboard.memberEmpty.hint")}
        />
      )}
    </div>
  )
}
