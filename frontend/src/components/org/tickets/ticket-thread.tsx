import type { PendingAttachment } from "@/components/org/tickets/attachments"

import { useQueryClient } from "@tanstack/react-query"
import { ArrowLeftIcon, GraduationCapIcon, LockIcon, SendIcon, TicketIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetTicketsIdQueryKey,
  getGetTicketsQueryKey,
  useGetTicketsId,
  usePostTicketsIdClose,
  usePostTicketsIdMessages,
} from "@/api/tickets/tickets"
import { AttachmentChips, AttachmentPicker } from "@/components/org/tickets/attachments"
import { TicketStatusBadge, TicketTypeBadge } from "@/components/org/tickets/ticket-badges"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { formatRelativeTime } from "@/lib/relative-time"
import { cn } from "@/lib/utils"

export function TicketThread({
  ticketId,
  currentUserId,
  onBack,
}: {
  ticketId?: string
  currentUserId: string
  onBack: () => void
}) {
  const { t, i18n } = useTranslation()
  const queryClient = useQueryClient()

  const { data, isLoading } = useGetTicketsId(ticketId ?? "", {
    query: { enabled: !!ticketId },
  })
  const ticket = (data?.status === 200 && data.data.data) || undefined

  const [body, setBody] = useState("")
  const [attachments, setAttachments] = useState<PendingAttachment[]>([])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetTicketsQueryKey() })
    if (ticketId) queryClient.invalidateQueries({ queryKey: getGetTicketsIdQueryKey(ticketId) })
  }

  const replyMutation = usePostTicketsIdMessages({
    mutation: {
      onSuccess: () => {
        setBody("")
        setAttachments([])
        invalidate()
      },
      onError: () => toast.error(t("tickets.error")),
    },
  })

  const closeMutation = usePostTicketsIdClose({
    mutation: {
      onSuccess: invalidate,
      onError: () => toast.error(t("tickets.error")),
    },
  })

  if (!ticketId) {
    return (
      <div className="text-muted-foreground flex h-full flex-col items-center justify-center gap-3 p-8 text-center text-sm">
        <TicketIcon className="size-8 opacity-40" />
        {t("tickets.thread.selectPrompt")}
      </div>
    )
  }

  if (isLoading || !ticket) {
    return (
      <div className="space-y-3 p-4">
        <Skeleton className="h-6 w-2/3" />
        <Skeleton className="h-4 w-1/3" />
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
      </div>
    )
  }

  const isClosed = ticket.status === "closed"
  const target =
    ticket.quiz_room?.quiz?.title ??
    ticket.gradebook_column?.title ??
    (ticket.type === "grade_objection" ? t("tickets.form.targetGeneral") : undefined)

  const submitReply = () => {
    const trimmed = body.trim()
    if (!trimmed || !ticket.id) return
    replyMutation.mutate({
      id: ticket.id,
      data: { body: trimmed, media_ids: attachments.map((a) => a.id) },
    })
  }

  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex flex-wrap items-start justify-between gap-2 border-b p-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon-sm" className="md:hidden" onClick={onBack} aria-label={t("tickets.thread.back")}>
              <ArrowLeftIcon className="size-4 rtl:rotate-180" />
            </Button>
            <h2 className="min-w-0 truncate text-base font-semibold">{ticket.title}</h2>
          </div>
          <div className="mt-2 flex flex-wrap items-center gap-2">
            <TicketStatusBadge status={ticket.status} />
            <TicketTypeBadge type={ticket.type} />
            <span className="text-muted-foreground text-xs">{ticket.class?.name}</span>
          </div>
          {target && (
            <div className="text-muted-foreground mt-2 inline-flex items-center gap-1.5 rounded-md border px-2 py-1 text-xs">
              <GraduationCapIcon className="size-3.5" />
              {target}
            </div>
          )}
        </div>
        {!isClosed && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => ticket.id && closeMutation.mutate({ id: ticket.id })}
            disabled={closeMutation.isPending}
          >
            <LockIcon className="size-4" />
            {t("tickets.thread.close")}
          </Button>
        )}
      </div>

      <div className="min-h-0 flex-1 space-y-3 overflow-y-auto p-4">
        {(ticket.messages ?? []).map((msg) => {
          const own = msg.user_id === currentUserId
          return (
            <div key={msg.id} className={cn("flex flex-col", own ? "items-end" : "items-start")}>
              <div
                className={cn(
                  "max-w-[85%] rounded-xl px-3 py-2 text-sm whitespace-pre-wrap",
                  own ? "bg-primary text-primary-foreground rounded-ee-sm" : "bg-muted rounded-es-sm"
                )}
              >
                {msg.body}
                <AttachmentChips mediaIds={msg.media_ids} />
              </div>
              <span className="text-muted-foreground mt-1 text-xs">
                {own ? t("tickets.thread.you") : (msg.user?.name ?? "")} ·{" "}
                {formatRelativeTime(msg.created_at, i18n.language)}
              </span>
            </div>
          )
        })}
      </div>

      {isClosed ? (
        <div className="text-muted-foreground flex items-center justify-center gap-2 border-t p-4 text-sm">
          <LockIcon className="size-4" />
          {t("tickets.thread.closed")}
        </div>
      ) : (
        <div className="space-y-2 border-t p-3">
          <Textarea
            value={body}
            onChange={(e) => setBody(e.target.value)}
            placeholder={t("tickets.thread.replyPlaceholder")}
            rows={2}
            className="resize-none"
            onKeyDown={(e) => {
              if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) submitReply()
            }}
          />
          <div className="flex items-center justify-between gap-2">
            <AttachmentPicker
              classId={ticket.class_id}
              attachments={attachments}
              onChange={setAttachments}
              disabled={replyMutation.isPending}
            />
            <Button size="sm" onClick={submitReply} disabled={replyMutation.isPending || !body.trim()}>
              <SendIcon className="size-4 rtl:rotate-180" />
              {t("tickets.thread.reply")}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
