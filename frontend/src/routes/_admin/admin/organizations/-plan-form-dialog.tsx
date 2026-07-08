import type { GithubCom4H1RZooraInternalDomainOrganization as Organization } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { Controller, useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetAdminOrganizationsQueryKey,
  getGetAdminOrganizationsStatsQueryKey,
  usePutAdminOrganizationsIdPlan,
} from "@/api/admin-organizations/admin-organizations"
import { GithubCom4H1RZooraInternalDomainPlan as Plan } from "@/api/model"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { DateTimePicker } from "@/components/ui/date-time-picker"
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

const planSchema = z.object({
  plan: z.nativeEnum(Plan),
  expires_at: z.string().optional(),
})

type PlanFormValues = z.infer<typeof planSchema>

interface PlanFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  organization?: Organization | null
}

// Plan/expiry live behind a dedicated admin endpoint (PUT /admin/organizations/:id/plan),
// separate from the org-details update, so they get their own modal. Clearing the expiry
// date sends no expires_at, which the backend treats as a perpetual plan.
export function PlanFormDialog({ open, onOpenChange, organization }: PlanFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const {
    handleSubmit,
    reset,
    control,
    watch,
    setValue,
    formState: { errors },
  } = useForm<PlanFormValues>({
    resolver: zodResolver(planSchema),
    defaultValues: { plan: Plan.PlanFree, expires_at: "" },
  })

  useEffect(() => {
    if (open) {
      reset({
        plan: (organization?.plan as Plan) ?? Plan.PlanFree,
        expires_at: organization?.plan_expires_at ?? "",
      })
    }
  }, [open, organization, reset])

  const mutation = usePutAdminOrganizationsIdPlan({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.orgs.plan.updateSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetAdminOrganizationsQueryKey() })
        queryClient.invalidateQueries({ queryKey: getGetAdminOrganizationsStatsQueryKey() })
        onOpenChange(false)
      },
    },
  })

  const onSubmit = handleSubmit((values) => {
    if (!organization?.id) return
    mutation.mutate({
      id: organization.id,
      data: { plan: values.plan, expires_at: values.expires_at || undefined },
    })
  })

  const planValue = watch("plan")
  const planOptions = [
    { value: Plan.PlanFree, label: t("admin.orgs.planLabels.free") },
    { value: Plan.PlanPro, label: t("admin.orgs.planLabels.pro") },
    { value: Plan.PlanEnterprise, label: t("admin.orgs.planLabels.enterprise") },
  ]

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("admin.orgs.plan.title")}
      description={t("admin.orgs.plan.description", { name: organization?.name ?? "" })}
      onSubmit={onSubmit}
      isLoading={mutation.isPending}
      submitLabel={t("common.save")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.plan || undefined}>
          <FieldLabel>{t("admin.orgs.plan.plan")}</FieldLabel>
          <Select
            value={planValue ?? null}
            onValueChange={(val) => {
              if (val) setValue("plan", val as Plan)
            }}
          >
            <SelectTrigger className="w-full">
              <SelectValue placeholder={t("admin.orgs.plan.planPlaceholder")}>
                {(value: Plan) => t(`admin.orgs.planLabels.${value}`, { defaultValue: value })}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {planOptions.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FieldError errors={[errors.plan]} />
        </Field>
        <Field data-invalid={!!errors.expires_at || undefined}>
          <FieldLabel>{t("admin.orgs.plan.expiresAt")}</FieldLabel>
          <Controller
            control={control}
            name="expires_at"
            render={({ field, fieldState }) => (
              <DateTimePicker
                value={field.value || undefined}
                onChange={(v) => field.onChange(v ?? "")}
                showTime={false}
                clearable
                invalid={fieldState.invalid}
                minDate={new Date()}
              />
            )}
          />
          <FieldDescription>{t("admin.orgs.plan.expiresAtHint")}</FieldDescription>
          <FieldError errors={[errors.expires_at]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}
