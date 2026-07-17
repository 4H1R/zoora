import {
  isTrackReference,
  LayoutContextProvider,
  LiveKitRoom,
  RoomAudioRenderer,
  setLogLevel,
  useCreateLayoutContext,
  useLocalParticipant,
  useParticipants,
  useTracks,
} from "@livekit/components-react"
import { useQueryClient } from "@tanstack/react-query"
import { DisconnectReason, Track } from "livekit-client"
import { Users } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import "@livekit/components-styles"
import "./livekit-overrides.css"

import type { RoomRole } from "./room-role"
import type { PreJoinChoices, RoomTab } from "./types"

import {
  getGetLiveRoomsIdRecordingsQueryKey,
  useDeleteLiveRoomsIdParticipantsIdentity,
  useGetLiveRoomsIdRecordings,
  usePostLiveRoomsIdEnd,
  usePostLiveRoomsIdHand,
  usePostLiveRoomsIdLeave,
  usePostLiveRoomsIdParticipantsIdentityMute,
  usePostLiveRoomsIdRecordings,
  usePostLiveRoomsIdRecordingsRecordingIdStop,
  usePutLiveRoomsIdParticipantsIdentityHand,
  usePutLiveRoomsIdParticipantsIdentityRole,
} from "@/api/live-sessions/live-sessions"
import { getMediaIdDownloadUrl, postMediaPresign } from "@/api/media/media"
import { usePostPollsIdAnswer } from "@/api/polls/polls"
import { isPlanError } from "@/lib/plan-errors"
import { cn } from "@/lib/utils"

import { ControlBar } from "./control-bar"
import { VotePollModal } from "./panels/vote-poll-modal"
import { ReconnectOverlay } from "./reconnect-overlay"
import { SlidesUploadOverlay, type SlidesUpload } from "./slides-upload-overlay"
import { RoomHeader } from "./room-header"
import { RoomPanel } from "./room-panel"
import { canPublish, RoomRoleContext, useRoomRole } from "./room-role"
import { Stage } from "./stage"
import { usePublishPresence } from "./use-publish-presence"
import { useRoomChat } from "./use-room-chat"
import { useRoomPolls } from "./use-room-polls"
import { useRoomQa } from "./use-room-qa"
import { useRoomRoles } from "./use-room-roles"
import { useStage } from "./use-stage"
import { WebcamRail } from "./webcam-rail"

// Silence LiveKit's verbose signal/track console logs (info/debug); keep warnings + errors.
setLogLevel("warn")

// PUT the file to S3 via XHR (not fetch) so we get real upload-progress events —
// fetch/redaxios can't report byte progress. `xhrRef` exposes the request so the
// host can abort a large upload; an abort rejects with "aborted" (a silent cancel,
// not an error). Resolves on 2xx.
function putWithProgress(
  url: string,
  file: File,
  mime: string,
  onProgress: (pct: number) => void,
  xhrRef: React.RefObject<XMLHttpRequest | null>
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhrRef.current = xhr
    xhr.open("PUT", url)
    xhr.setRequestHeader("Content-Type", mime)
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100))
    }
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) resolve()
      else reject(new Error(`upload failed: ${xhr.status}`))
    }
    xhr.onerror = () => reject(new Error("network error"))
    xhr.onabort = () => reject(new Error("aborted"))
    xhr.send(file)
  })
}

interface ActiveRoomProps {
  token: string
  serverUrl: string
  choices: PreJoinChoices
  sessionName: string
  className?: string
  liveId: string
  chatId?: string
  role: RoomRole
  actualStartTime?: string
  onDisconnect: () => void
  onEnded: () => void
}

export function ActiveRoom({
  token,
  serverUrl,
  sessionName,
  className,
  liveId,
  chatId,
  role,
  actualStartTime,
  onDisconnect,
  onEnded,
}: ActiveRoomProps) {
  const { t } = useTranslation()
  const leaveMutation = usePostLiveRoomsIdLeave()
  const endMutation = usePostLiveRoomsIdEnd()

  const handleLeave = () => {
    leaveMutation.mutate({ id: liveId }, { onSettled: onDisconnect })
  }

  // A host kick surfaces as a LiveKit PARTICIPANT_REMOVED disconnect — tell the
  // booted user why before tearing the room down (which returns them to lobby).
  const handleDisconnected = (reason?: DisconnectReason) => {
    if (reason === DisconnectReason.PARTICIPANT_REMOVED) {
      toast.error(t("liveRoom.people.youWereRemoved"))
    }
    onDisconnect()
  }

  const handleEndRoom = () => {
    endMutation.mutate({ id: liveId }, { onSettled: onEnded })
  }

  return (
    <RoomRoleContext.Provider value={role}>
      <LiveKitRoom
        serverUrl={serverUrl}
        token={token}
        // Viewers can't publish; publishers start muted and enable in-room.
        audio={false}
        video={false}
        onDisconnected={handleDisconnected}
        data-lk-theme="default"
        className="zoora-live bg-background text-foreground relative flex flex-col overflow-hidden"
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
          actualStartTime={actualStartTime}
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
  actualStartTime,
}: {
  sessionName: string
  className?: string
  onLeave: () => void
  leavePending: boolean
  onEndRoom: () => void
  endPending: boolean
  liveId: string
  chatId?: string
  actualStartTime?: string
}) {
  const { t } = useTranslation()
  const layoutContext = useCreateLayoutContext()
  const chat = useRoomChat(chatId)
  const [tab, setTab] = useState<RoomTab | null>(null)
  const [readCount, setReadCount] = useState(0)
  const [railOpen, setRailOpen] = useState(false) // hidden by default (mobile-first)

  const { localParticipant } = useLocalParticipant()
  // Every client self-publishes device/OS/browser + live network stats into its
  // participant attributes so hosts can inspect any participant (people panel).
  usePublishPresence()
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
  // Q&A session lives at room level so its data-channel listener stays mounted
  // regardless of which tab is open (same rule as chat/polls).
  const qa = useRoomQa(liveId)
  const answerMutation = usePostPollsIdAnswer()
  const onVote = (value: string) => {
    if (!polls.activePoll) return
    answerMutation.mutate(
      { id: polls.activePoll.pollId, data: { options: [value] } },
      {
        onSuccess: () => polls.markAnswered(),
        onError: (err) => {
          // The only 409 on this endpoint is POLL_CLOSED — the room finished (or
          // the host closed the poll) between render and click.
          const status =
            (err as { status?: number; response?: { status?: number } })?.status ??
            (err as { response?: { status?: number } })?.response?.status
          toast.error(status === 409 ? t("liveRoom.polls.closed") : t("liveRoom.polls.voteError"))
        },
      }
    )
  }

  // Camera publishers drive the webcam rail; with none, the rail (and its toggle) has nothing to show.
  const hasCameras =
    useTracks([{ source: Track.Source.Camera, withPlaceholder: false }], {
      onlySubscribed: false,
    }).filter(isTrackReference).length > 0

  const onStartWhiteboard = () => setStage({ kind: "whiteboard" })

  // Upload progress for a host-shared PDF, surfaced as an overlay over the stage
  // so the host sees it's working (and can cancel). Null when idle.
  const [slidesUpload, setSlidesUpload] = useState<SlidesUpload | null>(null)
  const uploadXhrRef = useRef<XMLHttpRequest | null>(null)

  const onShareSlides = async (file: File) => {
    setSlidesUpload({ fileName: file.name, phase: "preparing", progress: 0 })
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

      setSlidesUpload({ fileName: file.name, phase: "uploading", progress: 0 })
      await putWithProgress(uploadUrl, file, mime, (pct) => {
        setSlidesUpload((u) => (u ? { ...u, progress: pct } : u))
      }, uploadXhrRef)

      // Uploaded — resolve a presigned download URL all clients can fetch.
      setSlidesUpload({ fileName: file.name, phase: "processing", progress: 100 })
      const dlRes = await getMediaIdDownloadUrl(mediaId)
      const dlUrl = dlRes.status === 200 ? dlRes.data.data?.url : undefined
      if (!dlUrl) throw new Error("download url failed")

      // Broadcast to every participant (reliable data channel + late-join resync).
      setStage({ kind: "slides", url: dlUrl, page: 1, numPages: 0 })
      setSlidesUpload(null)
      // Confirm to the host that it's now live for the whole room.
      toast.success(t("liveRoom.stage.shared"))
    } catch (err) {
      setSlidesUpload(null)
      // A user-initiated cancel is not an error — stay quiet.
      if ((err as Error)?.message !== "aborted") toast.error(t("liveRoom.errors.upload"))
    } finally {
      uploadXhrRef.current = null
    }
  }

  const onCancelSlidesUpload = () => {
    uploadXhrRef.current?.abort()
    setSlidesUpload(null)
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
  const lowerHandMutation = usePutLiveRoomsIdParticipantsIdentityHand()
  const removeMutation = useDeleteLiveRoomsIdParticipantsIdentity()

  // Recording — host only. Poll the list to know if an egress is live (status
  // "started") and to pick up the async webhook-driven transition to completed.
  const queryClient = useQueryClient()
  const recordingsQuery = useGetLiveRoomsIdRecordings(liveId, undefined, {
    query: { enabled: isHost, refetchInterval: 15000 },
  })
  const activeRecording =
    recordingsQuery.data?.status === 200
      ? recordingsQuery.data.data.data?.items?.find((r) => r.status === "started")
      : undefined
  const isRecording = Boolean(activeRecording)
  const invalidateRecordings = () =>
    queryClient.invalidateQueries({ queryKey: getGetLiveRoomsIdRecordingsQueryKey(liveId) })
  const startRecording = usePostLiveRoomsIdRecordings()
  const stopRecording = usePostLiveRoomsIdRecordingsRecordingIdStop()
  const recordingPending = startRecording.isPending || stopRecording.isPending

  const onToggleRecording = () => {
    if (recordingPending) return
    if (activeRecording?.id) {
      stopRecording.mutate(
        { id: liveId, recordingId: activeRecording.id },
        {
          onSuccess: () => {
            toast.success(t("liveRoom.recording.stopped"))
            invalidateRecordings()
          },
          onError: () => toast.error(t("liveRoom.recording.stopError")),
        }
      )
    } else {
      startRecording.mutate(
        { id: liveId },
        {
          onSuccess: () => {
            toast.success(t("liveRoom.recording.started"))
            invalidateRecordings()
          },
          onError: (error) => {
            // Plan-gate 402s get a central upgrade toast (see main.tsx); only
            // handle the non-plan failure here.
            if (isPlanError(error)) return
            toast.error(t("liveRoom.recording.startError"))
          },
        }
      )
    }
  }

  const onToggleHand = () => handMutation.mutate({ id: liveId, data: { raised: !handRaised } })
  const onSetRole = (identity: string, r: "presenter" | "viewer") =>
    roleMutation.mutate({ id: liveId, identity, data: { role: r } })
  const onMute = (identity: string, trackSid: string) =>
    muteMutation.mutate({ id: liveId, identity, data: { track_sid: trackSid, muted: true } })
  const onLowerHand = (identity: string) => lowerHandMutation.mutate({ id: liveId, identity, data: { raised: false } })
  const onRemove = (identity: string, name: string) =>
    removeMutation.mutate(
      { id: liveId, identity },
      {
        onSuccess: () => toast.success(t("liveRoom.people.removed", { name })),
        onError: () => toast.error(t("liveRoom.people.removeError")),
      }
    )

  // Remote identities with a raised hand — drives the People badge and the
  // host toast. Excludes self (you can't raise a hand at yourself).
  const raisedIdentities = Object.entries(states)
    .filter(([identity, s]) => s.handRaised && identity !== myIdentity)
    .map(([identity]) => identity)
  const raisedHandCount = raisedIdentities.length

  // Host-only toast on each new raised hand (edge-detect false→true). Toasts are
  // "armed" only after a short settle window: a host joining/reconnecting to a
  // room with hands already up receives them as a burst of late-join re-announce
  // events that trickle in just after mount — not synchronously — so a first-run
  // baseline can't catch them. During the window we still record the baseline,
  // so once armed only hands raised afterwards toast. The People queue covers the
  // pre-existing ones.
  const participants = useParticipants()
  const prevRaisedRef = useRef<Set<string>>(new Set())
  const toastArmedRef = useRef(false)
  useEffect(() => {
    if (!isHost) return
    const id = setTimeout(() => {
      toastArmedRef.current = true
    }, 3000)
    return () => {
      toastArmedRef.current = false
      clearTimeout(id)
    }
  }, [isHost])
  useEffect(() => {
    const raised = new Set(raisedIdentities)
    if (isHost && toastArmedRef.current) {
      for (const identity of raised) {
        if (!prevRaisedRef.current.has(identity)) {
          const p = participants.find((x) => x.identity === identity)
          toast(t("liveRoom.people.handRaisedToast", { name: p?.name || identity }))
        }
      }
    }
    prevRaisedRef.current = raised
  }, [raisedIdentities, isHost, participants, t])

  useEffect(() => {
    if (tab === "chat") setReadCount(chat.count)
  }, [tab, chat.count])

  const unread = Math.max(0, chat.count - readCount)

  return (
    <LayoutContextProvider value={layoutContext}>
      <RoomHeader
        sessionName={sessionName}
        className={className}
        actualStartTime={actualStartTime}
        onOpenPeople={() => setTab("people")}
      />

      <div className="flex min-h-0 flex-1">
        <div className="relative flex min-w-0 flex-1 flex-col">
          <div className="flex min-h-0 flex-1 gap-3 p-3 sm:p-4">
            {railOpen && (
              <div className="hidden md:block">
                <WebcamRail orientation="vertical" />
              </div>
            )}
            <div className="relative min-w-0 flex-1">
              <Stage
                stage={stage}
                isHost={isHost}
                liveId={liveId}
                canDraw={canDraw}
                onPageChange={onPageChange}
                onLoadNumPages={onLoadNumPages}
              />
              {slidesUpload && <SlidesUploadOverlay upload={slidesUpload} onCancel={onCancelSlidesUpload} />}
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
            raisedHandCount={raisedHandCount}
            handRaised={handRaised}
            onToggleHand={onToggleHand}
            canShareStage={canPublish(role)}
            stageKind={stage.kind}
            onShareSlides={onShareSlides}
            onStopStage={onStopStage}
            onStartWhiteboard={onStartWhiteboard}
            isRecording={isRecording}
            recordingPending={recordingPending}
            onToggleRecording={onToggleRecording}
            qaOpenCount={qa.openCount}
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
                "absolute end-4 z-20 flex size-9 items-center justify-center rounded-lg transition-colors",
                // In whiteboard mode tldraw owns the top-end corner (undo/redo/…),
                // so drop below its action row on phones; desktop has room at top.
                stage.kind === "whiteboard" ? "top-16 sm:top-4" : "top-4",
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
          onLowerHand={onLowerHand}
          onRemove={onRemove}
          polls={polls}
          onVote={onVote}
          answerPending={answerMutation.isPending}
          qa={qa}
          myId={myIdentity}
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
