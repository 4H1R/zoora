import type { AudienceMode } from "@/components/notifications/audience-picker"
import type { GithubCom4H1RZooraInternalDomainNotificationAudienceDTO as AudienceDTO } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { SendIcon } from "lucide-react"
import { useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { GithubCom4H1RZooraInternalDomainNotificationAudienceDTOType as AudienceType } from "@/api/model"
import {
  getGetNotificationsSentQueryKey,
  usePostNotifications,
} from "@/api/notifications/notifications"
import { AudiencePicker } from "@/components/notifications/audience-picker"
import { Button } from "@/components/ui/button"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"

const schema = z.object({
  title: z.string().min(1).max(255),
  body: z.string().min(1).max(4000),
  action_url: z.string().url().max(500).optional().or(z.literal("")),
})

type FormValues = z.infer<typeof schema>

function defaultAudience(mode: AudienceMode): AudienceDTO {
  if (mode === "admin") return { type: AudienceType.all }
  if (mode === "teacher") return { type: AudienceType.class }
  return { type: AudienceType.org }
}

// Guards that the type-specific target for an audience is actually filled in
// before we let the send through.
function audienceComplete(a: AudienceDTO, mode: AudienceMode): boolean {
  switch (a.type) {
    case AudienceType.all:
      return true
    case AudienceType.org:
      return mode === "admin" ? !!a.org_id : true
    case AudienceType.class:
      return !!a.class_id
    case AudienceType.role:
      return !!a.role_id
    case AudienceType.users:
      return !!a.user_ids && a.user_ids.length > 0
    default:
      return false
  }
}

interface NotificationComposerProps {
  mode: AudienceMode
}

/** Shared send form used by the org composer and the admin compose tab. The
 * audience picker adapts to `mode`; everything else (title, body, link) is
 * identical across panels. */
export function NotificationComposer({ mode }: NotificationComposerProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [audience, setAudience] = useState<AudienceDTO>(defaultAudience(mode))
  const [audienceError, setAudienceError] = useState(false)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { title: "", body: "", action_url: "" },
  })

  const sendMutation = usePostNotifications({
    mutation: {
      onSuccess: () => {
        toast.success(t("notifications.send.sent"))
        reset({ title: "", body: "", action_url: "" })
        setAudience(defaultAudience(mode))
        queryClient.invalidateQueries({ queryKey: getGetNotificationsSentQueryKey() })
      },
    },
  })

  const onSubmit = handleSubmit((values) => {
    if (!audienceComplete(audience, mode)) {
      setAudienceError(true)
      return
    }
    setAudienceError(false)
    sendMutation.mutate({
      data: {
        title: values.title,
        body: values.body,
        action_url: values.action_url || undefined,
        audience,
      },
    })
  })

  return (
    <form onSubmit={onSubmit}>
      <FieldGroup>
        <Field data-invalid={!!errors.title || undefined}>
          <FieldLabel>{t("notifications.send.titleField")}</FieldLabel>
          <Input {...register("title")} maxLength={255} />
          <FieldError errors={[errors.title]} />
        </Field>

        <Field data-invalid={!!errors.body || undefined}>
          <FieldLabel>{t("notifications.send.bodyField")}</FieldLabel>
          <Textarea {...register("body")} rows={4} maxLength={4000} />
          <FieldError errors={[errors.body]} />
        </Field>

        <Field data-invalid={!!errors.action_url || undefined}>
          <FieldLabel>{t("notifications.send.actionUrl")}</FieldLabel>
          <Input {...register("action_url")} inputMode="url" placeholder="https://" />
          <FieldError errors={[errors.action_url]} />
        </Field>

        <div className="flex flex-col gap-2">
          <Label className="text-sm font-medium">{t("notifications.send.audience")}</Label>
          <AudiencePicker
            mode={mode}
            value={audience}
            onChange={(a) => {
              setAudience(a)
              setAudienceError(false)
            }}
          />
          {audienceError && (
            <p className="text-destructive text-sm">{t("notifications.send.audience")}</p>
          )}
        </div>

        <div className="flex justify-end">
          <Button type="submit" disabled={sendMutation.isPending}>
            <SendIcon />
            {t("notifications.send.submit")}
          </Button>
        </div>
      </FieldGroup>
    </form>
  )
}
