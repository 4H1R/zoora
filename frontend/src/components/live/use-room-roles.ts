import { useEffect, useRef, useState } from "react"
import { useLocalParticipant, useRoomContext } from "@livekit/components-react"
import { RoomEvent } from "livekit-client"

import { decodeRoomEvent, encodeRoomEvent } from "./room-events"
import { useRoomChannel } from "./use-room-channel"
import type { RoomRole } from "./room-role"

export interface ParticipantState {
  role: RoomRole
  handRaised: boolean
  handRaisedAt?: number
}

// Tracks live role + raised-hand per participant identity. Seeded from a
// snapshot, then kept current by data-channel events from the backend.
export function useRoomRoles(seed: Record<string, ParticipantState>) {
  const [states, setStates] = useState<Record<string, ParticipantState>>(seed)
  const room = useRoomContext()
  const { localParticipant } = useLocalParticipant()
  const myIdentity = localParticipant.identity

  // Latest local hand state for the participant-connected re-announce (avoids
  // re-subscribing the connect listener on every state change).
  const myHandRef = useRef<{ raised: boolean; raisedAt?: number }>({ raised: false })

  const { send } = useRoomChannel(undefined, (msg) => {
    const event = decodeRoomEvent(msg.payload)
    if (!event) return
    if (event.type === "role_changed") {
      setStates((prev) => ({
        ...prev,
        [event.data.identity]: {
          role: event.data.role,
          handRaised: prev[event.data.identity]?.handRaised ?? false,
          handRaisedAt: prev[event.data.identity]?.handRaisedAt,
        },
      }))
    } else if (event.type === "hand") {
      setStates((prev) => ({
        ...prev,
        [event.data.identity]: {
          role: prev[event.data.identity]?.role ?? "viewer",
          handRaised: event.data.raised,
          handRaisedAt: event.data.raised ? event.data.raisedAt : undefined,
        },
      }))
    }
  })

  // Keep the ref current with the local participant's own hand state.
  const mine = states[myIdentity]
  myHandRef.current = { raised: mine?.handRaised ?? false, raisedAt: mine?.handRaisedAt }

  // Late-join sync: `hand` events are ephemeral, so a participant who joins,
  // reconnects, or whose panel remounts misses hands already up. When a NEW
  // participant connects, every client with its own hand raised re-broadcasts
  // that fact (with the persisted raisedAt, so ordering survives). Fan-out is
  // O(hands-up), not O(participants).
  useEffect(() => {
    if (!room) return
    const onConnected = () => {
      const h = myHandRef.current
      if (!h.raised) return
      send(
        encodeRoomEvent({
          type: "hand",
          data: { identity: myIdentity, raised: true, raisedAt: h.raisedAt },
        }),
        { reliable: true },
      )
    }
    room.on(RoomEvent.ParticipantConnected, onConnected)
    return () => {
      room.off(RoomEvent.ParticipantConnected, onConnected)
    }
  }, [room, send, myIdentity])

  return states
}
