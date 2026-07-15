import type { GithubCom4H1RZooraInternalDomainLead as Lead } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { Controller, useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { getGetAdminLeadsQueryKey, usePostAdminLeadsIdConvert } from "@/api/admin-leads/admin-leads"
import { GithubCom4H1RZooraInternalDomainPlan as Plan } from "@/api/model"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { DateTimePicker } from "@/components/ui/date-time-picker"
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { PLAN_SIZES, PLAN_TIERS, planKey, planSize, planTier } from "@/lib/plan"

const convertSchema = z.object({
  org_name: z.string().min(2).max(255),
  slug: z
    .string()
    .min(2)
    .max(63)
    .regex(/^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/, "invalid"),
  plan: z.string().min(1),
  plan_expires_at: z.string().optional(),
  owner_name: z.string().min(2).max(255),
  owner_username: z.string().min(3).max(255),
  owner_password: z.string().min(8).max(255),
})

type ConvertFormValues = z.infer<typeof convertSchema>

interface ConvertDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  lead?: Lead | null
}

// Converts a lead into a live org + owner account in one atomic call. Org name
// and owner name are pre-filled from the lead; the admin picks the slug, plan,
// and the owner's login credentials to hand off.
export function ConvertDialog({ open, onOpenChange, lead }: ConvertDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const {
    register,
    handleSubmit,
    reset,
    control,
    watch,
    setValue,
    formState: { errors },
  } = useForm<ConvertFormValues>({
    resolver: zodResolver(convertSchema),
    defaultValues: {
      org_name: "",
      slug: "",
      plan: Plan.PlanFree as string,
      plan_expires_at: "",
      owner_name: "",
      owner_username: "",
      owner_password: "",
    },
  })

  useEffect(() => {
    if (open) {
      reset({
        org_name: lead?.org_name ?? "",
        slug: "",
        plan: Plan.PlanFree,
        plan_expires_at: "",
        owner_name: lead?.name ?? "",
        owner_username: "",
        owner_password: "",
      })
    }
  }, [open, lead, reset])

  const mutation = usePostAdminLeadsIdConvert({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.leads.convert.success"))
        queryClient.invalidateQueries({ queryKey: getGetAdminLeadsQueryKey() })
        onOpenChange(false)
      },
      onError: () => toast.error(t("admin.leads.convert.error")),
    },
  })

  const onSubmit = handleSubmit((values) => {
    if (!lead?.id) return
    mutation.mutate({
      id: lead.id,
      data: {
        org_name: values.org_name,
        slug: values.slug,
        plan: values.plan as Plan,
        plan_expires_at: values.plan_expires_at || undefined,
        owner_name: values.owner_name,
        owner_username: values.owner_username,
        owner_password: values.owner_password,
      },
    })
  })

  const planValue = watch("plan")
  const planOptions = PLAN_TIERS.flatMap((tier) =>
    PLAN_SIZES.map((size) => ({
      value: planKey(tier, size),
      label: `${t(`plans.tiers.${tier}`)} — ${t("plans.sizeSuffix", { size })}`,
    }))
  )

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("admin.leads.convert.title")}
      description={t("admin.leads.convert.description", { name: lead?.org_name ?? "" })}
      onSubmit={onSubmit}
      isLoading={mutation.isPending}
      submitLabel={t("admin.leads.convert.submit")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.org_name || undefined}>
          <FieldLabel>{t("admin.leads.convert.orgName")}</FieldLabel>
          <Input {...register("org_name")} placeholder={t("admin.leads.convert.orgNamePlaceholder")} />
          <FieldError errors={[errors.org_name]} />
        </Field>
        <Field data-invalid={!!errors.slug || undefined}>
          <FieldLabel>{t("admin.leads.convert.slug")}</FieldLabel>
          <Input {...register("slug")} dir="ltr" placeholder={t("admin.leads.convert.slugPlaceholder")} />
          <FieldDescription>{t("admin.leads.convert.slugHint")}</FieldDescription>
          <FieldError errors={[errors.slug]} />
        </Field>
        <Field data-invalid={!!errors.plan || undefined}>
          <FieldLabel>{t("admin.leads.convert.plan")}</FieldLabel>
          <Select
            value={planValue ?? null}
            onValueChange={(val) => {
              if (val) setValue("plan", val)
            }}
          >
            <SelectTrigger className="w-full">
              <SelectValue placeholder={t("admin.leads.convert.planPlaceholder")}>
                {(value: string) =>
                  `${t(`plans.tiers.${planTier(value)}`)} — ${t("plans.sizeSuffix", { size: planSize(value) })}`
                }
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
        <Field data-invalid={!!errors.plan_expires_at || undefined}>
          <FieldLabel>{t("admin.leads.convert.expiresAt")}</FieldLabel>
          <Controller
            control={control}
            name="plan_expires_at"
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
          <FieldDescription>{t("admin.leads.convert.expiresAtHint")}</FieldDescription>
        </Field>
        <Field data-invalid={!!errors.owner_name || undefined}>
          <FieldLabel>{t("admin.leads.convert.ownerName")}</FieldLabel>
          <Input {...register("owner_name")} placeholder={t("admin.leads.convert.ownerNamePlaceholder")} />
          <FieldError errors={[errors.owner_name]} />
        </Field>
        <Field data-invalid={!!errors.owner_username || undefined}>
          <FieldLabel>{t("admin.leads.convert.ownerUsername")}</FieldLabel>
          <Input
            {...register("owner_username")}
            dir="ltr"
            placeholder={t("admin.leads.convert.ownerUsernamePlaceholder")}
          />
          <FieldError errors={[errors.owner_username]} />
        </Field>
        <Field data-invalid={!!errors.owner_password || undefined}>
          <FieldLabel>{t("admin.leads.convert.ownerPassword")}</FieldLabel>
          <Input
            {...register("owner_password")}
            type="text"
            dir="ltr"
            autoComplete="off"
            placeholder={t("admin.leads.convert.ownerPasswordPlaceholder")}
          />
          <FieldDescription>{t("admin.leads.convert.ownerPasswordHint")}</FieldDescription>
          <FieldError errors={[errors.owner_password]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}
