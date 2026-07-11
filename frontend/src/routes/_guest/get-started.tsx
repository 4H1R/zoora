import { zodResolver } from "@hookform/resolvers/zod"
import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowRight, CheckCircle2 } from "lucide-react"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { z } from "zod"

import i18n from "@/i18n"

import { usePostLeads } from "@/api/leads/leads"
import GridBackground from "@/components/auth/gradient-background"
import { LanguageSwitcher } from "@/components/language-switcher"
import { Logo } from "@/components/logo"
import { ThemeToggle } from "@/components/theme-toggle"
import { Button, buttonVariants } from "@/components/ui/button"
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { Spinner } from "@/components/ui/spinner"
import { Textarea } from "@/components/ui/textarea"
import { PLAN_TIERS } from "@/lib/plan"
import { cn } from "@/lib/utils"

// Plan carried from the pricing card the visitor clicked. Constrained to the
// known tiers; anything else is dropped so the field stays advisory-only.
const searchSchema = z.object({
  plan: z
    .enum(PLAN_TIERS)
    .optional()
    .catch(undefined),
})

export const Route = createFileRoute("/_guest/get-started")({
  head: () => {
    const title = `${i18n.t("landing.getStarted.title")} — ${i18n.t("common.brandName")}`
    return { meta: [{ title }, { name: "description", content: title }] }
  },
  validateSearch: searchSchema,
  component: GetStartedComponent,
})

const leadSchema = z.object({
  name: z.string().min(2).max(255),
  phone: z.string().min(3).max(32),
  org_name: z.string().min(2).max(255),
  note: z.string().max(2000).optional(),
  // Honeypot: hidden from humans; a filled value flags a bot server-side.
  website: z.string().optional(),
})

type LeadFormValues = z.infer<typeof leadSchema>

function GetStartedComponent() {
  const { t } = useTranslation()
  const { plan } = Route.useSearch()

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitSuccessful },
  } = useForm<LeadFormValues>({
    resolver: zodResolver(leadSchema),
    defaultValues: { name: "", phone: "", org_name: "", note: "", website: "" },
  })

  const submit = usePostLeads({
    mutation: {
      onError: () => toast.error(t("landing.getStarted.error")),
    },
  })

  const onSubmit = handleSubmit((values) => {
    if (values.website) return // client-side honeypot short-circuit
    submit.mutate({
      data: {
        name: values.name,
        phone: values.phone,
        org_name: values.org_name,
        note: values.note || undefined,
        plan: plan || undefined,
      },
    })
  })

  const succeeded = submit.isSuccess && isSubmitSuccessful

  return (
    <div className="bg-muted/50 relative min-h-svh">
      <GridBackground />
      <header className="bg-background relative z-10 flex items-center justify-between px-6 py-5.5 md:px-8">
        <Link to="/" className="flex items-center">
          <Logo className="text-xl" />
        </Link>
        <div className="flex items-center gap-2">
          <ThemeToggle />
          <div className="bg-border/50 mx-1 h-4 w-px" />
          <LanguageSwitcher />
        </div>
      </header>

      <main className="relative z-10 grid min-h-[calc(100svh-72px)] place-items-center px-5 pt-6 pb-20">
        <div className="border-border bg-background w-full max-w-lg rounded-xl border p-8 shadow-sm">
          {succeeded ? (
            <div className="flex flex-col items-center py-6 text-center">
              <div className="bg-primary/10 text-primary flex size-14 items-center justify-center rounded-full">
                <CheckCircle2 className="size-7" />
              </div>
              <h1 className="font-heading mt-6 text-2xl font-semibold tracking-tight">
                {t("landing.getStarted.successTitle")}
              </h1>
              <p className="text-muted-foreground mt-3 max-w-sm text-sm leading-relaxed">
                {t("landing.getStarted.successDescription")}
              </p>
              <Link to="/" className={cn(buttonVariants({ variant: "outline" }), "mt-7 rounded-full")}>
                {t("landing.getStarted.backHome")}
              </Link>
            </div>
          ) : (
            <>
              <h1 className="text-2xl font-semibold tracking-tight">{t("landing.getStarted.title")}</h1>
              <p className="text-muted-foreground mt-2 text-sm leading-relaxed">
                {t("landing.getStarted.subtitle")}
              </p>
              {plan ? (
                <div className="border-primary/30 bg-primary/5 text-primary mt-5 inline-flex items-center gap-2 rounded-full border px-3.5 py-1 text-xs font-medium">
                  {t("landing.getStarted.selectedPlan", { plan: t(`plans.tiers.${plan}`) })}
                </div>
              ) : null}

              <form onSubmit={onSubmit} className="mt-6">
                <FieldGroup>
                  <Field data-invalid={!!errors.name || undefined}>
                    <FieldLabel>{t("landing.getStarted.name")}</FieldLabel>
                    <Input {...register("name")} placeholder={t("landing.getStarted.namePlaceholder")} />
                    <FieldError errors={[errors.name]} />
                  </Field>
                  <Field data-invalid={!!errors.phone || undefined}>
                    <FieldLabel>{t("landing.getStarted.phone")}</FieldLabel>
                    <Input
                      {...register("phone")}
                      type="tel"
                      dir="ltr"
                      placeholder={t("landing.getStarted.phonePlaceholder")}
                    />
                    <FieldError errors={[errors.phone]} />
                  </Field>
                  <Field data-invalid={!!errors.org_name || undefined}>
                    <FieldLabel>{t("landing.getStarted.orgName")}</FieldLabel>
                    <Input {...register("org_name")} placeholder={t("landing.getStarted.orgNamePlaceholder")} />
                    <FieldError errors={[errors.org_name]} />
                  </Field>
                  <Field>
                    <FieldLabel>{t("landing.getStarted.note")}</FieldLabel>
                    <Textarea {...register("note")} placeholder={t("landing.getStarted.notePlaceholder")} rows={3} />
                  </Field>

                  {/* Honeypot — visually removed, off the tab order, hidden from AT. */}
                  <div aria-hidden className="pointer-events-none absolute -left-[9999px] opacity-0">
                    <label>
                      Website
                      <input {...register("website")} type="text" tabIndex={-1} autoComplete="off" />
                    </label>
                  </div>

                  <Button type="submit" size="lg" className="mt-2 w-full rounded-full" disabled={submit.isPending}>
                    {submit.isPending ? <Spinner /> : null}
                    {t("landing.getStarted.submit")}
                    {!submit.isPending ? (
                      <ArrowRight className="rtl:-scale-x-100" data-icon="inline-end" />
                    ) : null}
                  </Button>
                </FieldGroup>
              </form>
            </>
          )}
        </div>
      </main>
    </div>
  )
}
