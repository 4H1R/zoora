import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useRef } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { getGetAdminLiveRoomsQueryKey } from "@/api/admin-livesessions/admin-livesessions"
import { getGetLiveRoomsQueryKey, usePostLiveRooms } from "@/api/live-sessions/live-sessions"
import { ClassPicker, SessionPicker } from "@/components/admin/forms/ClassSessionPicker"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { useSessionTitle } from "@/lib/session-title"

const schema = z.object({
  class_id: z.string().uuid(),
  class_session_id: z.string().uuid(),
  name: z.string().min(1).max(255),
})

type FormInput = z.input<typeof schema>
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
  name: "",
}

export function LiveRoomCreateModal({
  open,
  onOpenChange,
  defaultClassId,
  defaultSessionId,
}: LiveRoomCreateModalProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const genTitle = useSessionTitle()
  const autoNameRef = useRef("")

  const form = useForm<FormInput, unknown, FormValues>({
    resolver: zodResolver(schema),
    defaultValues: DEFAULTS,
  })

  useEffect(() => {
    if (open) {
      const name = genTitle(new Date())
      autoNameRef.current = name
      form.reset({
        ...DEFAULTS,
        name,
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
        queryClient.invalidateQueries({ queryKey: getGetLiveRoomsQueryKey() })
        onOpenChange(false)
      },
    },
  })

  const errors = form.formState.errors
  const selectedClassId = form.watch("class_id")
  const selectedSessionId = form.watch("class_session_id")

  const onSubmit = form.handleSubmit((values) => {
    createMutation.mutate({
      data: {
        class_session_id: values.class_session_id,
        name: values.name.trim(),
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

        <Field data-invalid={!!errors.name || undefined}>
          <FieldLabel>{t("admin.liveRooms.form.name")}</FieldLabel>
          <Input {...form.register("name")} placeholder={t("admin.liveRooms.form.namePlaceholder")} />
          <FieldError errors={[errors.name]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}
