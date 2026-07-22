import type { GithubCom4H1RZooraInternalDomainClassSession as Session } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useRef } from "react"
import { Controller, useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { getGetAdminSessionsQueryKey } from "@/api/admin-classes/admin-classes"
import {
  getGetClassesIdSessionsQueryKey,
  usePostClassesIdSessions,
  usePutClassesSessionsSessionId,
} from "@/api/classes/classes"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { DateTimePicker } from "@/components/ui/date-time-picker"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { useSessionTitle } from "@/lib/session-title"

const schema = z.object({
  name: z.string().min(2),
  description: z.string().optional(),
  start_time: z.string().min(1),
})

type FormValues = z.infer<typeof schema>

interface SessionCreateModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  classId: string
  session?: Session | null
}

export function SessionCreateModal({ open, onOpenChange, classId, session }: SessionCreateModalProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!session
  const genTitle = useSessionTitle()

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      description: "",
      start_time: "",
    },
  })

  // Tracks the last auto-filled title so we only overwrite the name field
  // while the user hasn't typed their own.
  const autoNameRef = useRef("")

  useEffect(() => {
    if (!open) return
    autoNameRef.current = ""
    if (isEdit && session) {
      form.reset({
        name: session.name ?? "",
        description: session.description ?? "",
        start_time: session.start_time ?? "",
      })
    } else {
      form.reset({ name: "", description: "", start_time: "" })
    }
  }, [open, session, isEdit, form])

  const startTime = form.watch("start_time")

  // On create, default the name to a readable label derived from the start
  // time ("کلاس دوشنبه ۲۰ تیر ساعت ۱۱:۳۰"), unless the user typed their own.
  useEffect(() => {
    if (!open || isEdit || !startTime) return
    const base = new Date(startTime)
    if (Number.isNaN(base.getTime())) return
    const current = form.getValues("name") ?? ""
    if (current !== "" && current !== autoNameRef.current) return
    const next = genTitle(base)
    autoNameRef.current = next
    form.setValue("name", next, { shouldValidate: true })
  }, [open, isEdit, startTime, form, genTitle])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetClassesIdSessionsQueryKey(classId) })
    queryClient.invalidateQueries({ queryKey: getGetAdminSessionsQueryKey() })
  }

  const createMutation = usePostClassesIdSessions({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.sessions.form.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutClassesSessionsSessionId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.sessions.form.updateSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const isLoading = createMutation.isPending || updateMutation.isPending
  const errors = form.formState.errors

  const onSubmit = form.handleSubmit((values) => {
    if (isEdit && session?.id) {
      updateMutation.mutate({
        sessionId: session.id,
        data: {
          name: values.name,
          description: values.description,
          start_time: values.start_time,
        },
      })
    } else {
      createMutation.mutate({
        id: classId,
        data: {
          name: values.name,
          description: values.description,
          start_time: values.start_time,
        },
      })
    }
  })

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("admin.sessions.form.editTitle") : t("admin.sessions.form.createTitle")}
      description={isEdit ? t("admin.sessions.form.editDescription") : t("admin.sessions.form.createDescription")}
      onSubmit={onSubmit}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.name || undefined}>
          <FieldLabel>{t("admin.sessions.form.name")}</FieldLabel>
          <Input {...form.register("name")} placeholder={t("admin.sessions.form.namePlaceholder")} />
          <FieldError errors={[errors.name]} />
        </Field>

        <Field>
          <FieldLabel>{t("admin.sessions.form.description")}</FieldLabel>
          <Textarea
            {...form.register("description")}
            placeholder={t("admin.sessions.form.descriptionPlaceholder")}
            rows={3}
          />
        </Field>

        <Field data-invalid={!!errors.start_time || undefined}>
          <FieldLabel>{t("admin.sessions.form.startTime")}</FieldLabel>
          <Controller
            control={form.control}
            name="start_time"
            render={({ field, fieldState }) => (
              <DateTimePicker
                value={field.value || undefined}
                onChange={(v) => field.onChange(v ?? "")}
                invalid={fieldState.invalid}
              />
            )}
          />
          <FieldError errors={[errors.start_time]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}
