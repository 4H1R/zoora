import type { GithubCom4H1RZooraInternalDomainClass } from "@/api/model"

import { Link } from "@tanstack/react-router"
import { Users } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { UserAvatar } from "@/components/user-avatar"
import { cn } from "@/lib/utils"

const coverGradients = [
  "from-emerald-600 to-teal-500",
  "from-violet-600 to-indigo-500",
  "from-rose-600 to-pink-500",
  "from-amber-600 to-orange-500",
  "from-sky-600 to-cyan-500",
  "from-fuchsia-600 to-purple-500",
] as const

function getGradient(id: string) {
  let hash = 0
  for (let i = 0; i < id.length; i++) {
    hash = id.charCodeAt(i) + ((hash << 5) - hash)
  }
  return coverGradients[Math.abs(hash) % coverGradients.length]
}

interface ClassCardProps {
  cls: GithubCom4H1RZooraInternalDomainClass
  orgId: string
  isLive?: boolean
}

export function ClassCard({ cls, orgId, isLive }: ClassCardProps) {
  const { t } = useTranslation()

  const gradient = getGradient(cls.id ?? "")
  const teacherName = cls.user?.name ?? ""
  const studentCount = cls.total_users ?? 0

  return (
    <div className="group/class-card ring-border bg-card overflow-hidden rounded-xl ring-1 transition-shadow hover:shadow-md">
      <Link to="/org/$orgId/classes/$classId" params={{ orgId, classId: cls.id! }} className="block">
        <div className={cn("relative flex h-28 flex-col justify-between bg-gradient-to-br p-3.5", gradient)}>
          <div className="flex items-start justify-between">
            <span className="rounded-md bg-white/20 px-1.5 py-0.5 text-[10px] font-semibold tracking-wide text-white/90 backdrop-blur-sm">
              {cls.id?.slice(0, 8).toUpperCase()}
            </span>
            {isLive && (
              <span className="inline-flex items-center gap-1 rounded-full bg-red-600 px-2 py-0.5 text-[10px] font-bold tracking-widest text-white">
                <span className="size-1.5 animate-pulse rounded-full bg-white" />
                {t("status.live")}
              </span>
            )}
          </div>
          <p className="line-clamp-2 text-sm leading-snug font-semibold text-white drop-shadow-sm">{cls.name}</p>
        </div>
      </Link>

      <div className="flex flex-col gap-2.5 p-3.5">
        <div className="text-muted-foreground flex items-center gap-2 text-xs">
          <div className="flex items-center gap-1">
            <Users className="size-3.5" />
            <span>
              {studentCount} {t("nav.students").toLowerCase()}
            </span>
          </div>
        </div>

        {cls.description && <p className="text-muted-foreground line-clamp-1 text-xs">{cls.description}</p>}

        <div className="border-border flex items-center justify-between border-t pt-2.5">
          {teacherName ? (
            <div className="flex items-center gap-1.5">
              <UserAvatar name={teacherName} size="sm" />
              <span className="text-muted-foreground text-xs">{teacherName}</span>
            </div>
          ) : (
            <div />
          )}
          <Button
            variant="outline"
            size="xs"
            render={<Link to="/org/$orgId/classes/$classId" params={{ orgId, classId: cls.id! }} />}
          >
            {t("common.manage")}
          </Button>
        </div>
      </div>
    </div>
  )
}

export function ClassCardSkeleton() {
  return (
    <div className="ring-border bg-card overflow-hidden rounded-xl ring-1">
      <Skeleton className="h-28 rounded-none" />
      <div className="flex flex-col gap-2.5 p-3.5">
        <div className="flex items-center gap-2">
          <Skeleton className="h-3.5 w-3.5 rounded-full" />
          <Skeleton className="h-3 w-20" />
        </div>
        <Skeleton className="h-3 w-3/4" />
        <div className="border-border flex items-center justify-between border-t pt-2.5">
          <div className="flex items-center gap-1.5">
            <Skeleton className="size-6 rounded-full" />
            <Skeleton className="h-3 w-16" />
          </div>
          <Skeleton className="h-6 w-16 rounded-lg" />
        </div>
      </div>
    </div>
  )
}
