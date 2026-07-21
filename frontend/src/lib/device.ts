// Best-effort device/OS/browser detection from the current client's user-agent.
// Shared by the live room (participant presence) and the quiz take-flow (device
// snapshot recorded on submit) so both surface identical labels. Hand-rolled —
// no UA-parsing dependency. Client-side only: touches `navigator`, so call it
// from an event handler / effect, never at module or SSR render time.

export type DeviceType = "mobile" | "tablet" | "desktop"

export interface DeviceInfo {
  device: DeviceType
  os: string
  browser: string
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
