export type WsEvent = { type: string; data: any }

type Status = "online" | "offline"

const BACKOFF_MIN = 1000
const BACKOFF_MAX = 30000

/**
 * Framework-agnostic WebSocket client for the conversations realtime hub.
 * Handles token auth, room join/leave/typing frames, and transparent
 * reconnection with exponential backoff plus room rejoin on reconnect.
 */
export class ChatWsClient {
  private ws: WebSocket | null = null
  private rooms = new Set<string>()
  private backoff = BACKOFF_MIN
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private closedByUser = false

  constructor(
    private url: string,
    private getToken: () => string | null,
    private onEvent: (e: WsEvent) => void,
    private onStatus: (s: Status) => void
  ) {}

  connect(): void {
    const token = this.getToken()
    if (!token) {
      // The token can be transiently null during a refresh window. If this isn't
      // a user-initiated close, keep the reconnect chain alive (with backoff, so
      // it can't loop-storm) so we retry once a token is available again. Without
      // this, a single null read would kill the chain permanently.
      if (!this.closedByUser) this.scheduleReconnect()
      return
    }

    this.closedByUser = false
    const ws = new WebSocket(`${this.url}?token=${encodeURIComponent(token)}`)
    this.ws = ws

    ws.onopen = () => {
      this.backoff = BACKOFF_MIN
      this.onStatus("online")
      // Rejoin every tracked room.
      for (const convId of this.rooms) {
        this.rawSend({ type: "join", conversation_id: convId })
      }
    }

    ws.onmessage = (e: MessageEvent) => {
      try {
        this.onEvent(JSON.parse(e.data))
      } catch {
        // Swallow malformed frames.
      }
    }

    ws.onclose = () => {
      this.onStatus("offline")
      if (!this.closedByUser) {
        this.scheduleReconnect()
      }
    }

    ws.onerror = () => {
      // Funnel errors into the close -> reconnect path.
      ws.close()
    }
  }

  join(convId: string): void {
    this.rooms.add(convId)
    this.rawSend({ type: "join", conversation_id: convId })
  }

  leave(convId: string): void {
    this.rooms.delete(convId)
    this.rawSend({ type: "leave", conversation_id: convId })
  }

  typing(convId: string): void {
    this.rawSend({ type: "typing", conversation_id: convId })
  }

  close(): void {
    this.closedByUser = true
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    this.ws?.close()
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return
    const delay = this.backoff
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, delay)
    this.backoff = Math.min(this.backoff * 2, BACKOFF_MAX)
  }

  private rawSend(payload: unknown): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(payload))
    }
  }
}
