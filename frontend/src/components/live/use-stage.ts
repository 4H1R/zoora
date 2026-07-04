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
  const [stage, setStageLocal] = useState<StageContent>({ kind: "none" })
  const stageRef = useRef<StageContent>({ kind: "none" })
  const requestedRef = useRef(false)

  const { send } = useRoomChannel(undefined, (msg) => {
    const event = decodeRoomEvent(msg.payload)
    if (!event) return

    if (event.type === "stage") {
      const next = event.data as StageContent
      stageRef.current = next
      setStageLocal(next)
    } else if (event.type === "request_stage" && isHost) {
      // Re-broadcast current stage state to the late-joining participant
      send(encodeRoomEvent({ type: "stage", data: stageRef.current }), { reliable: true })
    }
  })

  // Non-host: ask the host for current stage once after mount (short delay so
  // the data channel is ready). Guard with a ref so it fires exactly once.
  useEffect(() => {
    if (isHost || requestedRef.current) return
    requestedRef.current = true
    const timer = setTimeout(() => {
      send(encodeRoomEvent({ type: "request_stage", data: {} }), { reliable: true })
    }, 800)
    return () => clearTimeout(timer)
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
