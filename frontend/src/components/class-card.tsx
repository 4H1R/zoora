import type { GithubCom4H1RZooraInternalDomainClass } from "@/api/model"

import { Link } from "@tanstack/react-router"
import { ChevronRight, UserIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { UserAvatar } from "@/components/user-avatar"
import { cn } from "@/lib/utils"

const coverGradients = [
  "from-emerald-700 to-teal-500",
  "from-violet-700 to-indigo-500",
  "from-rose-700 to-pink-500",
  "from-amber-700 to-orange-500",
  "from-sky-700 to-cyan-500",
  "from-fuchsia-700 to-purple-500",
  "from-red-800 to-rose-500",
  "from-indigo-800 to-blue-500",
] as const

export function getGradient(id: string) {
  let hash = 0
  for (let i = 0; i < id.length; i++) {
    hash = id.charCodeAt(i) + ((hash << 5) - hash)
  }
  return coverGradients[Math.abs(hash) % coverGradients.length]
}

interface ClassCardProps {
  cls: GithubCom4H1RZooraInternalDomainClass
}

export function ClassCard({ cls }: ClassCardProps) {
  const { t } = useTranslation()

  const gradient = getGradient(cls.id ?? "")
  const teacherName = cls.user?.name ?? ""

  return (
    <div className="group/card ring-foreground/10 hover:ring-foreground/30 bg-card flex flex-col overflow-hidden rounded-xl ring-1 transition-all">
      <Link to="/org/classes/$classId" params={{ classId: cls.id! }} className="block">
        <div className={cn("relative flex h-28 flex-col justify-end bg-gradient-to-br p-3.5", gradient)}>
          <p className="line-clamp-2 text-sm leading-snug font-semibold text-white drop-shadow-sm">{cls.name}</p>
        </div>
      </Link>

      <div className="flex flex-1 flex-col gap-3 p-3.5">
        {cls.user_id && (
          <div className="flex items-center gap-2">
            {teacherName ? (
              <UserAvatar name={teacherName} size="md" />
            ) : (
              <UserIcon className="text-muted-foreground size-4" />
            )}
            <span className="text-foreground text-xs font-medium">{teacherName}</span>
          </div>
        )}

        {cls.description && <p className="text-muted-foreground line-clamp-2 text-xs">{cls.description}</p>}

        <div className="mt-auto border-t pt-3">
          <div className="flex items-center justify-end">
            <Button
              variant="outline"
              size="xs"
              render={<Link to="/org/classes/$classId" params={{ classId: cls.id! }} />}
            >
              {t("common.continue")}
              <ChevronRight className="size-3.5 rtl:rotate-180" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

export function ClassCardSkeleton() {
  return (
    <div className="ring-border bg-card overflow-hidden rounded-xl ring-1">
      <Skeleton className="h-28 rounded-none" />
      <div className="flex flex-col gap-3 p-3.5">
        <div className="flex items-center gap-1.5">
          <Skeleton className="size-6 rounded-full" />
          <Skeleton className="h-3 w-20" />
        </div>
        <Skeleton className="h-3 w-3/4" />
        <div className="border-t pt-3">
          <div className="flex items-center justify-end">
            <Skeleton className="h-6 w-20 rounded-lg" />
          </div>
        </div>
      </div>
    </div>
  )
}
