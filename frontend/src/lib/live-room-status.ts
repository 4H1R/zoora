import type { GithubCom4H1RZooraInternalDomainLiveRoomStatus } from "@/api/model"

// Backend LiveRoom.status is authoritative (set when the host starts/ends the
// room). It drives the badge — not the time-window heuristic in session-status.ts.
export type LiveRoomStatus = GithubCom4H1RZooraInternalDomainLiveRoomStatus

// CTA mode the card renders. "join" => anyone with join permission can enter an
// active room; "start" => host opens a not-started room; "waiting" => not-started,
// no manage permission; "ended" => finished, no action.
export type CtaMode = "join" | "start" | "waiting" | "ended"

// badgeStatus maps to the StatusBadge component's supported status union.
export function badgeStatus(status: LiveRoomStatus | undefined): "live" | "scheduled" | "ended" {
  switch (status) {
    case "active":
      return "live"
    case "finished":
      return "ended"
    default:
      return "scheduled"
  }
}

// badgeLabelKey is the i18n key for the badge text (overrides StatusBadge default).
export function badgeLabelKey(status: LiveRoomStatus | undefined): string {
  switch (status) {
    case "active":
      return "onlineClassesPage.status.live"
    case "finished":
      return "onlineClassesPage.status.finished"
    default:
      return "onlineClassesPage.status.notStarted"
  }
}

// ctaMode picks the CTA based on status and whether the caller can host (manage).
export function ctaMode(status: LiveRoomStatus | undefined, canManage: boolean): CtaMode {
  if (status === "active") return "join"
  if (status === "finished") return "ended"
  // created (not started)
  return canManage ? "start" : "waiting"
}
