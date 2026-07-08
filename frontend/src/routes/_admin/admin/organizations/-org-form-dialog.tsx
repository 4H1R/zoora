import type { GithubCom4H1RZooraInternalDomainOrganization as Organization } from "@/api/model"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetAdminOrganizationsQueryKey,
  getGetAdminOrganizationsStatsQueryKey,
  usePatchAdminOrganizationsIdSettings,
  usePostAdminOrganizations,
  usePutAdminOrganizationsId,
} from "@/api/admin-organizations/admin-organizations"
import { GithubCom4H1RZooraInternalDomainOrganizationStatus as OrgStatus } from "@/api/model"
import { ResourceFormDialog } from "@/components/form/resource-form-dialog"
import { Field, FieldContent, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"

const orgSchema = z.object({
  name: z.string().min(2),
  slug: z
    .string()
    .min(2)
    .max(63)
    .regex(/^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/),
  description: z.string().optional(),
  status: z.nativeEnum(OrgStatus).optional(),
})

type OrgFormValues = z.infer<typeof orgSchema>

interface OrgFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  organization?: Organization | null
}

export function OrgFormDialog({ open, onOpenChange, organization }: OrgFormDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const isEdit = !!organization

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<OrgFormValues>({
    resolver: zodResolver(orgSchema),
    defaultValues: {
      name: "",
      slug: "",
      description: "",
      status: undefined,
    },
  })

  useEffect(() => {
    if (open) {
      reset({
        name: organization?.name ?? "",
        slug: organization?.slug ?? "",
        description: organization?.description ?? "",
        status: (organization?.status as OrgStatus) ?? undefined,
      })
    }
  }, [open, organization, reset])

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetAdminOrganizationsQueryKey() })
    queryClient.invalidateQueries({ queryKey: getGetAdminOrganizationsStatsQueryKey() })
  }

  const createMutation = usePostAdminOrganizations({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.orgs.form.createSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const updateMutation = usePutAdminOrganizationsId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.orgs.form.updateSuccess"))
        invalidate()
        onOpenChange(false)
      },
    },
  })

  const isLoading = createMutation.isPending || updateMutation.isPending

  const onSubmit = handleSubmit((values) => {
    if (isEdit && organization?.id) {
      updateMutation.mutate({ id: organization.id, data: values })
    } else {
      createMutation.mutate({ data: values })
    }
  })

  const statusValue = watch("status")
  const statusOptions = [
    { value: OrgStatus.OrganizationStatusActive, label: t("admin.orgs.statusLabels.active") },
    { value: OrgStatus.OrganizationStatusTrial, label: t("admin.orgs.statusLabels.trial") },
    { value: OrgStatus.OrganizationStatusSuspended, label: t("admin.orgs.statusLabels.suspended") },
    { value: OrgStatus.OrganizationStatusArchived, label: t("admin.orgs.statusLabels.archived") },
  ]

  return (
    <ResourceFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("admin.orgs.form.editTitle") : t("admin.orgs.form.createTitle")}
      description={isEdit ? t("admin.orgs.form.editDescription") : t("admin.orgs.form.createDescription")}
      onSubmit={onSubmit}
      isLoading={isLoading}
      submitLabel={isEdit ? t("common.save") : t("common.create")}
    >
      <FieldGroup>
        <Field data-invalid={!!errors.name || undefined}>
          <FieldLabel>{t("admin.orgs.form.name")}</FieldLabel>
          <Input {...register("name")} placeholder={t("admin.orgs.form.namePlaceholder")} />
          <FieldError errors={[errors.name]} />
        </Field>
        <Field data-invalid={!!errors.slug || undefined}>
          <FieldLabel>{t("admin.orgs.form.slug")}</FieldLabel>
          <Input {...register("slug")} placeholder={t("admin.orgs.form.slugPlaceholder")} />
          <FieldError errors={[errors.slug]} />
        </Field>
        <Field>
          <FieldLabel>{t("admin.orgs.form.description")}</FieldLabel>
          <Textarea {...register("description")} placeholder={t("admin.orgs.form.descriptionPlaceholder")} rows={3} />
        </Field>
        <Field data-invalid={!!errors.status || undefined}>
          <FieldLabel>{t("admin.orgs.form.status")}</FieldLabel>
          <Select
            value={statusValue ?? null}
            onValueChange={(val) => {
              if (val) setValue("status", val as OrgStatus)
            }}
          >
            <SelectTrigger className="w-full">
              <SelectValue placeholder={t("admin.orgs.form.statusPlaceholder")}>
                {(v: OrgStatus) => statusOptions.find((o) => o.value === v)?.label}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {statusOptions.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <FieldError errors={[errors.status]} />
        </Field>
        {isEdit && organization?.id && <OrgSmsGate orgId={organization.id} organization={organization} />}
      </FieldGroup>
    </ResourceFormDialog>
  )
}

// OrgSmsGate is the super-admin toggle for the per-org SMS delivery channel.
// The org list/detail responses don't carry settings, so we seed from any
// runtime-embedded value and keep in sync with the patch response.
function OrgSmsGate({ orgId, organization }: { orgId: string; organization: Organization }) {
  const { t } = useTranslation()
  const initial = (organization as { settings?: { sms_enabled?: boolean } }).settings?.sms_enabled ?? false
  const [enabled, setEnabled] = useState(initial)

  const mutation = usePatchAdminOrganizationsIdSettings({
    mutation: {
      onSuccess: (res) => {
        if (res.status === 200 && typeof res.data.data?.sms_enabled === "boolean") {
          setEnabled(res.data.data.sms_enabled)
        }
        toast.success(t("common.save"))
      },
      onError: () => setEnabled((prev) => !prev),
    },
  })

  return (
    <Field orientation="horizontal">
      <FieldContent>
        <FieldLabel htmlFor="org-sms-gate">{t("notifications.smsGate.label")}</FieldLabel>
        <FieldDescription>{t("notifications.smsGate.description")}</FieldDescription>
      </FieldContent>
      <Switch
        id="org-sms-gate"
        checked={enabled}
        disabled={mutation.isPending}
        onCheckedChange={(next) => {
          setEnabled(next)
          mutation.mutate({ id: orgId, data: { sms_enabled: next } })
        }}
      />
    </Field>
  )
}
