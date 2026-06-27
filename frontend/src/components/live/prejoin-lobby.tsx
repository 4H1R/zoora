import { Link, useRouter } from "@tanstack/react-router"
import { ChevronLeft, Hourglass, Radio } from "lucide-react"
import { useTranslation } from "react-i18next"

import type {
  GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse,
  GithubCom4H1RZooraInternalDomainLiveRoom,
} from "@/api/model"
import type { PreJoinChoices } from "./types"

import { usePostLiveRoomsIdJoin } from "@/api/live-sessions/live-sessions"
import { useGetUsersMe } from "@/api/users/users"
import { StatusBadge } from "@/components/status-badge"
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
    <div className="flex min-h-screen flex-col bg-zinc-950 text-zinc-100">
      <header className="flex items-center px-5 py-4 sm:px-8">
        <Link
          to={orgId ? "/org" : "/"}
          className="inline-flex items-center gap-1.5 text-[13px] text-zinc-400 transition-colors hover:text-zinc-100"
        >
          <ChevronLeft className="size-3.5 rtl:rotate-180" />
          <span>{t("liveRoom.backToDashboard")}</span>
        </Link>
      </header>

      <main className="flex flex-1 items-center justify-center px-5 pb-16">
        <div className="w-full max-w-md rounded-3xl border border-white/10 bg-white/[0.03] p-7 text-center backdrop-blur-xl">
          <div className="flex justify-center">
            <StatusBadge status={room?.status === "active" ? "live" : isFinished ? "ended" : "scheduled"} />
          </div>

          {className && (
            <div className="mt-5 text-xs font-medium tracking-wide text-primary uppercase">{className}</div>
          )}
          <h1 className="mt-1.5 text-2xl leading-tight font-semibold tracking-tight text-white">
            {sessionName ?? t("liveRoom.session")}
          </h1>

          {teacherName && (
            <div className="mt-4 inline-flex items-center gap-2 rounded-full bg-white/5 px-3 py-1.5">
              <UserAvatar name={teacherName} size="sm" />
              <span className="text-sm text-zinc-200">{teacherName}</span>
            </div>
          )}

          <div className="mt-7">
            {isWaiting ? (
              <div className="flex flex-col items-center gap-3 rounded-2xl border border-amber-400/20 bg-amber-400/5 px-5 py-6">
                <span className="flex size-11 items-center justify-center rounded-full bg-amber-400/15 text-amber-300">
                  <Hourglass className="size-5 animate-pulse" />
                </span>
                <p className="text-sm font-medium text-zinc-100">{t("liveRoom.waitingForHost")}</p>
                {scheduledIso && now < new Date(scheduledIso).getTime() && (
                  <p className="font-mono text-2xl font-semibold tracking-tight text-amber-200 tabular-nums" dir="ltr">
                    {formatCountdown(scheduledIso, now)}
                  </p>
                )}
                <p className="text-xs leading-relaxed text-zinc-400">{t("liveRoom.waitingHint")}</p>
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
          </div>

          {joinMutation.isError && !isWaiting && (
            <p className="mt-4 text-sm text-red-400">{t("liveRoom.joinError")}</p>
          )}

          <Button
            variant="ghost"
            onClick={() => router.history.back()}
            className="mt-2 w-full text-zinc-400 hover:bg-white/5 hover:text-zinc-100"
          >
            {t("liveRoom.back")}
          </Button>
        </div>
      </main>
    </div>
  )
}
