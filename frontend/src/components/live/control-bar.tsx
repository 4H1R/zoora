import type { RoomTab } from "./types"
import type { LucideIcon } from "lucide-react"

import { useLocalParticipant } from "@livekit/components-react"
import {
  BarChart3,
  Circle,
  Hand,
  LogOut,
  MessageCircleQuestion,
  MessageSquare,
  Mic,
  MicOff,
  MonitorUp,
  MoreHorizontal,
  PenLine,
  Presentation,
  Square,
  Users,
  Video,
  VideoOff,
} from "lucide-react"
import { useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Sheet, SheetContent, SheetTitle, SheetTrigger } from "@/components/ui/sheet"
import { Spinner } from "@/components/ui/spinner"
import { cn } from "@/lib/utils"

import { canPublish, useRoomRole } from "./room-role"

interface ControlBarProps {
  tab: RoomTab | null
  openTab: (tab: RoomTab) => void
  closePanel: () => void
  onLeave: () => void
  leavePending: boolean
  onEndRoom: () => void
  endPending: boolean
  unread: number
  raisedHandCount: number
  handRaised: boolean
  onToggleHand: () => void
  canShareStage: boolean
  stageKind: "none" | "slides" | "whiteboard"
  onShareSlides: (file: File) => void
  onStopStage: () => void
  onStartWhiteboard: () => void
  isRecording: boolean
  recordingPending: boolean
  onToggleRecording: () => void
  qaOpenCount: number
}

export function ControlBar({
  tab,
  openTab,
  closePanel,
  onLeave,
  leavePending,
  onEndRoom,
  endPending,
  unread,
  raisedHandCount,
  handRaised,
  onToggleHand,
  canShareStage,
  stageKind,
  onShareSlides,
  onStopStage,
  onStartWhiteboard,
  isRecording,
  recordingPending,
  onToggleRecording,
  qaOpenCount,
}: ControlBarProps) {
  const { t } = useTranslation()
  const { localParticipant, isMicrophoneEnabled, isCameraEnabled, isScreenShareEnabled } = useLocalParticipant()
  const role = useRoomRole()
  const isHost = role === "host"
  const publisher = localParticipant.permissions?.canPublish ?? canPublish(role)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [leaveOpen, setLeaveOpen] = useState(false)
  const [recordOpen, setRecordOpen] = useState(false)

  const handleSlidesClick = () => {
    if (stageKind === "slides") {
      onStopStage()
    } else {
      fileInputRef.current?.click()
    }
  }

  const handleWhiteboardClick = () => {
    if (stageKind === "whiteboard") {
      onStopStage()
    } else {
      onStartWhiteboard()
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
      {/* Solid bg, no backdrop-blur: this bar floats over the <video> stage, and a
          backdrop-filter pass over a video makes it paint black on some GPUs. */}
      <div className="border-border bg-popover/95 pointer-events-auto relative flex items-center gap-1.5 rounded-2xl border p-1.5 shadow-2xl shadow-black/30 sm:gap-2">
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
              toggle(() => localParticipant.setScreenShareEnabled(!isScreenShareEnabled), "liveRoom.errors.screenShare")
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

        {/* Whiteboard — publishers (hosts) only, desktop only */}
        {canShareStage && (
          <CtrlButton
            icon={PenLine}
            on
            active={stageKind === "whiteboard"}
            label={
              stageKind === "whiteboard" ? t("liveRoom.controls.stopWhiteboard") : t("liveRoom.controls.whiteboard")
            }
            className="hidden sm:flex"
            onClick={handleWhiteboardClick}
          />
        )}

        {/* Record — host only, desktop only */}
        {isHost && (
          <RecordButton
            recording={isRecording}
            pending={recordingPending}
            label={isRecording ? t("liveRoom.controls.stopRecording") : t("liveRoom.controls.startRecording")}
            className="hidden sm:flex"
            onClick={() => setRecordOpen(true)}
          />
        )}

        {/* Divider — only when publisher-side controls precede it, else it
            orphans at the bar edge next to the hand icon for viewers */}
        {publisher && <span className="bg-border mx-0.5 h-7 w-px" />}

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

        {/* Q&A — always visible */}
        <CtrlButton
          icon={MessageCircleQuestion}
          on
          active={tab === "qa"}
          badge={tab !== "qa" ? qaOpenCount : 0}
          label={t("liveRoom.controls.qa")}
          onClick={() => togglePanel("qa")}
        />

        {/* People — desktop only */}
        <CtrlButton
          icon={Users}
          on
          active={tab === "people"}
          badge={tab !== "people" ? raisedHandCount : 0}
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
                className="text-foreground hover:bg-accent relative flex size-11 items-center justify-center rounded-xl transition-colors sm:hidden"
              />
            }
          >
            <MoreHorizontal className="size-5" />
          </SheetTrigger>
          <SheetContent side="bottom" className="bg-popover text-foreground p-0" showCloseButton={false}>
            <SheetTitle className="sr-only">{t("liveRoom.controls.more")}</SheetTitle>
            <div className="divide-border flex flex-col divide-y py-2">
              {publisher && (
                <button
                  type="button"
                  onClick={() =>
                    toggle(
                      () => localParticipant.setScreenShareEnabled(!isScreenShareEnabled),
                      "liveRoom.errors.screenShare"
                    )
                  }
                  className="text-foreground hover:bg-accent flex items-center gap-3 px-5 py-3.5 text-sm"
                >
                  <MonitorUp className="size-5 shrink-0" />
                  <span>
                    {isScreenShareEnabled ? t("liveRoom.controls.stopShare") : t("liveRoom.controls.shareScreen")}
                  </span>
                </button>
              )}
              {canShareStage && (
                <button
                  type="button"
                  onClick={handleSlidesClick}
                  className="text-foreground hover:bg-accent flex items-center gap-3 px-5 py-3.5 text-sm"
                >
                  <Presentation className="size-5 shrink-0" />
                  <span>
                    {stageKind === "slides" ? t("liveRoom.controls.stopSlides") : t("liveRoom.controls.shareSlides")}
                  </span>
                </button>
              )}
              {canShareStage && (
                <button
                  type="button"
                  onClick={handleWhiteboardClick}
                  className="text-foreground hover:bg-accent flex items-center gap-3 px-5 py-3.5 text-sm"
                >
                  <PenLine className="size-5 shrink-0" />
                  <span>
                    {stageKind === "whiteboard"
                      ? t("liveRoom.controls.stopWhiteboard")
                      : t("liveRoom.controls.whiteboard")}
                  </span>
                </button>
              )}
              {/* Record — host only */}
              {isHost && (
                <button
                  type="button"
                  onClick={() => setRecordOpen(true)}
                  disabled={recordingPending}
                  className="text-foreground hover:bg-accent flex items-center gap-3 px-5 py-3.5 text-sm disabled:opacity-60"
                >
                  {isRecording ? (
                    <Square className="size-5 shrink-0 fill-red-600 text-red-600" />
                  ) : (
                    <Circle className="size-5 shrink-0 fill-red-600 text-red-600" />
                  )}
                  <span>{isRecording ? t("liveRoom.controls.stopRecording") : t("liveRoom.controls.startRecording")}</span>
                </button>
              )}
              <button
                type="button"
                onClick={() => togglePanel("people")}
                className="text-foreground hover:bg-accent flex items-center gap-3 px-5 py-3.5 text-sm"
              >
                <Users className="size-5 shrink-0" />
                <span>{t("liveRoom.controls.people")}</span>
              </button>
              <button
                type="button"
                onClick={() => togglePanel("polls")}
                className="text-foreground hover:bg-accent flex items-center gap-3 px-5 py-3.5 text-sm"
              >
                <BarChart3 className="size-5 shrink-0" />
                <span>{t("liveRoom.controls.polls")}</span>
              </button>
            </div>
          </SheetContent>
        </Sheet>

        <span className="bg-border mx-0.5 h-7 w-px" />

        {/* Leave */}
        <button
          type="button"
          onClick={() => (isHost ? setLeaveOpen(true) : onLeave())}
          disabled={leavePending || endPending}
          title={t("liveRoom.leave")}
          className="bg-destructive hover:bg-destructive/90 flex h-11 items-center gap-2 rounded-xl px-4 text-sm font-semibold text-white transition-colors disabled:opacity-60"
        >
          {leavePending || endPending ? <Spinner className="size-4" /> : <LogOut className="size-4 rtl:rotate-180" />}
          <span className="hidden sm:inline">{t("liveRoom.leave")}</span>
        </button>
      </div>

      {isHost && (
        <AlertDialog open={recordOpen} onOpenChange={setRecordOpen}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>
                {isRecording ? t("liveRoom.recording.stopDialog.title") : t("liveRoom.recording.startDialog.title")}
              </AlertDialogTitle>
              <AlertDialogDescription>
                {isRecording
                  ? t("liveRoom.recording.stopDialog.description")
                  : t("liveRoom.recording.startDialog.description")}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel disabled={recordingPending}>{t("common.cancel")}</AlertDialogCancel>
              <AlertDialogAction
                variant={isRecording ? "destructive" : "default"}
                disabled={recordingPending}
                onClick={() => {
                  setRecordOpen(false)
                  onToggleRecording()
                }}
              >
                {recordingPending ? <Spinner className="size-4" /> : null}
                {isRecording ? t("liveRoom.controls.stopRecording") : t("liveRoom.controls.startRecording")}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}

      {isHost && (
        <AlertDialog open={leaveOpen} onOpenChange={setLeaveOpen}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{t("liveRoom.leaveDialog.title")}</AlertDialogTitle>
              <AlertDialogDescription>{t("liveRoom.leaveDialog.description")}</AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel disabled={leavePending || endPending}>{t("common.cancel")}</AlertDialogCancel>
              <AlertDialogAction
                variant="outline"
                disabled={leavePending || endPending}
                onClick={() => {
                  setLeaveOpen(false)
                  onLeave()
                }}
              >
                {leavePending ? <Spinner className="size-4" /> : null}
                {t("liveRoom.leaveDialog.leaveOnly")}
              </AlertDialogAction>
              <AlertDialogAction
                variant="destructive"
                disabled={leavePending || endPending}
                onClick={() => {
                  setLeaveOpen(false)
                  onEndRoom()
                }}
              >
                {endPending ? <Spinner className="size-4" /> : null}
                {t("liveRoom.leaveDialog.endRoom")}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      )}
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
          ? "bg-destructive hover:bg-destructive/90 text-white"
          : active
            ? "bg-primary text-primary-foreground hover:bg-primary/90"
            : "text-foreground hover:bg-accent",
        className
      )}
    >
      <ShownIcon className="size-5" />
      {badge != null && badge > 0 && (
        <span className="bg-primary text-primary-foreground absolute -end-0.5 -top-0.5 flex min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-semibold">
          {badge > 9 ? "9+" : badge}
        </span>
      )}
    </button>
  )
}

// Recording toggle. Idle: a solid red dot that grows on hover ("arm to record").
// Live: a red-filled tile with a stop glyph under a slow pulsing halo, so the
// recording state reads at a glance across the bar.
function RecordButton({
  recording,
  pending,
  label,
  className,
  onClick,
}: {
  recording: boolean
  pending: boolean
  label: string
  className?: string
  onClick: () => void
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={pending}
      aria-label={label}
      aria-pressed={recording}
      title={label}
      className={cn(
        "group relative flex size-11 items-center justify-center rounded-xl transition-colors disabled:opacity-60",
        recording ? "bg-red-600 text-white hover:bg-red-600/90" : "text-foreground hover:bg-accent",
        className
      )}
    >
      {pending ? (
        <Spinner className="size-5" />
      ) : recording ? (
        <span className="relative flex size-5 items-center justify-center">
          <span className="absolute inline-flex size-5 animate-ping rounded-full bg-white/40" />
          <Square className="relative size-2.5 fill-current" />
        </span>
      ) : (
        <Circle className="size-5 fill-red-600 text-red-600 transition-transform group-hover:scale-110" />
      )}
    </button>
  )
}
