import type { MentionCandidate } from "./lib/mentions"

import { useTranslation } from "react-i18next"

import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { cn } from "@/lib/utils"

import { initials } from "./lib/avatar"

interface MentionPopoverProps {
  /** Filtered members to offer for the in-progress `@token`. */
  members: MentionCandidate[]
  /** Index of the keyboard-highlighted row (wraps in the input handler). */
  activeIndex: number
  /** Commit the given member into the composer. */
  onSelect: (member: MentionCandidate) => void
  /** Sync the active row to the hovered member (keeps mouse + keyboard in step). */
  onHover: (index: number) => void
}

/**
 * Presentational `@mention` autocomplete panel. It floats just above the
 * composer (the parent positions it) and renders the filtered member list; all
 * detection, filtering and keyboard state live in `<MessageInput>` so the
 * Enter-guard and caret math stay in one place. Selection is mouse- or
 * keyboard-driven — the highlighted row is owned by the parent via `activeIndex`.
 */
export function MentionPopover({ members, activeIndex, onSelect, onHover }: MentionPopoverProps) {
  const { t } = useTranslation()

  return (
    <div
      role="listbox"
      aria-label={t("conversations.composer.mentions")}
      className="bg-popover text-popover-foreground ring-foreground/10 absolute start-0 bottom-full z-50 mb-2 max-h-56 w-64 overflow-y-auto rounded-lg p-1 shadow-md ring-1"
    >
      {members.map((member, index) => (
        <button
          key={member.id}
          type="button"
          role="option"
          aria-selected={index === activeIndex}
          // Keep focus on the textarea — commit on mousedown before blur fires.
          onMouseDown={(e) => {
            e.preventDefault()
            onSelect(member)
          }}
          onMouseMove={() => onHover(index)}
          className={cn(
            "flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-start text-sm transition-colors",
            index === activeIndex ? "bg-accent text-accent-foreground" : "hover:bg-accent/60"
          )}
        >
          <Avatar className="size-6 shrink-0">
            <AvatarFallback className="bg-muted text-[10px] font-medium">{initials(member.name)}</AvatarFallback>
          </Avatar>
          <span className="truncate">{member.name}</span>
          <span className="text-muted-foreground ms-auto shrink-0 font-mono text-xs">@{member.username}</span>
        </button>
      ))}
    </div>
  )
}
