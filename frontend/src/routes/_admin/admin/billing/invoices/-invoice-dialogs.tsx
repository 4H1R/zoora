import type { GithubCom4H1RZooraInternalDomainInvoice as Invoice } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { Trash2Icon } from "lucide-react"
import { useEffect } from "react"
import { useFieldArray, useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetAdminBillingInvoicesQueryKey,
  usePostAdminBillingInvoices,
  usePostAdminBillingInvoicesIdIssue,
  usePostAdminBillingInvoicesIdMarkPaid,
  usePostAdminBillingInvoicesIdRefund,
} from "@/api/admin-billing/admin-billing"
import { useGetAdminOrganizations } from "@/api/admin-organizations/admin-organizations"
import { GithubCom4H1RZooraInternalDomainAdminCreateInvoiceItemDTOKind as ItemKind } from "@/api/model"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Button } from "@/components/ui/button"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { tomanToRial } from "@/lib/billing"

function useInvalidateInvoices() {
  const queryClient = useQueryClient()
  return () => queryClient.invalidateQueries({ queryKey: getGetAdminBillingInvoicesQueryKey() })
}

// ---------------------------------------------------------------------------
// Mark as paid
// ---------------------------------------------------------------------------

const markPaidSchema = z.object({
  note: z.string().optional(),
  ref_id: z.string().optional(),
})
type MarkPaidValues = z.infer<typeof markPaidSchema>

export function MarkPaidDialog({
  open,
  onOpenChange,
  invoice,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  invoice?: Invoice | null
}) {
  const { t } = useTranslation()
  const invalidate = useInvalidateInvoices()

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<MarkPaidValues>({
    resolver: zodResolver(markPaidSchema),
    defaultValues: { note: "", ref_id: "" },
  })

  useEffect(() => {
    if (open) reset({ note: "", ref_id: "" })
  }, [open, reset])

  const mutation = usePostAdminBillingInvoicesIdMarkPaid({
    mutation: {
      onSuccess: () => {
        toast.success(t("billing.admin.markedPaid"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const onSubmit = handleSubmit((values) => {
    if (!invoice?.id) return
    mutation.mutate({
      id: invoice.id,
      data: { note: values.note || undefined, ref_id: values.ref_id || undefined },
    })
  })

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("billing.admin.markPaidTitle")}
      onSubmit={onSubmit}
      isLoading={mutation.isPending}
      submitLabel={t("billing.admin.markPaid")}
    >
      <FieldGroup>
        <Field>
          <FieldLabel>{t("billing.admin.markPaidNote")}</FieldLabel>
          <Textarea rows={2} {...register("note")} />
        </Field>
        <Field>
          <FieldLabel>{t("billing.admin.markPaidRef")}</FieldLabel>
          <Input {...register("ref_id")} />
          <FieldError errors={[errors.ref_id]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

// ---------------------------------------------------------------------------
// Refund
// ---------------------------------------------------------------------------

const refundSchema = z.object({
  reason: z.string().min(2),
})
type RefundValues = z.infer<typeof refundSchema>

export function RefundDialog({
  open,
  onOpenChange,
  invoice,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  invoice?: Invoice | null
}) {
  const { t } = useTranslation()
  const invalidate = useInvalidateInvoices()

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<RefundValues>({
    resolver: zodResolver(refundSchema),
    defaultValues: { reason: "" },
  })

  useEffect(() => {
    if (open) reset({ reason: "" })
  }, [open, reset])

  const mutation = usePostAdminBillingInvoicesIdRefund({
    mutation: {
      onSuccess: () => {
        toast.success(t("billing.admin.refunded"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const onSubmit = handleSubmit((values) => {
    if (!invoice?.id) return
    mutation.mutate({ id: invoice.id, data: { reason: values.reason } })
  })

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("billing.admin.refundTitle")}
      onSubmit={onSubmit}
      isLoading={mutation.isPending}
      submitLabel={t("billing.admin.refund")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.reason || undefined}>
          <FieldLabel>{t("billing.admin.refundReason")}</FieldLabel>
          <Textarea rows={2} {...register("reason")} />
          <FieldError errors={[errors.reason]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}

// ---------------------------------------------------------------------------
// Create invoice (draft → issue)
// ---------------------------------------------------------------------------

const createSchema = z.object({
  organization_id: z.string().min(1),
  tax_percent: z.number().min(0).max(100).optional(),
  items: z
    .array(
      z.object({
        kind: z.nativeEnum(ItemKind),
        description: z.string().min(2),
        quantity: z.number().int().positive(),
        unit_amount: z.number().int().positive(),
      })
    )
    .min(1),
})
type CreateValues = z.infer<typeof createSchema>

export function CreateInvoiceDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (open: boolean) => void }) {
  const { t } = useTranslation()
  const invalidate = useInvalidateInvoices()

  const { data: orgsResponse } = useGetAdminOrganizations({ page_size: 100 })
  const organizations = (orgsResponse?.status === 200 && orgsResponse.data.data?.items) || []

  const {
    register,
    handleSubmit,
    reset,
    control,
    setValue,
    watch,
    formState: { errors },
  } = useForm<CreateValues>({
    resolver: zodResolver(createSchema),
    defaultValues: {
      organization_id: "",
      tax_percent: undefined,
      items: [{ kind: ItemKind.custom, description: "", quantity: 1, unit_amount: undefined as unknown as number }],
    },
  })

  const { fields, append, remove } = useFieldArray({ control, name: "items" })

  useEffect(() => {
    if (open) {
      reset({
        organization_id: "",
        tax_percent: undefined,
        items: [{ kind: ItemKind.custom, description: "", quantity: 1, unit_amount: undefined as unknown as number }],
      })
    }
  }, [open, reset])

  const issueMutation = usePostAdminBillingInvoicesIdIssue({
    mutation: {
      onSuccess: () => {
        toast.success(t("billing.admin.issued"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const createMutation = usePostAdminBillingInvoices({
    mutation: {
      onSuccess: (res) => {
        toast.success(t("billing.admin.created"))
        invalidate()
        const id = res.status === 201 ? res.data.data?.id : undefined
        if (id) issueMutation.mutate({ id })
        else onOpenChange(false)
      },
    },
  })

  const onSubmit = handleSubmit((values) => {
    createMutation.mutate({
      data: {
        organization_id: values.organization_id,
        tax_percent: values.tax_percent,
        items: values.items.map((it) => ({
          kind: it.kind,
          description: it.description,
          quantity: it.quantity,
          unit_amount: tomanToRial(it.unit_amount),
        })),
      },
    })
  })

  const orgValue = watch("organization_id")
  const kindOptions = [
    { value: ItemKind.custom, label: t("billing.admin.kinds.custom") },
    { value: ItemKind.plan_subscription, label: t("billing.admin.kinds.plan_subscription") },
    { value: ItemKind.addon, label: t("billing.admin.kinds.addon") },
  ]

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("billing.admin.createInvoice")}
      onSubmit={onSubmit}
      isLoading={createMutation.isPending || issueMutation.isPending}
      submitLabel={t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.organization_id || undefined}>
          <FieldLabel>{t("billing.admin.org")}</FieldLabel>
          <Select
            value={orgValue || null}
            onValueChange={(val) => {
              if (val) setValue("organization_id", val, { shouldValidate: true })
            }}
          >
            <SelectTrigger className="w-full">
              <SelectValue placeholder={t("billing.admin.selectOrg")}>
                {(v: string) => organizations.find((o) => o.id === v)?.name ?? v}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {organizations.map((org) => (
                <SelectItem key={org.id} value={org.id!}>
                  {org.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FieldError errors={[errors.organization_id]} />
        </Field>

        <Field>
          <FieldLabel>{t("billing.admin.lineItems")}</FieldLabel>
          <div className="flex flex-col gap-3">
            {fields.map((field, index) => (
              <div key={field.id} className="border-border flex flex-col gap-2 rounded-lg border p-3">
                <div className="flex items-center gap-2">
                  <Input
                    className="flex-1"
                    placeholder={t("billing.admin.description")}
                    {...register(`items.${index}.description` as const)}
                  />
                  {fields.length > 1 && (
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      className="text-destructive hover:bg-destructive/10 hover:text-destructive"
                      onClick={() => remove(index)}
                    >
                      <Trash2Icon />
                    </Button>
                  )}
                </div>
                <div className="grid grid-cols-3 gap-2">
                  <Select
                    value={watch(`items.${index}.kind`) ?? null}
                    onValueChange={(val) => {
                      if (val) setValue(`items.${index}.kind`, val as ItemKind)
                    }}
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue>{(v: ItemKind) => kindOptions.find((o) => o.value === v)?.label}</SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      {kindOptions.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value}>
                          {opt.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <Input
                    type="number"
                    min={1}
                    placeholder={t("billing.admin.quantity")}
                    {...register(`items.${index}.quantity` as const, { valueAsNumber: true })}
                  />
                  <Input
                    type="number"
                    min={1}
                    placeholder={t("billing.admin.unitAmount")}
                    {...register(`items.${index}.unit_amount` as const, { valueAsNumber: true })}
                  />
                </div>
              </div>
            ))}
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() =>
                append({
                  kind: ItemKind.custom,
                  description: "",
                  quantity: 1,
                  unit_amount: undefined as unknown as number,
                })
              }
            >
              {t("billing.admin.addItem")}
            </Button>
          </div>
        </Field>

        <Field data-invalid={!!errors.tax_percent || undefined}>
          <FieldLabel>{t("billing.admin.taxPercent")}</FieldLabel>
          <Input
            type="number"
            min={0}
            max={100}
            {...register("tax_percent", {
              setValueAs: (v) => (v === "" || v == null ? undefined : Number(v)),
            })}
          />
          <FieldError errors={[errors.tax_percent]} />
        </Field>
      </FieldGroup>
    </ResourceFormDialog>
  )
}
