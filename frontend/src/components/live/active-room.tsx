import {
  LayoutContextProvider,
  LiveKitRoom,
  RoomAudioRenderer,
  useCreateLayoutContext,
  useLocalParticipant,
} from "@livekit/components-react"
import { Users } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import "@livekit/components-styles"
import "./livekit-overrides.css"

import {
  usePostLiveRoomsIdHand,
  usePostLiveRoomsIdLeave,
  usePostLiveRoomsIdParticipantsIdentityMute,
  usePutLiveRoomsIdParticipantsIdentityRole,
} from "@/api/live-sessions/live-sessions"
import { postMediaPresign, getMediaIdDownloadUrl } from "@/api/media/media"

import { ControlBar } from "./control-bar"
import { RoomHeader } from "./room-header"
import { RoomPanel } from "./room-panel"
import { canPublish, RoomRoleContext, type RoomRole, useRoomRole } from "./room-role"
import { Stage } from "./stage"
import type { PreJoinChoices, RoomTab } from "./types"
import { useRoomChat } from "./use-room-chat"
import { useRoomRoles } from "./use-room-roles"
import { useStage } from "./use-stage"
import { WebcamRail } from "./webcam-rail"

interface ActiveRoomProps {
  token: string
  serverUrl: string
  choices: PreJoinChoices
  sessionName: string
  className?: string
  liveId: string
  chatId?: string
  role: RoomRole
  onDisconnect: () => void
}

export function ActiveRoom({
  token,
  serverUrl,
  sessionName,
  className,
  liveId,
  chatId,
  role,
  onDisconnect,
}: ActiveRoomProps) {
  const leaveMutation = usePostLiveRoomsIdLeave()

  const handleLeave = () => {
    leaveMutation.mutate({ id: liveId }, { onSettled: onDisconnect })
  }

  return (
    <RoomRoleContext.Provider value={role}>
      <LiveKitRoom
        serverUrl={serverUrl}
        token={token}
        // Viewers can't publish; publishers start muted and enable in-room.
        audio={false}
        video={false}
        onDisconnected={onDisconnect}
        data-lk-theme="default"
        className="zoora-live relative flex flex-col overflow-hidden bg-zinc-950 text-zinc-100"
      >
        <RoomShell
          sessionName={sessionName}
          className={className}
          onLeave={handleLeave}
          leavePending={leaveMutation.isPending}
          liveId={liveId}
          chatId={chatId}
        />
        <RoomAudioRenderer />
      </LiveKitRoom>
    </RoomRoleContext.Provider>
  )
}

function RoomShell({
  sessionName,
  className,
  onLeave,
  leavePending,
  liveId,
  chatId,
}: {
  sessionName: string
  className?: string
  onLeave: () => void
  leavePending: boolean
  liveId: string
  chatId?: string
}) {
  const { t } = useTranslation()
  const layoutContext = useCreateLayoutContext()
  const chat = useRoomChat(chatId)
  const [tab, setTab] = useState<RoomTab | null>(null)
  const [readCount, setReadCount] = useState(0)
  const [railOpen, setRailOpen] = useState(false) // hidden by default (mobile-first)

  const { localParticipant } = useLocalParticipant()
  const states = useRoomRoles({})
  const role = useRoomRole()
  const isHost = role === "host"
  const myIdentity = localParticipant.identity
  const handRaised = states[myIdentity]?.handRaised ?? false

  const { stage, setStage } = useStage(isHost)
  const canDraw = localParticipant.permissions?.canPublish ?? false

  const onStartWhiteboard = () => setStage({ kind: "whiteboard" })

  const onShareSlides = async (file: File) => {
    try {
      const mime = file.type || "application/pdf"
      const presignRes = await postMediaPresign({
        model_type: "live_room",
        model_id: liveId,
        collection_name: "slides",
        file_name: file.name,
        mime_type: mime,
        size: file.size,
      })
      const uploadUrl = presignRes.status === 201 ? presignRes.data.data?.upload_url : undefined
      const mediaId = presignRes.status === 201 ? presignRes.data.data?.media?.id : undefined
      if (!uploadUrl || !mediaId) throw new Error("presign failed")

      const put = await fetch(uploadUrl, {
        method: "PUT",
        body: file,
        headers: { "Content-Type": mime },
      })
      if (!put.ok) throw new Error(`upload failed: ${put.status}`)

      // Get a presigned download URL so all clients can fetch the PDF
      const dlRes = await getMediaIdDownloadUrl(mediaId)
      const dlUrl = dlRes.status === 200 ? dlRes.data.data?.url : undefined
      if (!dlUrl) throw new Error("download url failed")

      setStage({ kind: "slides", url: dlUrl, page: 1, numPages: 0 })
    } catch {
      toast.error(t("liveRoom.errors.upload"))
    }
  }

  const onStopStage = () => setStage({ kind: "none" })

  const onPageChange = (page: number) => {
    if (stage.kind === "slides") {
      setStage({ ...stage, page })
    }
  }

  const onLoadNumPages = (numPages: number) => {
    if (stage.kind === "slides") {
      setStage({ ...stage, numPages })
    }
  }

  const roleMutation = usePutLiveRoomsIdParticipantsIdentityRole()
  const muteMutation = usePostLiveRoomsIdParticipantsIdentityMute()
  const handMutation = usePostLiveRoomsIdHand()

  const onToggleHand = () => handMutation.mutate({ id: liveId, data: { raised: !handRaised } })
  const onSetRole = (identity: string, r: "presenter" | "viewer") =>
    roleMutation.mutate({ id: liveId, identity, data: { role: r } })
  const onMute = (identity: string, trackSid: string) =>
    muteMutation.mutate({ id: liveId, identity, data: { track_sid: trackSid, muted: true } })

  useEffect(() => {
    if (tab === "chat") setReadCount(chat.count)
  }, [tab, chat.count])

  const unread = Math.max(0, chat.count - readCount)

  return (
    <LayoutContextProvider value={layoutContext}>
      <RoomHeader sessionName={sessionName} className={className} />

      <div className="flex min-h-0 flex-1">
        <div className="relative flex min-w-0 flex-1 flex-col">
          <div className="flex min-h-0 flex-1 gap-3 p-3 sm:p-4">
            {railOpen && (
              <div className="hidden md:block">
                <WebcamRail orientation="vertical" />
              </div>
            )}
            <div className="min-w-0 flex-1">
              <Stage
                stage={stage}
                isHost={isHost}
                liveId={liveId}
                canDraw={canDraw}
                onPageChange={onPageChange}
                onLoadNumPages={onLoadNumPages}
              />
            </div>
          </div>

          {railOpen && (
            <div className="px-3 pb-24 md:hidden">
              <WebcamRail orientation="horizontal" />
            </div>
          )}

          <ControlBar
            tab={tab}
            openTab={(next) => setTab(next)}
            closePanel={() => setTab(null)}
            onLeave={onLeave}
            leavePending={leavePending}
            unread={unread}
            handRaised={handRaised}
            onToggleHand={onToggleHand}
            canShareStage={canPublish(role)}
            stageKind={stage.kind}
            onShareSlides={onShareSlides}
            onStopStage={onStopStage}
            onStartWhiteboard={onStartWhiteboard}
          />

          <button
            type="button"
            onClick={() => setRailOpen((v) => !v)}
            aria-label={t("liveRoom.toggleRail")}
            className="absolute end-4 top-4 z-20 flex size-9 items-center justify-center rounded-lg bg-black/50 text-zinc-200 backdrop-blur-md transition-colors hover:bg-black/70"
          >
            <Users className="size-4" />
          </button>
        </div>

        <RoomPanel
          tab={tab ?? "chat"}
          setTab={setTab}
          open={tab !== null}
          onClose={() => setTab(null)}
          chat={chat}
          unread={unread}
          states={states}
          isHost={role === "host"}
          liveId={liveId}
          onSetRole={onSetRole}
          onMute={onMute}
        />
      </div>
    </LayoutContextProvider>
  )
}
