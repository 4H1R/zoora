import { useParticipants } from "@livekit/components-react"
import { Track } from "livekit-client"
import { Hand, Mic, MicOff, MoreVertical, Users, Video, VideoOff } from "lucide-react"
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

export function PeoplePanel({ states, isHost, onSetRole, onMute }: PeoplePanelProps) {
  const { t } = useTranslation()
  const participants = useParticipants()

  return (
    <ScrollArea className="min-h-0 flex-1">
      <div className="px-3 pt-3">
        <span className="font-mono text-[11px] tracking-[0.2em] text-zinc-400 uppercase">
          {t("liveRoom.peopleCount", { count: participants.length })}
        </span>
      </div>
      {participants.length === 0 ? (
        <div className="flex flex-col items-center gap-2 py-12 text-center text-zinc-500">
          <Users className="size-7 opacity-40" />
          <p className="text-sm">{t("liveRoom.controls.people")}</p>
        </div>
      ) : (
        <ul className="space-y-1 p-2.5">
          {participants.map((p) => {
            const name = p.name || p.identity
            const state = states[p.identity]
            const participantRole = state?.role ?? "viewer"
            const micPub = p.getTrackPublication(Track.Source.Microphone)
            const micSid = micPub?.trackSid

            return (
              <li key={p.sid} className="flex items-center gap-3 rounded-xl px-2 py-2 hover:bg-white/5">
                <UserAvatar name={name} size="sm" online={true} />
                <span className="min-w-0 flex-1 truncate text-sm text-zinc-100">
                  {name}
                  {p.isLocal && <span className="ms-1.5 text-xs text-zinc-500">({t("liveRoom.you")})</span>}
                </span>
                <span className="flex items-center gap-1.5 text-zinc-400">
                  {state?.handRaised && (
                    <Hand
                      className="size-4 text-primary"
                      aria-label={t("liveRoom.people.handRaised")}
                    />
                  )}
                  {p.isMicrophoneEnabled ? <Mic className="size-4" /> : <MicOff className="size-4 text-red-400/80" />}
                  {p.isCameraEnabled ? <Video className="size-4" /> : <VideoOff className="size-4 text-zinc-600" />}
                  {isHost && !p.isLocal && (
                    <DropdownMenu>
                      <DropdownMenuTrigger
                        render={
                          <button
                            type="button"
                            aria-label={t("liveRoom.people.actions")}
                            className={cn(
                              "flex size-7 items-center justify-center rounded-lg text-zinc-400 transition-colors hover:bg-white/10 hover:text-zinc-100",
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
