import { useLocalParticipant } from "@livekit/components-react"
import { useRef } from "react"
import {
  BarChart3,
  Hand,
  LogOut,
  type LucideIcon,
  MessageSquare,
  Mic,
  MicOff,
  MonitorUp,
  MoreHorizontal,
  Presentation,
  Users,
  Video,
  VideoOff,
} from "lucide-react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { Sheet, SheetContent, SheetTitle, SheetTrigger } from "@/components/ui/sheet"
import { Spinner } from "@/components/ui/spinner"
import { cn } from "@/lib/utils"

import { canPublish, useRoomRole } from "./room-role"
import type { RoomTab } from "./types"

interface ControlBarProps {
  tab: RoomTab | null
  openTab: (tab: RoomTab) => void
  closePanel: () => void
  onLeave: () => void
  leavePending: boolean
  unread: number
  handRaised: boolean
  onToggleHand: () => void
  canShareStage: boolean
  stageKind: "none" | "slides"
  onShareSlides: (file: File) => void
  onStopStage: () => void
}

export function ControlBar({ tab, openTab, closePanel, onLeave, leavePending, unread, handRaised, onToggleHand, canShareStage, stageKind, onShareSlides, onStopStage }: ControlBarProps) {
  const { t } = useTranslation()
  const { localParticipant, isMicrophoneEnabled, isCameraEnabled, isScreenShareEnabled } = useLocalParticipant()
  const role = useRoomRole()
  const publisher = localParticipant.permissions?.canPublish ?? canPublish(role)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleSlidesClick = () => {
    if (stageKind === "slides") {
      onStopStage()
    } else {
      fileInputRef.current?.click()
    }
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) onShareSlides(file)
    // Reset so the same file can be re-selected later
    e.target.value = ""
  }

  // Swallow OS-permission dismissals; surface real errors via toast
  const toggle = async (fn: () => Promise<unknown>, errorKey: string) => {
    try {
      await fn()
    } catch (err) {
      if (err instanceof DOMException && (err.name === "NotAllowedError" || err.name === "AbortError")) return
      toast.error(t(errorKey))
    }
  }

  const togglePanel = (next: RoomTab) => {
    if (tab === next) {
      closePanel()
    } else {
      openTab(next)
    }
  }

  return (
    <div className="pointer-events-none absolute inset-x-0 bottom-0 z-20 flex justify-center pb-4 sm:pb-5">
      <div className="pointer-events-none absolute inset-x-0 bottom-0 h-28 bg-gradient-to-t from-black/55 to-transparent" />
      <div className="pointer-events-auto relative flex items-center gap-1.5 rounded-2xl border border-white/10 bg-zinc-900/85 p-1.5 shadow-2xl shadow-black/50 backdrop-blur-xl sm:gap-2">

        {/* Mic / Cam — publishers only */}
        {publisher && (
          <>
            <CtrlButton
              icon={Mic}
              offIcon={MicOff}
              on={isMicrophoneEnabled}
              danger
              label={isMicrophoneEnabled ? t("liveRoom.controls.micOff") : t("liveRoom.controls.micOn")}
              onClick={() =>
                toggle(
                  () => localParticipant.setMicrophoneEnabled(!isMicrophoneEnabled),
                  "liveRoom.errors.microphone",
                )
              }
            />
            <CtrlButton
              icon={Video}
              offIcon={VideoOff}
              on={isCameraEnabled}
              danger
              label={isCameraEnabled ? t("liveRoom.controls.cameraOff") : t("liveRoom.controls.cameraOn")}
              onClick={() =>
                toggle(
                  () => localParticipant.setCameraEnabled(!isCameraEnabled),
                  "liveRoom.errors.camera",
                )
              }
            />
          </>
        )}

        {/* Screenshare — publishers only, desktop only */}
        {publisher && (
          <CtrlButton
            icon={MonitorUp}
            on={isScreenShareEnabled}
            active={isScreenShareEnabled}
            label={isScreenShareEnabled ? t("liveRoom.controls.stopShare") : t("liveRoom.controls.shareScreen")}
            className="hidden sm:flex"
            onClick={() =>
              toggle(
                () => localParticipant.setScreenShareEnabled(!isScreenShareEnabled),
                "liveRoom.errors.screenShare",
              )
            }
          />
        )}

        {/* Slides — publishers (hosts) only, desktop only */}
        {canShareStage && (
          <>
            <input
              ref={fileInputRef}
              type="file"
              accept="application/pdf"
              className="hidden"
              onChange={handleFileChange}
            />
            <CtrlButton
              icon={Presentation}
              on
              active={stageKind === "slides"}
              label={stageKind === "slides" ? t("liveRoom.controls.stopSlides") : t("liveRoom.controls.shareSlides")}
              className="hidden sm:flex"
              onClick={handleSlidesClick}
            />
          </>
        )}

        <span className="mx-0.5 h-7 w-px bg-white/10" />

        {/* Raise hand — viewers only */}
        {!publisher && (
          <CtrlButton
            icon={Hand}
            on
            active={handRaised}
            label={t(handRaised ? "liveRoom.controls.lowerHand" : "liveRoom.controls.raiseHand")}
            onClick={onToggleHand}
          />
        )}

        {/* Chat — always visible */}
        <CtrlButton
          icon={MessageSquare}
          on
          active={tab === "chat"}
          badge={tab !== "chat" ? unread : 0}
          label={t("liveRoom.controls.chat")}
          onClick={() => togglePanel("chat")}
        />

        {/* People — desktop only */}
        <CtrlButton
          icon={Users}
          on
          active={tab === "people"}
          label={t("liveRoom.controls.people")}
          className="hidden sm:flex"
          onClick={() => togglePanel("people")}
        />

        {/* Polls — desktop only */}
        <CtrlButton
          icon={BarChart3}
          on
          active={tab === "polls"}
          label={t("liveRoom.controls.polls")}
          className="hidden sm:flex"
          onClick={() => togglePanel("polls")}
        />

        {/* Mobile "More" Sheet — sm:hidden */}
        <Sheet>
          <SheetTrigger
            render={
              <button
                type="button"
                aria-label={t("liveRoom.controls.more")}
                title={t("liveRoom.controls.more")}
                className="relative flex size-11 items-center justify-center rounded-xl text-zinc-200 transition-colors hover:bg-white/10 sm:hidden"
              />
            }
          >
            <MoreHorizontal className="size-5" />
          </SheetTrigger>
          <SheetContent side="bottom" className="bg-zinc-900 p-0 text-zinc-100" showCloseButton={false}>
            <SheetTitle className="sr-only">{t("liveRoom.controls.more")}</SheetTitle>
            <div className="flex flex-col divide-y divide-white/10 py-2">
              {publisher && (
                <button
                  type="button"
                  onClick={() =>
                    toggle(
                      () => localParticipant.setScreenShareEnabled(!isScreenShareEnabled),
                      "liveRoom.errors.screenShare",
                    )
                  }
                  className="flex items-center gap-3 px-5 py-3.5 text-sm text-zinc-200 hover:bg-white/5"
                >
                  <MonitorUp className="size-5 shrink-0" />
                  <span>{isScreenShareEnabled ? t("liveRoom.controls.stopShare") : t("liveRoom.controls.shareScreen")}</span>
                </button>
              )}
              {canShareStage && (
                <button
                  type="button"
                  onClick={handleSlidesClick}
                  className="flex items-center gap-3 px-5 py-3.5 text-sm text-zinc-200 hover:bg-white/5"
                >
                  <Presentation className="size-5 shrink-0" />
                  <span>{stageKind === "slides" ? t("liveRoom.controls.stopSlides") : t("liveRoom.controls.shareSlides")}</span>
                </button>
              )}
              <button
                type="button"
                onClick={() => togglePanel("people")}
                className="flex items-center gap-3 px-5 py-3.5 text-sm text-zinc-200 hover:bg-white/5"
              >
                <Users className="size-5 shrink-0" />
                <span>{t("liveRoom.controls.people")}</span>
              </button>
              <button
                type="button"
                onClick={() => togglePanel("polls")}
                className="flex items-center gap-3 px-5 py-3.5 text-sm text-zinc-200 hover:bg-white/5"
              >
                <BarChart3 className="size-5 shrink-0" />
                <span>{t("liveRoom.controls.polls")}</span>
              </button>
            </div>
          </SheetContent>
        </Sheet>

        <span className="mx-0.5 h-7 w-px bg-white/10" />

        {/* Leave */}
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
            ? "bg-primary text-primary-foreground hover:bg-primary/90"
            : "text-zinc-200 hover:bg-white/10",
        className,
      )}
    >
      <ShownIcon className="size-5" />
      {badge != null && badge > 0 && (
        <span className="absolute -end-0.5 -top-0.5 flex min-w-4 items-center justify-center rounded-full bg-primary px-1 text-[10px] font-semibold text-primary-foreground">
          {badge > 9 ? "9+" : badge}
        </span>
      )}
    </button>
  )
}
