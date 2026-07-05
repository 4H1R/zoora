import { useConnectionQualityIndicator, useLocalParticipant, useRoomContext } from "@livekit/components-react"
import { useEffect, useRef, useState } from "react"

import type { NetStats } from "./presence"
import { deriveRates, readTransportStats, transportSources, type TransportSample } from "./webrtc-stats"

const POLL_MS = 2000

type LocalStats = Omit<NetStats, "quality">

const EMPTY: LocalStats = { rtt: null, jitter: null, packetLoss: null, downKbps: null, upKbps: null }

/**
 * Polls WebRTC transport stats from the local peer connections so the header can
 * surface live connection health (ping, jitter, loss, down/up bitrate). Quality
 * comes from LiveKit's own indicator for the local participant.
 */
export function useConnectionStats(): NetStats {
  const room = useRoomContext()
  // Scope to the local participant explicitly: RoomHeader renders outside any
  // ParticipantContext, so without this useEnsureParticipant() throws.
  const { localParticipant } = useLocalParticipant()
  const { quality } = useConnectionQualityIndicator({ participant: localParticipant })

  const [stats, setStats] = useState<LocalStats>(EMPTY)
  const prev = useRef<TransportSample | null>(null)

  useEffect(() => {
    let active = true

    const read = async () => {
      const pcs = transportSources(room.engine?.pcManager)
      if (!pcs.length) return
      const cur = await readTransportStats(pcs)
      if (!active) return
      const rates = deriveRates(prev.current, cur)
      prev.current = cur
      setStats({ rtt: cur.rtt, jitter: cur.jitter, ...rates })
    }

    void read()
    const id = setInterval(read, POLL_MS)
    return () => {
      active = false
      clearInterval(id)
    }
  }, [room])

  return { quality, ...stats }
}
