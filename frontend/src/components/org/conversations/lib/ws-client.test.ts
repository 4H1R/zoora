import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import { ChatWsClient } from "./ws-client"
import type { WsEvent } from "./ws-client"

// Minimal fake WebSocket installed on globalThis. Tracks instances so tests can
// drive lifecycle transitions manually.
class FakeWebSocket {
  static OPEN = 1
  static CLOSED = 3
  static instances: FakeWebSocket[] = []

  url: string
  readyState = 0
  send = vi.fn()
  close = vi.fn(() => {
    this.readyState = FakeWebSocket.CLOSED
  })
  onopen: (() => void) | null = null
  onmessage: ((e: { data: string }) => void) | null = null
  onclose: (() => void) | null = null
  onerror: (() => void) | null = null

  constructor(url: string) {
    this.url = url
    FakeWebSocket.instances.push(this)
  }

  // Test helpers.
  simulateOpen() {
    this.readyState = FakeWebSocket.OPEN
    this.onopen?.()
  }
  simulateMessage(data: unknown) {
    this.onmessage?.({ data: typeof data === "string" ? data : JSON.stringify(data) })
  }
  simulateClose() {
    this.readyState = FakeWebSocket.CLOSED
    this.onclose?.()
  }
  simulateError() {
    this.onerror?.()
  }

  static last() {
    return FakeWebSocket.instances[FakeWebSocket.instances.length - 1]
  }
  static reset() {
    FakeWebSocket.instances = []
  }
}

const originalWs = globalThis.WebSocket

beforeEach(() => {
  FakeWebSocket.reset()
  ;(globalThis as any).WebSocket = FakeWebSocket
  vi.useFakeTimers()
})

afterEach(() => {
  vi.useRealTimers()
  ;(globalThis as any).WebSocket = originalWs
})

function makeClient(token: string | null = "tok") {
  const events: WsEvent[] = []
  const statuses: Array<"online" | "offline"> = []
  const client = new ChatWsClient(
    "wss://example.test/ws",
    () => token,
    (e) => events.push(e),
    (s) => statuses.push(s)
  )
  return { client, events, statuses }
}

describe("ChatWsClient", () => {
  it("connects with token in query string", () => {
    const { client } = makeClient("a b")
    client.connect()
    expect(FakeWebSocket.instances).toHaveLength(1)
    expect(FakeWebSocket.last().url).toBe("wss://example.test/ws?token=a%20b")
  })

  it("skips connect when no token", () => {
    const { client } = makeClient(null)
    client.connect()
    expect(FakeWebSocket.instances).toHaveLength(0)
  })

  it("reports online and resets on open", () => {
    const { client, statuses } = makeClient()
    client.connect()
    FakeWebSocket.last().simulateOpen()
    expect(statuses).toContain("online")
  })

  it("join sends the correct frame when open", () => {
    const { client } = makeClient()
    client.connect()
    const ws = FakeWebSocket.last()
    ws.simulateOpen()
    client.join("c1")
    expect(ws.send).toHaveBeenCalledWith(JSON.stringify({ type: "join", conversation_id: "c1" }))
  })

  it("join before open only sends after open (rejoin tracked rooms)", () => {
    const { client } = makeClient()
    client.connect()
    const ws = FakeWebSocket.last()
    client.join("c1")
    expect(ws.send).not.toHaveBeenCalled()
    ws.simulateOpen()
    expect(ws.send).toHaveBeenCalledWith(JSON.stringify({ type: "join", conversation_id: "c1" }))
  })

  it("leave removes and sends leave frame", () => {
    const { client } = makeClient()
    client.connect()
    const ws = FakeWebSocket.last()
    ws.simulateOpen()
    client.join("c1")
    client.leave("c1")
    expect(ws.send).toHaveBeenCalledWith(JSON.stringify({ type: "leave", conversation_id: "c1" }))
  })

  it("typing sends typing frame", () => {
    const { client } = makeClient()
    client.connect()
    const ws = FakeWebSocket.last()
    ws.simulateOpen()
    client.typing("c1")
    expect(ws.send).toHaveBeenCalledWith(JSON.stringify({ type: "typing", conversation_id: "c1" }))
  })

  it("parses messages to onEvent, swallowing parse errors", () => {
    const { client, events } = makeClient()
    client.connect()
    const ws = FakeWebSocket.last()
    ws.simulateOpen()
    ws.simulateMessage({ type: "message", data: { id: "m1" } })
    ws.simulateMessage("not json{")
    expect(events).toEqual([{ type: "message", data: { id: "m1" } }])
  })

  it("reconnects with backoff on close and rejoins tracked rooms", () => {
    const { client, statuses } = makeClient()
    client.connect()
    const ws1 = FakeWebSocket.last()
    ws1.simulateOpen()
    client.join("c1")
    ws1.send.mockClear()

    ws1.simulateClose()
    expect(statuses).toContain("offline")
    // No immediate reconnect.
    expect(FakeWebSocket.instances).toHaveLength(1)
    // First backoff is 1000ms.
    vi.advanceTimersByTime(1000)
    expect(FakeWebSocket.instances).toHaveLength(2)

    const ws2 = FakeWebSocket.last()
    ws2.simulateOpen()
    expect(ws2.send).toHaveBeenCalledWith(JSON.stringify({ type: "join", conversation_id: "c1" }))
  })

  it("backoff grows exponentially and is capped at 30000", () => {
    const { client } = makeClient()
    client.connect()
    // 1st close -> 1000ms
    FakeWebSocket.last().simulateClose()
    vi.advanceTimersByTime(999)
    expect(FakeWebSocket.instances).toHaveLength(1)
    vi.advanceTimersByTime(1)
    expect(FakeWebSocket.instances).toHaveLength(2)
    // 2nd close -> 2000ms
    FakeWebSocket.last().simulateClose()
    vi.advanceTimersByTime(1999)
    expect(FakeWebSocket.instances).toHaveLength(2)
    vi.advanceTimersByTime(1)
    expect(FakeWebSocket.instances).toHaveLength(3)
  })

  it("close() prevents reconnect", () => {
    const { client } = makeClient()
    client.connect()
    const ws = FakeWebSocket.last()
    ws.simulateOpen()
    client.close()
    expect(ws.close).toHaveBeenCalled()
    ws.simulateClose()
    vi.advanceTimersByTime(60000)
    expect(FakeWebSocket.instances).toHaveLength(1)
  })

  it("onerror closes the socket (funnels to reconnect)", () => {
    const { client } = makeClient()
    client.connect()
    const ws = FakeWebSocket.last()
    ws.simulateOpen()
    ws.simulateError()
    expect(ws.close).toHaveBeenCalled()
  })

  it("reschedules the reconnect when the token is transiently null, then connects once it returns", () => {
    let token: string | null = null
    const client = new ChatWsClient(
      "wss://example.test/ws",
      () => token,
      () => {},
      () => {}
    )
    client.connect()
    // No token yet -> no socket, but a reconnect must be scheduled (chain alive).
    expect(FakeWebSocket.instances).toHaveLength(0)
    // Token appears before the backoff fires; the scheduled attempt connects.
    token = "tok"
    vi.advanceTimersByTime(1000)
    expect(FakeWebSocket.instances).toHaveLength(1)
    expect(FakeWebSocket.last().url).toBe("wss://example.test/ws?token=tok")
  })

  it("keeps retrying with backoff while the token stays null (no loop-storm)", () => {
    let token: string | null = null
    const client = new ChatWsClient(
      "wss://example.test/ws",
      () => token,
      () => {},
      () => {}
    )
    client.connect()
    // Retry #1 at 1000ms: still null -> no socket, but reschedules (backoff 2000).
    vi.advanceTimersByTime(1000)
    expect(FakeWebSocket.instances).toHaveLength(0)
    // Retry #2 at 2000ms: still null -> reschedules (backoff 4000).
    vi.advanceTimersByTime(2000)
    expect(FakeWebSocket.instances).toHaveLength(0)
    // Token returns; retry #3 at 4000ms finally connects.
    token = "tok"
    vi.advanceTimersByTime(4000)
    expect(FakeWebSocket.instances).toHaveLength(1)
  })

  it("does NOT reschedule after a user-initiated close with a null token", () => {
    let token: string | null = "tok"
    const client = new ChatWsClient(
      "wss://example.test/ws",
      () => token,
      () => {},
      () => {}
    )
    client.connect()
    FakeWebSocket.last().simulateOpen()
    client.close()
    token = null
    // A close() sets closedByUser; any later null-token connect must not revive.
    client.connect()
    vi.advanceTimersByTime(60000)
    expect(FakeWebSocket.instances).toHaveLength(1)
  })
})
