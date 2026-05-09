import { LiveKitRoom, RoomAudioRenderer, VideoConference } from "@livekit/components-react"
import { createFileRoute } from "@tanstack/react-router"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import "@livekit/components-styles"

import type { GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse } from "@/api/model"

import { LogOut, MonitorPlay } from "lucide-react"

import { useGetLiveRoomsId, usePostLiveRoomsIdLeave } from "@/api/live-sessions/live-sessions"
import { PreJoinLobby } from "@/components/pre-join-lobby"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
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
      <div className="bg-muted/40 flex h-screen items-center justify-center">
        <Spinner className="size-8" />
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
