import { usePreviewTracks } from "@livekit/components-react"
import { Link, useRouter } from "@tanstack/react-router"
import { LocalVideoTrack } from "livekit-client"
import {
  CalendarDays,
  ChevronLeft,
  Hourglass,
  Info,
  Lock,
  Mic,
  MicOff,
  Radio,
  Settings2,
  ShieldCheck,
  Users,
  Video,
  VideoOff,
} from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import type {
  GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse,
  GithubCom4H1RZooraInternalDomainLiveRoom,
} from "@/api/model"
import type { PreJoinChoices } from "./types"

import { usePostLiveRoomsIdJoin } from "@/api/live-sessions/live-sessions"
import { useGetUsersMe } from "@/api/users/users"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Spinner } from "@/components/ui/spinner"
import { UserAvatar } from "@/components/user-avatar"
import { userHasAny } from "@/lib/access"
import { formatDate } from "@/lib/format-date"
import { formatCountdown, useNow } from "@/lib/session-status"
import { cn } from "@/lib/utils"

interface PreJoinLobbyProps {
  room: GithubCom4H1RZooraInternalDomainLiveRoom | undefined
  liveId: string
  onJoined: (data: GithubCom4H1RZooraInternalDomainJoinLiveRoomResponse, choices: PreJoinChoices) => void
}

export function PreJoinLobby({ room, liveId, onJoined }: PreJoinLobbyProps) {
  const { t, i18n } = useTranslation()
  const router = useRouter()
  const joinMutation = usePostLiveRoomsIdJoin()
  const { data: meData } = useGetUsersMe()
  const me = meData?.status === 200 ? meData.data.data : undefined
  const orgId = me?.organization_id
  const myName = me?.name ?? t("liveRoom.you")
  // The /live route renders outside the org AccessProvider, so the useAccess
  // hooks aren't available — userHasAny reads permissions straight off /users/me.
  const isModerator = userHasAny(me, [
    "live_sessions:manage",
    "live_sessions:manage_any",
    "live_sessions:create",
  ])
  const now = useNow(1000)

  const [audioEnabled, setAudioEnabled] = useState(true)
  const [videoEnabled, setVideoEnabled] = useState(true)
  const [videoDeviceId, setVideoDeviceId] = useState<string | undefined>(undefined)
  const [audioDeviceId, setAudioDeviceId] = useState<string | undefined>(undefined)
  const [permError, setPermError] = useState(false)
  const [cams, setCams] = useState<MediaDeviceInfo[]>([])
  const [mics, setMics] = useState<MediaDeviceInfo[]>([])

  const tracks = usePreviewTracks(
    {
      audio: audioEnabled ? { deviceId: audioDeviceId } : false,
      video: videoEnabled ? { deviceId: videoDeviceId } : false,
    },
    () => setPermError(true)
  )

  const videoTrack = tracks?.find((tr) => tr instanceof LocalVideoTrack) as LocalVideoTrack | undefined
  const videoRef = useRef<HTMLVideoElement>(null)

  useEffect(() => {
    const el = videoRef.current
    if (el && videoTrack) {
      videoTrack.unmute()
      videoTrack.attach(el)
      return () => {
        videoTrack.detach(el)
      }
    }
  }, [videoTrack])

  // Enumerate devices once permission has been granted (labels appear then).
  useEffect(() => {
    if (!navigator.mediaDevices?.enumerateDevices) return
    const update = () => {
      void navigator.mediaDevices.enumerateDevices().then((list) => {
        setCams(list.filter((d) => d.kind === "videoinput" && d.deviceId))
        setMics(list.filter((d) => d.kind === "audioinput" && d.deviceId))
      })
    }
    update()
    navigator.mediaDevices.addEventListener("devicechange", update)
    return () => navigator.mediaDevices.removeEventListener("devicechange", update)
  }, [tracks])

  const isFinished = room?.status === "finished"
  const isActive = room?.status === "active"
  const isCreated = room?.status === "created"
  const session = room?.class_session
  const className = session?.class?.name
  const teacherName = session?.class?.user?.name
  const sessionName = session?.name

  // The room's own scheduled time wins; fall back to the session start time.
  const scheduledIso = room?.scheduled_start_time ?? session?.start_time
  // Students can't enter a not-yet-started (created) room; only the host can.
  const isWaiting = isCreated && !isModerator

  // Format in the active app language so Persian users get Jalali dates.
  const startTime = scheduledIso ? formatDate(scheduledIso, i18n.language, "time") : undefined
  const startDate = scheduledIso ? formatDate(scheduledIso, i18n.language, "weekday-long") : undefined

  const handleJoin = () => {
    joinMutation.mutate(
      { id: liveId },
      {
        onSuccess: (res) => {
          const joinData = (res.status === 200 && res.data.data) || undefined
          if (joinData?.token && joinData?.livekit_url) {
            onJoined(joinData, { videoEnabled, audioEnabled, videoDeviceId, audioDeviceId })
          }
        },
      }
    )
  }

  return (
    <div className="relative flex min-h-screen flex-col overflow-hidden bg-zinc-950 text-zinc-100">
      <Atmosphere />

      <header className="relative z-10 flex items-center justify-between px-5 py-4 sm:px-8">
        <Link
          to={orgId ? "/org" : "/"}
          className="inline-flex items-center gap-1.5 text-[13px] text-zinc-400 transition-colors hover:text-zinc-100"
        >
          <ChevronLeft className="size-3.5 rtl:rotate-180" />
          <span>{t("liveRoom.backToDashboard")}</span>
        </Link>
        <div className="flex items-center gap-2 text-[13px] font-medium text-zinc-400">
          <ShieldCheck className="size-3.5 text-indigo-400" />
          <span>{t("liveRoom.secureEntry")}</span>
        </div>
      </header>

      <main className="relative z-10 mx-auto flex w-full max-w-6xl flex-1 flex-col items-center justify-center gap-6 px-5 pb-10 sm:px-8 lg:flex-row lg:items-stretch lg:gap-8">
        <section className="flex w-full flex-col gap-4 lg:max-w-[58%]">
          <div className="group relative aspect-video w-full overflow-hidden rounded-3xl border border-white/10 bg-zinc-900 shadow-2xl shadow-black/50">
            {videoEnabled && videoTrack && !permError ? (
              <video
                ref={videoRef}
                className="h-full w-full -scale-x-100 object-cover"
                disablePictureInPicture
                muted
                playsInline
              />
            ) : (
              <div className="flex h-full w-full flex-col items-center justify-center gap-4">
                <UserAvatar name={myName} size="lg" className="size-20 text-3xl" />
                <p className="text-sm text-zinc-400">
                  {permError ? t("liveRoom.cameraBlocked") : t("liveRoom.cameraOff")}
                </p>
              </div>
            )}

            <div className="absolute bottom-3 start-3 inline-flex items-center gap-2 rounded-full bg-black/55 px-3 py-1.5 text-xs font-medium text-white backdrop-blur-md">
              <span className={cn("size-1.5 rounded-full", audioEnabled ? "bg-emerald-400" : "bg-zinc-500")} />
              {myName}
            </div>

            <div className="absolute inset-x-0 bottom-3 flex items-center justify-center gap-2.5">
              <DeviceToggle
                on={audioEnabled}
                onToggle={() => setAudioEnabled((v) => !v)}
                onIcon={<Mic className="size-5" />}
                offIcon={<MicOff className="size-5" />}
                label={audioEnabled ? t("liveRoom.controls.micOff") : t("liveRoom.controls.micOn")}
                devices={mics}
                deviceId={audioDeviceId}
                onSelectDevice={setAudioDeviceId}
                menuLabel={t("liveRoom.device.microphone")}
              />
              <DeviceToggle
                on={videoEnabled}
                onToggle={() => setVideoEnabled((v) => !v)}
                onIcon={<Video className="size-5" />}
                offIcon={<VideoOff className="size-5" />}
                label={videoEnabled ? t("liveRoom.controls.cameraOff") : t("liveRoom.controls.cameraOn")}
                devices={cams}
                deviceId={videoDeviceId}
                onSelectDevice={setVideoDeviceId}
                menuLabel={t("liveRoom.device.camera")}
              />
            </div>
          </div>

          <p className="text-center text-[13px] text-zinc-500">{t("liveRoom.checkDevices")}</p>
        </section>

        <section className="flex w-full flex-col lg:max-w-[42%]">
          <div className="flex flex-1 flex-col rounded-3xl border border-white/10 bg-white/[0.03] p-6 backdrop-blur-xl sm:p-7">
            <div className="flex items-center gap-2.5">
              <StatusBadge status={isActive ? "live" : isFinished ? "ended" : "scheduled"} />
              {isActive && <span className="text-[13px] text-zinc-400">{t("liveRoom.classStarted")}</span>}
            </div>

            {className && (
              <div className="mt-5 text-xs font-medium tracking-wide text-indigo-300/80 uppercase">{className}</div>
            )}
            <h1 className="mt-1.5 text-2xl leading-tight font-semibold tracking-tight text-white sm:text-[28px]">
              {sessionName ?? t("liveRoom.session")}
            </h1>
            {session?.description && <p className="mt-2 text-sm text-zinc-400">{session.description}</p>}

            <div className="mt-6 space-y-4 border-t border-white/10 pt-6">
              {teacherName && (
                <MetaRow icon={<Users className="size-4" />} label={t("liveRoom.instructor")}>
                  <div className="flex items-center gap-2">
                    <UserAvatar name={teacherName} size="sm" />
                    <span className="text-sm font-medium text-zinc-100">{teacherName}</span>
                  </div>
                </MetaRow>
              )}
              {startDate && (
                <MetaRow icon={<CalendarDays className="size-4" />} label={t("liveRoom.startTime")}>
                  <span className="text-sm text-zinc-200">
                    {startDate}
                    {startTime && (
                      <>
                        {" · "}
                        <span className="font-mono" dir="ltr">
                          {startTime}
                        </span>
                      </>
                    )}
                  </span>
                </MetaRow>
              )}
            </div>

            {joinMutation.isError && !isWaiting && (
              <p className="mt-4 text-center text-sm text-red-400">{t("liveRoom.joinError")}</p>
            )}

            <div className="mt-auto flex flex-col gap-3 pt-6">
              {isWaiting ? (
                <div className="flex flex-col items-center gap-3 rounded-2xl border border-amber-400/20 bg-amber-400/5 px-5 py-6 text-center">
                  <span className="flex size-11 items-center justify-center rounded-full bg-amber-400/15 text-amber-300">
                    <Hourglass className="size-5 animate-pulse" />
                  </span>
                  <p className="text-sm font-medium text-zinc-100">{t("liveRoom.waitingForHost")}</p>
                  {scheduledIso && now < new Date(scheduledIso).getTime() && (
                    <p
                      className="font-mono text-2xl font-semibold tracking-tight text-amber-200 tabular-nums"
                      dir="ltr"
                    >
                      {formatCountdown(scheduledIso, now)}
                    </p>
                  )}
                  <p className="text-xs leading-relaxed text-zinc-400">{t("liveRoom.waitingHint")}</p>
                </div>
              ) : (
                <Button
                  size="lg"
                  onClick={handleJoin}
                  disabled={joinMutation.isPending || isFinished}
                  className="h-12 w-full gap-2 bg-indigo-500 text-base font-semibold text-white hover:bg-indigo-400"
                >
                  {joinMutation.isPending ? (
                    <Spinner className="size-4" />
                  ) : isFinished ? (
                    t("liveRoom.sessionEnded")
                  ) : isCreated ? (
                    <>
                      <Radio className="size-4" />
                      {t("liveRoom.startSession")}
                    </>
                  ) : (
                    t("liveRoom.joinNow")
                  )}
                </Button>
              )}
              <Button
                variant="ghost"
                onClick={() => router.history.back()}
                className="text-zinc-400 hover:bg-white/5 hover:text-zinc-100"
              >
                {t("liveRoom.back")}
              </Button>
            </div>
          </div>

          <div className="mt-4 flex items-center justify-center gap-3 text-xs text-zinc-500">
            <span className="inline-flex items-center gap-1.5">
              <Lock className="size-3" />
              {t("liveRoom.secureEntry")}
            </span>
            <span className="inline-block h-3 w-px bg-white/10" />
            <span className="inline-flex items-center gap-1.5">
              <Info className="size-3" />
              {t("liveRoom.avSettingsNote")}
            </span>
          </div>
        </section>
      </main>
    </div>
  )
}

function DeviceToggle({
  on,
  onToggle,
  onIcon,
  offIcon,
  label,
  devices,
  deviceId,
  onSelectDevice,
  menuLabel,
}: {
  on: boolean
  onToggle: () => void
  onIcon: React.ReactNode
  offIcon: React.ReactNode
  label: string
  devices: MediaDeviceInfo[]
  deviceId: string | undefined
  onSelectDevice: (id: string) => void
  menuLabel: string
}) {
  const { t } = useTranslation()
  return (
    <div className="flex items-center overflow-hidden rounded-full bg-black/55 backdrop-blur-md">
      <button
        type="button"
        onClick={onToggle}
        aria-label={label}
        title={label}
        className={cn(
          "flex size-11 items-center justify-center transition-colors",
          on ? "text-white hover:bg-white/10" : "bg-red-500 text-white hover:bg-red-400"
        )}
      >
        {on ? onIcon : offIcon}
      </button>
      {devices.length > 0 && (
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <button
                type="button"
                aria-label={menuLabel}
                className="flex h-11 items-center justify-center px-1.5 text-white/70 transition-colors hover:bg-white/10 hover:text-white"
              >
                <Settings2 className="size-3.5" />
              </button>
            }
          />
          <DropdownMenuContent align="center" className="max-w-[260px]">
            <DropdownMenuLabel>{menuLabel}</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuRadioGroup value={deviceId ?? devices[0]?.deviceId} onValueChange={onSelectDevice}>
              {devices.map((d) => (
                <DropdownMenuRadioItem key={d.deviceId} value={d.deviceId} className="truncate">
                  {d.label || t("liveRoom.device.unknown")}
                </DropdownMenuRadioItem>
              ))}
            </DropdownMenuRadioGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      )}
    </div>
  )
}

function MetaRow({ icon, label, children }: { icon: React.ReactNode; label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <div className="flex items-center gap-2 text-xs font-medium text-zinc-500">
        {icon}
        <span>{label}</span>
      </div>
      <div className="text-end">{children}</div>
    </div>
  )
}

function Atmosphere() {
  return (
    <div aria-hidden className="pointer-events-none absolute inset-0 overflow-hidden">
      <div className="absolute -top-1/4 start-1/2 size-[60rem] -translate-x-1/2 rounded-full bg-indigo-600/15 blur-[120px]" />
      <div className="absolute -bottom-1/3 -start-1/4 size-[40rem] rounded-full bg-fuchsia-600/10 blur-[120px]" />
      <div
        className="absolute inset-0 opacity-[0.15]"
        style={{
          backgroundImage:
            "linear-gradient(to right, rgba(255,255,255,0.04) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.04) 1px, transparent 1px)",
          backgroundSize: "48px 48px",
          maskImage: "radial-gradient(ellipse 80% 60% at 50% 40%, black, transparent)",
        }}
      />
    </div>
  )
}
