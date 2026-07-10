import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

import { Link } from "@tanstack/react-router"
import { BellOffIcon, HashIcon, UsersIcon } from "lucide-react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { formatRelativeTime } from "@/lib/relative-time"
import { cn } from "@/lib/utils"

import { conversationTitle } from "./lib/conversation-title"
import { PresenceDot } from "./presence-dot"

// Server hands us a stable `color_index` per conversation; map it onto a fixed
// palette of soft tints (Tailwind scale, light + dark) so a given conversation
// keeps the same accent everywhere. No hashing, no arbitrary hex.
const AVATAR_TINTS = [
  "bg-rose-100 text-rose-700 dark:bg-rose-500/20 dark:text-rose-200",
  "bg-sky-100 text-sky-700 dark:bg-sky-500/20 dark:text-sky-200",
  "bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-200",
  "bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-200",
  "bg-violet-100 text-violet-700 dark:bg-violet-500/20 dark:text-violet-200",
  "bg-cyan-100 text-cyan-700 dark:bg-cyan-500/20 dark:text-cyan-200",
  "bg-pink-100 text-pink-700 dark:bg-pink-500/20 dark:text-pink-200",
  "bg-indigo-100 text-indigo-700 dark:bg-indigo-500/20 dark:text-indigo-200",
] as const

function tintFor(index?: number): string {
  const len = AVATAR_TINTS.length
  return AVATAR_TINTS[(((index ?? 0) % len) + len) % len]
}

// Up to two initials (first + last token) for the avatar fallback. Works for
// Persian and Latin names alike; toUpperCase is a no-op on scripts without case.
function initials(name?: string): string {
  const parts = (name ?? "").trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return "؟"
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase()
}

// Group / channel conversations get a small type glyph badged onto the avatar so
// they read differently from 1:1 DMs at a glance.
const TYPE_GLYPH = {
  group: UsersIcon,
  channel: HashIcon,
} as const

interface ConversationItemProps {
  conversation: Conversation
  isActive: boolean
  /** DM partner online state; `undefined` for non-DM rows or unknown presence. */
  presenceOnline?: boolean
  /** Whether the viewer has muted this conversation — shows a muted glyph. */
  muted?: boolean
}

export function ConversationItem({ conversation, isActive, presenceOnline, muted }: ConversationItemProps) {
  const { t, i18n } = useTranslation()
  const { user } = useAccess()

  const id = conversation.id ?? ""
  const name = conversationTitle(conversation, user.id)
  const unreadCount = conversation.unread_count ?? 0
  const unread = unreadCount > 0
  const last = conversation.last_message
  const TypeIcon = conversation.type ? TYPE_GLYPH[conversation.type as keyof typeof TYPE_GLYPH] : undefined

  // In groups/channels prefix the sender; DMs need no sender label.
  const senderName = last?.sender?.name
  const showSender = conversation.type !== "direct" && Boolean(senderName)
  const timestamp = last?.created_at ?? conversation.updated_at

  return (
    <Link
      to="/org/conversations/$conversationId"
      params={{ conversationId: id }}
      className={cn(
        "group relative flex items-center gap-3 rounded-xl px-2.5 py-2.5 transition-colors",
        "hover:bg-muted/60 focus-visible:bg-muted/60 focus-visible:outline-hidden",
        isActive ? "bg-muted" : unread && "bg-primary/[0.035]"
      )}
    >
      {/* Active accent bar pinned to the start edge (RTL-safe via `start-0`). */}
      {isActive && <span className="bg-primary absolute inset-y-2 start-0 w-0.5 rounded-full" aria-hidden />}

      <div className="relative shrink-0">
        <Avatar className="size-10">
          {conversation.avatar_url && <AvatarImage src={conversation.avatar_url} alt={name} />}
          <AvatarFallback className={cn("text-xs font-semibold", tintFor(conversation.color_index))}>
            {initials(name)}
          </AvatarFallback>
        </Avatar>
        {TypeIcon && (
          <span className="bg-background ring-background text-muted-foreground absolute -end-0.5 -bottom-0.5 flex size-4 items-center justify-center rounded-full ring-2">
            <TypeIcon className="size-2.5" />
          </span>
        )}
        {/* Online dot for DMs (no type glyph to collide with). */}
        {presenceOnline !== undefined && (
          <PresenceDot online={presenceOnline} className="absolute -end-0.5 -bottom-0.5" />
        )}
      </div>

      <div className="flex min-w-0 flex-1 flex-col gap-0.5">
        <div className="flex items-center justify-between gap-2">
          <span className="flex min-w-0 items-center gap-1.5">
            <span
              className={cn(
                "truncate text-sm leading-tight",
                unread ? "text-foreground font-semibold" : "text-foreground/90 font-medium"
              )}
            >
              {name}
            </span>
            {muted && (
              <BellOffIcon
                className="text-muted-foreground size-3.5 shrink-0"
                aria-label={t("conversations.item.muted")}
              />
            )}
          </span>
          {timestamp && (
            <time className="text-muted-foreground shrink-0 font-mono text-[11px] tabular-nums">
              {formatRelativeTime(timestamp, i18n.language)}
            </time>
          )}
        </div>

        <div className="flex items-center justify-between gap-2">
          <p
            className={cn(
              "min-w-0 flex-1 truncate text-xs leading-tight",
              unread ? "text-foreground/75" : "text-muted-foreground"
            )}
          >
            {last?.content ? (
              <>
                {showSender && <span className="font-medium">{senderName}: </span>}
                {last.content}
              </>
            ) : (
              <span className="italic">{t("conversations.list.noMessages")}</span>
            )}
          </p>
          {unread && (
            <Badge className="h-5 min-w-5 shrink-0 justify-center rounded-full px-1.5 text-[11px] font-semibold tabular-nums">
              {unreadCount > 99 ? "99+" : unreadCount}
            </Badge>
          )}
        </div>
      </div>
    </Link>
  )
}
