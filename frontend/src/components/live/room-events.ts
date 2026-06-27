// Typed envelope for the LiveKit data channel. The Go backend broadcasts the
// same shape ({ type, data }) via SendData for role/hand UI notifications.
export type RoomEvent =
  | { type: "role_changed"; data: { identity: string; role: "host" | "presenter" | "viewer" } }
  | { type: "hand"; data: { identity: string; raised: boolean } }
  | { type: "stage"; data: { kind: "none" | "slides"; url?: string; page?: number; numPages?: number } }
  | { type: "request_stage"; data: Record<string, never> }

const encoder = new TextEncoder()
const decoder = new TextDecoder()

export function encodeRoomEvent(event: RoomEvent): Uint8Array {
  return encoder.encode(JSON.stringify(event))
}

export function decodeRoomEvent(payload: Uint8Array): RoomEvent | null {
  try {
    const parsed = JSON.parse(decoder.decode(payload))
    if (parsed && typeof parsed.type === "string") return parsed as RoomEvent
  } catch {
    // non-JSON or unknown packet (e.g. LiveKit chat) — ignore
  }
  return null
}
