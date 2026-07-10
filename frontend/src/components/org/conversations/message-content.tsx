import type { MentionCandidate } from "./lib/mentions"

import { cn } from "@/lib/utils"

import { highlightMentions } from "./lib/mentions"

interface MessageContentProps {
  content: string
  /** Conversation members, used to resolve `@<DisplayName>` highlight spans. */
  members: MentionCandidate[]
  /** Own bubbles sit on the accent surface, so mentions get a lighter tint. */
  isOwn: boolean
}

/**
 * Renders message text with best-effort @mention highlighting. The content is
 * split into plain + mention segments by `highlightMentions` (longest-name-first,
 * every occurrence), then each mention is wrapped in a subtly-tinted span.
 * Whitespace is preserved verbatim (`whitespace-pre-wrap`) so multi-line and
 * indented messages read exactly as typed.
 */
export function MessageContent({ content, members, isOwn }: MessageContentProps) {
  const segments = highlightMentions(content, members)

  return (
    <p className="break-words whitespace-pre-wrap">
      {segments.map((segment, index) =>
        segment.isMention ? (
          <span
            key={index}
            className={cn(
              "rounded-sm px-0.5 font-medium",
              isOwn ? "bg-primary-foreground/20 text-primary-foreground" : "bg-primary/10 text-primary"
            )}
          >
            {segment.text}
          </span>
        ) : (
          <span key={index}>{segment.text}</span>
        )
      )}
    </p>
  )
}
