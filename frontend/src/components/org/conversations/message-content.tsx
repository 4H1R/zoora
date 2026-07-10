import { useProfileCard } from "@/stores/profile-card"
import { cn } from "@/lib/utils"

import { splitMentions } from "./lib/mentions"

interface MessageContentProps {
  content: string
  /** Own bubbles sit on the accent surface, so mentions get a lighter tint. */
  isOwn: boolean
}

/**
 * Renders message text with clickable `@username` mentions. Every `@username`
 * token (3-30 charset) is tinted and, on click, opens the profile card which
 * lazily resolves the handle through the directory. Whitespace preserved.
 */
export function MessageContent({ content, isOwn }: MessageContentProps) {
  const openCard = useProfileCard((s) => s.open)
  const segments = splitMentions(content)

  return (
    <p dir="auto" className="break-words whitespace-pre-wrap text-start">
      {segments.map((segment, index) =>
        segment.isMention ? (
          <button
            key={index}
            type="button"
            onClick={() => segment.username && openCard({ username: segment.username })}
            className={cn(
              "cursor-pointer rounded-sm px-0.5 font-medium",
              isOwn
                ? "bg-primary-foreground/20 text-primary-foreground"
                : "bg-primary/10 text-primary hover:bg-primary/20"
            )}
          >
            {segment.text}
          </button>
        ) : (
          <span key={index}>{segment.text}</span>
        )
      )}
    </p>
  )
}
