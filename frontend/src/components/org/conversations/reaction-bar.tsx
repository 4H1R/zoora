import type { ChatMessage } from "./lib/messages"

import { useTranslation } from "react-i18next"

import { cn } from "@/lib/utils"
import { useChatReactions } from "@/stores/chat-reactions"

import { useToggleReaction } from "./use-toggle-reaction"

interface ReactionBarProps {
  message: ChatMessage
  convId: string
}

/**
 * Compact pills for a message's reactions — emoji + count — sorted by count
 * desc, then emoji for a stable tiebreak. Clicking a pill re-toggles that emoji
 * through the shared mutation. Pills the signed-in user toggled on THIS SESSION
 * are highlighted (the server stores counts only, so this cannot survive a
 * reload). Renders nothing when the message has no reactions.
 */
export function ReactionBar({ message, convId }: ReactionBarProps) {
  const { t } = useTranslation()
  const { toggle } = useToggleReaction(convId)

  const messageId = message.id ?? ""
  const ownSet = useChatReactions((s) => s.own[messageId])

  const entries = Object.entries(message.reactions ?? {})
    .filter(([, count]) => count > 0)
    .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))

  if (entries.length === 0) return null

  return (
    <div className="mt-1 flex flex-wrap items-center gap-1">
      {entries.map(([emoji, count]) => {
        const reacted = ownSet?.has(emoji) ?? false
        return (
          <button
            key={emoji}
            type="button"
            onClick={() => toggle(messageId, emoji)}
            aria-label={t("conversations.reactions.toggle", { emoji })}
            aria-pressed={reacted}
            className={cn(
              "flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs leading-none transition",
              reacted
                ? "border-primary bg-primary/10 text-primary font-medium"
                : "bg-muted text-muted-foreground hover:bg-accent border-transparent"
            )}
          >
            <span className="text-sm leading-none">{emoji}</span>
            <span className="tabular-nums">{count}</span>
          </button>
        )
      })}
    </div>
  )
}
