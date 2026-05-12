import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { getGetAdminLiveRoomsQueryKey } from "@/api/admin-livesessions/admin-livesessions"
import { usePostLiveRooms } from "@/api/live-sessions/live-sessions"
import { ClassPicker, SessionPicker } from "@/components/admin/forms/ClassSessionPicker"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Checkbox } from "@/components/ui/checkbox"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"

const schema = z.object({
  class_id: z.string().uuid(),
  class_session_id: z.string().uuid(),
  max_participants: z.coerce.number().int().min(1).max(1000).optional(),
  auto_record: z.boolean().optional(),
  allow_mic_default: z.boolean().optional(),
  allow_camera_default: z.boolean().optional(),
  allow_screen_share_default: z.boolean().optional(),
})

type FormValues = z.infer<typeof schema>

interface LiveRoomCreateModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  defaultClassId?: string
  defaultSessionId?: string
}

const DEFAULTS: FormValues = {
  class_id: "",
  class_session_id: "",
  max_participants: 100,
  auto_record: false,
  allow_mic_default: true,
  allow_camera_default: true,
  allow_screen_share_default: false,
}

export function LiveRoomCreateModal({
  open,
  onOpenChange,
  defaultClassId,
  defaultSessionId,
}: LiveRoomCreateModalProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: DEFAULTS,
  })

  useEffect(() => {
    if (open) {
      form.reset({
        ...DEFAULTS,
        class_id: defaultClassId ?? "",
        class_session_id: defaultSessionId ?? "",
      })
    }
  }, [open, defaultClassId, defaultSessionId])

  const createMutation = usePostLiveRooms({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.liveRooms.form.createSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetAdminLiveRoomsQueryKey() })
        onOpenChange(false)
      },
    },
  })

  const errors = form.formState.errors
  const selectedClassId = form.watch("class_id")
  const selectedSessionId = form.watch("class_session_id")
  const autoRecord = form.watch("auto_record")
  const allowMic = form.watch("allow_mic_default")
  const allowCamera = form.watch("allow_camera_default")
  const allowScreen = form.watch("allow_screen_share_default")

  const onSubmit = form.handleSubmit((values) => {
    createMutation.mutate({
      data: {
        class_session_id: values.class_session_id,
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
      title={t("admin.liveRooms.form.createTitle")}
      description={t("admin.liveRooms.form.createDescription")}
      onSubmit={onSubmit}
      isLoading={createMutation.isPending}
      submitLabel={t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.class_id || undefined}>
          <FieldLabel>{t("admin.liveRooms.form.class")}</FieldLabel>
          <ClassPicker
            value={selectedClassId || undefined}
            onChange={(id) => {
              form.setValue("class_id", id, { shouldValidate: true })
              form.setValue("class_session_id", "", { shouldValidate: true })
            }}
            placeholder={t("admin.liveRooms.form.classPlaceholder")}
          />
          <FieldError errors={[errors.class_id]} />
        </Field>

        <Field data-invalid={!!errors.class_session_id || undefined}>
          <FieldLabel>{t("admin.liveRooms.form.session")}</FieldLabel>
          <SessionPicker
            classId={selectedClassId || undefined}
            value={selectedSessionId || undefined}
            onChange={(id) => form.setValue("class_session_id", id, { shouldValidate: true })}
            placeholder={t("admin.liveRooms.form.sessionPlaceholder")}
          />
          <FieldError errors={[errors.class_session_id]} />
        </Field>

        <Field data-invalid={!!errors.max_participants || undefined}>
          <FieldLabel>{t("admin.liveRooms.form.maxParticipants")}</FieldLabel>
          <Input
            type="number"
            min={1}
            max={1000}
            {...form.register("max_participants")}
            placeholder={t("admin.liveRooms.form.maxParticipantsPlaceholder")}
          />
          <FieldError errors={[errors.max_participants]} />
        </Field>

        <Field orientation="horizontal">
          <Checkbox
            checked={!!allowMic}
            onCheckedChange={(c) => form.setValue("allow_mic_default", !!c)}
          />
          <FieldLabel className="cursor-pointer text-start">
            {t("admin.liveRooms.form.allowMicDefault")}
          </FieldLabel>
        </Field>

        <Field orientation="horizontal">
          <Checkbox
            checked={!!allowCamera}
            onCheckedChange={(c) => form.setValue("allow_camera_default", !!c)}
          />
          <FieldLabel className="cursor-pointer text-start">
            {t("admin.liveRooms.form.allowCameraDefault")}
          </FieldLabel>
        </Field>

        <Field orientation="horizontal">
          <Checkbox
            checked={!!allowScreen}
            onCheckedChange={(c) => form.setValue("allow_screen_share_default", !!c)}
          />
          <FieldLabel className="cursor-pointer text-start">
            {t("admin.liveRooms.form.allowScreenShareDefault")}
          </FieldLabel>
        </Field>

        <Field orientation="horizontal">
          <Checkbox
            checked={!!autoRecord}
            onCheckedChange={(c) => form.setValue("auto_record", !!c)}
          />
          <FieldLabel className="cursor-pointer text-start">
            {t("admin.liveRooms.form.autoRecord")}
          </FieldLabel>
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}
