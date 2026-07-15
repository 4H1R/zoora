import { useNavigate } from "@tanstack/react-router"
import { SearchIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { useDebounce } from "use-debounce"

import { useGetConversationsDirectory, usePostConversationsDirect } from "@/api/conversations/conversations"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Spinner } from "@/components/ui/spinner"
import { cn } from "@/lib/utils"

import { avatarTint, initials } from "./lib/avatar"
import { useChatCache } from "./use-chat-cache"

interface NewDirectDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

/**
 * Simple "start a direct message" picker: search the member-safe user directory
 * (works for any viewer, no users:view needed) and pick one person. Selecting a
 * row hits the idempotent DM endpoint and navigates into the conversation —
 * reusing an existing DM if one already exists. Available to every viewer;
 * group/channel creation stays behind conversations:manage.
 */
export function NewDirectDialog({ open, onOpenChange }: NewDirectDialogProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { invalidateConversations } = useChatCache()

  const [search, setSearch] = useState("")
  const [debouncedSearch] = useDebounce(search, 300)

  useEffect(() => {
    if (open) setSearch("")
  }, [open])

  // Directory membership changes rarely; a modest staleTime serves quick
  // close/reopen cycles (and the reset-to-empty term) from cache instead of
  // refetching the whole roster each time the dialog opens.
  const directoryQuery = useGetConversationsDirectory(
    { search: debouncedSearch || undefined },
    { query: { enabled: open, staleTime: 30_000 } }
  )
  const people = directoryQuery.data?.status === 200 ? (directoryQuery.data.data.data ?? []) : []

  const directMutation = usePostConversationsDirect({
    mutation: {
      onSuccess: (res) => {
        if (res.status === 201 && res.data.data?.id) {
          const id = res.data.data.id
          onOpenChange(false)
          invalidateConversations()
          navigate({
            to: "/org/conversations/$conversationId",
            params: { conversationId: id },
            search: {},
          })
        } else {
          toast.error(t("conversations.direct.error"))
        }
      },
      onError: () => toast.error(t("conversations.direct.error")),
    },
  })

  const isLoading = directoryQuery.isFetching
  const isEmpty = !isLoading && people.length === 0

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[90dvh] flex-col sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("conversations.direct.title")}</DialogTitle>
          <DialogDescription>{t("conversations.direct.description")}</DialogDescription>
        </DialogHeader>

        <div className="relative">
          <SearchIcon className="text-muted-foreground pointer-events-none absolute start-2.5 top-1/2 size-4 -translate-y-1/2" />
          <Input
            autoFocus
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t("conversations.direct.searchPlaceholder")}
            className="h-9 ps-9"
          />
        </div>

        <ScrollArea className="min-h-0 flex-1">
          <div className="flex flex-col gap-0.5 py-1">
            {isLoading && (
              <div className="flex items-center justify-center py-8">
                <Spinner className="text-muted-foreground size-5" />
              </div>
            )}

            {isEmpty && (
              <p className="text-muted-foreground py-8 text-center text-sm">
                {debouncedSearch ? t("conversations.direct.empty") : t("conversations.direct.hint")}
              </p>
            )}

            {!isLoading &&
              people.map((person) => (
                <button
                  key={person.id}
                  type="button"
                  disabled={directMutation.isPending}
                  onClick={() => person.id && directMutation.mutate({ data: { user_id: person.id } })}
                  className={cn(
                    "hover:bg-muted flex items-center gap-3 rounded-lg px-2 py-2 text-start transition",
                    "disabled:pointer-events-none disabled:opacity-60"
                  )}
                >
                  <Avatar className="size-9">
                    <AvatarFallback className={cn("text-xs font-semibold", avatarTint(person.id))}>
                      {initials(person.name ?? "")}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex min-w-0 flex-col">
                    <span className="truncate text-sm font-medium">{person.name}</span>
                    {person.username && (
                      <span className="text-muted-foreground truncate font-mono text-xs">@{person.username}</span>
                    )}
                  </div>
                </button>
              ))}
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  )
}
