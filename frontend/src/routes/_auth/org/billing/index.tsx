import type { GithubCom4H1RZooraInternalDomainPlanPrice as PlanPrice } from "@/api/model"
import type { ErrorType } from "@/api/mutator/custom-instance"

import { createFileRoute, Link } from "@tanstack/react-router"
import { CheckIcon, HistoryIcon, SparklesIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { useGetBillingPlans, usePostBillingCheckout } from "@/api/billing/billing"
import {
  GithubCom4H1RZooraInternalDomainBillingInterval as BillingInterval,
  GithubCom4H1RZooraInternalDomainGatewayName as GatewayName,
  GithubCom4H1RZooraInternalDomainPlan as Plan,
} from "@/api/model"
import { useGetOrganizationsId } from "@/api/organizations/organizations"
import { useGetUsersMe } from "@/api/users/users"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { Spinner } from "@/components/ui/spinner"
import { useOrgGuard } from "@/lib/access"
import { useFormatToman } from "@/lib/billing"
import { useFormatDate } from "@/lib/data-table"
import { orgHead } from "@/lib/org-head"
import { PLAN_SIZES, planRank, planSize, planTier, TIER_ICON } from "@/lib/plan"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org/billing/")({
  head: () => orgHead("billing.title"),
  component: BillingPage,
})

function BillingPage() {
  const { t } = useTranslation()
  const allowed = useOrgGuard("billing:manage")
  const formatDate = useFormatDate()
  const formatToman = useFormatToman()

  const [interval, setInterval] = useState<BillingInterval>(BillingInterval.BillingIntervalMonthly)
  const [size, setSize] = useState<number | null>(null)

  const { data: meResponse } = useGetUsersMe()
  const orgId = (meResponse?.status === 200 && meResponse.data.data?.organization_id) || ""

  const { data: orgResponse } = useGetOrganizationsId(orgId)
  const org = orgResponse?.status === 200 ? orgResponse.data.data : undefined
  const currentPlan = (org?.plan ?? Plan.PlanFree) as string
  // Default the size picker to the org's current capacity once it loads.
  const selectedSize = size ?? planSize(currentPlan)

  const { data: plansResponse, isLoading } = useGetBillingPlans()
  const prices = (plansResponse?.status === 200 && plansResponse.data.data) || []

  const checkout = usePostBillingCheckout({
    mutation: {
      onSuccess: (res) => {
        const url = res.status === 201 ? res.data.data?.redirect_url : undefined
        if (url) window.location.href = url
      },
      onError: (err: ErrorType<unknown>) => {
        const code = (err.response?.data as { error?: { code?: string } } | undefined)?.error?.code
        if (code === "DOWNGRADE_NOT_ALLOWED") toast.error(t("billing.downgradeNotAllowed"))
        else toast.error(t("billing.checkoutError"))
      },
    },
  })

  const visiblePrices = prices
    .filter((p) => p.interval === interval && p.plan && planSize(p.plan) === selectedSize)
    .sort((a, b) => planRank(a.plan) - planRank(b.plan))

  const handleCheckout = (plan: string) => {
    checkout.mutate({
      data: { plan: plan as Plan, interval, gateway: GatewayName.GatewayZarinpal },
    })
  }

  if (!allowed) return null

  return (
    <div className="mx-auto flex w-full max-w-5xl flex-col gap-6">
      <PageHeader
        title={t("billing.title")}
        actions={
          <Button variant="outline" size="sm" render={<Link to="/org/billing/invoices" />}>
            <HistoryIcon />
            {t("billing.viewHistory")}
          </Button>
        }
      />

      <CurrentPlanCard plan={currentPlan} expiresAt={org?.plan_expires_at} formatDate={formatDate} />

      <div className="flex flex-wrap items-center justify-between gap-4">
        <h2 className="font-heading text-lg font-semibold tracking-tight">{t("billing.choosePlan")}</h2>
        <div className="flex flex-wrap items-center gap-3">
          <SizeToggle value={selectedSize} onChange={setSize} />
          <IntervalToggle value={interval} onChange={setInterval} />
        </div>
      </div>

      {isLoading ? (
        <PlansSkeleton />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {visiblePrices.map((price) => (
            <PlanCard
              key={price.id}
              price={price}
              interval={interval}
              currentPlan={currentPlan}
              formatToman={formatToman}
              onCheckout={handleCheckout}
              isPending={checkout.isPending && checkout.variables?.data.plan === price.plan}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function SizeToggle({ value, onChange }: { value: number; onChange: (v: number) => void }) {
  const { t } = useTranslation()
  return (
    <div className="bg-muted inline-flex rounded-lg p-1">
      {PLAN_SIZES.map((size) => (
        <button
          key={size}
          type="button"
          onClick={() => onChange(size)}
          className={cn(
            "rounded-md px-3 py-1.5 text-sm font-medium transition-colors tabular-nums",
            value === size
              ? "bg-background text-foreground shadow-sm"
              : "text-muted-foreground hover:text-foreground"
          )}
        >
          {t("billing.sizeOption", { size })}
        </button>
      ))}
    </div>
  )
}

function IntervalToggle({ value, onChange }: { value: BillingInterval; onChange: (v: BillingInterval) => void }) {
  const { t } = useTranslation()
  const options: { value: BillingInterval; label: string }[] = [
    { value: BillingInterval.BillingIntervalMonthly, label: t("billing.monthly") },
    { value: BillingInterval.BillingIntervalYearly, label: t("billing.yearly") },
  ]
  return (
    <div className="bg-muted inline-flex rounded-lg p-1">
      {options.map((opt) => (
        <button
          key={opt.value}
          type="button"
          onClick={() => onChange(opt.value)}
          className={cn(
            "rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
            value === opt.value
              ? "bg-background text-foreground shadow-sm"
              : "text-muted-foreground hover:text-foreground"
          )}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}

function CurrentPlanCard({
  plan,
  expiresAt,
  formatDate,
}: {
  plan: string
  expiresAt?: string
  formatDate: (value?: string) => string
}) {
  const { t } = useTranslation()
  const Icon = TIER_ICON[planTier(plan)] ?? SparklesIcon
  return (
    <section className="bg-card ring-foreground/10 flex items-center gap-4 rounded-2xl border p-5 ring-1">
      <div className="bg-primary/10 text-primary grid size-12 shrink-0 place-items-center rounded-xl">
        <Icon className="size-6" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-muted-foreground text-xs font-medium tracking-wide uppercase">{t("billing.currentPlan")}</p>
        <h2 className="font-heading text-xl font-semibold tracking-tight">
          {t(`plans.tiers.${planTier(plan)}`)}
          <span className="text-muted-foreground ms-2 text-sm font-medium tabular-nums">
            {t("plans.sizeSuffix", { size: planSize(plan) })}
          </span>
        </h2>
      </div>
      <div className="text-end text-sm">
        <p className="text-muted-foreground text-xs">{t("billing.expiresAt")}</p>
        <p className="text-foreground font-medium tabular-nums">
          {expiresAt ? formatDate(expiresAt) : t("billing.perpetual")}
        </p>
      </div>
    </section>
  )
}

function PlanCard({
  price,
  interval,
  currentPlan,
  formatToman,
  onCheckout,
  isPending,
}: {
  price: PlanPrice
  interval: BillingInterval
  currentPlan: string
  formatToman: (amountRial?: number) => string
  onCheckout: (plan: string) => void
  isPending: boolean
}) {
  const { t } = useTranslation()
  const plan = price.plan!
  const Icon = TIER_ICON[planTier(plan)] ?? SparklesIcon

  const isCurrent = plan === currentPlan
  const isDowngrade = planRank(plan) < planRank(currentPlan)
  const perLabel = interval === BillingInterval.BillingIntervalYearly ? t("billing.perYear") : t("billing.perMonth")

  return (
    <div
      className={cn(
        "bg-card ring-foreground/10 relative flex flex-col gap-4 rounded-2xl border p-6 ring-1",
        isCurrent && "ring-primary/40 ring-2"
      )}
    >
      <div className="flex items-center gap-3">
        <div className="bg-muted grid size-10 place-items-center rounded-xl">
          <Icon className="size-5" />
        </div>
        <h3 className="font-heading text-lg font-semibold tracking-tight">
          {t(`plans.tiers.${planTier(plan)}`)}
          <span className="text-muted-foreground ms-2 text-sm font-medium tabular-nums">
            {t("plans.sizeSuffix", { size: planSize(plan) })}
          </span>
        </h3>
      </div>

      <div className="flex items-baseline gap-1.5">
        <span className="text-foreground text-3xl font-bold tabular-nums">{formatToman(price.amount)}</span>
        <span className="text-muted-foreground text-sm">{t("billing.toman")}</span>
        <span className="text-muted-foreground text-sm">{perLabel}</span>
      </div>

      <Button
        className="mt-auto w-full"
        disabled={isDowngrade || isPending}
        variant={isCurrent ? "outline" : "default"}
        onClick={() => onCheckout(plan)}
      >
        {isPending && <Spinner />}
        {isCurrent ? t("billing.renew") : t("billing.subscribe")}
      </Button>
      {isDowngrade && (
        <p className="text-muted-foreground flex items-start gap-1.5 text-xs leading-relaxed">
          <CheckIcon className="mt-0.5 size-3.5 shrink-0 opacity-0" />
          {t("billing.downgradeNotAllowed")}
        </p>
      )}
    </div>
  )
}

/** Loading placeholder mirroring the plan cards grid. */
function PlansSkeleton() {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: 3 }, (_, i) => (
        <div key={i} className="bg-card ring-foreground/10 flex flex-col gap-4 rounded-2xl border p-6 ring-1">
          <div className="flex items-center gap-3">
            <Skeleton className="size-10 rounded-xl" />
            <Skeleton className="h-6 w-24" />
          </div>
          <Skeleton className="h-9 w-32" />
          <Skeleton className="mt-auto h-9 w-full rounded-md" />
        </div>
      ))}
    </div>
  )
}
