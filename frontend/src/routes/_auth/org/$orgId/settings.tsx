import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import {
  CalendarDaysIcon,
  CheckIcon,
  CopyIcon,
  HistoryIcon,
  UsersIcon,
} from "lucide-react"
import { useEffect, useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetOrganizationsIdQueryKey,
  useGetOrganizationsId,
  usePutOrganizationsId,
} from "@/api/organizations/organizations"
import type { GithubCom4H1RZooraInternalDomainOrganizationStatus as OrgStatus } from "@/api/model/githubCom4H1RZooraInternalDomainOrganizationStatus"
import { FormSaveBar } from "@/components/form-save-bar"
import { Button } from "@/components/ui/button"
import { Field, FieldError, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"
import { Textarea } from "@/components/ui/textarea"
import { useFormatDate } from "@/lib/data-table"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org/$orgId/settings")({
  head: () => orgHead("org.nav.settings"),
  component: RouteComponent,
})

const settingsSchema = z.object({
  name: z.string().min(2),
  description: z.string().optional(),
})

type SettingsFormValues = z.infer<typeof settingsSchema>

const STATUS_STYLES: Record<OrgStatus, { dot: string; chip: string; key: string }> = {
  active: {
    dot: "bg-success",
    chip: "border-success/25 bg-success/10 text-success",
    key: "org.settings.statusActive",
  },
  trial: {
    dot: "bg-warning",
    chip: "border-warning/25 bg-warning/10 text-warning",
    key: "org.settings.statusTrial",
  },
  suspended: {
    dot: "bg-destructive",
    chip: "border-destructive/25 bg-destructive/10 text-destructive",
    key: "org.settings.statusSuspended",
  },
  archived: {
    dot: "bg-muted-foreground",
    chip: "border-border bg-muted text-muted-foreground",
    key: "org.settings.statusArchived",
  },
}

function initials(name?: string) {
  if (!name) return "—"
  const words = name.trim().split(/\s+/).filter(Boolean)
  if (words.length === 0) return "—"
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase()
  return (words[0][0] + words[words.length - 1][0]).toUpperCase()
}

function RouteComponent() {
  const { orgId } = Route.useParams()
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const formatDate = useFormatDate()
  const allowed = useOrgGuard("organizations:update")
  const [copied, setCopied] = useState(false)

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

  const form = useForm<SettingsFormValues>({
    resolver: zodResolver(settingsSchema),
    defaultValues: { name: "", description: "" },
  })
  const {
    register,
    handleSubmit,
    reset,
    watch,
    formState: { errors },
  } = form

  useEffect(() => {
    if (org) {
      reset({ name: org.name ?? "", description: org.description ?? "" })
    }
  }, [org, reset])

  const onSubmit = handleSubmit((values) => {
    updateMutation.mutate({ id: orgId, data: values })
  })

  const handleReset = () => {
    if (org) reset({ name: org.name ?? "", description: org.description ?? "" })
  }

  const copyId = async () => {
    try {
      await navigator.clipboard.writeText(orgId)
      setCopied(true)
      toast.success(t("org.settings.copied"))
      window.setTimeout(() => setCopied(false), 1600)
    } catch {
      /* clipboard unavailable */
    }
  }

  const isPending = updateMutation.isPending
  const liveName = watch("name")
  const status = (org?.status ?? "active") as OrgStatus
  const statusStyle = STATUS_STYLES[status] ?? STATUS_STYLES.active
  const memberCount = org?.total_users ?? 0

  if (!allowed) return null

  if (isLoading) {
    return (
      <div className="flex min-h-60 items-center justify-center">
        <Spinner className="text-muted-foreground size-6" />
      </div>
    )
  }

  return (
    <div className="mx-auto w-full max-w-4xl pb-24">
      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-6">
        <section className="relative overflow-hidden rounded-2xl border">
          <div
            aria-hidden
            className="pointer-events-none absolute inset-0 bg-gradient-to-br from-success/8 via-card to-card"
          />
          <div
            aria-hidden
            className="pointer-events-none absolute inset-0 opacity-[0.4]"
            style={{
              backgroundImage:
                "radial-gradient(circle at 1px 1px, var(--border) 1px, transparent 0)",
              backgroundSize: "22px 22px",
              maskImage: "radial-gradient(120% 100% at 100% 0%, black, transparent 65%)",
              WebkitMaskImage: "radial-gradient(120% 100% at 100% 0%, black, transparent 65%)",
            }}
          />
          <div
            aria-hidden
            className="pointer-events-none absolute -end-16 -top-20 size-56 rounded-full bg-success/15 blur-3xl"
          />

          <div className="relative flex flex-col gap-5 p-6 sm:flex-row sm:items-center sm:gap-6 sm:p-8">
            <div className="relative shrink-0">
              <div className="from-success to-green-700 text-success-foreground ring-foreground/10 grid size-18 place-items-center rounded-2xl bg-gradient-to-br font-heading text-2xl font-semibold tracking-tight shadow-lg ring-1 select-none">
                {initials(liveName || org?.name)}
              </div>
            </div>

            <div className="min-w-0 flex-1">
              <div className="mb-2 flex flex-wrap items-center gap-2.5">
                <h1 className="font-heading text-foreground truncate text-2xl font-semibold tracking-tight sm:text-3xl">
                  {liveName || org?.name || "—"}
                </h1>
                <span
                  className={cn(
                    "inline-flex h-6 items-center gap-1.5 rounded-full border px-2.5 text-xs font-medium",
                    statusStyle.chip
                  )}
                >
                  <span className={cn("size-1.5 rounded-full", statusStyle.dot)} />
                  {t(statusStyle.key)}
                </span>
              </div>

              <div className="text-muted-foreground flex flex-wrap items-center gap-x-5 gap-y-1.5 text-sm">
                <span className="inline-flex items-center gap-1.5">
                  <UsersIcon className="size-4" />
                  <span className="text-foreground font-medium tabular-nums">
                    {memberCount.toLocaleString()}
                  </span>
                  {t("org.settings.members")}
                </span>
                <span className="inline-flex items-center gap-1.5">
                  <CalendarDaysIcon className="size-4" />
                  {t("org.settings.created")} {formatDate(org?.created_at)}
                </span>
              </div>
            </div>
          </div>
        </section>

        <Section
          title={t("org.settings.general")}
          hint={t("org.settings.generalHint")}
        >
          <Field data-invalid={!!errors.name || undefined}>
            <FieldLabel htmlFor="org-name" className="text-xs">
              {t("org.settings.name")}
            </FieldLabel>
            <Input
              id="org-name"
              type="text"
              placeholder={t("org.settings.namePlaceholder")}
              className="h-10"
              aria-invalid={!!errors.name}
              {...register("name")}
            />
            {errors.name && <FieldError>{errors.name.message}</FieldError>}
          </Field>

          <Field>
            <FieldLabel htmlFor="org-description" className="text-xs">
              {t("org.settings.description")}
            </FieldLabel>
            <Textarea
              id="org-description"
              placeholder={t("org.settings.descriptionPlaceholder")}
              rows={4}
              className="resize-none"
              {...register("description")}
            />
          </Field>
        </Section>

        <Section
          title={t("org.settings.details")}
          hint={t("org.settings.detailsHint")}
        >
          <DetailRow label={t("org.settings.orgId")}>
            <code className="bg-muted text-muted-foreground rounded-md px-2 py-1 font-mono text-xs">
              {orgId}
            </code>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              onClick={copyId}
              aria-label={t("org.settings.copyId")}
              className="ms-1"
            >
              {copied ? (
                <CheckIcon className="text-success size-3.5" />
              ) : (
                <CopyIcon className="size-3.5" />
              )}
            </Button>
          </DetailRow>

          <Separator />

          <DetailRow label={t("org.settings.created")} icon={<CalendarDaysIcon className="size-4" />}>
            <span className="text-foreground text-sm tabular-nums">
              {formatDate(org?.created_at)}
            </span>
          </DetailRow>

          <Separator />

          <DetailRow label={t("org.settings.updated")} icon={<HistoryIcon className="size-4" />}>
            <span className="text-foreground text-sm tabular-nums">
              {formatDate(org?.updated_at)}
            </span>
          </DetailRow>
        </Section>
      </form>

      <FormSaveBar
        form={form}
        onSave={onSubmit}
        onReset={handleReset}
        isPending={isPending}
      />
    </div>
  )
}

function Section({
  title,
  hint,
  children,
}: {
  title: string
  hint: string
  children: React.ReactNode
}) {
  return (
    <section className="bg-card ring-foreground/10 grid gap-x-8 gap-y-5 rounded-2xl border p-6 ring-1 sm:p-8 md:grid-cols-[14rem_1fr]">
      <div className="md:pe-4">
        <h2 className="font-heading text-foreground text-base font-semibold tracking-tight">
          {title}
        </h2>
        <p className="text-muted-foreground mt-1 text-sm leading-relaxed">{hint}</p>
      </div>
      <div className="flex flex-col gap-5">{children}</div>
    </section>
  )
}

function DetailRow({
  label,
  icon,
  children,
}: {
  label: string
  icon?: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <div className="flex items-center justify-between gap-3">
      <span className="text-muted-foreground inline-flex items-center gap-2 text-sm">
        {icon}
        {label}
      </span>
      <span className="inline-flex items-center">{children}</span>
    </div>
  )
}

function Separator() {
  return <div className="bg-border h-px w-full" />
}
