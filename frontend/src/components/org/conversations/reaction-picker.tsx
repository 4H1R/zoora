import { EmojiPicker } from "frimousse"
import { SmilePlusIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"

// One-tap row of the reactions people actually reach for, in a deliberate order
// (approval → love → laughter → surprise → sadness → celebration).
const QUICK_EMOJIS = ["👍", "❤️", "😂", "😮", "😢", "🎉"]

interface ReactionPickerProps {
  /** Fired with the chosen emoji from either the quick row or the full picker. */
  onSelect: (emoji: string) => void
  /** Alignment of the popover — mirror the message side so it opens inward. */
  align?: "start" | "center" | "end"
  className?: string
}

/**
 * The "add reaction" affordance: a small smile-plus icon button that opens a
 * popover with a one-tap quick row of common emojis above the full frimousse
 * picker (the same picker the composer uses). Selecting from either row fires
 * `onSelect` and closes.
 */
export function ReactionPicker({ onSelect, align = "start", className }: ReactionPickerProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  function pick(emoji: string) {
    onSelect(emoji)
    setOpen(false)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className={cn("text-muted-foreground hover:text-foreground size-7 rounded-full", className)}
            aria-label={t("conversations.reactions.add")}
          />
        }
      >
        <SmilePlusIcon />
      </PopoverTrigger>
      <PopoverContent align={align} side="top" className="w-fit p-0">
        {/* Quick row — big tap targets with a playful lift on hover. */}
        <div className="flex items-center gap-0.5 border-b p-1.5">
          {QUICK_EMOJIS.map((emoji) => (
            <button
              key={emoji}
              type="button"
              onClick={() => pick(emoji)}
              aria-label={t("conversations.reactions.toggle", { emoji })}
              className="hover:bg-accent flex size-9 items-center justify-center rounded-lg text-xl transition hover:-translate-y-0.5"
            >
              {emoji}
            </button>
          ))}
        </div>

        <EmojiPicker.Root onEmojiSelect={({ emoji }) => pick(emoji)} className="isolate flex h-72 w-72 flex-col">
          <EmojiPicker.Search
            placeholder={t("conversations.composer.emojiSearch")}
            className="bg-muted/60 placeholder:text-muted-foreground focus-visible:ring-ring/40 m-2 rounded-lg px-2.5 py-2 text-sm outline-none focus-visible:ring-2"
          />
          <EmojiPicker.Viewport className="relative flex-1 outline-hidden">
            <EmojiPicker.Loading className="text-muted-foreground absolute inset-0 flex items-center justify-center text-sm">
              {t("conversations.composer.emojiLoading")}
            </EmojiPicker.Loading>
            <EmojiPicker.Empty className="text-muted-foreground absolute inset-0 flex items-center justify-center text-sm">
              {t("conversations.composer.emojiEmpty")}
            </EmojiPicker.Empty>
            <EmojiPicker.List
              className="pb-2 select-none"
              components={{
                CategoryHeader: ({ category, ...props }) => (
                  <div className="bg-popover text-muted-foreground px-2 pt-2 pb-1 text-xs font-medium" {...props}>
                    {category.label}
                  </div>
                ),
                Row: ({ children, ...props }) => (
                  <div className="scroll-my-1 px-1" {...props}>
                    {children}
                  </div>
                ),
                Emoji: ({ emoji, ...props }) => (
                  <button
                    className={cn(
                      "flex size-8 items-center justify-center rounded-md text-lg",
                      emoji.isActive && "bg-accent"
                    )}
                    {...props}
                  >
                    {emoji.emoji}
                  </button>
                ),
              }}
            />
          </EmojiPicker.Viewport>
        </EmojiPicker.Root>
      </PopoverContent>
    </Popover>
  )
}
