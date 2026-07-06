import { MessageSquare, SendHorizonal, Trash2 } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useFormatDate } from "@/lib/format-date"

import type { useRoomChat } from "../use-room-chat"

interface ChatPanelProps {
  chat: ReturnType<typeof useRoomChat>
  canModerate: boolean
}

export function ChatPanel({ chat, canModerate }: ChatPanelProps) {
  const { t } = useTranslation()
  const formatDate = useFormatDate()
  const { messages, send, isSending, deleteMessage } = chat
  const [text, setText] = useState("")
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [messages.length])

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault()
    const value = text.trim()
    if (!value) return
    send(value)
    setText("")
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <ScrollArea className="min-h-0 flex-1">
        <div className="flex flex-col gap-3 p-3">
          {messages.length === 0 && (
            <div className="flex flex-col items-center gap-2 py-12 text-center text-muted-foreground">
              <MessageSquare className="size-7 opacity-40" />
              <p className="text-sm">{t("liveRoom.chat.empty")}</p>
            </div>
          )}
          {messages.map((msg) => {
            const time = formatDate(msg.createdAt || undefined, "time")
            return (
              <div key={msg.id} className="group flex flex-col gap-0.5">
                <div className="flex items-baseline gap-2">
                  <span className="text-xs font-medium text-primary">{msg.senderName}</span>
                  <span className="font-mono text-[10px] text-muted-foreground" dir="ltr">
                    {time}
                  </span>
                  {canModerate && (
                    <button
                      type="button"
                      onClick={() => deleteMessage(msg.id)}
                      aria-label={t("liveRoom.chat.delete")}
                      className="ms-auto flex size-5 shrink-0 items-center justify-center rounded text-muted-foreground opacity-0 transition-opacity hover:text-red-400 focus-visible:opacity-100 group-hover:opacity-100"
                    >
                      <Trash2 className="size-3.5" />
                    </button>
                  )}
                </div>
                <p className="break-words text-sm text-foreground">{msg.content}</p>
              </div>
            )
          })}
          <div ref={bottomRef} />
        </div>
      </ScrollArea>

      <form onSubmit={handleSend} className="flex items-center gap-2 border-t border-border p-2.5">
        <Input
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder={t("liveRoom.chat.placeholder")}
          className="h-10 border-border bg-input text-foreground placeholder:text-muted-foreground focus-visible:ring-ring/40"
        />
        <Button type="submit" size="icon" disabled={isSending || !text.trim()} className="size-10 shrink-0">
          <SendHorizonal className="size-4 rtl:rotate-180" />
        </Button>
      </form>
    </div>
  )
}
