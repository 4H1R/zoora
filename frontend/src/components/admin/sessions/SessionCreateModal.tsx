import type { GithubCom4H1RZooraInternalDomainClassSession as Session } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetClassesIdSessionsQueryKey,
  usePostClassesIdSessions,
  usePutClassesSessionsSessionId,
} from "@/api/classes/classes"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"

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

function isoToLocalInput(iso?: string): string {
  if (!iso) return ""
  const d = new Date(iso)
  if (isNaN(d.getTime())) return ""
  const pad = (n: number) => String(n).padStart(2, "0")
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function SessionCreateModal({ open, onOpenChange, classId, session }: SessionCreateModalProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!session

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      description: "",
      start_time: "",
    },
  })

  useEffect(() => {
    if (!open) return
    if (isEdit && session) {
      form.reset({
        name: session.name ?? "",
        description: session.description ?? "",
        start_time: isoToLocalInput(session.start_time),
      })
    } else {
      form.reset({ name: "", description: "", start_time: "" })
    }
  }, [open, session, isEdit])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetClassesIdSessionsQueryKey(classId) })
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
    const startISO = new Date(values.start_time).toISOString()
    if (isEdit && session?.id) {
      updateMutation.mutate({
        sessionId: session.id,
        data: {
          name: values.name,
          description: values.description,
          start_time: startISO,
        },
      })
    } else {
      createMutation.mutate({
        id: classId,
        data: {
          name: values.name,
          description: values.description,
          start_time: startISO,
        },
      })
    }
  })

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("admin.sessions.form.editTitle") : t("admin.sessions.form.createTitle")}
      description={
        isEdit ? t("admin.sessions.form.editDescription") : t("admin.sessions.form.createDescription")
      }
      onSubmit={onSubmit}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.name || undefined}>
          <FieldLabel>{t("admin.sessions.form.name")}</FieldLabel>
          <Input
            {...form.register("name")}
            placeholder={t("admin.sessions.form.namePlaceholder")}
          />
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
          <Input type="datetime-local" {...form.register("start_time")} />
          <FieldError errors={[errors.start_time]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}
