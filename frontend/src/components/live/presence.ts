import { ConnectionQuality } from "livekit-client"

// Device/OS/browser and live network stats are not exposed by the LiveKit server
// SDK or by remote Participant objects, so each client detects its own and
// publishes them into its participant attributes. Every other client (the host)
// then reads them straight off `participant.attributes` — no backend round-trip.
export const PRESENCE_KEYS = {
  device: "z_device",
  os: "z_os",
  browser: "z_browser",
  net: "z_net",
} as const

export type DeviceType = "mobile" | "tablet" | "desktop"

export interface DeviceInfo {
  device: DeviceType
  os: string
  browser: string
}

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

interface NavigatorUAData {
  mobile?: boolean
  platform?: string
}

function match(ua: string, re: RegExp): string | null {
  const m = ua.match(re)
  return m ? m.slice(1).filter(Boolean).join(".") : null
}

function detectType(ua: string, uaMobile?: boolean): DeviceType {
  const touch = navigator.maxTouchPoints ?? 0
  // Modern iPadOS Safari masquerades as desktop macOS but reports touch points.
  if (/iPad/.test(ua) || (/Macintosh/.test(ua) && touch > 1)) return "tablet"
  if (/Tablet/.test(ua) || (/Android/.test(ua) && !/Mobile/.test(ua))) return "tablet"
  if (uaMobile === true || /Mobi|iPhone|iPod|Android.*Mobile/.test(ua)) return "mobile"
  return "desktop"
}

function detectOs(ua: string, platform?: string): string {
  const touch = navigator.maxTouchPoints ?? 0
  if (/iPhone|iPad|iPod/.test(ua) || (/Macintosh/.test(ua) && touch > 1)) {
    const v = match(ua, /OS (\d+)[._](\d+)/)
    return v ? `iOS ${v}` : "iOS"
  }
  if (/Android/.test(ua)) {
    const v = match(ua, /Android (\d+(?:\.\d+)?)/)
    return v ? `Android ${v}` : "Android"
  }
  if (/Windows NT/.test(ua)) {
    const v = match(ua, /Windows NT (\d+\.\d+)/)
    const names: Record<string, string> = { "10.0": "10/11", "6.3": "8.1", "6.2": "8", "6.1": "7" }
    return v ? `Windows ${names[v] ?? v}` : "Windows"
  }
  if (/Mac OS X/.test(ua)) {
    const v = match(ua, /Mac OS X (\d+)[._](\d+)/)
    return v ? `macOS ${v}` : "macOS"
  }
  if (/Linux/.test(ua)) return "Linux"
  return platform || "Unknown"
}

function detectBrowser(ua: string): string {
  let v: string | null
  if ((v = match(ua, /Edg\/(\d+)/))) return `Edge ${v}`
  if ((v = match(ua, /OPR\/(\d+)/))) return `Opera ${v}`
  if ((v = match(ua, /Firefox\/(\d+)/))) return `Firefox ${v}`
  if (/Chrome\//.test(ua) && (v = match(ua, /Chrome\/(\d+)/))) return `Chrome ${v}`
  if (/Safari\//.test(ua) && (v = match(ua, /Version\/(\d+)/))) return `Safari ${v}`
  return "Unknown"
}

/** Best-effort detection of the current client's device, OS and browser. */
export function detectDevice(): DeviceInfo {
  const nav = navigator as Navigator & { userAgentData?: NavigatorUAData }
  const ua = navigator.userAgent
  const uaData = nav.userAgentData
  return {
    device: detectType(ua, uaData?.mobile),
    os: detectOs(ua, uaData?.platform),
    browser: detectBrowser(ua),
  }
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
