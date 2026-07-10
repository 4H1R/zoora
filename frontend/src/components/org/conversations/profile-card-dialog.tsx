import { useNavigate } from "@tanstack/react-router"
import { useRef } from "react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  useGetConversationsDirectoryUsername,
  useGetConversationsPresence,
  usePostConversationsDirect,
} from "@/api/conversations/conversations"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Spinner } from "@/components/ui/spinner"
import { useProfileCard } from "@/stores/profile-card"
import { cn } from "@/lib/utils"

import { avatarTint, initials } from "./lib/avatar"
import { PresenceDot } from "./presence-dot"

/**
 * Single reusable profile card. Opens from search, avatars, member rows and
 * @mention clicks via `useProfileCard`. When only a username is known (mention
 * click) it resolves the user through the directory endpoint first. "Send
 * message" hits the idempotent DM endpoint and navigates into the conversation.
 */
export function ProfileCardDialog() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { user } = useAccess()
  const { target, close } = useProfileCard()

  const open = target !== null
  const needsResolve = open && !target?.userId && !!target?.username

  const resolveQuery = useGetConversationsDirectoryUsername(target?.username ?? "", {
    query: { enabled: needsResolve },
  })
  const resolved = resolveQuery.data?.status === 200 ? resolveQuery.data.data.data : undefined

  const userId = target?.userId ?? resolved?.id
  const name = resolved?.name ?? target?.name ?? ""
  const username = resolved?.username ?? target?.username

  const presenceQuery = useGetConversationsPresence(
    { user_ids: userId ?? "" },
    { query: { enabled: open && !!userId } }
  )
  const presence =
    userId && presenceQuery.data?.status === 200 ? presenceQuery.data.data.data?.[userId] : undefined

  const isSelf = !!userId && userId === user.id

  const directMutation = usePostConversationsDirect({
    mutation: {
      onSuccess: (res) => {
        if (res.status === 201 && res.data.data?.id) {
          const id = res.data.data.id
          close()
          navigate({
            to: "/org/conversations/$conversationId",
            params: { conversationId: id },
            search: {},
          })
        } else {
          toast.error(t("conversations.profile.error"))
        }
      },
      onError: () => toast.error(t("conversations.profile.error")),
    },
  })

  const notFound = needsResolve && resolveQuery.data?.status === 404
  const loading = needsResolve && resolveQuery.isFetching

  // Keep the last resolved card so the "not found" fallback never flashes
  // during the dialog's close animation, when `target` is already null.
  const lastCard = useRef<{
    userId: string
    name: string
    username?: string
    isSelf: boolean
  } | null>(null)
  if (open && userId) {
    lastCard.current = { userId, name, username, isSelf }
  }
  const card = open ? (userId ? { userId, name, username, isSelf } : null) : lastCard.current

  return (
    <Dialog open={open} onOpenChange={(next) => !next && close()}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader className="sr-only">
          <DialogTitle>
            {card?.name || card?.username || t("conversations.profile.title")}
          </DialogTitle>
        </DialogHeader>

        {loading ? (
          <div className="flex items-center justify-center py-10">
            <Spinner className="text-muted-foreground size-5" />
          </div>
        ) : !card ? (
          notFound ? (
            <p className="text-muted-foreground py-8 text-center text-sm">
              {t("conversations.profile.notFound")}
            </p>
          ) : null
        ) : (
          <div className="flex flex-col items-center gap-3 py-2">
            <div className="relative">
              <Avatar className="size-20">
                <AvatarFallback className={cn("text-2xl font-semibold", avatarTint(card.userId))}>
                  {initials(card.name)}
                </AvatarFallback>
              </Avatar>
              {presence && (
                <span className="absolute bottom-1 end-1">
                  <PresenceDot online={presence.online} />
                </span>
              )}
            </div>

            <div className="flex flex-col items-center gap-0.5 text-center">
              <span className="text-lg font-semibold">{card.name}</span>
              {card.username && (
                <span className="text-muted-foreground font-mono text-sm">@{card.username}</span>
              )}
            </div>

            {card.isSelf ? (
              <span className="text-muted-foreground text-sm">{t("conversations.profile.you")}</span>
            ) : (
              <Button
                className="w-full"
                disabled={directMutation.isPending}
                onClick={() => directMutation.mutate({ data: { user_id: card.userId } })}
              >
                {directMutation.isPending && <Spinner />}
                {t("conversations.profile.sendMessage")}
              </Button>
            )}
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
