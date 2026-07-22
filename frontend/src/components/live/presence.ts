import type { DeviceInfo, DeviceType } from "@/lib/device"

import { ConnectionQuality } from "livekit-client"

import { detectDevice } from "@/lib/device"

// Device/OS/browser and live network stats are not exposed by the LiveKit server
// SDK or by remote Participant objects, so each client detects its own and
// publishes them into its participant attributes. Every other client (the host)
// then reads them straight off `participant.attributes` — no backend round-trip.
// The device detector itself lives in `@/lib/device` (shared with the quiz
// take-flow); re-exported here so existing live-room imports keep working.
export const PRESENCE_KEYS = {
  device: "z_device",
  os: "z_os",
  browser: "z_browser",
  net: "z_net",
} as const

export { detectDevice }
export type { DeviceInfo, DeviceType }

export interface NetStats {
  quality: ConnectionQuality
  /** Round-trip time in ms */
  rtt: number | null
  /** Jitter in ms */
  jitter: number | null
  /** Packet loss over the last window, percent (0-100) */
  packetLoss: number | null
  /** Inbound bitrate in kbps (what this participant is receiving) */
  downKbps: number | null
  /** Outbound bitrate in kbps (what this participant is sending) */
  upKbps: number | null
}


// Compact wire format for the `z_net` attribute — short keys keep the attribute
// small since it is republished every few seconds.
interface NetWire {
  q: ConnectionQuality
  rtt: number | null
  jit: number | null
  loss: number | null
  down: number | null
  up: number | null
}

export function serializeNet(stats: NetStats): string {
  const wire: NetWire = {
    q: stats.quality,
    rtt: stats.rtt,
    jit: stats.jitter,
    loss: stats.packetLoss,
    down: stats.downKbps,
    up: stats.upKbps,
  }
  return JSON.stringify(wire)
}

export function parseNet(raw: string | undefined): NetStats | null {
  if (!raw) return null
  try {
    const w = JSON.parse(raw) as Partial<NetWire>
    return {
      quality: (w.q as ConnectionQuality) ?? ConnectionQuality.Unknown,
      rtt: w.rtt ?? null,
      jitter: w.jit ?? null,
      packetLoss: w.loss ?? null,
      downKbps: w.down ?? null,
      upKbps: w.up ?? null,
    }
  } catch {
    return null
  }
}

export interface PresenceInfo {
  device: DeviceInfo | null
  net: NetStats | null
}

/** Read published presence off a participant's attributes map. */
export function readPresence(attributes: Record<string, string> | undefined): PresenceInfo {
  const attrs = attributes ?? {}
  const device = attrs[PRESENCE_KEYS.device]
    ? {
        device: attrs[PRESENCE_KEYS.device] as DeviceType,
        os: attrs[PRESENCE_KEYS.os] ?? "Unknown",
        browser: attrs[PRESENCE_KEYS.browser] ?? "Unknown",
      }
    : null
  return { device, net: parseNet(attrs[PRESENCE_KEYS.net]) }
}
