import type { RoomRole } from "../room-role"
import type { Participant } from "livekit-client"

import { useParticipants } from "@livekit/components-react"
import { Track } from "livekit-client"
import {
  ArrowUp,
  Crown,
  Eye,
  Hand,
  Info,
  Mic,
  MicOff,
  MonitorUp,
  MoreVertical,
  Users,
  UserX,
  Video,
  VideoOff,
} from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { ScrollArea } from "@/components/ui/scroll-area"
import { UserAvatar } from "@/components/user-avatar"
import { cn } from "@/lib/utils"

import { ParticipantInfoDialog } from "./participant-info-dialog"

interface PeoplePanelProps {
  states: Record<string, { role: RoomRole; handRaised: boolean; handRaisedAt?: number }>
  isHost: boolean
  onSetRole: (identity: string, role: "presenter" | "viewer") => void
  onMute: (identity: string, trackSid: string) => void
  onLowerHand: (identity: string) => void
  onRemove: (identity: string, name: string) => void
}

// The backend stamps each participant's room role into LiveKit metadata at join
// time (`{"role":"host"}`), so every client can read it straight off the
// participant — no snapshot fetch. Live role changes still arrive via the
// data-channel `role_changed` event (the `states` map), which takes precedence.
function readMetaRole(metadata: string | undefined): RoomRole | undefined {
  if (!metadata) return undefined
  try {
    const role = (JSON.parse(metadata) as { role?: string }).role
    if (role === "host" || role === "presenter" || role === "viewer") return role
  } catch {
    // malformed metadata — fall through to default
  }
  return undefined
}

function resolveRole(p: Participant, states: PeoplePanelProps["states"]): RoomRole {
  return states[p.identity]?.role ?? readMetaRole(p.metadata) ?? "viewer"
}

const ROLE_STYLES: Record<RoomRole, { icon: typeof Crown; className: string }> = {
  host: { icon: Crown, className: "bg-primary/12 text-primary" },
  presenter: { icon: MonitorUp, className: "bg-[var(--green-50)] text-[var(--green-800)]" },
  viewer: { icon: Eye, className: "bg-muted text-muted-foreground" },
}

function RoleBadge({ role }: { role: RoomRole }) {
  const { t } = useTranslation()
  const { icon: Icon, className } = ROLE_STYLES[role]
  return (
    <span
      className={cn(
        "inline-flex shrink-0 items-center gap-1 rounded-md px-1.5 py-0.5 font-mono text-[9px] font-medium tracking-[0.12em] uppercase",
        className
      )}
    >
      <Icon className="size-2.5" />
      {t(`liveRoom.people.roles.${role}`)}
    </span>
  )
}

export function PeoplePanel({ states, isHost, onSetRole, onMute, onLowerHand, onRemove }: PeoplePanelProps) {
  const { t } = useTranslation()
  const participants = useParticipants()
  // Host-only: clicking a participant opens a details dialog (device, OS, network).
  const [selected, setSelected] = useState<Participant | null>(null)

  const raisedQueue = participants
    .filter((p) => states[p.identity]?.handRaised)
    .sort((a, b) => (states[a.identity]?.handRaisedAt ?? 0) - (states[b.identity]?.handRaisedAt ?? 0))

  return (
    <>
      <ScrollArea className="min-h-0 flex-1">
        {isHost && raisedQueue.length > 0 && (
          <div className="border-border bg-primary/5 border-b px-3 py-3">
            <div className="text-primary mb-2 flex items-center gap-1.5 font-mono text-[11px] tracking-[0.2em] uppercase">
              <Hand className="size-3.5" />
              {t("liveRoom.people.raisedHands", { count: raisedQueue.length })}
            </div>
            <ul className="space-y-1">
              {raisedQueue.map((p) => {
                const name = p.name || p.identity
                return (
                  <li key={p.sid} className="hover:bg-accent flex items-center gap-2.5 rounded-lg px-1.5 py-1.5">
                    <UserAvatar name={name} size="sm" online={true} />
                    <span className="text-foreground min-w-0 flex-1 truncate text-sm">{name}</span>
                    <button
                      type="button"
                      onClick={() => onSetRole(p.identity, "presenter")}
                      title={t("liveRoom.people.makePresenter")}
                      className="bg-primary text-primary-foreground hover:bg-primary/90 flex size-7 items-center justify-center rounded-lg transition-colors"
                    >
                      <ArrowUp className="size-4" />
                    </button>
                    <button
                      type="button"
                      onClick={() => onLowerHand(p.identity)}
                      title={t("liveRoom.people.lowerHand")}
                      className="text-muted-foreground hover:bg-accent hover:text-foreground flex size-7 items-center justify-center rounded-lg transition-colors"
                    >
                      <Hand className="size-4" />
                    </button>
                  </li>
                )
              })}
            </ul>
          </div>
        )}
        <div className="px-3 pt-3">
          <span className="text-muted-foreground font-mono text-[11px] tracking-[0.2em] uppercase">
            {t("liveRoom.peopleCount", { count: participants.length })}
          </span>
        </div>
        {participants.length === 0 ? (
          <div className="text-muted-foreground flex flex-col items-center gap-2 py-12 text-center">
            <Users className="size-7 opacity-40" />
            <p className="text-sm">{t("liveRoom.controls.people")}</p>
          </div>
        ) : (
          <ul className="space-y-1 p-2.5">
            {participants.map((p) => {
              const name = p.name || p.identity
              const state = states[p.identity]
              const participantRole = resolveRole(p, states)
              const micPub = p.getTrackPublication(Track.Source.Microphone)
              const micSid = micPub?.trackSid
              // Hosts are immutable for the session — no host may change or mute
              // another host. Actions only surface for non-host participants.
              const canManage = isHost && !p.isLocal && participantRole !== "host"

              return (
                <li key={p.sid} className="hover:bg-accent flex items-center gap-2.5 rounded-xl px-2 py-2">
                  <UserAvatar name={name} size="sm" online={true} />
                  <span className="flex min-w-0 flex-1 items-center gap-1.5">
                    <span className="text-foreground min-w-0 truncate text-sm">{name}</span>
                    {p.isLocal && <span className="text-muted-foreground shrink-0 text-xs">({t("liveRoom.you")})</span>}
                  </span>
                  <RoleBadge role={participantRole} />
                  <span className="text-muted-foreground flex items-center gap-1.5">
                    {state?.handRaised && (
                      <Hand className="text-primary size-4" aria-label={t("liveRoom.people.handRaised")} />
                    )}
                    {/* Mic/cam status only for publishers — viewers can't publish,
                      so their always-off icons are meaningless noise. */}
                    {participantRole !== "viewer" && (
                      <>
                        {p.isMicrophoneEnabled ? (
                          <Mic className="size-4" />
                        ) : (
                          <MicOff className="size-4 text-red-400/80" />
                        )}
                        {p.isCameraEnabled ? (
                          <Video className="size-4" />
                        ) : (
                          <VideoOff className="text-muted-foreground size-4" />
                        )}
                      </>
                    )}
                    {isHost && (
                      <DropdownMenu>
                        <DropdownMenuTrigger
                          render={
                            <button
                              type="button"
                              aria-label={t("liveRoom.people.actions")}
                              className={cn(
                                "text-muted-foreground hover:bg-accent hover:text-foreground flex size-7 items-center justify-center rounded-lg transition-colors"
                              )}
                            >
                              <MoreVertical className="size-4" />
                            </button>
                          }
                        />
                        <DropdownMenuContent align="end" className="min-w-44 [&_[role=menuitem]]:whitespace-nowrap">
                          <DropdownMenuItem onClick={() => setSelected(p)}>
                            <Info className="size-4" />
                            {t("liveRoom.people.info.view")}
                          </DropdownMenuItem>
                          {canManage && (
                            <>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                onClick={() =>
                                  onSetRole(p.identity, participantRole === "presenter" ? "viewer" : "presenter")
                                }
                              >
                                {participantRole === "presenter" ? (
                                  <Eye className="size-4" />
                                ) : (
                                  <ArrowUp className="size-4" />
                                )}
                                {participantRole === "presenter"
                                  ? t("liveRoom.people.makeViewer")
                                  : t("liveRoom.people.makePresenter")}
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                disabled={!micSid}
                                onClick={() => {
                                  if (micSid) onMute(p.identity, micSid)
                                }}
                              >
                                <MicOff className="size-4" />
                                {t("liveRoom.people.mute")}
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem variant="destructive" onClick={() => onRemove(p.identity, name)}>
                                <UserX className="size-4" />
                                {t("liveRoom.people.remove")}
                              </DropdownMenuItem>
                            </>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    )}
                  </span>
                </li>
              )
            })}
          </ul>
        )}
      </ScrollArea>
      <ParticipantInfoDialog
        participant={selected}
        role={selected ? resolveRole(selected, states) : "viewer"}
        onClose={() => setSelected(null)}
      />
    </>
  )
}
