import { useDataChannel } from "@livekit/components-react"
import { useState } from "react"

import { decodeRoomEvent } from "./room-events"
import type { RoomRole } from "./room-role"

export interface ParticipantState {
  role: RoomRole
  handRaised: boolean
}

// Tracks live role + raised-hand per participant identity. Seeded from a
// snapshot, then kept current by data-channel events from the backend.
export function useRoomRoles(seed: Record<string, ParticipantState>) {
  const [states, setStates] = useState<Record<string, ParticipantState>>(seed)

  useDataChannel((msg) => {
    const event = decodeRoomEvent(msg.payload)
    if (!event) return
    if (event.type === "role_changed") {
      setStates((prev) => ({
        ...prev,
        [event.data.identity]: {
          role: event.data.role,
          handRaised: prev[event.data.identity]?.handRaised ?? false,
        },
      }))
    } else if (event.type === "hand") {
      setStates((prev) => ({
        ...prev,
        [event.data.identity]: {
          role: prev[event.data.identity]?.role ?? "viewer",
          handRaised: event.data.raised,
        },
      }))
    }
  })

  return states
}
