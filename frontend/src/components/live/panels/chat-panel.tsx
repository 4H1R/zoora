import type { useChat } from "@livekit/components-react"
import { MessageSquare, SendHorizonal } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useFormatDate } from "@/lib/format-date"

export function ChatPanel({ chat }: { chat: ReturnType<typeof useChat> }) {
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
    <div className="flex min-h-0 flex-1 flex-col">
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
                  <span className="text-xs font-medium text-primary">{sender}</span>
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
          className="h-10 border-white/10 bg-white/5 text-zinc-100 placeholder:text-zinc-500 focus-visible:ring-ring/40"
        />
        <Button type="submit" size="icon" disabled={isSending || !text.trim()} className="size-10 shrink-0">
          <SendHorizonal className="size-4 rtl:rotate-180" />
        </Button>
      </form>
    </div>
  )
}
