import type { PendingAttachment } from "@/components/org/tickets/attachments"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { useGetClasses } from "@/api/classes/classes"
import { useGetClassesIdGradebookColumns } from "@/api/gradebook/gradebook"
import { useGetQuizzesMe } from "@/api/quizzes/quizzes"
import { getGetTicketsQueryKey, usePostTickets } from "@/api/tickets/tickets"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { AttachmentPicker } from "@/components/org/tickets/attachments"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"

const schema = z
  .object({
    class_id: z.string().uuid(),
    type: z.enum(["question", "grade_objection", "other"]),
    title: z.string().min(2).max(255),
    body: z.string().min(1).max(10000),
    target_kind: z.enum(["general", "quiz", "column"]),
    quiz_room_id: z.string().optional(),
    gradebook_column_id: z.string().optional(),
  })
  .refine((v) => v.target_kind !== "quiz" || !!v.quiz_room_id, { path: ["quiz_room_id"] })
  .refine((v) => v.target_kind !== "column" || !!v.gradebook_column_id, { path: ["gradebook_column_id"] })

type FormValues = z.infer<typeof schema>

const DEFAULTS: FormValues = {
  class_id: "",
  type: "question",
  title: "",
  body: "",
  target_kind: "general",
  quiz_room_id: undefined,
  gradebook_column_id: undefined,
}

export function CreateTicketDialog({
  open,
  onOpenChange,
  onCreated,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreated: (ticketId: string) => void
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [attachments, setAttachments] = useState<PendingAttachment[]>([])

  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: DEFAULTS })
  const errors = form.formState.errors
  const classId = form.watch("class_id")
  const type = form.watch("type")
  const targetKind = form.watch("target_kind")

  useEffect(() => {
    if (!open) {
      form.reset(DEFAULTS)
      setAttachments([])
    }
  }, [open, form])

  const { data: classesData } = useGetClasses(undefined, { query: { enabled: open } })
  const classes = (classesData?.status === 200 && classesData.data.data?.items) || []

  const { data: examsData } = useGetQuizzesMe(undefined, {
    query: { enabled: open && type === "grade_objection" && targetKind === "quiz" },
  })
  const exams = ((examsData?.status === 200 && examsData.data.data?.items) || []).filter(
    (e) => e.class_id === classId && !!e.room?.id
  )

  const { data: columnsData } = useGetClassesIdGradebookColumns(classId, undefined, {
    query: { enabled: open && !!classId && type === "grade_objection" && targetKind === "column" },
  })
  const columns = (columnsData?.status === 200 && columnsData.data.data?.items) || []

  const mutation = usePostTickets({
    mutation: {
      onSuccess: (res) => {
        queryClient.invalidateQueries({ queryKey: getGetTicketsQueryKey() })
        onOpenChange(false)
        const id = res.status === 201 ? res.data.data?.id : undefined
        if (id) onCreated(id)
      },
      onError: () => toast.error(t("tickets.error")),
    },
  })

  const onSubmit = form.handleSubmit((values) => {
    mutation.mutate({
      data: {
        class_id: values.class_id,
        type: values.type,
        title: values.title,
        body: values.body,
        quiz_room_id:
          values.type === "grade_objection" && values.target_kind === "quiz" ? values.quiz_room_id : undefined,
        gradebook_column_id:
          values.type === "grade_objection" && values.target_kind === "column" ? values.gradebook_column_id : undefined,
        media_ids: attachments.map((a) => a.id),
      },
    })
  })

  const classItems = classes.map((c) => ({ value: c.id ?? "", label: c.name ?? "" }))
  const typeItems = (["question", "grade_objection", "other"] as const).map((v) => ({
    value: v,
    label: t(`tickets.type.${v}`),
  }))
  const targetItems = [
    { value: "general", label: t("tickets.form.targetGeneral") },
    { value: "quiz", label: t("tickets.form.targetQuiz") },
    { value: "column", label: t("tickets.form.targetColumn") },
  ]
  const examItems = exams.map((e) => ({ value: e.room?.id ?? "", label: e.title ?? "" }))
  const columnItems = columns.map((c) => ({ value: c.id ?? "", label: c.title ?? "" }))

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("tickets.new")}
      onSubmit={onSubmit}
      isLoading={mutation.isPending}
      submitLabel={t("tickets.form.submit")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.class_id || undefined}>
          <FieldLabel>{t("tickets.form.class")}</FieldLabel>
          <Select
            items={classItems}
            value={classId || null}
            onValueChange={(v) => {
              form.setValue("class_id", v ?? "", { shouldValidate: true })
              form.setValue("quiz_room_id", undefined)
              form.setValue("gradebook_column_id", undefined)
            }}
          >
            <SelectTrigger>
              <SelectValue placeholder={t("tickets.form.classPlaceholder")} />
            </SelectTrigger>
            <SelectContent>
              {classItems.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FieldError errors={[errors.class_id]} />
        </Field>

        <Field>
          <FieldLabel>{t("tickets.form.type")}</FieldLabel>
          <Select
            items={typeItems}
            value={type}
            onValueChange={(v) => form.setValue("type", (v ?? "question") as FormValues["type"])}
          >
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {typeItems.map((item) => (
                <SelectItem key={item.value} value={item.value}>
                  {item.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </Field>

        {type === "grade_objection" && (
          <>
            <Field>
              <FieldLabel>{t("tickets.form.target")}</FieldLabel>
              <Select
                items={targetItems}
                value={targetKind}
                onValueChange={(v) => {
                  form.setValue("target_kind", (v ?? "general") as FormValues["target_kind"])
                  form.setValue("quiz_room_id", undefined)
                  form.setValue("gradebook_column_id", undefined)
                }}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {targetItems.map((item) => (
                    <SelectItem key={item.value} value={item.value}>
                      {item.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>

            {targetKind === "quiz" && (
              <Field data-invalid={!!errors.quiz_room_id || undefined}>
                <FieldLabel>{t("tickets.form.targetQuiz")}</FieldLabel>
                <Select
                  items={examItems}
                  value={form.watch("quiz_room_id") || null}
                  onValueChange={(v) => form.setValue("quiz_room_id", v ?? undefined, { shouldValidate: true })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder={t("tickets.form.targetQuizPlaceholder")} />
                  </SelectTrigger>
                  <SelectContent>
                    {examItems.map((item) => (
                      <SelectItem key={item.value} value={item.value}>
                        {item.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FieldError errors={[errors.quiz_room_id]} />
              </Field>
            )}

            {targetKind === "column" && (
              <Field data-invalid={!!errors.gradebook_column_id || undefined}>
                <FieldLabel>{t("tickets.form.targetColumn")}</FieldLabel>
                <Select
                  items={columnItems}
                  value={form.watch("gradebook_column_id") || null}
                  onValueChange={(v) => form.setValue("gradebook_column_id", v ?? undefined, { shouldValidate: true })}
                >
                  <SelectTrigger>
                    <SelectValue placeholder={t("tickets.form.targetColumnPlaceholder")} />
                  </SelectTrigger>
                  <SelectContent>
                    {columnItems.map((item) => (
                      <SelectItem key={item.value} value={item.value}>
                        {item.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FieldError errors={[errors.gradebook_column_id]} />
              </Field>
            )}
          </>
        )}

        <Field data-invalid={!!errors.title || undefined}>
          <FieldLabel>{t("tickets.form.title")}</FieldLabel>
          <Input {...form.register("title")} placeholder={t("tickets.form.titlePlaceholder")} />
          <FieldError errors={[errors.title]} />
        </Field>

        <Field data-invalid={!!errors.body || undefined}>
          <FieldLabel>{t("tickets.form.body")}</FieldLabel>
          <Textarea {...form.register("body")} rows={4} placeholder={t("tickets.form.bodyPlaceholder")} />
          <FieldError errors={[errors.body]} />
        </Field>

        <Field>
          <FieldLabel>{t("tickets.form.attachments")}</FieldLabel>
          <AttachmentPicker classId={classId || undefined} attachments={attachments} onChange={setAttachments} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}
