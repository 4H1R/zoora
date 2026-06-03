import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { getGetLiveRoomsQueryKey, usePostLiveRooms } from "@/api/live-sessions/live-sessions"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Checkbox } from "@/components/ui/checkbox"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"

const schema = z.object({
  max_participants: z.coerce.number().int().min(1).max(1000).optional(),
  auto_record: z.boolean().optional(),
  allow_mic_default: z.boolean().optional(),
  allow_camera_default: z.boolean().optional(),
  allow_screen_share_default: z.boolean().optional(),
})

type FormValues = z.infer<typeof schema>

const DEFAULTS: FormValues = {
  max_participants: 100,
  auto_record: false,
  allow_mic_default: true,
  allow_camera_default: true,
  allow_screen_share_default: false,
}

interface LiveRoomFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  classSessionId: string
}

export function LiveRoomFormDialog({ open, onOpenChange, classSessionId }: LiveRoomFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: DEFAULTS,
  })

  useEffect(() => {
    if (open) form.reset(DEFAULTS)
  }, [open])

  const createMutation = usePostLiveRooms({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.session.liveRooms.form.createSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetLiveRoomsQueryKey() })
        onOpenChange(false)
      },
    },
  })

  const errors = form.formState.errors
  const autoRecord = form.watch("auto_record")
  const allowMic = form.watch("allow_mic_default")
  const allowCamera = form.watch("allow_camera_default")
  const allowScreen = form.watch("allow_screen_share_default")

  const onSubmit = form.handleSubmit((values) => {
    createMutation.mutate({
      data: {
        class_session_id: classSessionId,
        config: {
          max_participants: values.max_participants ?? 100,
          auto_record: !!values.auto_record,
          allow_mic_default: !!values.allow_mic_default,
          allow_camera_default: !!values.allow_camera_default,
          allow_screen_share_default: !!values.allow_screen_share_default,
        },
      },
    })
  })

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("org.session.liveRooms.form.createTitle")}
      description={t("org.session.liveRooms.form.createDescription")}
      onSubmit={onSubmit}
      isLoading={createMutation.isPending}
      submitLabel={t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.max_participants || undefined}>
          <FieldLabel>{t("org.session.liveRooms.form.maxParticipants")}</FieldLabel>
          <Input type="number" min={1} max={1000} {...form.register("max_participants")} placeholder="100" />
          <FieldError errors={[errors.max_participants]} />
        </Field>

        <Field orientation="horizontal">
          <Checkbox checked={!!allowMic} onCheckedChange={(c) => form.setValue("allow_mic_default", !!c)} />
          <FieldLabel className="cursor-pointer text-start">{t("org.session.liveRooms.form.allowMic")}</FieldLabel>
        </Field>

        <Field orientation="horizontal">
          <Checkbox checked={!!allowCamera} onCheckedChange={(c) => form.setValue("allow_camera_default", !!c)} />
          <FieldLabel className="cursor-pointer text-start">{t("org.session.liveRooms.form.allowCamera")}</FieldLabel>
        </Field>

        <Field orientation="horizontal">
          <Checkbox checked={!!allowScreen} onCheckedChange={(c) => form.setValue("allow_screen_share_default", !!c)} />
          <FieldLabel className="cursor-pointer text-start">{t("org.session.liveRooms.form.allowScreen")}</FieldLabel>
        </Field>

        <Field orientation="horizontal">
          <Checkbox checked={!!autoRecord} onCheckedChange={(c) => form.setValue("auto_record", !!c)} />
          <FieldLabel className="cursor-pointer text-start">{t("org.session.liveRooms.form.autoRecord")}</FieldLabel>
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}
