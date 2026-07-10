import type { ChatMessage } from "./lib/messages"

import { PinIcon, PinOffIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { ScrollArea } from "@/components/ui/scroll-area"

import { useJumpToMessage } from "./jump-context"
import { usePinActions, usePins } from "./use-pins"

/** Collapse whitespace into a single-line snippet; fall back for media-only pins. */
function snippet(message: ChatMessage, fallback: string): string {
  const text = (message.content ?? "").replace(/\s+/g, " ").trim()
  return text || fallback
}

interface PinnedBarProps {
  convId: string
}

/**
 * A compact strip beneath the thread header surfacing the conversation's pinned
 * messages. Hidden entirely when nothing is pinned. With a single pin the whole
 * bar jumps to it (plus an inline unpin). With several, the bar previews the
 * most recent and opens a popover listing them all — each row jumps to its
 * message and carries its own unpin affordance. Pin state refreshes via the
 * invalidations in `usePinActions` (there is no realtime echo for pinning).
 */
export function PinnedBar({ convId }: PinnedBarProps) {
  const { t } = useTranslation()
  const { pins } = usePins(convId)
  const { unpin } = usePinActions(convId)
  const jumpToMessage = useJumpToMessage()
  const [open, setOpen] = useState(false)

  if (pins.length === 0) return null

  const latest = pins[0]
  const attachmentLabel = t("conversations.pinned.attachment")

  // One-line "Sender: snippet" preview shared by the bar and its trigger.
  function preview(message: ChatMessage) {
    return (
      <span className="flex min-w-0 items-baseline gap-1.5 text-xs">
        <span className="text-foreground shrink-0 font-medium">{message.sender?.name ?? ""}</span>
        <span className="text-muted-foreground truncate">{snippet(message, attachmentLabel)}</span>
      </span>
    )
  }

  // Single pin: the whole bar is one jump target with an inline unpin.
  if (pins.length === 1) {
    return (
      <div className="bg-accent/40 flex items-center gap-1 border-b py-1.5 pe-2 ps-4">
        <PinIcon className="text-primary size-3.5 shrink-0" />
        <button
          type="button"
          onClick={() => jumpToMessage(latest.id ?? "")}
          className="hover:bg-accent/60 flex min-w-0 flex-1 items-center rounded-md px-1.5 py-1 text-start transition"
          aria-label={t("conversations.pinned.jump")}
        >
          {preview(latest)}
        </button>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          className="text-muted-foreground hover:text-foreground size-7 shrink-0 rounded-full"
          aria-label={t("conversations.actions.unpin")}
          onClick={() => unpin(latest.id ?? "")}
        >
          <PinOffIcon />
        </Button>
      </div>
    )
  }

  function jumpAndClose(id: string) {
    setOpen(false)
    jumpToMessage(id)
  }

  // Several pins: preview the latest, reveal the full list in a popover.
  return (
    <Popover open={open} onOpenChange={setOpen}>
      <div className="bg-accent/40 flex items-center border-b py-1.5 pe-2 ps-4">
        <PinIcon className="text-primary me-1 size-3.5 shrink-0" />
        <PopoverTrigger
          render={
            <button
              type="button"
              className="hover:bg-accent/60 flex min-w-0 flex-1 items-center gap-2 rounded-md px-1.5 py-1 text-start transition"
              aria-label={t("conversations.pinned.count", { count: pins.length })}
            />
          }
        >
          {preview(latest)}
          <span className="text-primary bg-primary/10 ms-auto shrink-0 rounded-full px-2 py-0.5 text-xs font-medium">
            {pins.length}
          </span>
        </PopoverTrigger>
      </div>

      <PopoverContent align="start" side="bottom" className="w-80 p-0">
        <p className="text-muted-foreground border-b px-3 py-2 text-xs font-medium">
          {t("conversations.pinned.count", { count: pins.length })}
        </p>
        <ScrollArea className="max-h-72">
          <ul className="p-1">
            {pins.map((message) => (
              <li key={message.id} className="flex items-center gap-1">
                <button
                  type="button"
                  onClick={() => jumpAndClose(message.id ?? "")}
                  className="hover:bg-accent flex min-w-0 flex-1 flex-col items-start gap-0.5 rounded-md px-2 py-1.5 text-start transition"
                >
                  <span className="text-foreground text-xs font-medium">{message.sender?.name ?? ""}</span>
                  <span className="text-muted-foreground line-clamp-2 text-xs">
                    {snippet(message, attachmentLabel)}
                  </span>
                </button>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  className="text-muted-foreground hover:text-foreground size-7 shrink-0 rounded-full"
                  aria-label={t("conversations.actions.unpin")}
                  onClick={() => unpin(message.id ?? "")}
                >
                  <PinOffIcon />
                </Button>
              </li>
            ))}
          </ul>
        </ScrollArea>
      </PopoverContent>
    </Popover>
  )
}
