import type { MuteDuration } from "./lib/mute"
import type { GithubCom4H1RZooraInternalDomainConversation as Conversation } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useNavigate } from "@tanstack/react-router"
import { BellIcon, BellOffIcon, LogOutIcon, Trash2Icon } from "lucide-react"
import { useEffect, useState } from "react"
import { useAccess } from "react-access-engine"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  useDeleteConversationsId,
  useGetConversationsIdMembers,
  usePatchConversationsId,
  usePostConversationsIdLeave,
  usePostConversationsIdMute,
} from "@/api/conversations/conversations"
import { FormSaveBar } from "@/components/form-save-bar"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { Field, FieldError, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { Spinner } from "@/components/ui/spinner"

import { isMuted, muteUntilISO } from "./lib/mute"
import { useChatCache } from "./use-chat-cache"

const detailsSchema = z.object({
  name: z.string().min(1).max(255),
})

type DetailsValues = z.infer<typeof detailsSchema>

const MUTE_DURATIONS: MuteDuration[] = ["1h", "8h", "1w", "forever"]

interface ConversationSettingsProps {
  conversation: Conversation
  open: boolean
  onOpenChange: (open: boolean) => void
  /** Whether the viewer may rename / delete (conversations:manage). */
  canManage: boolean
}

/**
 * Per-conversation settings slide-over: rename (managers only),
 * notification muting (any member, with quick presets), and the leave / delete
 * actions. Leave and delete confirm, then navigate back to the list.
 */
export function ConversationSettings({ conversation, open, onOpenChange, canManage }: ConversationSettingsProps) {
  const { t } = useTranslation()
  const { user } = useAccess()
  const navigate = useNavigate()
  const { invalidateConversations, invalidateMembers } = useChatCache()
  const [confirmLeave, setConfirmLeave] = useState(false)
  const [confirmDelete, setConfirmDelete] = useState(false)

  const convId = conversation.id ?? ""
  const isDirect = conversation.type === "direct"
  const canEditDetails = canManage && !isDirect

  // Viewer's own mute state, read from the members roster.
  const { data: membersData } = useGetConversationsIdMembers(convId, { query: { enabled: open } })
  const members = membersData?.status === 200 ? (membersData.data.data ?? []) : []
  const selfMutedUntil = members.find((m) => (m.user_id ?? m.user?.id) === user.id)?.muted_until
  const muted = isMuted(selfMutedUntil)

  const form = useForm<DetailsValues>({
    resolver: zodResolver(detailsSchema),
    defaultValues: { name: conversation.name ?? "" },
  })

  useEffect(() => {
    if (open) form.reset({ name: conversation.name ?? "" })
  }, [open, conversation.id])

  const patchMutation = usePatchConversationsId({
    mutation: {
      onSuccess: (res, vars) => {
        if (res.status === 200) {
          toast.success(t("conversations.settings.saved"))
          form.reset({ name: vars.data.name ?? "" })
          invalidateConversations()
        } else {
          toast.error(t("conversations.settings.saveError"))
        }
      },
      onError: () => toast.error(t("conversations.settings.saveError")),
    },
  })

  const muteMutation = usePostConversationsIdMute({
    mutation: {
      onSuccess: () => {
        invalidateMembers(convId)
        invalidateConversations()
      },
      onError: () => toast.error(t("conversations.settings.mute.error")),
    },
  })

  const leaveMutation = usePostConversationsIdLeave({
    mutation: {
      onSuccess: () => {
        toast.success(t("conversations.settings.leave.success"))
        invalidateConversations()
        onOpenChange(false)
        navigate({ to: "/org/conversations" })
      },
      onError: () => toast.error(t("conversations.settings.leave.error")),
    },
  })

  const deleteMutation = useDeleteConversationsId({
    mutation: {
      onSuccess: () => {
        toast.success(t("conversations.settings.delete.success"))
        invalidateConversations()
        onOpenChange(false)
        navigate({ to: "/org/conversations" })
      },
      onError: () => toast.error(t("conversations.settings.delete.error")),
    },
  })

  const onSaveDetails = form.handleSubmit((values) => {
    patchMutation.mutate({
      id: convId,
      data: { name: values.name.trim() },
    })
  })

  function applyMute(duration: MuteDuration) {
    muteMutation.mutate({ id: convId, data: { muted_until: muteUntilISO(duration) } })
  }

  function unmute() {
    muteMutation.mutate({ id: convId, data: { muted_until: undefined } })
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full gap-0 overflow-y-auto p-0 sm:max-w-sm">
        <SheetHeader className="border-b px-4 py-4">
          <SheetTitle>{t("conversations.settings.title")}</SheetTitle>
          <SheetDescription>{t("conversations.settings.description")}</SheetDescription>
        </SheetHeader>

        <div className="flex flex-col gap-6 p-4">
          {canEditDetails && (
            <form onSubmit={onSaveDetails} className="flex flex-col gap-3">
              <Field data-invalid={!!form.formState.errors.name || undefined}>
                <FieldLabel htmlFor="settings-name">{t("conversations.new.fields.name")}</FieldLabel>
                <Input id="settings-name" {...form.register("name")} />
                <FieldError errors={[form.formState.errors.name]} />
              </Field>
            </form>
          )}

          {/* Notifications: mute presets + unmute — available to any member. */}
          <div className="flex items-center justify-between gap-3">
            <div className="flex min-w-0 flex-col">
              <span className="text-sm font-medium">{t("conversations.settings.mute.label")}</span>
              <span className="text-muted-foreground text-xs">
                {muted ? t("conversations.settings.mute.on") : t("conversations.settings.mute.off")}
              </span>
            </div>
            {muted ? (
              <Button type="button" variant="outline" size="sm" disabled={muteMutation.isPending} onClick={unmute}>
                <BellIcon />
                {t("conversations.settings.mute.unmute")}
              </Button>
            ) : (
              <DropdownMenu>
                <DropdownMenuTrigger
                  render={
                    <Button type="button" variant="outline" size="sm" disabled={muteMutation.isPending}>
                      <BellOffIcon />
                      {t("conversations.settings.mute.action")}
                    </Button>
                  }
                />
                <DropdownMenuContent align="end" className="min-w-44">
                  {MUTE_DURATIONS.map((duration) => (
                    <DropdownMenuItem key={duration} onClick={() => applyMute(duration)}>
                      {t(`conversations.settings.mute.duration.${duration}`)}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            )}
          </div>

          {/* Destructive actions. Leaving a DM is not a thing (the pair is
              fixed server-side and the backend rejects it), so the action only
              renders for groups/channels. */}
          <div className="flex flex-col gap-2 border-t pt-4">
            {!isDirect && (
              <Button
                type="button"
                variant="outline"
                className="text-destructive hover:text-destructive justify-start"
                onClick={() => setConfirmLeave(true)}
              >
                <LogOutIcon className="rtl:-scale-x-100" />
                {t("conversations.settings.leave.action")}
              </Button>
            )}
            {canManage && (
              <Button
                type="button"
                variant="outline"
                className="text-destructive hover:bg-destructive/10 hover:text-destructive justify-start"
                onClick={() => setConfirmDelete(true)}
              >
                <Trash2Icon />
                {t("conversations.settings.delete.action")}
              </Button>
            )}
          </div>
        </div>

        {canEditDetails && (
          <FormSaveBar form={form} onSave={onSaveDetails} isPending={patchMutation.isPending} />
        )}
      </SheetContent>

      <AlertDialog open={confirmLeave} onOpenChange={setConfirmLeave}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("conversations.settings.leave.confirmTitle")}</AlertDialogTitle>
            <AlertDialogDescription>{t("conversations.settings.leave.confirmDescription")}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={leaveMutation.isPending}>{t("common.cancel")}</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={leaveMutation.isPending}
              onClick={() => leaveMutation.mutate({ id: convId })}
            >
              {leaveMutation.isPending && <Spinner />}
              {t("conversations.settings.leave.action")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={confirmDelete} onOpenChange={setConfirmDelete}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("conversations.settings.delete.confirmTitle")}</AlertDialogTitle>
            <AlertDialogDescription>{t("conversations.settings.delete.confirmDescription")}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteMutation.isPending}>{t("common.cancel")}</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              disabled={deleteMutation.isPending}
              onClick={() => deleteMutation.mutate({ id: convId })}
            >
              {deleteMutation.isPending && <Spinner />}
              {t("common.delete")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Sheet>
  )
}
