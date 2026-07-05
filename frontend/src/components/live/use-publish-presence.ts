import { useLocalParticipant, useRoomContext } from "@livekit/components-react"
import { ConnectionState } from "livekit-client"
import { useEffect, useRef } from "react"

import { detectDevice, PRESENCE_KEYS, serializeNet, type NetStats } from "./presence"
import { deriveRates, readTransportStats, transportSources, type TransportSample } from "./webrtc-stats"

const PUBLISH_MS = 5000

/**
 * Publishes this client's device/OS/browser (once) and live network stats
 * (every few seconds) into its LiveKit participant attributes, so a host can
 * inspect any participant by reading `participant.attributes`. Each client must
 * self-publish because a remote's uplink stats are unknowable from the host side.
 */
export function usePublishPresence(): void {
  const room = useRoomContext()
  const { localParticipant } = useLocalParticipant()
  const prev = useRef<TransportSample | null>(null)

  useEffect(() => {
    let active = true
    const device = detectDevice()
    let deviceSent = false

    const publish = async () => {
      if (!active || room.state !== ConnectionState.Connected) return

      const cur = await readTransportStats(transportSources(room.engine?.pcManager))
      if (!active) return
      const rates = deriveRates(prev.current, cur)
      prev.current = cur

      const net: NetStats = {
        quality: localParticipant.connectionQuality,
        rtt: cur.rtt,
        jitter: cur.jitter,
        ...rates,
      }

      const attrs: Record<string, string> = { [PRESENCE_KEYS.net]: serializeNet(net) }
      if (!deviceSent) {
        attrs[PRESENCE_KEYS.device] = device.device
        attrs[PRESENCE_KEYS.os] = device.os
        attrs[PRESENCE_KEYS.browser] = device.browser
      }

      try {
        await localParticipant.setAttributes(attrs)
        deviceSent = true
      } catch {
        // transient — retried on the next tick, device attrs still pending
      }
    }

    void publish()
    const id = setInterval(() => void publish(), PUBLISH_MS)
    return () => {
      active = false
      clearInterval(id)
    }
  }, [room, localParticipant])
}
