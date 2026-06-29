import { Link, useRouter } from "@tanstack/react-router"
import { Hourglass, Radio } from "lucide-react"
import { useTranslation } from "react-i18next"

import type {
  GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse,
  GithubCom4H1RZooraInternalDomainLiveRoom,
} from "@/api/model"
import type { PreJoinChoices } from "./types"

import { usePostLiveRoomsIdJoin } from "@/api/live-sessions/live-sessions"
import { useGetUsersMe } from "@/api/users/users"
import GridBackground from "@/components/auth/gradient-background"
import { LanguageSwitcher } from "@/components/language-switcher"
import { Logo } from "@/components/logo"
import { SessionStatusPill } from "@/components/session/status-pill"
import { ThemeToggle } from "@/components/theme-toggle"
import { Button } from "@/components/ui/button"
import { Spinner } from "@/components/ui/spinner"
import { UserAvatar } from "@/components/user-avatar"
import { deriveRoomRole } from "./room-role"
import { formatCountdown, useNow } from "@/lib/session-status"

const NO_MEDIA: PreJoinChoices = {
  audioEnabled: false,
  videoEnabled: false,
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
    <div className="relative flex min-h-svh flex-col bg-muted/50 text-foreground">
      {/* Shared auth atmosphere: faint grid + brand halo, same as the login page */}
      <GridBackground />

      <header className="relative z-10 flex items-center justify-between px-5 py-4 sm:px-8">
        <Link
          to={orgId ? "/org" : "/"}
          className="flex items-center"
          aria-label={t("liveRoom.backToDashboard")}
        >
          <Logo className="text-xl" />
        </Link>
        <div className="flex items-center gap-2">
          <ThemeToggle />
          <div className="mx-1 h-4 w-px bg-border/50" />
          <LanguageSwitcher />
        </div>
      </header>

      <main className="relative z-10 flex flex-1 items-center justify-center px-5 pb-16">
        <div className="flex w-full max-w-sm flex-col items-center rounded-3xl border border-border bg-card px-8 py-10 text-center shadow-sm">
          <SessionStatusPill status={room?.status === "active" ? "live" : isFinished ? "ended" : "scheduled"} />

          {className && (
            <div className="mt-6 text-[0.7rem] font-semibold tracking-caps text-primary uppercase">{className}</div>
          )}
          <h1 className="mt-2 text-2xl leading-tight font-semibold tracking-tight text-balance text-foreground">
            {sessionName ?? t("liveRoom.session")}
          </h1>

          {teacherName && (
            <div className="mt-5 inline-flex items-center gap-2 text-sm text-muted-foreground">
              <UserAvatar name={teacherName} size="sm" />
              <span>{teacherName}</span>
            </div>
          )}

          <div className="mt-8 w-full">
            {isWaiting ? (
              <div className="flex flex-col items-center gap-3 rounded-2xl border border-amber-400/20 bg-amber-400/5 px-5 py-6">
                <span className="flex size-11 items-center justify-center rounded-full bg-amber-400/15 text-amber-500 dark:text-amber-300">
                  <Hourglass className="size-5" />
                </span>
                <p className="text-sm font-medium text-foreground">{t("liveRoom.waitingForHost")}</p>
                {scheduledIso && now < new Date(scheduledIso).getTime() && (
                  <p className="font-mono text-2xl font-semibold tracking-tight text-amber-600 tabular-nums dark:text-amber-200" dir="ltr">
                    {formatCountdown(scheduledIso, now)}
                  </p>
                )}
                <p className="text-xs leading-relaxed text-muted-foreground">{t("liveRoom.waitingHint")}</p>
              </div>
            ) : (
              <Button
                size="lg"
                onClick={handleJoin}
                disabled={joinMutation.isPending || isFinished}
                className="h-12 w-full gap-2 text-base font-semibold"
              >
                {joinMutation.isPending ? (
                  <Spinner className="size-4" />
                ) : isFinished ? (
                  t("liveRoom.sessionEnded")
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
              <p className="mt-4 text-sm text-destructive">{t("liveRoom.joinError")}</p>
            )}

            <Button
              variant="ghost"
              onClick={() => router.history.back()}
              className="mt-2 w-full text-muted-foreground hover:bg-muted hover:text-foreground"
            >
              {t("liveRoom.back")}
            </Button>
          </div>
        </div>
      </main>
    </div>
  )
}
