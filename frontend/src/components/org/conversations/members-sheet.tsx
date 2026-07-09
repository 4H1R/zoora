import type { GithubCom4H1RZooraInternalDomainConversationMember as ConversationMember } from "@/api/model"

import { UserMinusIcon } from "lucide-react"
import { useState } from "react"
import { useAccess } from "react-access-engine"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  useDeleteConversationsIdMembersUserId,
  useGetConversationsIdMembers,
  usePostConversationsIdMembers,
} from "@/api/conversations/conversations"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { UserSelect } from "@/components/form/user-select"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { Spinner } from "@/components/ui/spinner"
import { cn } from "@/lib/utils"

import { avatarTint, initials } from "./lib/avatar"
import { useChatCache } from "./use-chat-cache"

interface MembersSheetProps {
  convId: string
  open: boolean
  onOpenChange: (open: boolean) => void
  /** Whether the viewer may add/remove members (conversations:manage). */
  canManage: boolean
}

/**
 * Slide-over roster for a conversation: every member with avatar, name and role.
 * Managers get an add-member picker at the top and a per-row remove action
 * (confirmed); the viewer's own row is never removable here (they Leave instead).
 */
export function MembersSheet({ convId, open, onOpenChange, canManage }: MembersSheetProps) {
  const { t } = useTranslation()
  const { user } = useAccess()
  const { invalidateMembers, invalidateConversations } = useChatCache()
  const [pendingRemoval, setPendingRemoval] = useState<ConversationMember | null>(null)

  const { data, isLoading } = useGetConversationsIdMembers(convId, {
    query: { enabled: open },
  })
  const members = data?.status === 200 ? (data.data.data ?? []) : []

  const addMutation = usePostConversationsIdMembers({
    mutation: {
      onSuccess: (res) => {
        if (res.status === 201) {
          toast.success(t("conversations.members.addSuccess"))
          invalidateMembers(convId)
          invalidateConversations()
        } else {
          toast.error(t("conversations.members.addError"))
        }
      },
      onError: () => toast.error(t("conversations.members.addError")),
    },
  })

  const removeMutation = useDeleteConversationsIdMembersUserId({
    mutation: {
      onSuccess: () => {
        toast.success(t("conversations.members.removeSuccess"))
        invalidateMembers(convId)
        invalidateConversations()
        setPendingRemoval(null)
      },
      onError: () => toast.error(t("conversations.members.removeError")),
    },
  })

  function handleAdd(userId: string) {
    if (members.some((m) => (m.user_id ?? m.user?.id) === userId)) return
    addMutation.mutate({ id: convId, data: { user_id: userId, role: "member" } })
  }

  function confirmRemove() {
    const targetId = pendingRemoval?.user_id ?? pendingRemoval?.user?.id
    if (targetId) removeMutation.mutate({ id: convId, userId: targetId })
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full gap-0 p-0 sm:max-w-sm">
        <SheetHeader className="border-b px-4 py-4">
          <SheetTitle>{t("conversations.members.title")}</SheetTitle>
          <SheetDescription>{t("conversations.members.count", { count: members.length })}</SheetDescription>
        </SheetHeader>

        {canManage && (
          <div className="border-b px-4 py-3">
            <UserSelect onChange={handleAdd} placeholder={t("conversations.members.addPlaceholder")} />
          </div>
        )}

        <div className="flex min-h-0 flex-1 flex-col gap-0.5 overflow-y-auto p-2">
          {isLoading ? (
            <div className="flex flex-1 items-center justify-center py-10">
              <Spinner className="text-muted-foreground size-5" />
            </div>
          ) : (
            members.map((member) => {
              const memberUserId = member.user_id ?? member.user?.id ?? ""
              const name = member.user?.name ?? ""
              const isSelf = memberUserId === user.id
              return (
                <div
                  key={member.id ?? memberUserId}
                  className="group hover:bg-muted/50 flex items-center gap-3 rounded-lg px-2 py-2"
                >
                  <Avatar className="size-9 shrink-0">
                    <AvatarFallback className={cn("text-xs font-semibold", avatarTint(memberUserId))}>
                      {initials(name)}
                    </AvatarFallback>
                  </Avatar>

                  <div className="flex min-w-0 flex-1 flex-col">
                    <span className="truncate text-sm font-medium">
                      {name}
                      {isSelf && (
                        <span className="text-muted-foreground ms-1 font-normal">{t("conversations.members.you")}</span>
                      )}
                    </span>
                    {member.user?.username && (
                      <span className="text-muted-foreground truncate font-mono text-xs">{member.user.username}</span>
                    )}
                  </div>

                  {member.role && (
                    <Badge variant="secondary" className="shrink-0 capitalize">
                      {t(`conversations.members.role.${member.role}`)}
                    </Badge>
                  )}

                  {canManage && !isSelf && (
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      className="text-muted-foreground hover:text-destructive shrink-0"
                      aria-label={t("conversations.members.remove")}
                      onClick={() => setPendingRemoval(member)}
                    >
                      <UserMinusIcon />
                    </Button>
                  )}
                </div>
              )
            })
          )}
        </div>
      </SheetContent>

      <DeleteConfirmDialog
        open={pendingRemoval !== null}
        onOpenChange={(next) => !next && setPendingRemoval(null)}
        resourceName={pendingRemoval?.user?.name ?? ""}
        onConfirm={confirmRemove}
        isLoading={removeMutation.isPending}
      />
    </Sheet>
  )
}
