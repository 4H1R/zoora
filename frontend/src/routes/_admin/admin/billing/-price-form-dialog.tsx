import type { GithubCom4H1RZooraInternalDomainPlanPrice as PlanPrice } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { getGetAdminBillingPricesQueryKey, usePutAdminBillingPrices } from "@/api/admin-billing/admin-billing"
import {
  GithubCom4H1RZooraInternalDomainBillingInterval as BillingInterval,
  GithubCom4H1RZooraInternalDomainPlan as Plan,
} from "@/api/model"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { rialToToman, tomanToRial } from "@/lib/billing"
import { PAID_TIERS, PLAN_SIZES, planKey } from "@/lib/plan"

const priceSchema = z.object({
  plan: z.string().min(1),
  interval: z.nativeEnum(BillingInterval),
  amount: z.number().int().positive(),
})

type PriceFormValues = z.infer<typeof priceSchema>

interface PriceFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  price?: PlanPrice | null
}

// Amounts are entered in Toman and stored in Rial (×10). Currency defaults to IRR
// and is hidden in v1. Upsert is keyed by (plan, interval) server-side.
export function PriceFormDialog({ open, onOpenChange, price }: PriceFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<PriceFormValues>({
    resolver: zodResolver(priceSchema),
    defaultValues: {
      plan: planKey("pro", 50),
      interval: BillingInterval.BillingIntervalMonthly,
      amount: undefined as unknown as number,
    },
  })

  useEffect(() => {
    if (open) {
      reset({
        plan: price?.plan ?? planKey("pro", 50),
        interval: (price?.interval as BillingInterval) ?? BillingInterval.BillingIntervalMonthly,
        amount: price?.amount ? rialToToman(price.amount) : (undefined as unknown as number),
      })
    }
  }, [open, price, reset])

  const mutation = usePutAdminBillingPrices({
    mutation: {
      onSuccess: () => {
        toast.success(t("billing.admin.priceSaved"))
        queryClient.invalidateQueries({ queryKey: getGetAdminBillingPricesQueryKey() })
        onOpenChange(false)
      },
    },
  })

  const onSubmit = handleSubmit((values) => {
    mutation.mutate({
      data: {
        plan: values.plan as Plan,
        interval: values.interval,
        amount: tomanToRial(values.amount),
        currency: "IRR",
      },
    })
  })

  const planValue = watch("plan")
  const intervalValue = watch("interval")

  // Free tiers have no price row — only paid tier × size combos are sellable.
  const planOptions = PAID_TIERS.flatMap((tier) =>
    PLAN_SIZES.map((size) => ({
      value: planKey(tier, size),
      label: `${t(`plans.tiers.${tier}`)} — ${t("plans.sizeSuffix", { size })}`,
    }))
  )
  const intervalOptions = [
    { value: BillingInterval.BillingIntervalMonthly, label: t("billing.intervals.monthly") },
    { value: BillingInterval.BillingIntervalYearly, label: t("billing.intervals.yearly") },
  ]

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("billing.admin.savePrice")}
      onSubmit={onSubmit}
      isLoading={mutation.isPending}
      submitLabel={t("common.save")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.plan || undefined}>
          <FieldLabel>{t("billing.admin.plan")}</FieldLabel>
          <Select
            value={planValue ?? null}
            onValueChange={(val) => {
              if (val) setValue("plan", val as Plan)
            }}
          >
            <SelectTrigger className="w-full">
              <SelectValue>{(v: Plan) => planOptions.find((o) => o.value === v)?.label}</SelectValue>
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

        <Field data-invalid={!!errors.interval || undefined}>
          <FieldLabel>{t("billing.admin.interval")}</FieldLabel>
          <Select
            value={intervalValue ?? null}
            onValueChange={(val) => {
              if (val) setValue("interval", val as BillingInterval)
            }}
          >
            <SelectTrigger className="w-full">
              <SelectValue>{(v: BillingInterval) => intervalOptions.find((o) => o.value === v)?.label}</SelectValue>
            </SelectTrigger>
            <SelectContent>
              {intervalOptions.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FieldError errors={[errors.interval]} />
        </Field>

        <Field data-invalid={!!errors.amount || undefined}>
          <FieldLabel>{t("billing.admin.amountToman")}</FieldLabel>
          <Input type="number" min={1} {...register("amount", { valueAsNumber: true })} />
          <FieldError errors={[errors.amount]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}
