import { Check, ChevronUp, MessageCircleQuestion, RotateCcw, SendHorizonal, Trash2, X } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useFormatDate } from "@/lib/format-date"
import { cn } from "@/lib/utils"

import type { RoomQuestion } from "../use-room-qa"
import type { useRoomQa } from "../use-room-qa"

interface QaPanelProps {
  qa: ReturnType<typeof useRoomQa>
  isHost: boolean
  myId: string
}

export function QaPanel({ qa, isHost, myId }: QaPanelProps) {
  const { t } = useTranslation()
  const { questions, ask, isAsking, vote, resolve, dismiss, reopen, remove } = qa
  const [text, setText] = useState("")

  const handleAsk = (e: React.FormEvent) => {
    e.preventDefault()
    const value = text.trim()
    if (value.length < 2) return
    ask(value, () => toast.error(t("liveRoom.qa.askError")))
    setText("")
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <ScrollArea className="min-h-0 flex-1">
        <div className="flex flex-col gap-2.5 p-3">
          {questions.length === 0 && (
            <div className="flex flex-col items-center gap-2 py-12 text-center text-muted-foreground">
              <MessageCircleQuestion className="size-7 opacity-40" />
              <p className="text-sm">{t("liveRoom.qa.empty")}</p>
              <p className="text-xs opacity-70">{t("liveRoom.qa.emptyHint")}</p>
            </div>
          )}
          {questions.map((q) => (
            <QuestionCard
              key={q.id}
              q={q}
              isHost={isHost}
              isMine={q.authorId === myId}
              onVote={() => vote(q.id)}
              onResolve={() => resolve(q.id)}
              onDismiss={() => dismiss(q.id)}
              onReopen={() => reopen(q.id)}
              onRemove={() => remove(q.id)}
            />
          ))}
        </div>
      </ScrollArea>

      <form onSubmit={handleAsk} className="flex items-center gap-2 border-t border-border p-2.5">
        <Input
          value={text}
          onChange={(e) => setText(e.target.value)}
          maxLength={500}
          placeholder={t("liveRoom.qa.placeholder")}
          className="h-10 border-border bg-input text-foreground placeholder:text-muted-foreground focus-visible:ring-ring/40"
        />
        <Button type="submit" size="icon" disabled={isAsking || text.trim().length < 2} className="size-10 shrink-0">
          <SendHorizonal className="size-4 rtl:rotate-180" />
        </Button>
      </form>
    </div>
  )
}

interface QuestionCardProps {
  q: RoomQuestion
  isHost: boolean
  isMine: boolean
  onVote: () => void
  onResolve: () => void
  onDismiss: () => void
  onReopen: () => void
  onRemove: () => void
}

function QuestionCard({ q, isHost, isMine, onVote, onResolve, onDismiss, onReopen, onRemove }: QuestionCardProps) {
  const { t, i18n } = useTranslation()
  const formatDate = useFormatDate()
  const isOpen = q.status === "open"
  const canVote = isOpen && !isMine
  const time = formatDate(q.createdAt || undefined, "time")

  return (
    <div
      className={cn(
        "flex gap-2.5 rounded-lg border border-border bg-muted/40 p-2.5",
        !isOpen && "opacity-60",
      )}
    >
      {/* Vote pill */}
      <button
        type="button"
        onClick={onVote}
        disabled={!canVote}
        aria-label={t("liveRoom.qa.upvote")}
        className={cn(
          "flex h-fit w-10 shrink-0 flex-col items-center gap-0.5 rounded-md border px-1 py-1.5 transition-colors",
          q.votedByMe
            ? "border-primary bg-primary/15 text-primary"
            : "border-border bg-background text-muted-foreground",
          canVote ? "hover:border-primary hover:text-primary" : "cursor-default opacity-70",
        )}
      >
        <ChevronUp className="size-4" />
        <span className="text-xs font-semibold tabular-nums">{q.voteCount}</span>
      </button>

      {/* Body */}
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <div className="flex items-baseline gap-2">
          <span className="truncate text-xs font-medium text-primary">{q.authorName}</span>
          <span className="font-mono text-[10px] text-muted-foreground" dir="ltr">
            {time}
          </span>
          {q.status === "resolved" && (
            <span className="ms-auto shrink-0 rounded-full bg-green-500/15 px-2 py-0.5 text-[10px] font-semibold text-green-500">
              {t("liveRoom.qa.resolved")}
            </span>
          )}
          {q.status === "dismissed" && (
            <span className="ms-auto shrink-0 rounded-full bg-muted px-2 py-0.5 text-[10px] font-semibold text-muted-foreground">
              {t("liveRoom.qa.dismissed")}
            </span>
          )}
        </div>
        <p className="break-words text-sm text-foreground">{q.text}</p>

        {/* Actions — pinned to the inline-end (left in RTL, right in LTR), the
            side opposite the text, regardless of any inherited CSS direction
            inside the LiveKit room. */}
        {(isHost || isMine) && (
          <div dir={i18n.dir()} className="mt-0.5 flex items-center justify-end gap-1">
            {isHost && isOpen && (
              <>
                <ActionButton label={t("liveRoom.qa.resolve")} onClick={onResolve} tone="positive">
                  <Check className="size-3.5" />
                  {t("liveRoom.qa.resolve")}
                </ActionButton>
                <ActionButton label={t("liveRoom.qa.dismiss")} onClick={onDismiss}>
                  <X className="size-3.5" />
                </ActionButton>
              </>
            )}
            {isHost && !isOpen && (
              <ActionButton label={t("liveRoom.qa.reopen")} onClick={onReopen}>
                <RotateCcw className="size-3.5" />
                {t("liveRoom.qa.reopen")}
              </ActionButton>
            )}
            {(isHost || isMine) && (
              <ActionButton label={t("liveRoom.qa.delete")} onClick={onRemove} tone="danger">
                <Trash2 className="size-3.5" />
              </ActionButton>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

function ActionButton({
  children,
  label,
  onClick,
  tone,
}: {
  children: React.ReactNode
  label: string
  onClick: () => void
  tone?: "positive" | "danger"
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      title={label}
      className={cn(
        "flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent",
        tone === "positive" && "hover:text-green-500",
        tone === "danger" && "hover:text-red-400",
      )}
    >
      {children}
    </button>
  )
}
