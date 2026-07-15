import type { DataPublishOptions } from "livekit-client"

import { useDataChannel } from "@livekit/components-react"
import { useRef } from "react"

export interface ChannelMessage {
  payload: Uint8Array
  topic?: string
}

export type ChannelSend = (data: Uint8Array, options?: DataPublishOptions) => void

/**
 * Thin wrapper over LiveKit's `useDataChannel` that hands back a STABLE message
 * callback and a STABLE `send`.
 *
 * Why this exists: passing a fresh inline callback straight to `useDataChannel`
 * makes it rebuild its internal send/isSending observables on every render. Its
 * "isSending" observer is wired lazily on subscribe, so a packet sent during the
 * re-subscribe window hits `o.next()` on an undefined observer and throws
 * "Cannot read properties of undefined (reading 'next')" — silently dropping the
 * message. Stabilizing the callback keeps the channel built exactly once.
 *
 * The returned `send` is also stable, so effects that broadcast on
 * ParticipantConnected (late-join re-sync) don't re-subscribe every render.
 *
 * Callers may freely close over changing props/state in `onMessage`: the latest
 * closure is always invoked via a ref, without re-subscribing the channel.
 *
 * @param topic optional data-channel topic; `undefined` receives all untopiced
 *   messages (LiveKit treats an undefined topic as the default room channel).
 */
export function useRoomChannel(
  topic: string | undefined,
  onMessage: (msg: ChannelMessage) => void
): { send: ChannelSend } {
  const onMessageRef = useRef(onMessage)
  onMessageRef.current = onMessage

  // Identity-stable across renders; always dispatches to the latest onMessage.
  const handler = useRef((msg: ChannelMessage) => onMessageRef.current(msg)).current

  // `topic` is a compile-time constant per call site, so this is never a
  // conditional hook. Undefined is passed through as the default channel.
  const { send } = useDataChannel(topic as string, handler)

  const sendRef = useRef(send)
  sendRef.current = send

  const stableSend = useRef<ChannelSend>((data, options) => {
    void sendRef.current(data, options ?? {})
  }).current

  return { send: stableSend }
}
