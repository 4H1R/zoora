import type { GithubCom4H1RZooraInternalDomainUserCustomFieldDefinition as CustomFieldDefinition } from "@/api/model"
import type { ErrorType } from "@/api/mutator/custom-instance"

import type { Resolver } from "react-hook-form"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetCustomFieldDefinitionsQueryKey,
  usePatchCustomFieldDefinitionsId,
  usePostCustomFieldDefinitions,
} from "@/api/custom-fields/custom-fields"
import { Button } from "@/components/ui/button"
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { Spinner } from "@/components/ui/spinner"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"

const FIELD_TYPES = ["text", "number", "date", "boolean", "select"] as const

const schema = z.object({
  label: z.string().min(1).max(255),
  field_type: z.enum(FIELD_TYPES),
  optionsText: z.string().optional().default(""),
  is_required: z.boolean().default(false),
  is_unique: z.boolean().default(false),
  visible_to_user: z.boolean().default(false),
  description: z.string().max(1000).optional().default(""),
})
type FormValues = z.infer<typeof schema>

interface Props {
  open: boolean
  onOpenChange: (o: boolean) => void
  definition: CustomFieldDefinition | null
}

export function CustomFieldSheet({ open, onOpenChange, definition }: Props) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!definition

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema) as Resolver<FormValues>,
    defaultValues: {
      label: "",
      field_type: "text",
      optionsText: "",
      is_required: false,
      is_unique: false,
      visible_to_user: false,
      description: "",
    },
  })

  const fieldType = watch("field_type")

  useEffect(() => {
    if (!open) return
    reset({
      label: definition?.label ?? "",
      field_type: (definition?.field_type as FormValues["field_type"]) ?? "text",
      optionsText: (definition?.options ?? []).join("\n"),
      is_required: definition?.is_required ?? false,
      is_unique: definition?.is_unique ?? false,
      visible_to_user: definition?.visible_to_user ?? false,
      description: definition?.description ?? "",
    })
  }, [open, definition, reset])

  const invalidate = () => queryClient.invalidateQueries({ queryKey: getGetCustomFieldDefinitionsQueryKey() })

  const onError = (err: unknown) => {
    const status = (err as ErrorType<unknown>).response?.status
    if (status === 409) {
      toast.error(t("org.customFields.errors.optionInUse"))
      return
    }
    toast.error(t("org.customFields.errors.generic"))
  }

  const createMutation = usePostCustomFieldDefinitions({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.customFields.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
      onError,
    },
  })
  const updateMutation = usePatchCustomFieldDefinitionsId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.customFields.updateSuccess"))
        invalidate()
        onOpenChange(false)
      },
      onError,
    },
  })

  const submit = handleSubmit((v) => {
    const options =
      v.field_type === "select"
        ? v.optionsText
            .split("\n")
            .map((s: string) => s.trim())
            .filter(Boolean)
        : []
    if (isEdit && definition?.id) {
      updateMutation.mutate({
        id: definition.id,
        data: {
          label: v.label,
          options,
          is_required: v.is_required,
          is_unique: v.is_unique,
          visible_to_user: v.visible_to_user,
          description: v.description || undefined,
        },
      })
    } else {
      createMutation.mutate({
        data: {
          label: v.label,
          field_type: v.field_type,
          options,
          is_required: v.is_required,
          is_unique: v.is_unique,
          visible_to_user: v.visible_to_user,
          description: v.description || undefined,
        },
      })
    }
  })

  const isLoading = createMutation.isPending || updateMutation.isPending

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="flex w-full flex-col gap-0 p-0 sm:max-w-md">
        <SheetHeader>
          <SheetTitle>{t(isEdit ? "org.customFields.editTitle" : "org.customFields.createTitle")}</SheetTitle>
          <SheetDescription>{t("org.customFields.subtitle")}</SheetDescription>
        </SheetHeader>

        <form onSubmit={submit} className="flex min-h-0 flex-1 flex-col">
          <div className="flex-1 space-y-5 overflow-y-auto px-4 py-4">
            <FieldGroup>
              <Field data-invalid={!!errors.label || undefined}>
                <FieldLabel>{t("org.customFields.label")}</FieldLabel>
                <Input {...register("label")} />
              </Field>

              <Field>
                <FieldLabel>{t("org.customFields.type")}</FieldLabel>
                <Select
                  value={fieldType}
                  onValueChange={(v) => setValue("field_type", v as FormValues["field_type"])}
                  disabled={isEdit}
                  items={FIELD_TYPES.map((ft) => ({ value: ft, label: t(`org.customFields.types.${ft}`) }))}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {FIELD_TYPES.map((ft) => (
                      <SelectItem key={ft} value={ft}>
                        {t(`org.customFields.types.${ft}`)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                {isEdit ? <p className="text-muted-foreground text-xs">{t("org.customFields.typeLocked")}</p> : null}
              </Field>

              {fieldType === "select" ? (
                <Field>
                  <FieldLabel>{t("org.customFields.options")}</FieldLabel>
                  <Textarea rows={4} {...register("optionsText")} placeholder={t("org.customFields.optionsHint")} />
                  <p className="text-muted-foreground text-xs">{t("org.customFields.optionsHint")}</p>
                </Field>
              ) : null}

              <Field>
                <FieldLabel>{t("org.customFields.description")}</FieldLabel>
                <Textarea rows={2} {...register("description")} />
              </Field>

              <ToggleRow
                label={t("org.customFields.required")}
                checked={watch("is_required")}
                onChange={(c) => setValue("is_required", c)}
              />
              <ToggleRow
                label={t("org.customFields.unique")}
                checked={watch("is_unique")}
                onChange={(c) => setValue("is_unique", c)}
              />
              <ToggleRow
                label={t("org.customFields.visibleToUser")}
                checked={watch("visible_to_user")}
                onChange={(c) => setValue("visible_to_user", c)}
              />
            </FieldGroup>
          </div>

          <SheetFooter className="flex-row justify-end gap-2 border-t">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {t("common.cancel")}
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? <Spinner className="size-4" /> : t("common.save")}
            </Button>
          </SheetFooter>
        </form>
      </SheetContent>
    </Sheet>
  )
}

function ToggleRow({ label, checked, onChange }: { label: string; checked: boolean; onChange: (c: boolean) => void }) {
  return (
    <label className="flex items-center justify-between gap-3">
      <span className="text-sm">{label}</span>
      <Switch checked={checked} onCheckedChange={onChange} />
    </label>
  )
}
