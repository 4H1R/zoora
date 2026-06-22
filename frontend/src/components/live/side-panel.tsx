import { useChat, useParticipants } from "@livekit/components-react"
import { MessageSquare, Mic, MicOff, SendHorizonal, Users, Video, VideoOff, X } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { UserAvatar } from "@/components/user-avatar"
import { useFormatDate } from "@/lib/format-date"
import { cn } from "@/lib/utils"

import type { SidePanelTab } from "./types"

interface SidePanelProps {
  tab: SidePanelTab
  setTab: (tab: SidePanelTab) => void
  onClose: () => void
  chat: ReturnType<typeof useChat>
}

export function SidePanel({ tab, setTab, onClose, chat }: SidePanelProps) {
  const { t } = useTranslation()

  return (
    <aside className="flex h-full w-full shrink-0 flex-col border-white/10 bg-zinc-900/70 backdrop-blur-xl sm:w-80 sm:border-s">
      <div className="flex items-center justify-between gap-2 border-b border-white/10 p-2.5">
        <div className="flex items-center gap-1 rounded-xl bg-white/5 p-1">
          <TabButton active={tab === "people"} onClick={() => setTab("people")} icon={<Users className="size-4" />}>
            {t("liveRoom.controls.people")}
          </TabButton>
          <TabButton active={tab === "chat"} onClick={() => setTab("chat")} icon={<MessageSquare className="size-4" />}>
            {t("liveRoom.controls.chat")}
          </TabButton>
        </div>
        <button
          type="button"
          onClick={onClose}
          aria-label={t("common.close")}
          className="flex size-8 items-center justify-center rounded-lg text-zinc-400 transition-colors hover:bg-white/5 hover:text-zinc-100"
        >
          <X className="size-4" />
        </button>
      </div>

      {tab === "people" ? <PeopleList /> : <ChatPanel chat={chat} />}
    </aside>
  )
}

function TabButton({
  active,
  onClick,
  icon,
  children,
}: {
  active: boolean
  onClick: () => void
  icon: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-[13px] font-medium transition-colors",
        active ? "bg-white/10 text-white" : "text-zinc-400 hover:text-zinc-200"
      )}
    >
      {icon}
      {children}
    </button>
  )
}

function PeopleList() {
  const { t } = useTranslation()
  const participants = useParticipants()

  return (
    <ScrollArea className="flex-1">
      <div className="px-3 pt-3">
        <span className="text-zinc-400 font-mono text-[11px] tracking-[0.2em] uppercase">
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
            return (
              <li key={p.sid} className="flex items-center gap-3 rounded-xl px-2 py-2 hover:bg-white/5">
                <UserAvatar name={name} size="sm" online={p.isMicrophoneEnabled} />
                <span className="min-w-0 flex-1 truncate text-sm text-zinc-100">
                  {name}
                  {p.isLocal && <span className="ms-1.5 text-xs text-zinc-500">({t("liveRoom.you")})</span>}
                </span>
                <span className="flex items-center gap-1.5 text-zinc-400">
                  {p.isMicrophoneEnabled ? (
                    <Mic className="size-4" />
                  ) : (
                    <MicOff className="size-4 text-red-400/80" />
                  )}
                  {p.isCameraEnabled ? <Video className="size-4" /> : <VideoOff className="size-4 text-zinc-600" />}
                </span>
              </li>
            )
          })}
        </ul>
      )}
    </ScrollArea>
  )
}

function ChatPanel({ chat }: { chat: ReturnType<typeof useChat> }) {
  const { t } = useTranslation()
  const formatDate = useFormatDate()
  const { chatMessages, send, isSending } = chat
  const [text, setText] = useState("")
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [chatMessages.length])

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault()
    const value = text.trim()
    if (!value) return
    void send(value)
    setText("")
  }

  return (
    <>
      <ScrollArea className="flex-1">
        <div className="flex flex-col gap-3 p-3">
          {chatMessages.length === 0 && (
            <div className="flex flex-col items-center gap-2 py-12 text-center text-zinc-500">
              <MessageSquare className="size-7 opacity-40" />
              <p className="text-sm">{t("liveRoom.chat.empty")}</p>
            </div>
          )}
          {chatMessages.map((msg, i) => {
            const sender = msg.from?.name || msg.from?.identity || "—"
            const time = formatDate(msg.timestamp, "time")
            return (
              <div key={`${msg.timestamp}-${i}`} className="flex flex-col gap-0.5">
                <div className="flex items-baseline gap-2">
                  <span className="text-xs font-medium text-indigo-300">{sender}</span>
                  <span className="font-mono text-[10px] text-zinc-600" dir="ltr">
                    {time}
                  </span>
                </div>
                <p className="text-sm break-words text-zinc-200">{msg.message}</p>
              </div>
            )
          })}
          <div ref={bottomRef} />
        </div>
      </ScrollArea>

      <form onSubmit={handleSend} className="flex items-center gap-2 border-t border-white/10 p-2.5">
        <Input
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder={t("liveRoom.chat.placeholder")}
          className="h-10 border-white/10 bg-white/5 text-zinc-100 placeholder:text-zinc-500 focus-visible:ring-indigo-500/40"
        />
        <Button
          type="submit"
          size="icon"
          disabled={isSending || !text.trim()}
          className="size-10 shrink-0 bg-indigo-500 text-white hover:bg-indigo-400"
        >
          <SendHorizonal className="size-4 rtl:rotate-180" />
        </Button>
      </form>
    </>
  )
}
