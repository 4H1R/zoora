import { LiveKitRoom, RoomAudioRenderer, VideoConference } from "@livekit/components-react"
import { createFileRoute } from "@tanstack/react-router"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import "@livekit/components-styles"

import type {
  GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse,
  GithubCom4H1RZooraInternalDomainLiveRoom,
} from "@/api/model"

import { LogIn, LogOut, Mic, MicOff, MonitorPlay, Users, Video, VideoOff } from "lucide-react"

import { useGetLiveRoomsId, usePostLiveRoomsIdJoin, usePostLiveRoomsIdLeave } from "@/api/live-sessions/live-sessions"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Spinner } from "@/components/ui/spinner"

export const Route = createFileRoute("/live/$liveId")({
  component: RouteComponent,
})

function RouteComponent() {
  const { liveId } = Route.useParams()
  const { t } = useTranslation()
  const { data, isPending } = useGetLiveRoomsId(liveId)
  const [connectionData, setConnectionData] = useState<GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse | null>(
    null
  )

  const room = (data?.status === 200 && data.data.data) || undefined

  if (isPending) {
    return (
      <div className="flex h-screen items-center justify-center bg-zinc-950">
        <Spinner className="size-8 text-white" />
      </div>
    )
  }

  if (connectionData?.token && connectionData?.livekit_url) {
    return (
      <ActiveRoom
        token={connectionData.token}
        serverUrl={connectionData.livekit_url}
        sessionName={room?.class_session?.name ?? t("liveRoom.session")}
        liveId={liveId}
        onDisconnect={() => setConnectionData(null)}
      />
    )
  }

  return <PreJoinLobby room={room} liveId={liveId} onJoined={setConnectionData} />
}

function PreJoinLobby({
  room,
  liveId,
  onJoined,
}: {
  room: GithubCom4H1RZooraInternalDomainLiveRoom | undefined
  liveId: string
  onJoined: (data: GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse) => void
}) {
  const { t } = useTranslation()
  const [audioEnabled, setAudioEnabled] = useState(true)
  const [videoEnabled, setVideoEnabled] = useState(true)
  const joinMutation = usePostLiveRoomsIdJoin()

  const statusColor: Record<string, string> = {
    created: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30",
    active: "bg-emerald-500/20 text-emerald-400 border-emerald-500/30",
    finished: "bg-zinc-500/20 text-zinc-400 border-zinc-500/30",
  }

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

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-950 p-4">
      <Card className="w-full max-w-lg border-zinc-800 bg-zinc-900">
        <CardHeader className="space-y-4 text-center">
          <div className="mx-auto flex size-16 items-center justify-center rounded-2xl bg-indigo-500/10">
            <MonitorPlay className="size-8 text-indigo-400" />
          </div>
          <div className="space-y-2">
            <CardTitle className="text-2xl text-white">{room?.class_session?.name ?? t("liveRoom.session")}</CardTitle>
            {room?.status && (
              <Badge variant="outline" className={statusColor[room.status] ?? ""}>
                {t(`status.${room.status}`, room.status)}
              </Badge>
            )}
          </div>
          {room?.config && (
            <div className="flex items-center justify-center gap-4 text-sm text-zinc-400">
              <span className="flex items-center gap-1">
                <Users className="size-4" />
                {t("liveRoom.maxParticipants", {
                  count: room.config.max_participants,
                })}
              </span>
            </div>
          )}
        </CardHeader>

        <CardContent className="space-y-6">
          <div className="flex items-center justify-center gap-3">
            <Button
              variant={audioEnabled ? "secondary" : "destructive"}
              size="icon"
              className="size-12 rounded-full"
              onClick={() => setAudioEnabled(!audioEnabled)}
            >
              {audioEnabled ? <Mic className="size-5" /> : <MicOff className="size-5" />}
            </Button>
            <Button
              variant={videoEnabled ? "secondary" : "destructive"}
              size="icon"
              className="size-12 rounded-full"
              onClick={() => setVideoEnabled(!videoEnabled)}
            >
              {videoEnabled ? <Video className="size-5" /> : <VideoOff className="size-5" />}
            </Button>
          </div>

          {joinMutation.isError && <p className="text-center text-sm text-red-400">{t("liveRoom.joinError")}</p>}

          <Button className="w-full" size="lg" onClick={handleJoin} disabled={joinMutation.isPending || isFinished}>
            {joinMutation.isPending ? <Spinner className="me-2 size-4" /> : <LogIn className="me-2 size-4" />}
            {isFinished ? t("liveRoom.sessionEnded") : t("liveRoom.joinSession")}
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}

function ActiveRoom({
  token,
  serverUrl,
  sessionName,
  liveId,
  onDisconnect,
}: {
  token: string
  serverUrl: string
  sessionName: string
  liveId: string
  onDisconnect: () => void
}) {
  const { t } = useTranslation()
  const leaveMutation = usePostLiveRoomsIdLeave()

  const handleDisconnect = () => {
    leaveMutation.mutate({ id: liveId }, { onSettled: onDisconnect })
  }

  return (
    <div className="flex h-screen flex-col bg-zinc-950" data-lk-theme="default">
      <header className="flex items-center justify-between border-b border-zinc-800 px-4 py-2">
        <div className="flex items-center gap-3">
          <MonitorPlay className="size-5 text-indigo-400" />
          <span className="text-sm font-medium text-white">{sessionName}</span>
          <Badge variant="outline" className="border-emerald-500/30 bg-emerald-500/20 text-emerald-400">
            {t("status.live")}
          </Badge>
        </div>
        <Button variant="destructive" size="sm" onClick={handleDisconnect} disabled={leaveMutation.isPending}>
          <LogOut className="me-2 size-4" />
          {t("liveRoom.leave")}
        </Button>
      </header>

      <div className="flex-1 overflow-hidden">
        <LiveKitRoom serverUrl={serverUrl} token={token} audio={true} video={true} className="h-full">
          <VideoConference />
          <RoomAudioRenderer />
        </LiveKitRoom>
      </div>
    </div>
  )
}
