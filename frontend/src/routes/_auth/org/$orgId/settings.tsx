import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { Building2Icon } from "lucide-react"
import { useEffect } from "react"
import { useAccess } from "react-access-engine"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetOrganizationsIdQueryKey,
  useGetOrganizationsId,
  usePutOrganizationsId,
} from "@/api/organizations/organizations"
import { Button } from "@/components/ui/button"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"
import { Textarea } from "@/components/ui/textarea"

export const Route = createFileRoute("/_auth/org/$orgId/settings")({
  component: RouteComponent,
})

const settingsSchema = z.object({
  name: z.string().min(2, "org.settings.nameError"),
  description: z.string().optional(),
})

type SettingsFormValues = z.infer<typeof settingsSchema>

function RouteComponent() {
  const { orgId } = Route.useParams()
  const { t } = useTranslation()
  const { can } = useAccess()
  const queryClient = useQueryClient()

  const canUpdate = can("organizations:update")
  const { data: orgResponse, isLoading } = useGetOrganizationsId(orgId)
  const org = orgResponse?.status === 200 ? orgResponse.data.data : undefined

  const updateMutation = usePutOrganizationsId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.settings.updateSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetOrganizationsIdQueryKey(orgId) })
      },
    },
  })

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isDirty },
  } = useForm<SettingsFormValues>({
    resolver: zodResolver(settingsSchema),
    defaultValues: { name: "", description: "" },
  })

  useEffect(() => {
    if (org) {
      reset({
        name: org.name ?? "",
        description: org.description ?? "",
      })
    }
  }, [org, reset])

  const onSubmit = handleSubmit((values) => {
    updateMutation.mutate({ id: orgId, data: values })
  })

  const isPending = updateMutation.isPending

  if (isLoading) {
    return (
      <div className="flex min-h-60 items-center justify-center">
        <Spinner className="text-muted-foreground size-6" />
      </div>
    )
  }

  return (
    <div className="mx-auto w-full max-w-xl py-4 lg:py-8">
      <div className="mb-8 flex items-start gap-4">
        <div className="bg-primary/10 text-primary flex size-11 shrink-0 items-center justify-center rounded-lg">
          <Building2Icon className="size-5" />
        </div>
        <div>
          <h1 className="text-xl font-semibold tracking-tight">{t("org.settings.title")}</h1>
          <p className="text-muted-foreground mt-1 text-sm">{t("org.settings.subtitle")}</p>
        </div>
      </div>

      <div className="border-border bg-card rounded-xl border p-6 shadow-sm">
        <form onSubmit={onSubmit} noValidate>
          <FieldGroup>
            <Field data-invalid={!!errors.name || undefined}>
              <FieldLabel htmlFor="org-name" className="text-xs">
                {t("org.settings.name")}
              </FieldLabel>
              <Input
                id="org-name"
                type="text"
                placeholder={t("org.settings.namePlaceholder")}
                className="h-10.5"
                disabled={!canUpdate}
                aria-invalid={!!errors.name}
                {...register("name")}
              />
              {errors.name && <FieldError>{t(errors.name.message ?? "org.settings.nameError")}</FieldError>}
            </Field>

            <Field>
              <FieldLabel htmlFor="org-description" className="text-xs">
                {t("org.settings.description")}
              </FieldLabel>
              <Textarea
                id="org-description"
                placeholder={t("org.settings.descriptionPlaceholder")}
                rows={4}
                disabled={!canUpdate}
                {...register("description")}
              />
            </Field>

            {canUpdate && (
              <Button
                type="submit"
                disabled={isPending || !isDirty}
                className="mt-2 h-10.5 w-full text-sm font-semibold sm:w-auto sm:px-8"
              >
                {isPending && <Spinner />}
                {isPending ? t("org.settings.saving") : t("common.save")}
              </Button>
            )}
          </FieldGroup>
        </form>
      </div>
    </div>
  )
}
