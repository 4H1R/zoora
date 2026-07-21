import type { ErrorType } from "@/api/mutator/custom-instance"

import { zodResolver } from "@hookform/resolvers/zod"
import { BuildingIcon, CalendarDaysIcon, EyeIcon, EyeOffIcon, LockIcon, ShieldCheckIcon, UserIcon } from "lucide-react"
import { useState } from "react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import { AUTH_TOKEN_KEY } from "@/api/mutator/custom-instance"
import { useGetOrganizationsId } from "@/api/organizations/organizations"
import { useGetUsersMe, usePostUsersMePassword } from "@/api/users/users"
import { AccountCustomFields } from "@/components/account/account-custom-fields"
import { LanguageSwitcher } from "@/components/language-switcher"
import { ThemeToggle } from "@/components/theme-toggle"
import { Button } from "@/components/ui/button"
import { Field, FieldError, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { useFormatDate } from "@/lib/data-table"
import { useRoleName } from "@/lib/permissions"

function initials(name?: string) {
  if (!name) return "—"
  const words = name.trim().split(/\s+/).filter(Boolean)
  if (words.length === 0) return "—"
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase()
  return (words[0][0] + words[words.length - 1][0]).toUpperCase()
}

/** Shared account settings surface: profile hero, password change, preferences.
    Rendered by both the org (`/org/account`) and admin (`/admin/account`) routes.
    Admin hides the profile hero (`showProfile={false}`) — platform admins manage
    their identity elsewhere. */
export function AccountSettings({ showProfile = true }: { showProfile?: boolean }) {
  const { t } = useTranslation()
  const formatDate = useFormatDate()
  const roleName = useRoleName()

  const { data: meResponse, isLoading } = useGetUsersMe()
  const me = meResponse?.status === 200 ? meResponse.data.data : undefined

  const orgId = me?.organization_id ?? ""
  const { data: orgResponse } = useGetOrganizationsId(orgId, { query: { enabled: !!orgId } })
  const orgName = orgResponse?.status === 200 ? orgResponse.data.data?.name : undefined

  if (isLoading) return <AccountSkeleton showProfile={showProfile} />

  const role = me?.is_admin ? t("admin.roleAdmin") : me?.role?.name ? roleName(me.role.name) : t("admin.roleMember")

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 pb-24">
      {/* Profile — read-only. Name/username are admin-managed (anti-troll); this
          card only displays them. Hidden for platform admins. */}
      {showProfile && (
        <section className="ring-foreground/10 relative overflow-hidden rounded-2xl border ring-1">
          <div
            aria-hidden
            className="from-primary/8 via-card to-card pointer-events-none absolute inset-0 bg-gradient-to-br"
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
            className="bg-primary/15 pointer-events-none absolute -end-16 -top-20 size-56 rounded-full blur-3xl"
          />

          <div className="relative flex flex-col gap-5 p-6 sm:flex-row sm:items-center sm:gap-6 sm:p-8">
            <div className="from-primary text-primary-foreground ring-foreground/10 font-heading grid size-18 shrink-0 place-items-center rounded-2xl bg-gradient-to-br to-indigo-700 text-2xl font-semibold tracking-tight shadow-lg ring-1 select-none">
              {initials(me?.name)}
            </div>

            <div className="min-w-0 flex-1">
              <div className="mb-2 flex flex-wrap items-center gap-2.5">
                <h1 className="font-heading text-foreground truncate text-2xl font-semibold tracking-tight sm:text-3xl">
                  {me?.name || "—"}
                </h1>
                <span className="border-primary/25 bg-primary/10 text-primary inline-flex h-6 items-center gap-1.5 rounded-full border px-2.5 text-xs font-medium">
                  <ShieldCheckIcon className="size-3.5" />
                  {role}
                </span>
              </div>

              <div className="text-muted-foreground flex flex-wrap items-center gap-x-5 gap-y-1.5 text-sm">
                <span className="inline-flex items-center gap-1.5">
                  <UserIcon className="size-4" />
                  <span className="text-foreground font-medium">@{me?.username}</span>
                </span>
                {orgName && (
                  <span className="inline-flex items-center gap-1.5">
                    <BuildingIcon className="size-4" />
                    {orgName}
                  </span>
                )}
                {me?.created_at && (
                  <span className="inline-flex items-center gap-1.5">
                    <CalendarDaysIcon className="size-4" />
                    {t("account.created")} {formatDate(me.created_at)}
                  </span>
                )}
              </div>

              <p className="text-muted-foreground mt-3 text-xs leading-relaxed">{t("account.lockedNote")}</p>
            </div>
          </div>
        </section>
      )}

      {showProfile && <AccountCustomFields />}

      <SecuritySection />

      <Section
        title={t("account.preferences.title")}
        hint={t("account.preferences.hint")}
        icon={<UserIcon className="size-4" />}
      >
        <PreferenceRow label={t("account.preferences.theme")} hint={t("account.preferences.themeHint")}>
          <ThemeToggle />
        </PreferenceRow>
        <div className="bg-border h-px w-full" />
        <PreferenceRow label={t("account.preferences.language")} hint={t("account.preferences.languageHint")}>
          <LanguageSwitcher />
        </PreferenceRow>
      </Section>
    </div>
  )
}

function makePasswordSchema(t: (k: string) => string) {
  return z
    .object({
      current_password: z.string().min(1),
      new_password: z.string().min(8, t("account.security.errors.min")),
      confirm_password: z.string(),
    })
    .refine((v) => v.new_password === v.confirm_password, {
      path: ["confirm_password"],
      message: t("account.security.errors.mismatch"),
    })
    .refine((v) => v.new_password !== v.current_password, {
      path: ["new_password"],
      message: t("account.security.errors.sameAsCurrent"),
    })
}

type PasswordValues = z.infer<ReturnType<typeof makePasswordSchema>>

function SecuritySection() {
  const { t } = useTranslation()

  const form = useForm<PasswordValues>({
    resolver: zodResolver(makePasswordSchema(t)),
    defaultValues: { current_password: "", new_password: "", confirm_password: "" },
  })
  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = form

  const mutation = usePostUsersMePassword({
    mutation: {
      onSuccess: (res) => {
        // Swap in the freshly-issued token so this device stays signed in while
        // every other session was revoked server-side.
        if (res.status === 200 && res.data.data?.token) {
          localStorage.setItem(AUTH_TOKEN_KEY, res.data.data.token)
        }
        toast.success(t("account.security.success"))
        reset()
      },
      onError: (err) => {
        const status = (err as ErrorType<unknown>).response?.status
        if (status === 401) {
          // Backend rejected the current password.
          setError("current_password", { message: t("account.security.errors.currentWrong") })
          return
        }
        toast.error(t("account.security.errors.generic"))
      },
    },
  })

  const onSubmit = handleSubmit((values) => {
    mutation.mutate({
      data: { current_password: values.current_password, new_password: values.new_password },
    })
  })

  return (
    <form onSubmit={onSubmit} noValidate>
      <Section
        title={t("account.security.title")}
        hint={t("account.security.hint")}
        icon={<LockIcon className="size-4" />}
      >
        <PasswordField
          id="current-password"
          label={t("account.security.currentPassword")}
          autoComplete="current-password"
          error={errors.current_password?.message}
          invalid={!!errors.current_password}
          {...register("current_password")}
        />
        <PasswordField
          id="new-password"
          label={t("account.security.newPassword")}
          autoComplete="new-password"
          error={errors.new_password?.message}
          invalid={!!errors.new_password}
          {...register("new_password")}
        />
        <PasswordField
          id="confirm-password"
          label={t("account.security.confirmPassword")}
          autoComplete="new-password"
          error={errors.confirm_password?.message}
          invalid={!!errors.confirm_password}
          {...register("confirm_password")}
        />
        <div className="flex justify-end">
          <Button type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? t("account.security.submitting") : t("account.security.submit")}
          </Button>
        </div>
      </Section>
    </form>
  )
}

type PasswordFieldProps = React.ComponentProps<"input"> & {
  id: string
  label: string
  error?: string
  invalid?: boolean
}

function PasswordField({ id, label, error, invalid, ...props }: PasswordFieldProps) {
  const { t } = useTranslation()
  const [show, setShow] = useState(false)

  return (
    <Field data-invalid={invalid || undefined}>
      <FieldLabel htmlFor={id} className="text-xs">
        {label}
      </FieldLabel>
      <div className="relative">
        <Input id={id} type={show ? "text" : "password"} className="h-10 pe-10" aria-invalid={invalid} {...props} />
        <button
          type="button"
          onClick={() => setShow((s) => !s)}
          aria-label={show ? t("account.security.hide") : t("account.security.show")}
          className="text-muted-foreground hover:text-foreground absolute end-1 top-1/2 grid size-8 -translate-y-1/2 place-items-center rounded-md transition-colors"
        >
          {show ? <EyeOffIcon className="size-4" /> : <EyeIcon className="size-4" />}
        </button>
      </div>
      {error && <FieldError>{error}</FieldError>}
    </Field>
  )
}

function Section({
  title,
  hint,
  icon,
  children,
}: {
  title: string
  hint: string
  icon?: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <section className="bg-card ring-foreground/10 grid gap-x-8 gap-y-5 rounded-2xl border p-6 ring-1 sm:p-8 md:grid-cols-[14rem_1fr]">
      <div className="md:pe-4">
        <h2 className="font-heading text-foreground flex items-center gap-2 text-base font-semibold tracking-tight">
          {icon}
          {title}
        </h2>
        <p className="text-muted-foreground mt-1 text-sm leading-relaxed">{hint}</p>
      </div>
      <div className="flex flex-col gap-5">{children}</div>
    </section>
  )
}

function PreferenceRow({ label, hint, children }: { label: string; hint: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <div className="min-w-0">
        <p className="text-foreground text-sm font-medium">{label}</p>
        <p className="text-muted-foreground mt-0.5 text-xs leading-relaxed">{hint}</p>
      </div>
      <div className="shrink-0">{children}</div>
    </div>
  )
}

/** Loading placeholder mirroring the account layout: profile hero + two sections. */
function AccountSkeleton({ showProfile = true }: { showProfile?: boolean }) {
  return (
    <div className="mx-auto flex w-full max-w-3xl flex-col gap-6 pb-24">
      {showProfile && (
        <section className="rounded-2xl border p-6 sm:p-8">
          <div className="flex flex-col gap-5 sm:flex-row sm:items-center sm:gap-6">
            <Skeleton className="size-18 shrink-0 rounded-2xl" />
            <div className="min-w-0 flex-1">
              <Skeleton className="h-7 w-48" />
              <div className="mt-3 flex flex-wrap gap-x-5 gap-y-2">
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-4 w-32" />
              </div>
              <Skeleton className="mt-3 h-3 w-64" />
            </div>
          </div>
        </section>
      )}

      {Array.from({ length: 2 }, (_, i) => (
        <section key={i} className="grid gap-x-8 gap-y-5 rounded-2xl border p-6 sm:p-8 md:grid-cols-[14rem_1fr]">
          <div className="md:pe-4">
            <Skeleton className="h-5 w-28" />
            <Skeleton className="mt-2 h-4 w-40" />
          </div>
          <div className="flex flex-col gap-5">
            <div className="flex flex-col gap-2">
              <Skeleton className="h-3 w-20" />
              <Skeleton className="h-10 w-full rounded-md" />
            </div>
            <div className="flex flex-col gap-2">
              <Skeleton className="h-3 w-24" />
              <Skeleton className="h-10 w-full rounded-md" />
            </div>
          </div>
        </section>
      ))}
    </div>
  )
}
