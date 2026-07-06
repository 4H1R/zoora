// Typed envelope for the LiveKit data channel. The Go backend broadcasts the
// same shape ({ type, data }) via SendData for role/hand UI notifications.
export type RoomEvent =
  | { type: "role_changed"; data: { identity: string; role: "host" | "presenter" | "viewer" } }
  | { type: "hand"; data: { identity: string; raised: boolean; raisedAt?: number } }
  | { type: "stage"; data: { kind: "none" | "slides" | "whiteboard"; url?: string; page?: number; numPages?: number } }
  | { type: "request_stage"; data: Record<string, never> }
  | { type: "poll_launched"; data: { pollId: string; name: string; options: { label: string; value: string }[]; allowedAnswersCount: number } }
  | { type: "poll_results"; data: { pollId: string; counts: Record<string, number>; total: number } }
  | { type: "poll_closed"; data: { pollId: string } }
  | {
      type: "chat_message"
      data: {
        id: string
        chat_id: string
        sender_id: string
        sender: { id: string; name: string }
        message_type: string
        content: string
        created_at: string
        parent_message_id: string | null
      }
    }
  | { type: "chat_message_deleted"; data: { id: string } }

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
