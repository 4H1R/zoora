import {
  isTrackReference,
  LayoutContextProvider,
  LiveKitRoom,
  RoomAudioRenderer,
  setLogLevel,
  useCreateLayoutContext,
  useLocalParticipant,
  useTracks,
} from "@livekit/components-react"
import { Track } from "livekit-client"
import { Users } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import "@livekit/components-styles"
import "./livekit-overrides.css"

// Silence LiveKit's verbose signal/track console logs (info/debug); keep warnings + errors.
setLogLevel("warn")

import {
  usePostLiveRoomsIdEnd,
  usePostLiveRoomsIdHand,
  usePostLiveRoomsIdLeave,
  usePostLiveRoomsIdParticipantsIdentityMute,
  usePutLiveRoomsIdParticipantsIdentityRole,
} from "@/api/live-sessions/live-sessions"
import { postMediaPresign, getMediaIdDownloadUrl } from "@/api/media/media"
import { usePostPollsIdAnswer } from "@/api/polls/polls"
import { cn } from "@/lib/utils"

import { ControlBar } from "./control-bar"
import { VotePollModal } from "./panels/vote-poll-modal"
import { ReconnectOverlay } from "./reconnect-overlay"
import { RoomHeader } from "./room-header"
import { RoomPanel } from "./room-panel"
import { canPublish, RoomRoleContext, type RoomRole, useRoomRole } from "./room-role"
import { Stage } from "./stage"
import type { PreJoinChoices, RoomTab } from "./types"
import { useRoomChat } from "./use-room-chat"
import { useRoomPolls } from "./use-room-polls"
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
  const endMutation = usePostLiveRoomsIdEnd()

  const handleLeave = () => {
    leaveMutation.mutate({ id: liveId }, { onSettled: onDisconnect })
  }

  const handleEndRoom = () => {
    endMutation.mutate({ id: liveId }, { onSettled: onDisconnect })
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
        className="zoora-live relative flex flex-col overflow-hidden bg-background text-foreground"
      >
        <RoomShell
          sessionName={sessionName}
          className={className}
          onLeave={handleLeave}
          leavePending={leaveMutation.isPending}
          onEndRoom={handleEndRoom}
          endPending={endMutation.isPending}
          liveId={liveId}
          chatId={chatId}
        />
        <ReconnectOverlay />
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
  onEndRoom,
  endPending,
  liveId,
  chatId,
}: {
  sessionName: string
  className?: string
  onLeave: () => void
  leavePending: boolean
  onEndRoom: () => void
  endPending: boolean
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

  // Poll session lives at room level so its data-channel listener stays mounted
  // regardless of which tab/panel is open. Viewers get a modal popup to vote.
  const polls = useRoomPolls(isHost)
  const answerMutation = usePostPollsIdAnswer()
  const onVote = (value: string) => {
    if (!polls.activePoll) return
    answerMutation.mutate(
      { id: polls.activePoll.pollId, data: { options: [value] } },
      {
        onSuccess: () => polls.markAnswered(),
        onError: () => toast.error(t("liveRoom.polls.voteError")),
      },
    )
  }

  // Camera publishers drive the webcam rail; with none, the rail (and its toggle) has nothing to show.
  const hasCameras =
    useTracks([{ source: Track.Source.Camera, withPlaceholder: false }], {
      onlySubscribed: false,
    }).filter(isTrackReference).length > 0

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
            onEndRoom={onEndRoom}
            endPending={endPending}
            unread={unread}
            handRaised={handRaised}
            onToggleHand={onToggleHand}
            canShareStage={canPublish(role)}
            stageKind={stage.kind}
            onShareSlides={onShareSlides}
            onStopStage={onStopStage}
            onStartWhiteboard={onStartWhiteboard}
          />

          {hasCameras && (
            <button
              type="button"
              onClick={() => setRailOpen((v) => !v)}
              aria-label={t("liveRoom.toggleRail")}
              aria-pressed={railOpen}
              className={cn(
                // Solid bg, no backdrop-blur: floats over the <video> stage, and a
                // backdrop-filter pass over a video paints it black on some GPUs.
                "absolute end-4 top-4 z-20 flex size-9 items-center justify-center rounded-lg transition-colors",
                railOpen
                  ? "bg-primary text-primary-foreground hover:bg-primary/90"
                  : "bg-black/70 text-zinc-200 hover:bg-black/80"
              )}
            >
              <Users className="size-4" />
            </button>
          )}
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
          polls={polls}
          onVote={onVote}
          answerPending={answerMutation.isPending}
        />
      </div>

      {!isHost && (
        <VotePollModal
          activePoll={polls.activePoll}
          results={polls.results}
          hasAnswered={polls.hasAnswered}
          isPending={answerMutation.isPending}
          onVote={onVote}
        />
      )}
    </LayoutContextProvider>
  )
}
