import type { PreJoinChoices } from "./types"
import type {
  GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse,
  GithubCom4H1RZooraInternalDomainLiveRoom,
} from "@/api/model"
import type { LucideIcon } from "lucide-react"
import type { ReactNode } from "react"

import { Link, useRouter } from "@tanstack/react-router"
import { CheckCircle2, Hourglass, Radio } from "lucide-react"
import { useTranslation } from "react-i18next"

import { usePostLiveRoomsIdJoin } from "@/api/live-sessions/live-sessions"
import { useGetUsersMe } from "@/api/users/users"
import GridBackground from "@/components/auth/gradient-background"
import { LanguageSwitcher } from "@/components/language-switcher"
import { Logo } from "@/components/logo"
import { SessionStatusPill } from "@/components/session/status-pill"
import { ThemeToggle } from "@/components/theme-toggle"
import { Button, buttonVariants } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import { UserAvatar } from "@/components/user-avatar"
import { formatCountdown, useNow } from "@/lib/session-status"
import { cn } from "@/lib/utils"

import { deriveRoomRole } from "./room-role"

const NO_MEDIA: PreJoinChoices = {
  audioEnabled: false,
  videoEnabled: false,
}

// Tone-tinted status boxes shown in place of the Join button (waiting / finished).
const NOTICE_TONES = {
  amber: {
    box: "border-amber-400/20 bg-amber-400/5",
    icon: "bg-amber-400/15 text-amber-500 dark:text-amber-300",
  },
  emerald: {
    box: "border-emerald-500/20 bg-emerald-500/5",
    icon: "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400",
  },
} as const

function LobbyNotice({
  tone,
  icon: Icon,
  title,
  children,
}: {
  tone: keyof typeof NOTICE_TONES
  icon: LucideIcon
  title: string
  children?: ReactNode
}) {
  const c = NOTICE_TONES[tone]
  return (
    <div className={cn("flex flex-col items-center gap-3 rounded-2xl border px-5 py-6", c.box)}>
      <span className={cn("flex size-11 items-center justify-center rounded-full", c.icon)}>
        <Icon className="size-5" />
      </span>
      <p className="text-foreground text-sm font-medium">{title}</p>
      {children}
    </div>
  )
}

interface PreJoinLobbyProps {
  room: GithubCom4H1RZooraInternalDomainLiveRoom | undefined
  liveId: string
  onJoined: (data: GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse, choices: PreJoinChoices) => void
}

export function PreJoinLobby({ room, liveId, onJoined }: PreJoinLobbyProps) {
  const { t } = useTranslation()
  const router = useRouter()
  const joinMutation = usePostLiveRoomsIdJoin()
  const { data: meData } = useGetUsersMe()
  const me = meData?.status === 200 ? meData.data.data : undefined
  const orgId = me?.organization_id
  const role = deriveRoomRole(me)
  const isHost = role === "host"
  const now = useNow(1000)

  const isFinished = room?.status === "finished"
  const isCreated = room?.status === "created"
  const session = room?.class_session
  const className = session?.class?.name
  const teacherName = session?.class?.user?.name
  const sessionName = session?.name
  const scheduledIso = room?.scheduled_start_time ?? session?.start_time
  const isWaiting = isCreated && !isHost

  const handleJoin = () => {
    joinMutation.mutate(
      { id: liveId },
      {
        onSuccess: (res) => {
          const joinData = (res.status === 200 && res.data.data) || undefined
          if (joinData?.token && joinData?.livekit_url) onJoined(joinData, NO_MEDIA)
        },
      }
    )
  }

  return (
    <div className="bg-muted/50 text-foreground relative flex min-h-svh flex-col">
      {/* Shared auth atmosphere: faint grid + brand halo, same as the login page */}
      <GridBackground />

      <header className="relative z-10 flex items-center justify-between px-5 py-4 sm:px-8">
        <Link to={orgId ? "/org" : "/"} className="flex items-center" aria-label={t("liveRoom.backToDashboard")}>
          <Logo className="text-xl" />
        </Link>
        <div className="flex items-center gap-2">
          <ThemeToggle />
          <div className="bg-border/50 mx-1 h-4 w-px" />
          <LanguageSwitcher />
        </div>
      </header>

      <main className="relative z-10 flex flex-1 items-center justify-center px-5 pb-16">
        <div className="border-border bg-card flex w-full max-w-sm flex-col items-center rounded-3xl border px-8 py-10 text-center shadow-sm">
          <SessionStatusPill status={room?.status === "active" ? "live" : isFinished ? "ended" : "scheduled"} />

          {className && (
            <div className="tracking-caps text-primary mt-6 text-[0.7rem] font-semibold uppercase">{className}</div>
          )}
          <h1 className="text-foreground mt-2 text-2xl leading-tight font-semibold tracking-tight text-balance">
            {sessionName ?? t("liveRoom.session")}
          </h1>

          {teacherName && (
            <div className="text-muted-foreground mt-5 inline-flex items-center gap-2 text-sm">
              <UserAvatar name={teacherName} size="sm" />
              <span>{teacherName}</span>
            </div>
          )}

          <div className="mt-8 w-full">
            {isFinished ? (
              <LobbyNotice tone="emerald" icon={CheckCircle2} title={t("liveRoom.finished.title")}>
                <p className="text-muted-foreground text-xs leading-relaxed">{t("liveRoom.finished.hint")}</p>
              </LobbyNotice>
            ) : isWaiting ? (
              <LobbyNotice tone="amber" icon={Hourglass} title={t("liveRoom.waitingForHost")}>
                {scheduledIso && now < new Date(scheduledIso).getTime() && (
                  <p
                    className="font-mono text-2xl font-semibold tracking-tight text-amber-600 tabular-nums dark:text-amber-200"
                    dir="ltr"
                  >
                    {formatCountdown(scheduledIso, now)}
                  </p>
                )}
                <p className="text-muted-foreground text-xs leading-relaxed">{t("liveRoom.waitingHint")}</p>
              </LobbyNotice>
            ) : (
              <Button
                size="lg"
                onClick={handleJoin}
                disabled={joinMutation.isPending}
                className="h-12 w-full gap-2 text-base font-semibold"
              >
                {joinMutation.isPending ? (
                  <Spinner className="size-4" />
                ) : isCreated ? (
                  <>
                    <Radio className="size-4" />
                    {t("liveRoom.startSession")}
                  </>
                ) : (
                  t("liveRoom.joinNow")
                )}
              </Button>
            )}

            {joinMutation.isError && !isWaiting && (
              <p className="text-destructive mt-4 text-sm">{t("liveRoom.joinError")}</p>
            )}

            {isFinished ? (
              <Link
                to={orgId ? "/org" : "/"}
                className={cn(buttonVariants({ size: "lg" }), "mt-2 h-12 w-full text-base font-semibold")}
              >
                {t("liveRoom.finished.backToDashboard")}
              </Link>
            ) : (
              <Button
                variant="ghost"
                onClick={() => router.history.back()}
                className="text-muted-foreground hover:bg-muted hover:text-foreground mt-2 w-full"
              >
                {t("liveRoom.back")}
              </Button>
            )}
          </div>
        </div>
      </main>
    </div>
  )
}
