import {
  LayoutContextProvider,
  LiveKitRoom,
  RoomAudioRenderer,
  useChat,
  useCreateLayoutContext,
} from "@livekit/components-react"
import { useEffect, useState } from "react"

import "@livekit/components-styles"
import "./livekit-overrides.css"

import { usePostLiveRoomsIdLeave } from "@/api/live-sessions/live-sessions"

import { ControlBar } from "./control-bar"
import { RoomHeader } from "./room-header"
import { SidePanel } from "./side-panel"
import type { PreJoinChoices, SidePanelTab } from "./types"
import { VideoGrid } from "./video-grid"

interface ActiveRoomProps {
  token: string
  serverUrl: string
  choices: PreJoinChoices
  sessionName: string
  className?: string
  liveId: string
  onDisconnect: () => void
}

export function ActiveRoom({
  token,
  serverUrl,
  choices,
  sessionName,
  className,
  liveId,
  onDisconnect,
}: ActiveRoomProps) {
  const leaveMutation = usePostLiveRoomsIdLeave()

  const handleLeave = () => {
    leaveMutation.mutate({ id: liveId }, { onSettled: onDisconnect })
  }

  return (
    <LiveKitRoom
      serverUrl={serverUrl}
      token={token}
      audio={choices.audioEnabled ? { deviceId: choices.audioDeviceId } : false}
      video={choices.videoEnabled ? { deviceId: choices.videoDeviceId } : false}
      onDisconnected={onDisconnect}
      data-lk-theme="default"
      className="zoora-live relative flex flex-col overflow-hidden bg-zinc-950 text-zinc-100"
    >
      <RoomShell
        sessionName={sessionName}
        className={className}
        onLeave={handleLeave}
        leavePending={leaveMutation.isPending}
      />
      <RoomAudioRenderer />
    </LiveKitRoom>
  )
}

function RoomShell({
  sessionName,
  className,
  onLeave,
  leavePending,
}: {
  sessionName: string
  className?: string
  onLeave: () => void
  leavePending: boolean
}) {
  const layoutContext = useCreateLayoutContext()
  // Chat lives here (always mounted) so messages accumulate even while the
  // chat panel is closed; the panel just reads from this shared state.
  const chat = useChat()
  const [panel, setPanel] = useState<SidePanelTab | null>(null)
  const [readCount, setReadCount] = useState(0)

  useEffect(() => {
    if (panel === "chat") setReadCount(chat.chatMessages.length)
  }, [panel, chat.chatMessages.length])

  const unread = Math.max(0, chat.chatMessages.length - readCount)

  return (
    <LayoutContextProvider value={layoutContext}>
      <RoomHeader sessionName={sessionName} className={className} />

      <div className="flex min-h-0 flex-1">
        <div className="relative min-w-0 flex-1">
          <div className="absolute inset-0 p-3 sm:p-4">
            <VideoGrid layoutContext={layoutContext} />
          </div>
          <ControlBar
            panel={panel}
            setPanel={setPanel}
            onLeave={onLeave}
            leavePending={leavePending}
            unread={unread}
          />
        </div>

        {panel && (
          <div className="hidden h-full sm:block">
            <SidePanel tab={panel} setTab={setPanel} onClose={() => setPanel(null)} chat={chat} />
          </div>
        )}
      </div>

      {/* Mobile: overlay panel */}
      {panel && (
        <div className="absolute inset-0 z-30 bg-zinc-950/95 backdrop-blur-sm sm:hidden">
          <SidePanel tab={panel} setTab={setPanel} onClose={() => setPanel(null)} chat={chat} />
        </div>
      )}
    </LayoutContextProvider>
  )
}
