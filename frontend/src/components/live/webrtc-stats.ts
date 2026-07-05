// Shared WebRTC transport-stats reader. Both the local header popover
// (use-connection-stats) and the presence publisher (use-publish-presence) parse
// the same getStats() reports, so the RTCStats quirks live in exactly one place.
// getStats() only ever reports the LOCAL peer connections — a remote peer's own
// uplink RTT/jitter/loss is unknowable, which is why each client self-publishes.

export type StatsSource = { getStats(): Promise<RTCStatsReport> }

export interface TransportSample {
  bytesReceived: number
  bytesSent: number
  packetsReceived: number
  packetsLost: number
  timestamp: number
}

export interface TransportRead extends TransportSample {
  /** Round-trip time in ms from the nominated ICE candidate pair */
  rtt: number | null
  /** Jitter in ms, max across inbound-rtp reports */
  jitter: number | null
}

/** Aggregate transport stats across one or more peer connections. */
export async function readTransportStats(pcs: StatsSource[]): Promise<TransportRead> {
  let rtt: number | null = null
  let maxJitter = 0
  let hasJitter = false
  const agg: TransportRead = {
    rtt: null,
    jitter: null,
    bytesReceived: 0,
    bytesSent: 0,
    packetsReceived: 0,
    packetsLost: 0,
    timestamp: 0,
  }

  for (const pc of pcs) {
    let report: RTCStatsReport
    try {
      report = await pc.getStats()
    } catch {
      continue
    }
    report.forEach((r) => {
      const s = r as Record<string, unknown>
      if (
        s.type === "candidate-pair" &&
        (s.nominated === true || s.selected === true) &&
        s.state === "succeeded" &&
        typeof s.currentRoundTripTime === "number"
      ) {
        const ms = Math.round(s.currentRoundTripTime * 1000)
        rtt = rtt == null ? ms : Math.min(rtt, ms)
      }
      if (s.type === "inbound-rtp") {
        if (typeof s.jitter === "number") {
          maxJitter = Math.max(maxJitter, s.jitter)
          hasJitter = true
        }
        if (typeof s.bytesReceived === "number") agg.bytesReceived += s.bytesReceived
        if (typeof s.packetsReceived === "number") agg.packetsReceived += s.packetsReceived
        if (typeof s.packetsLost === "number") agg.packetsLost += s.packetsLost
        if (typeof s.timestamp === "number") agg.timestamp = s.timestamp
      }
      if (s.type === "outbound-rtp" && typeof s.bytesSent === "number") {
        agg.bytesSent += s.bytesSent
        if (typeof s.timestamp === "number") agg.timestamp = s.timestamp
      }
    })
  }

  agg.rtt = rtt
  agg.jitter = hasJitter ? Math.round(maxJitter * 1000) : null
  return agg
}

export interface TransportRates {
  downKbps: number | null
  upKbps: number | null
  packetLoss: number | null
}

/** Derive bitrates and loss from the delta between two samples. */
export function deriveRates(prev: TransportSample | null, cur: TransportSample): TransportRates {
  if (!prev || cur.timestamp <= prev.timestamp) {
    return { downKbps: null, upKbps: null, packetLoss: null }
  }
  const dtSec = (cur.timestamp - prev.timestamp) / 1000
  const kbps = (bytes: number) => Math.max(0, Math.round((bytes * 8) / dtSec / 1000))
  const dRecv = cur.packetsReceived - prev.packetsReceived
  const dLost = cur.packetsLost - prev.packetsLost
  const total = dRecv + dLost
  return {
    downKbps: kbps(cur.bytesReceived - prev.bytesReceived),
    upKbps: kbps(cur.bytesSent - prev.bytesSent),
    packetLoss: total > 0 ? Math.max(0, Math.min(100, (dLost / total) * 100)) : 0,
  }
}

/** The two local peer connections, if the room engine is up. */
export function transportSources(pcm: { subscriber?: unknown; publisher?: unknown } | undefined): StatsSource[] {
  return [pcm?.subscriber, pcm?.publisher].filter(Boolean) as StatsSource[]
}
