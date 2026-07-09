import { zodResolver } from "@hookform/resolvers/zod"
import { useNavigate } from "@tanstack/react-router"
import { HashIcon, UserIcon, UsersIcon } from "lucide-react"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { usePostConversations, usePostConversationsDirect } from "@/api/conversations/conversations"
import { UserSelect } from "@/components/form/user-select"
import { UserMultiSelect } from "@/components/notifications/user-multi-select"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Field, FieldError, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"

import { useChatCache } from "./use-chat-cache"

// A single flat form spanning all three conversation types; `superRefine`
// applies the per-type required rules (partner for direct, name for group /
// channel). Keeping one schema lets the type toggle preserve entered fields.
const schema = z
  .object({
    type: z.enum(["direct", "group", "channel"]),
    user_id: z.string().optional(),
    name: z.string().max(255).optional(),
    description: z.string().max(1000).optional(),
    member_ids: z.array(z.string()),
  })
  .superRefine((v, ctx) => {
    if (v.type === "direct") {
      if (!v.user_id) {
        ctx.addIssue({ code: "custom", path: ["user_id"], params: { i18n: "validation.required" } })
      }
      return
    }
    if (!v.name || v.name.trim().length === 0) {
      ctx.addIssue({ code: "custom", path: ["name"], params: { i18n: "validation.required" } })
    }
  })

type FormValues = z.infer<typeof schema>

const DEFAULTS: FormValues = {
  type: "direct",
  user_id: undefined,
  name: "",
  description: "",
  member_ids: [],
}

interface NewConversationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  /** Whether the viewer may create groups/channels (conversations:manage). When
   * false only direct messages are offered. */
  canManage: boolean
}

/**
 * Start-a-conversation dialog. A type toggle (Direct / Group / Channel) swaps the
 * body: direct picks a single org user and hits the idempotent DM endpoint;
 * group/channel take a name (+ optional description) and members. On success we
 * refresh the sidebar and navigate straight into the new conversation.
 */
export function NewConversationDialog({ open, onOpenChange, canManage }: NewConversationDialogProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { invalidateConversations } = useChatCache()

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: DEFAULTS,
  })

  useEffect(() => {
    if (open) form.reset(DEFAULTS)
  }, [open])

  const type = form.watch("type")
  const memberIds = form.watch("member_ids")
  const userId = form.watch("user_id")

  function goToConversation(id?: string) {
    onOpenChange(false)
    invalidateConversations()
    if (id) navigate({ to: "/org/conversations/$conversationId", params: { conversationId: id } })
  }

  const directMutation = usePostConversationsDirect({
    mutation: {
      onSuccess: (res) => {
        if (res.status === 201) goToConversation(res.data.data?.id)
        else toast.error(t("conversations.new.error"))
      },
      onError: () => toast.error(t("conversations.new.error")),
    },
  })

  const createMutation = usePostConversations({
    mutation: {
      onSuccess: (res) => {
        if (res.status === 201) {
          toast.success(t("conversations.new.created"))
          goToConversation(res.data.data?.id)
        } else {
          toast.error(t("conversations.new.error"))
        }
      },
      onError: () => toast.error(t("conversations.new.error")),
    },
  })

  const isPending = directMutation.isPending || createMutation.isPending

  const onSubmit = form.handleSubmit((values) => {
    if (values.type === "direct") {
      if (!values.user_id) return
      directMutation.mutate({ data: { user_id: values.user_id } })
      return
    }
    createMutation.mutate({
      data: {
        type: values.type,
        name: values.name?.trim(),
        description: values.description?.trim() || undefined,
        member_ids: values.member_ids,
      },
    })
  })

  const errors = form.formState.errors

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[90dvh] flex-col sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("conversations.new.title")}</DialogTitle>
          <DialogDescription>{t("conversations.new.description")}</DialogDescription>
        </DialogHeader>

        <form onSubmit={onSubmit} className="flex min-h-0 flex-1 flex-col gap-4">
          {canManage && (
            <Tabs value={type} onValueChange={(value) => form.setValue("type", value as FormValues["type"])}>
              <TabsList className="w-full">
                <TabsTrigger value="direct">
                  <UserIcon data-icon="inline-start" />
                  {t("conversations.new.type.direct")}
                </TabsTrigger>
                <TabsTrigger value="group">
                  <UsersIcon data-icon="inline-start" />
                  {t("conversations.new.type.group")}
                </TabsTrigger>
                <TabsTrigger value="channel">
                  <HashIcon data-icon="inline-start" />
                  {t("conversations.new.type.channel")}
                </TabsTrigger>
              </TabsList>
            </Tabs>
          )}

          <div className="-mx-1 flex min-h-0 flex-1 flex-col gap-4 overflow-x-clip overflow-y-auto px-1">
            {type === "direct" ? (
              <Field data-invalid={!!errors.user_id || undefined}>
                <FieldLabel>{t("conversations.new.fields.user")}</FieldLabel>
                <UserSelect
                  value={userId}
                  onChange={(id) => form.setValue("user_id", id, { shouldValidate: true })}
                  placeholder={t("conversations.new.fields.userPlaceholder")}
                />
                <FieldError errors={[errors.user_id]} />
              </Field>
            ) : (
              <>
                <Field data-invalid={!!errors.name || undefined}>
                  <FieldLabel htmlFor="conversation-name">{t("conversations.new.fields.name")}</FieldLabel>
                  <Input
                    id="conversation-name"
                    {...form.register("name")}
                    placeholder={t("conversations.new.fields.namePlaceholder")}
                  />
                  <FieldError errors={[errors.name]} />
                </Field>

                <Field data-invalid={!!errors.description || undefined}>
                  <FieldLabel htmlFor="conversation-description">
                    {t("conversations.new.fields.description")}
                  </FieldLabel>
                  <Textarea
                    id="conversation-description"
                    rows={3}
                    {...form.register("description")}
                    placeholder={t("conversations.new.fields.descriptionPlaceholder")}
                  />
                  <FieldError errors={[errors.description]} />
                </Field>

                <Field>
                  <FieldLabel>{t("conversations.new.fields.members")}</FieldLabel>
                  <UserMultiSelect value={memberIds} onChange={(ids) => form.setValue("member_ids", ids)} />
                </Field>
              </>
            )}
          </div>

          <DialogFooter>
            <DialogClose render={<Button variant="outline" disabled={isPending} />}>{t("common.cancel")}</DialogClose>
            <Button type="submit" disabled={isPending}>
              {isPending && <Spinner />}
              {type === "direct" ? t("conversations.new.start") : t("common.create")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
