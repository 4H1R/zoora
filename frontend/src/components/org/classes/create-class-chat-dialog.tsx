import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { HashIcon, MessagesSquareIcon, UsersIcon } from "lucide-react"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { getGetClassesIdQueryKey, usePostClassesIdConversation } from "@/api/classes/classes"
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

const schema = z
  .object({
    type: z.enum(["group", "channel"]),
    name: z.string().max(255).optional(),
  })
  .superRefine((v, ctx) => {
    if (!v.name || v.name.trim().length === 0) {
      ctx.addIssue({ code: "custom", path: ["name"], params: { i18n: "validation.required" } })
    }
  })

type FormValues = z.infer<typeof schema>

interface CreateClassChatDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  classId: string
  /** Prefills the conversation name (defaults to the class name). */
  className: string
}

/**
 * Creates a class's group/channel chat and seeds it with the teacher (admin) and
 * every enrolled student in one action. Authorized by class ownership on the
 * backend, so it appears for teachers/managers even without conversations:manage.
 * On success we jump straight into the new conversation.
 */
export function CreateClassChatDialog({ open, onOpenChange, classId, className }: CreateClassChatDialogProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { type: "group", name: className },
  })

  // Re-seed the name from the class each time the dialog opens.
  useEffect(() => {
    if (open) form.reset({ type: "group", name: className })
  }, [open, className])

  const type = form.watch("type")
  const errors = form.formState.errors

  const createMutation = usePostClassesIdConversation({
    mutation: {
      onSuccess: (res) => {
        if (res.status === 201) {
          toast.success(t("org.class.chat.created"))
          onOpenChange(false)
          queryClient.invalidateQueries({ queryKey: getGetClassesIdQueryKey(classId) })
          const id = res.data.data?.id
          if (id) navigate({ to: "/org/conversations/$conversationId", params: { conversationId: id } })
        } else {
          toast.error(t("org.class.chat.error"))
        }
      },
      onError: () => toast.error(t("org.class.chat.error")),
    },
  })

  const isPending = createMutation.isPending

  const onSubmit = form.handleSubmit((values) => {
    createMutation.mutate({
      id: classId,
      data: { type: values.type, name: values.name?.trim() },
    })
  })

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <MessagesSquareIcon className="size-5" />
            {t("org.class.chat.title")}
          </DialogTitle>
          <DialogDescription>{t("org.class.chat.description")}</DialogDescription>
        </DialogHeader>

        <form onSubmit={onSubmit} className="flex flex-col gap-4">
          <Tabs value={type} onValueChange={(value) => form.setValue("type", value as FormValues["type"])}>
            <TabsList className="w-full">
              <TabsTrigger value="group">
                <UsersIcon data-icon="inline-start" />
                {t("org.class.chat.type.group")}
              </TabsTrigger>
              <TabsTrigger value="channel">
                <HashIcon data-icon="inline-start" />
                {t("org.class.chat.type.channel")}
              </TabsTrigger>
            </TabsList>
          </Tabs>

          <p className="text-muted-foreground text-xs leading-relaxed">
            {type === "channel" ? t("org.class.chat.type.channelHint") : t("org.class.chat.type.groupHint")}
          </p>

          <Field data-invalid={!!errors.name || undefined}>
            <FieldLabel htmlFor="class-chat-name">{t("org.class.chat.fields.name")}</FieldLabel>
            <Input
              id="class-chat-name"
              {...form.register("name")}
              placeholder={t("org.class.chat.fields.namePlaceholder")}
            />
            <FieldError errors={[errors.name]} />
          </Field>

          <DialogFooter>
            <DialogClose render={<Button variant="outline" disabled={isPending} />}>{t("common.cancel")}</DialogClose>
            <Button type="submit" disabled={isPending}>
              {isPending && <Spinner />}
              {t("org.class.chat.submit")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
