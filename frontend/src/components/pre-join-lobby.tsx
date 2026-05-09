import { Link, useRouter } from "@tanstack/react-router"
import { useTranslation } from "react-i18next"

import type {
  GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse,
  GithubCom4H1RZooraInternalDomainLiveRoom,
} from "@/api/model"

import { usePostLiveRoomsIdJoin } from "@/api/live-sessions/live-sessions"
import { useGetUsersMe } from "@/api/users/users"
import { StatusBadge } from "@/components/status-badge"
import { UserAvatar } from "@/components/user-avatar"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { Spinner } from "@/components/ui/spinner"
import {
  ArrowLeft,
  Calendar,
  ChevronLeft,
  Circle,
  Clock,
  Info,
  Lock,
  Users,
} from "lucide-react"

interface PreJoinLobbyProps {
  room: GithubCom4H1RZooraInternalDomainLiveRoom | undefined
  liveId: string
  onJoined: (data: GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse) => void
}

export function PreJoinLobby({ room, liveId, onJoined }: PreJoinLobbyProps) {
  const { t } = useTranslation()
  const router = useRouter()
  const joinMutation = usePostLiveRoomsIdJoin()
  const { data: meData } = useGetUsersMe()
  const orgId = meData?.status === 200 ? meData.data.data?.organization_id : undefined

  const handleJoin = () => {
    joinMutation.mutate(
      { id: liveId },
      {
        onSuccess: (res) => {
          const joinData = (res.status === 200 && res.data.data) || undefined
          if (joinData?.token && joinData?.livekit_url) {
            onJoined(joinData)
          }
        },
      }
    )
  }

  const isFinished = room?.status === "finished"
  const isActive = room?.status === "active"
  const session = room?.class_session
  const className = session?.class?.name
  const teacherName = session?.class?.user?.name
  const sessionName = session?.name
  const maxParticipants = room?.config?.max_participants
  const autoRecord = room?.config?.auto_record

  const startTime = session?.start_time
    ? new Date(session.start_time).toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      })
    : undefined

  const startDate = session?.start_time
    ? new Date(session.start_time).toLocaleDateString(undefined, {
        weekday: "long",
        month: "long",
        day: "numeric",
      })
    : undefined

  return (
    <div className="bg-muted/40 flex min-h-screen items-center justify-center p-6">
      <div className="w-full max-w-[640px]">
        <Link
          to={orgId ? "/org/$orgId" : "/"}
          params={orgId ? { orgId } : undefined}
          className="text-muted-foreground hover:text-foreground mb-5 inline-flex items-center gap-1.5 text-[13px]"
        >
          <ChevronLeft className="size-3.5" />
          <span>{t("liveRoom.backToDashboard")}</span>
        </Link>

        <Card className="overflow-hidden shadow-sm">
          {/* Eyebrow header */}
          {room?.status && (
            <div className="border-b px-6 py-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="flex items-center gap-2.5">
                  <StatusBadge
                    status={
                      isActive ? "live" : isFinished ? "ended" : "scheduled"
                    }
                  />
                  {isActive && (
                    <span className="text-muted-foreground text-[13px]">
                      {t("liveRoom.classStarted")}
                    </span>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* Body */}
          <CardContent className="p-8">
            {className && (
              <div className="text-muted-foreground mb-2 text-end text-xs font-medium tracking-wide uppercase ltr:text-start">
                {className}
              </div>
            )}

            <h1 className="text-foreground text-[30px] leading-[1.2] font-semibold tracking-tight">
              {sessionName ?? t("liveRoom.session")}
            </h1>

            {session?.description && (
              <p className="text-muted-foreground mt-2 text-[15px]">
                {session.description}
              </p>
            )}

            <Separator className="my-6" />

            {/* Meta grid */}
            <div className="grid grid-cols-2 gap-x-8 gap-y-4">
              {teacherName && (
                <MetaRow icon={<Users className="size-3.5" />} label={t("liveRoom.instructor")}>
                  <div className="flex items-center gap-2">
                    <UserAvatar name={teacherName} size="sm" />
                    <span className="text-foreground text-sm font-medium">
                      {teacherName}
                    </span>
                  </div>
                </MetaRow>
              )}

              {startDate && (
                <MetaRow icon={<Calendar className="size-3.5" />} label={t("liveRoom.startTime")}>
                  <span className="text-foreground text-sm">
                    {startDate}
                    {startTime && (
                      <>
                        {" · "}
                        <span className="font-mono ltr:inline-block" dir="ltr">
                          {startTime}
                        </span>
                      </>
                    )}
                  </span>
                </MetaRow>
              )}

              {maxParticipants != null && maxParticipants > 0 && (
                <MetaRow icon={<Users className="size-3.5" />} label={t("liveRoom.participants")}>
                  <span className="text-foreground text-sm">
                    <span className="font-mono ltr:inline-block" dir="ltr">
                      {maxParticipants}
                    </span>
                    <span className="text-muted-foreground ms-1.5">
                      {t("liveRoom.capacity")}
                    </span>
                  </span>
                </MetaRow>
              )}

              {autoRecord && (
                <MetaRow icon={<Clock className="size-3.5" />} label={t("common.recording")}>
                  <span className="text-foreground text-sm">
                    {t("liveRoom.recordingEnabled")}
                  </span>
                </MetaRow>
              )}
            </div>

            {/* Recording notice */}
            {autoRecord && (
              <div className="bg-muted/50 border-border mt-6 flex items-start gap-2.5 rounded-lg border p-3">
                <Circle className="mt-0.5 size-2.5 shrink-0 fill-[#dc2626] text-[#dc2626]" />
                <p className="text-muted-foreground text-[13px] leading-relaxed">
                  {t("liveRoom.recordingNotice")}
                </p>
              </div>
            )}

            {joinMutation.isError && (
              <p className="mt-4 text-center text-sm text-red-500">
                {t("liveRoom.joinError")}
              </p>
            )}
          </CardContent>

          {/* Footer */}
          <div className="bg-muted/30 flex items-center justify-between gap-3 border-t px-6 py-4">
            <Button variant="outline" size="sm" onClick={() => router.history.back()}>
              <ChevronLeft className="me-1.5 size-3.5" />
              {t("liveRoom.back")}
            </Button>

            <Button
              size="lg"
              onClick={handleJoin}
              disabled={joinMutation.isPending || isFinished}
              className="gap-2"
            >
              {joinMutation.isPending ? (
                <Spinner className="size-4" />
              ) : (
                <span>
                  {isFinished
                    ? t("liveRoom.sessionEnded")
                    : t("liveRoom.joinSession")}
                </span>
              )}
              {!joinMutation.isPending && <ArrowLeft className="size-4 rtl:rotate-180" />}
            </Button>
          </div>
        </Card>

        {/* Footer hints */}
        <div className="text-muted-foreground mt-4 flex items-center justify-center gap-3.5 text-xs">
          <span className="inline-flex items-center gap-1.5">
            <Lock className="size-3" />
            <span>{t("liveRoom.secureEntry")}</span>
          </span>
          <span className="bg-border inline-block h-3 w-px" />
          <span className="inline-flex items-center gap-1.5">
            <Info className="size-3" />
            <span>{t("liveRoom.avSettingsNote")}</span>
          </span>
        </div>
      </div>
    </div>
  )
}

function MetaRow({
  icon,
  label,
  children,
}: {
  icon: React.ReactNode
  label: string
  children: React.ReactNode
}) {
  return (
    <div>
      <div className="text-muted-foreground mb-1 flex items-center gap-1.5 text-xs font-medium">
        {icon}
        <span>{label}</span>
      </div>
      <div>{children}</div>
    </div>
  )
}
