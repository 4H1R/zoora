import type { GithubCom4H1RZooraInternalDomainOrganizationStatus as OrgStatus } from "@/api/model/githubCom4H1RZooraInternalDomainOrganizationStatus"
import type { PlanTier } from "@/lib/plan"
import type { LucideIcon } from "lucide-react"

import { zodResolver } from "@hookform/resolvers/zod"
import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import {
  CalendarClockIcon,
  CalendarDaysIcon,
  CheckIcon,
  CopyIcon,
  CrownIcon,
  HistoryIcon,
  InfinityIcon,
  InfoIcon,
  RocketIcon,
  SparklesIcon,
  UsersIcon,
  ZapIcon,
} from "lucide-react"
import { useEffect, useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import {
  getGetOrganizationsIdQueryKey,
  getGetOrganizationsIdSettingsQueryKey,
  useGetOrganizationsId,
  useGetOrganizationsIdSettings,
  usePutOrganizationsId,
  usePutOrganizationsIdSettings,
} from "@/api/organizations/organizations"
import { useGetUsersMe } from "@/api/users/users"
import { FormSaveBar } from "@/components/form-save-bar"
import { Button } from "@/components/ui/button"
import { Field, FieldError, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { Textarea } from "@/components/ui/textarea"
import { useOrgGuard } from "@/lib/access"
import { useFormatDate } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"
import { planSize, planTier } from "@/lib/plan"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org/settings/")({
  head: () => orgHead("org.nav.settings"),
  component: RouteComponent,
})

const settingsSchema = z.object({
  name: z.string().min(2),
  description: z.string().optional(),
})

type SettingsFormValues = z.infer<typeof settingsSchema>

const attendanceSchema = z.object({
  attendance_present_threshold_percent: z.number().int().min(1).max(100),
})

type AttendanceFormValues = z.infer<typeof attendanceSchema>

function orgFormDefaults(org?: { name?: string | null; description?: string | null }): SettingsFormValues {
  return { name: org?.name ?? "", description: org?.description ?? "" }
}

function attendanceFormDefaults(settings?: {
  attendance_present_threshold_percent?: number | null
}): AttendanceFormValues {
  return {
    attendance_present_threshold_percent: settings?.attendance_present_threshold_percent ?? 75,
  }
}

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

type PlanStyle = {
  icon: LucideIcon
  chip: string
  wash: string
  glow: string
  iconWrap: string
  check: string
}

// Each tier gets its own accent, echoing the STATUS_STYLES pattern: free = neutral,
// plus = success/green, pro = brand/primary, max = warning/gold. Feature bullets
// are qualitative (no hard numbers) since orgs can't read the admin plan catalog.
const PLAN_STYLES: Record<PlanTier, PlanStyle> = {
  free: {
    icon: SparklesIcon,
    chip: "border-border bg-muted text-muted-foreground",
    wash: "from-muted/50 via-card to-card",
    glow: "bg-muted-foreground/10",
    iconWrap: "bg-muted text-muted-foreground ring-border",
    check: "text-muted-foreground",
  },
  plus: {
    icon: ZapIcon,
    chip: "border-success/25 bg-success/10 text-success",
    wash: "from-success/8 via-card to-card",
    glow: "bg-success/15",
    iconWrap: "from-success to-success/70 text-success-foreground bg-gradient-to-br ring-foreground/10",
    check: "text-success",
  },
  pro: {
    icon: RocketIcon,
    chip: "border-primary/25 bg-primary/10 text-primary",
    wash: "from-primary/8 via-card to-card",
    glow: "bg-primary/15",
    iconWrap: "from-primary to-primary/70 text-primary-foreground bg-gradient-to-br ring-foreground/10",
    check: "text-primary",
  },
  max: {
    icon: CrownIcon,
    chip: "border-warning/25 bg-warning/10 text-warning",
    wash: "from-warning/8 via-card to-card",
    glow: "bg-warning/15",
    iconWrap: "from-warning text-warning-foreground bg-gradient-to-br to-amber-600 ring-foreground/10",
    check: "text-warning",
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
  const { data: meResponse } = useGetUsersMe()
  const orgId = (meResponse?.status === 200 && meResponse.data.data?.organization_id) || ""
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const formatDate = useFormatDate()
  const allowed = useOrgGuard("organizations:update")
  const [copied, setCopied] = useState(false)

  // Two independent settings resources, each its own query/mutation/form,
  // but unified under a single save bar at the bottom of the page.
  const { data: orgResponse, isLoading } = useGetOrganizationsId(orgId)
  const org = orgResponse?.status === 200 ? orgResponse.data.data : undefined

  const { data: settingsResponse } = useGetOrganizationsIdSettings(orgId)
  const settings = settingsResponse?.status === 200 ? settingsResponse.data.data : undefined

  const updateMutation = usePutOrganizationsId({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.settings.updateSuccess"))
        queryClient.invalidateQueries({ queryKey: getGetOrganizationsIdQueryKey(orgId) })
      },
    },
  })

  const attendanceMutation = usePutOrganizationsIdSettings({
    mutation: {
      onSuccess: () => {
        toast.success(t("org.settings.attendance.updateSuccess"))
        queryClient.invalidateQueries({
          queryKey: getGetOrganizationsIdSettingsQueryKey(orgId),
        })
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

  const attendanceForm = useForm<AttendanceFormValues>({
    resolver: zodResolver(attendanceSchema),
    defaultValues: { attendance_present_threshold_percent: 75 },
  })

  // Sync forms with server data once it loads.
  useEffect(() => {
    if (org) reset(orgFormDefaults(org))
  }, [org, reset])

  useEffect(() => {
    if (settings) attendanceForm.reset(attendanceFormDefaults(settings))
  }, [settings, attendanceForm])

  const onSubmit = handleSubmit((values) => {
    updateMutation.mutate({ id: orgId, data: values })
  })

  const onAttendanceSubmit = attendanceForm.handleSubmit((values) => {
    attendanceMutation.mutate({ id: orgId, data: values })
  })

  const dirty = form.formState.isDirty || attendanceForm.formState.isDirty
  const anyPending = updateMutation.isPending || attendanceMutation.isPending

  // Save bar drives both forms — submit only what changed, reset reverts both.
  const handleSaveAll = () => {
    if (form.formState.isDirty) onSubmit()
    if (attendanceForm.formState.isDirty) onAttendanceSubmit()
  }

  const handleReset = () => {
    if (org) reset(orgFormDefaults(org))
    if (settings) attendanceForm.reset(attendanceFormDefaults(settings))
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

  const liveName = watch("name")
  const status = (org?.status ?? "active") as OrgStatus
  const statusStyle = STATUS_STYLES[status] ?? STATUS_STYLES.active
  const memberCount = org?.total_users ?? 0
  const plan = org?.plan ?? "free_50"

  if (!allowed) return null

  if (isLoading) {
    return <SettingsSkeleton />
  }

  return (
    <div className="mx-auto w-full max-w-4xl pb-24">
      <form onSubmit={onSubmit} noValidate className="flex flex-col gap-6">
        <section className="relative overflow-hidden rounded-2xl border">
          <div
            aria-hidden
            className="from-success/8 via-card to-card pointer-events-none absolute inset-0 bg-gradient-to-br"
          />
          <div
            aria-hidden
            className="pointer-events-none absolute inset-0 opacity-[0.4]"
            style={{
              backgroundImage: "radial-gradient(circle at 1px 1px, var(--border) 1px, transparent 0)",
              backgroundSize: "22px 22px",
              maskImage: "radial-gradient(120% 100% at 100% 0%, black, transparent 65%)",
              WebkitMaskImage: "radial-gradient(120% 100% at 100% 0%, black, transparent 65%)",
            }}
          />
          <div
            aria-hidden
            className="bg-success/15 pointer-events-none absolute -end-16 -top-20 size-56 rounded-full blur-3xl"
          />

          <div className="relative flex flex-col gap-5 p-6 sm:flex-row sm:items-center sm:gap-6 sm:p-8">
            <div className="relative shrink-0">
              <div className="from-success text-success-foreground ring-foreground/10 font-heading grid size-18 place-items-center rounded-2xl bg-gradient-to-br to-green-700 text-2xl font-semibold tracking-tight shadow-lg ring-1 select-none">
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
                  <span className="text-foreground font-medium tabular-nums">{memberCount.toLocaleString()}</span>
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

        <PlanCard plan={plan} expiresAt={org?.plan_expires_at} formatDate={formatDate} />

        <Section title={t("org.settings.general")} hint={t("org.settings.generalHint")}>
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

        <Section title={t("org.settings.details")} hint={t("org.settings.detailsHint")}>
          <DetailRow label={t("org.settings.orgId")}>
            <code className="bg-muted text-muted-foreground rounded-md px-2 py-1 font-mono text-xs">{orgId}</code>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              onClick={copyId}
              aria-label={t("org.settings.copyId")}
              className="ms-1"
            >
              {copied ? <CheckIcon className="text-success size-3.5" /> : <CopyIcon className="size-3.5" />}
            </Button>
          </DetailRow>

          <Separator />

          <DetailRow label={t("org.settings.created")} icon={<CalendarDaysIcon className="size-4" />}>
            <span className="text-foreground text-sm tabular-nums">{formatDate(org?.created_at)}</span>
          </DetailRow>

          <Separator />

          <DetailRow label={t("org.settings.updated")} icon={<HistoryIcon className="size-4" />}>
            <span className="text-foreground text-sm tabular-nums">{formatDate(org?.updated_at)}</span>
          </DetailRow>
        </Section>
      </form>

      <form onSubmit={onAttendanceSubmit} noValidate className="mt-6">
        <Section title={t("org.settings.attendance.title")} hint={t("org.settings.attendance.hint")}>
          <Field data-invalid={!!attendanceForm.formState.errors.attendance_present_threshold_percent || undefined}>
            <FieldLabel htmlFor="attendance-threshold" className="text-xs">
              {t("org.settings.attendance.thresholdLabel")}
            </FieldLabel>
            <Input
              id="attendance-threshold"
              type="number"
              min={1}
              max={100}
              className="h-10 w-full"
              aria-invalid={!!attendanceForm.formState.errors.attendance_present_threshold_percent}
              {...attendanceForm.register("attendance_present_threshold_percent", {
                valueAsNumber: true,
              })}
            />
            <p className="text-muted-foreground text-sm leading-relaxed">
              {t("org.settings.attendance.thresholdHint")}
            </p>
            {attendanceForm.formState.errors.attendance_present_threshold_percent && (
              <FieldError>{attendanceForm.formState.errors.attendance_present_threshold_percent.message}</FieldError>
            )}
          </Field>
        </Section>
      </form>

      <FormSaveBar form={form} onSave={handleSaveAll} onReset={handleReset} visible={dirty} isPending={anyPending} />
    </div>
  )
}

function PlanCard({
  plan,
  expiresAt,
  formatDate,
}: {
  plan: string
  expiresAt?: string
  formatDate: (value?: string) => string
}) {
  const { t } = useTranslation()
  const tier = planTier(plan)
  const style = PLAN_STYLES[tier] ?? PLAN_STYLES.free
  const Icon = style.icon
  const features = t(`org.settings.plan.features.${tier}`, {
    returnObjects: true,
  }) as string[]
  const hasExpiry = Boolean(expiresAt)

  return (
    <section className="ring-foreground/10 relative overflow-hidden rounded-2xl border ring-1">
      <div aria-hidden className={cn("pointer-events-none absolute inset-0 bg-gradient-to-br", style.wash)} />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 opacity-[0.4]"
        style={{
          backgroundImage: "radial-gradient(circle at 1px 1px, var(--border) 1px, transparent 0)",
          backgroundSize: "22px 22px",
          maskImage: "radial-gradient(120% 100% at 100% 0%, black, transparent 65%)",
          WebkitMaskImage: "radial-gradient(120% 100% at 100% 0%, black, transparent 65%)",
        }}
      />
      <div
        aria-hidden
        className={cn("pointer-events-none absolute -end-16 -top-20 size-56 rounded-full blur-3xl", style.glow)}
      />

      <div className="relative p-6 sm:p-8">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex items-center gap-4">
            <div
              className={cn(
                "grid size-14 shrink-0 place-items-center rounded-2xl shadow-sm ring-1 select-none",
                style.iconWrap
              )}
            >
              <Icon className="size-7" />
            </div>
            <div className="min-w-0">
              <p className="text-muted-foreground text-xs font-medium tracking-wide uppercase">
                {t("org.settings.plan.current")}
              </p>
              <h2 className="font-heading text-foreground text-2xl font-semibold tracking-tight">
                {t(`org.settings.plan.tiers.${tier}.name`)}
                <span className="text-muted-foreground ms-2 text-sm font-medium tabular-nums">
                  {t("plans.sizeSuffix", { size: planSize(plan) })}
                </span>
              </h2>
              <p className="text-muted-foreground mt-0.5 text-sm leading-relaxed">
                {t(`org.settings.plan.tiers.${tier}.tagline`)}
              </p>
            </div>
          </div>

          <span
            className={cn(
              "inline-flex h-7 shrink-0 items-center gap-1.5 self-start rounded-full border px-3 text-xs font-medium",
              style.chip
            )}
          >
            {hasExpiry ? (
              <>
                <CalendarClockIcon className="size-3.5" />
                <span className="tabular-nums">
                  {t("org.settings.plan.expiresOn", { date: formatDate(expiresAt) })}
                </span>
              </>
            ) : (
              <>
                <InfinityIcon className="size-3.5" />
                {t("org.settings.plan.perpetual")}
              </>
            )}
          </span>
        </div>

        <ul className="mt-6 grid gap-2.5 sm:grid-cols-2">
          {features.map((feature) => (
            <li key={feature} className="text-foreground/90 flex items-center gap-2 text-sm">
              <CheckIcon className={cn("size-4 shrink-0", style.check)} />
              {feature}
            </li>
          ))}
        </ul>

        <div className="text-muted-foreground mt-6 flex items-start gap-2 border-t pt-4 text-xs leading-relaxed">
          <InfoIcon className="mt-0.5 size-3.5 shrink-0" />
          {t("org.settings.plan.managedNote")}
        </div>
      </div>
    </section>
  )
}

/** Loading placeholder mirroring the settings form: hero card, plan card and
 * the two labelled sections. */
function SettingsSkeleton() {
  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 pb-24">
      <section className="rounded-2xl border p-6 sm:p-8">
        <div className="flex flex-col gap-5 sm:flex-row sm:items-center sm:gap-6">
          <Skeleton className="size-18 shrink-0 rounded-2xl" />
          <div className="min-w-0 flex-1">
            <Skeleton className="h-7 w-56" />
            <div className="mt-3 flex flex-wrap gap-x-5 gap-y-2">
              <Skeleton className="h-4 w-28" />
              <Skeleton className="h-4 w-40" />
            </div>
          </div>
        </div>
      </section>

      <Skeleton className="h-28 rounded-2xl" />

      {Array.from({ length: 2 }, (_, i) => (
        <section key={i} className="grid gap-x-8 gap-y-5 rounded-2xl border p-6 sm:p-8 md:grid-cols-[14rem_1fr]">
          <div className="md:pe-4">
            <Skeleton className="h-5 w-32" />
            <Skeleton className="mt-2 h-4 w-44" />
          </div>
          <div className="flex flex-col gap-5">
            <div className="flex flex-col gap-2">
              <Skeleton className="h-3 w-20" />
              <Skeleton className="h-10 w-full rounded-md" />
            </div>
            <div className="flex flex-col gap-2">
              <Skeleton className="h-3 w-24" />
              <Skeleton className="h-24 w-full rounded-md" />
            </div>
          </div>
        </section>
      ))}
    </div>
  )
}

function Section({ title, hint, children }: { title: string; hint: string; children: React.ReactNode }) {
  return (
    <section className="bg-card ring-foreground/10 grid gap-x-8 gap-y-5 rounded-2xl border p-6 ring-1 sm:p-8 md:grid-cols-[14rem_1fr]">
      <div className="md:pe-4">
        <h2 className="font-heading text-foreground text-base font-semibold tracking-tight">{title}</h2>
        <p className="text-muted-foreground mt-1 text-sm leading-relaxed">{hint}</p>
      </div>
      <div className="flex flex-col gap-5">{children}</div>
    </section>
  )
}

function DetailRow({ label, icon, children }: { label: string; icon?: React.ReactNode; children: React.ReactNode }) {
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
