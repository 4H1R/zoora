import { createContext, useContext } from "react"

import type { GithubCom4H1RZooraInternalDomainUser } from "@/api/model"
import { userHasAny } from "@/lib/access"

// Phase 1: role is fixed at join. Host == can manage the live session.
// "Presenter" exists in the type from day one (Phase 2 adds live promotion),
// but in Phase 1 only Host and Viewer are ever assigned.
export type RoomRole = "host" | "presenter" | "viewer"

export function deriveRoomRole(
  me: GithubCom4H1RZooraInternalDomainUser | undefined
): RoomRole {
  // The /live route renders outside the org AccessProvider, so read perms
  // straight off /users/me (same approach the lobby already uses).
  const isHost = userHasAny(me, [
    "live_sessions:manage",
    "live_sessions:manage_any",
    "live_sessions:create",
  ])
  return isHost ? "host" : "viewer"
}

// A publisher may put media on the stage / appear in the webcam rail.
export function canPublish(role: RoomRole): boolean {
  return role === "host" || role === "presenter"
}

export const RoomRoleContext = createContext<RoomRole>("viewer")

export function useRoomRole(): RoomRole {
  return useContext(RoomRoleContext)
}
