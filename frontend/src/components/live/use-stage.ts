import { useRoomContext } from "@livekit/components-react"
import { RoomEvent } from "livekit-client"
import { useEffect, useRef, useState } from "react"

import { decodeRoomEvent, encodeRoomEvent } from "./room-events"
import { useRoomChannel } from "./use-room-channel"

export interface StageContent {
  kind: "none" | "slides" | "whiteboard"
  url?: string
  page?: number
  numPages?: number
}

// Tracks the shared stage content. Host mutators broadcast over the data
// channel; all clients apply incoming stage events. Late joiners send
// request_stage once after mount so the host can re-broadcast current state.
export function useStage(isHost: boolean) {
  const room = useRoomContext()
  const [stage, setStageLocal] = useState<StageContent>({ kind: "none" })
  const stageRef = useRef<StageContent>({ kind: "none" })
  const receivedRef = useRef(false)

  const { send } = useRoomChannel(undefined, (msg) => {
    const event = decodeRoomEvent(msg.payload)
    if (!event) return

    if (event.type === "stage") {
      const next = event.data as StageContent
      receivedRef.current = true
      stageRef.current = next
      setStageLocal(next)
    } else if (event.type === "request_stage" && isHost) {
      // Re-broadcast current stage state to the late-joining participant
      send(encodeRoomEvent({ type: "stage", data: stageRef.current }), { reliable: true })
    }
  })

  // Host: whenever a participant connects, re-broadcast the current stage so
  // late joiners get it without depending on their request timing. This is the
  // reliable path — the joiner's request below is a fallback.
  useEffect(() => {
    if (!isHost || !room) return
    const onConnected = () => {
      send(encodeRoomEvent({ type: "stage", data: stageRef.current }), { reliable: true })
    }
    room.on(RoomEvent.ParticipantConnected, onConnected)
    return () => {
      room.off(RoomEvent.ParticipantConnected, onConnected)
    }
  }, [isHost, room, send])

  // Non-host: ask the host for the current stage after mount, retrying until a
  // stage event arrives (the data channel may not be ready on the first try).
  useEffect(() => {
    if (isHost) return
    let attempts = 0
    const interval = setInterval(() => {
      if (receivedRef.current || attempts >= 6) {
        clearInterval(interval)
        return
      }
      attempts += 1
      send(encodeRoomEvent({ type: "request_stage", data: {} }), { reliable: true })
    }, 800)
    return () => clearInterval(interval)
  }, [isHost, send])

  // Host-only mutator: update local state AND broadcast to all participants.
  function setStage(next: StageContent) {
    if (!isHost) return
    stageRef.current = next
    setStageLocal(next)
    send(encodeRoomEvent({ type: "stage", data: next }), { reliable: true })
  }

  return { stage, setStage }
}
