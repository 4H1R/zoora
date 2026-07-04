import {
  useConnectionQualityIndicator,
  useLocalParticipant,
  useRoomContext,
} from "@livekit/components-react"
import { ConnectionQuality } from "livekit-client"
import { useEffect, useRef, useState } from "react"

export interface ConnectionStats {
  quality: ConnectionQuality
  /** Round-trip time in milliseconds (ping), null until measured */
  rtt: number | null
  /** Jitter in milliseconds */
  jitter: number | null
  /** Packet loss over the last sample window, as a percentage (0-100) */
  packetLoss: number | null
  /** Inbound bitrate in kilobits per second */
  downKbps: number | null
}

interface Sample {
  bytesReceived: number
  packetsReceived: number
  packetsLost: number
  timestamp: number
}

const POLL_MS = 2000

/**
 * Polls WebRTC transport stats from the subscriber peer connection so the UI can
 * surface live connection health (ping, jitter, loss, bitrate). RTT comes from the
 * nominated ICE candidate pair; the rest is aggregated across inbound-rtp reports.
 */
export function useConnectionStats(): ConnectionStats {
  const room = useRoomContext()
  // Scope to the local participant explicitly: RoomHeader renders outside any
  // ParticipantContext, so without this the hook's useEnsureParticipant() throws
  // "No participant provided" and crashes the whole room to the error boundary.
  const { localParticipant } = useLocalParticipant()
  const { quality } = useConnectionQualityIndicator({ participant: localParticipant })

  const [rtt, setRtt] = useState<number | null>(null)
  const [jitter, setJitter] = useState<number | null>(null)
  const [packetLoss, setPacketLoss] = useState<number | null>(null)
  const [downKbps, setDownKbps] = useState<number | null>(null)

  const prev = useRef<Sample | null>(null)

  useEffect(() => {
    let active = true

    const read = async () => {
      // subscriber carries the media we receive; fall back to publisher for hosts.
      const pc = room.engine?.pcManager?.subscriber ?? room.engine?.pcManager?.publisher
      if (!pc) return

      let report: RTCStatsReport
      try {
        report = await pc.getStats()
      } catch {
        return
      }
      if (!active) return

      let nextRtt: number | null = null
      let maxJitter = 0
      let hasJitter = false
      const agg: Sample = { bytesReceived: 0, packetsReceived: 0, packetsLost: 0, timestamp: 0 }

      report.forEach((r) => {
        const s = r as Record<string, unknown>
        if (
          s.type === "candidate-pair" &&
          (s.nominated === true || s.selected === true) &&
          s.state === "succeeded" &&
          typeof s.currentRoundTripTime === "number"
        ) {
          nextRtt = Math.round(s.currentRoundTripTime * 1000)
        }
        if (s.type === "inbound-rtp") {
          if (typeof s.jitter === "number") {
            maxJitter = Math.max(maxJitter, s.jitter)
            hasJitter = true
          }
          if (typeof s.bytesReceived === "number") agg.bytesReceived += s.bytesReceived
          if (typeof s.packetsReceived === "number") agg.packetsReceived += s.packetsReceived
          if (typeof s.packetsLost === "number") agg.packetsLost += s.packetsLost
          agg.timestamp = typeof s.timestamp === "number" ? s.timestamp : agg.timestamp
        }
      })

      setRtt(nextRtt)
      setJitter(hasJitter ? Math.round(maxJitter * 1000) : null)

      const last = prev.current
      if (last && agg.timestamp > last.timestamp) {
        const dtSec = (agg.timestamp - last.timestamp) / 1000
        const dBytes = agg.bytesReceived - last.bytesReceived
        setDownKbps(dtSec > 0 ? Math.max(0, Math.round((dBytes * 8) / dtSec / 1000)) : null)

        const dRecv = agg.packetsReceived - last.packetsReceived
        const dLost = agg.packetsLost - last.packetsLost
        const total = dRecv + dLost
        setPacketLoss(total > 0 ? Math.max(0, Math.min(100, (dLost / total) * 100)) : 0)
      }
      prev.current = agg
    }

    void read()
    const id = setInterval(read, POLL_MS)
    return () => {
      active = false
      clearInterval(id)
    }
  }, [room])

  return { quality, rtt, jitter, packetLoss, downKbps }
}
