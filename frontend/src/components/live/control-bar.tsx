import { useLocalParticipant } from "@livekit/components-react"
import {
  LogOut,
  type LucideIcon,
  MessageSquare,
  Mic,
  MicOff,
  MonitorUp,
  Users,
  Video,
  VideoOff,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { Spinner } from "@/components/ui/spinner"
import { cn } from "@/lib/utils"

import type { SidePanelTab } from "./types"

interface ControlBarProps {
  panel: SidePanelTab | null
  setPanel: (tab: SidePanelTab | null) => void
  onLeave: () => void
  leavePending: boolean
  unread: number
}

export function ControlBar({ panel, setPanel, onLeave, leavePending, unread }: ControlBarProps) {
  const { t } = useTranslation()
  const { localParticipant, isMicrophoneEnabled, isCameraEnabled, isScreenShareEnabled } = useLocalParticipant()

  // Previously the toggle fired-and-forgot (`void fn()`), so a rejected
  // setScreenShareEnabled (e.g. publish blocked) gave the user zero feedback —
  // the button "did nothing". Now we await and surface real failures; dismissing
  // the OS picker (NotAllowedError/AbortError) is a no-op, not an error.
  const toggle = async (fn: () => Promise<unknown>, errorKey: string) => {
    try {
      await fn()
    } catch (err) {
      if (err instanceof DOMException && (err.name === "NotAllowedError" || err.name === "AbortError")) return
      toast.error(t(errorKey))
    }
  }

  return (
    <div className="pointer-events-none absolute inset-x-0 bottom-0 z-20 flex justify-center pb-4 sm:pb-5">
      <div className="pointer-events-none absolute inset-x-0 bottom-0 h-28 bg-gradient-to-t from-black/55 to-transparent" />
      <div className="pointer-events-auto relative flex items-center gap-1.5 rounded-2xl border border-white/10 bg-zinc-900/85 p-1.5 shadow-2xl shadow-black/50 backdrop-blur-xl sm:gap-2">
        <CtrlButton
          icon={Mic}
          offIcon={MicOff}
          on={isMicrophoneEnabled}
          danger
          label={isMicrophoneEnabled ? t("liveRoom.controls.micOff") : t("liveRoom.controls.micOn")}
          onClick={() =>
            toggle(() => localParticipant.setMicrophoneEnabled(!isMicrophoneEnabled), "liveRoom.errors.microphone")
          }
        />
        <CtrlButton
          icon={Video}
          offIcon={VideoOff}
          on={isCameraEnabled}
          danger
          label={isCameraEnabled ? t("liveRoom.controls.cameraOff") : t("liveRoom.controls.cameraOn")}
          onClick={() =>
            toggle(() => localParticipant.setCameraEnabled(!isCameraEnabled), "liveRoom.errors.camera")
          }
        />
        <CtrlButton
          icon={MonitorUp}
          on={isScreenShareEnabled}
          active={isScreenShareEnabled}
          label={isScreenShareEnabled ? t("liveRoom.controls.stopShare") : t("liveRoom.controls.shareScreen")}
          className="hidden sm:flex"
          onClick={() =>
            toggle(() => localParticipant.setScreenShareEnabled(!isScreenShareEnabled), "liveRoom.errors.screenShare")
          }
        />

        <span className="mx-0.5 h-7 w-px bg-white/10" />

        <CtrlButton
          icon={Users}
          on
          active={panel === "people"}
          label={t("liveRoom.controls.people")}
          onClick={() => setPanel(panel === "people" ? null : "people")}
        />
        <CtrlButton
          icon={MessageSquare}
          on
          active={panel === "chat"}
          badge={panel !== "chat" ? unread : 0}
          label={t("liveRoom.controls.chat")}
          onClick={() => setPanel(panel === "chat" ? null : "chat")}
        />

        <span className="mx-0.5 h-7 w-px bg-white/10" />

        <button
          type="button"
          onClick={onLeave}
          disabled={leavePending}
          title={t("liveRoom.leave")}
          className="flex h-11 items-center gap-2 rounded-xl bg-red-500 px-4 text-sm font-semibold text-white transition-colors hover:bg-red-400 disabled:opacity-60"
        >
          {leavePending ? <Spinner className="size-4" /> : <LogOut className="size-4 rtl:rotate-180" />}
          <span className="hidden sm:inline">{t("liveRoom.leave")}</span>
        </button>
      </div>
    </div>
  )
}

function CtrlButton({
  icon: Icon,
  offIcon: OffIcon,
  on,
  active,
  danger,
  label,
  badge,
  className,
  onClick,
}: {
  icon: LucideIcon
  offIcon?: LucideIcon
  on: boolean
  active?: boolean
  danger?: boolean
  label: string
  badge?: number
  className?: string
  onClick: () => void
}) {
  const ShownIcon = on ? Icon : (OffIcon ?? Icon)
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      title={label}
      className={cn(
        "relative flex size-11 items-center justify-center rounded-xl transition-colors",
        danger && !on
          ? "bg-red-500 text-white hover:bg-red-400"
          : active
            ? "bg-indigo-500 text-white hover:bg-indigo-400"
            : "text-zinc-200 hover:bg-white/10",
        className
      )}
    >
      <ShownIcon className="size-5" />
      {badge != null && badge > 0 && (
        <span className="absolute -end-0.5 -top-0.5 flex min-w-4 items-center justify-center rounded-full bg-indigo-500 px-1 text-[10px] font-semibold text-white">
          {badge > 9 ? "9+" : badge}
        </span>
      )}
    </button>
  )
}
