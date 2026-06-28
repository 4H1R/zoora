import { createFileRoute } from "@tanstack/react-router"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import type { GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse } from "@/api/model"
import type { PreJoinChoices } from "@/components/live/types"

import { useGetLiveRoomsId } from "@/api/live-sessions/live-sessions"
import { useGetUsersMe } from "@/api/users/users"
import { ActiveRoom } from "@/components/live/active-room"
import { deriveRoomRole } from "@/components/live/room-role"
import { PreJoinLobby } from "@/components/live/prejoin-lobby"
import { Spinner } from "@/components/ui/spinner"

export const Route = createFileRoute("/_auth/live/$liveId")({
  component: RouteComponent,
})

interface Connection {
  data: GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse
  choices: PreJoinChoices
}

function RouteComponent() {
  const { liveId } = Route.useParams()
  const { t } = useTranslation()
  const [connection, setConnection] = useState<Connection | null>(null)
  const { data: meData } = useGetUsersMe()
  const me = meData?.status === 200 ? meData.data.data : undefined
  const role = deriveRoomRole(me)
  const { data, isPending } = useGetLiveRoomsId(liveId, {
    query: {
      // While a student waits in the lobby for a scheduled room, poll so the
      // status flips from "created" to "active" the moment the host starts —
      // no manual refresh. Stop polling once connected to the room.
      refetchInterval: (query) => {
        const res = query.state.data
        const status = res?.status === 200 ? res.data.data?.status : undefined
        return !connection && status === "created" ? 5000 : false
      },
    },
  })

  const room = (data?.status === 200 && data.data.data) || undefined

  if (isPending) {
    return (
      <div className="flex h-screen items-center justify-center bg-background">
        <Spinner className="size-8 text-primary" />
      </div>
    )
  }

  if (connection?.data.token && connection.data.livekit_url) {
    return (
      <ActiveRoom
        token={connection.data.token}
        serverUrl={connection.data.livekit_url}
        choices={connection.choices}
        sessionName={room?.class_session?.name ?? t("liveRoom.session")}
        className={room?.class_session?.class?.name}
        liveId={liveId}
        chatId={connection.data.chat_id}
        role={role}
        onDisconnect={() => setConnection(null)}
      />
    )
  }

  return (
    <PreJoinLobby
      room={room}
      liveId={liveId}
      onJoined={(joinData, choices) => setConnection({ data: joinData, choices })}
    />
  )
}
