export interface PreJoinChoices {
  // Phase 1: no device selection in the lobby. Publishers (host/presenter)
  // start muted and pick devices in-room. Kept as a struct for forward-compat.
  audioEnabled: boolean
  videoEnabled: boolean
  audioDeviceId?: string
  videoDeviceId?: string
}

export type RoomTab = "chat" | "people" | "polls"
