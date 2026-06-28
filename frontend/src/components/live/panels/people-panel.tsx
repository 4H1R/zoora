import { useParticipants } from "@livekit/components-react"
import { Track, type Participant } from "livekit-client"
import { Crown, Eye, Hand, Mic, MicOff, MonitorUp, MoreVertical, Users, Video, VideoOff } from "lucide-react"
import { useTranslation } from "react-i18next"

import { ScrollArea } from "@/components/ui/scroll-area"
import { UserAvatar } from "@/components/user-avatar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { cn } from "@/lib/utils"

import type { RoomRole } from "../room-role"

interface PeoplePanelProps {
  states: Record<string, { role: RoomRole; handRaised: boolean }>
  isHost: boolean
  onSetRole: (identity: string, role: "presenter" | "viewer") => void
  onMute: (identity: string, trackSid: string) => void
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
        className,
      )}
    >
      <Icon className="size-2.5" />
      {t(`liveRoom.people.roles.${role}`)}
    </span>
  )
}

export function PeoplePanel({ states, isHost, onSetRole, onMute }: PeoplePanelProps) {
  const { t } = useTranslation()
  const participants = useParticipants()

  return (
    <ScrollArea className="min-h-0 flex-1">
      <div className="px-3 pt-3">
        <span className="font-mono text-[11px] tracking-[0.2em] text-muted-foreground uppercase">
          {t("liveRoom.peopleCount", { count: participants.length })}
        </span>
      </div>
      {participants.length === 0 ? (
        <div className="flex flex-col items-center gap-2 py-12 text-center text-muted-foreground">
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
              <li key={p.sid} className="flex items-center gap-2.5 rounded-xl px-2 py-2 hover:bg-accent">
                <UserAvatar name={name} size="sm" online={true} />
                <span className="flex min-w-0 flex-1 items-center gap-1.5">
                  <span className="min-w-0 truncate text-sm text-foreground">{name}</span>
                  {p.isLocal && <span className="shrink-0 text-xs text-muted-foreground">({t("liveRoom.you")})</span>}
                </span>
                <RoleBadge role={participantRole} />
                <span className="flex items-center gap-1.5 text-muted-foreground">
                  {state?.handRaised && (
                    <Hand
                      className="size-4 text-primary"
                      aria-label={t("liveRoom.people.handRaised")}
                    />
                  )}
                  {p.isMicrophoneEnabled ? <Mic className="size-4" /> : <MicOff className="size-4 text-red-400/80" />}
                  {p.isCameraEnabled ? <Video className="size-4" /> : <VideoOff className="size-4 text-muted-foreground" />}
                  {canManage && (
                    <DropdownMenu>
                      <DropdownMenuTrigger
                        render={
                          <button
                            type="button"
                            aria-label={t("liveRoom.people.actions")}
                            className={cn(
                              "flex size-7 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-accent hover:text-foreground",
                            )}
                          >
                            <MoreVertical className="size-4" />
                          </button>
                        }
                      />
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          onClick={() =>
                            onSetRole(p.identity, participantRole === "presenter" ? "viewer" : "presenter")
                          }
                        >
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
                          {t("liveRoom.people.mute")}
                        </DropdownMenuItem>
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
  )
}
