import { Link } from "@tanstack/react-router"
import { CheckIcon, LockIcon, MessagesSquareIcon, RocketIcon, SparklesIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

// ChatUpgrade is the paywall shown at /org/conversations when the org's plan
// doesn't include the `chat` feature. It reaches only users whose nav surfaces
// the route despite the missing feature (conversations:manage) or who deep-link
// in. The CTA adapts: billing:manage holders get a checkout link; everyone else
// gets a "ask your admin" nudge, since they can't purchase.
export function ChatUpgrade({ canUpgrade }: { canUpgrade: boolean }) {
  const { t } = useTranslation()

  const features = [
    t("conversations.upgrade.features.groups"),
    t("conversations.upgrade.features.channels"),
    t("conversations.upgrade.features.reactions"),
    t("conversations.upgrade.features.search"),
  ]

  return (
    <div className="mx-auto flex w-full max-w-2xl flex-1 items-center py-8">
      <div className="ring-foreground/10 from-primary/[0.03] via-card to-card relative w-full overflow-hidden rounded-3xl bg-gradient-to-b ring-1">
        {/* Atmospheric glow — decorative, no Tailwind equivalent. */}
        <div className="bg-primary/[0.07] pointer-events-none absolute start-1/2 -top-32 size-48 -translate-x-1/2 rounded-full blur-3xl rtl:translate-x-1/2" />

        <div className="relative flex flex-col items-center px-6 py-12 text-center sm:px-12 sm:py-16">
          {/* Locked chat glyph */}
          <div className="relative mb-6">
            <div className="bg-primary/10 ring-primary/15 text-primary flex size-16 items-center justify-center rounded-2xl ring-1">
              <MessagesSquareIcon className="size-8" />
            </div>
            <span className="bg-background ring-foreground/10 text-muted-foreground absolute -end-1.5 -bottom-1.5 flex size-7 items-center justify-center rounded-full ring-1">
              <LockIcon className="size-3.5" />
            </span>
          </div>

          <Badge variant="outline" className="mb-4 gap-1.5 py-1 ps-1.5 pe-2.5">
            <RocketIcon className="text-primary" />
            {t("conversations.upgrade.availableOn")}
          </Badge>

          <h1 className="text-2xl font-bold tracking-tight text-balance sm:text-3xl">
            {t("conversations.upgrade.title")}
          </h1>
          <p className="text-muted-foreground mt-3 max-w-md text-sm leading-relaxed text-pretty">
            {t("conversations.upgrade.description")}
          </p>

          {/* Feature grid */}
          <ul className="mt-8 grid w-full max-w-md grid-cols-1 gap-3 text-start sm:grid-cols-2">
            {features.map((feature) => (
              <li
                key={feature}
                className="bg-card/60 ring-foreground/10 flex items-center gap-2.5 rounded-xl px-3 py-2.5 text-sm ring-1"
              >
                <CheckIcon className="text-primary size-4 shrink-0" />
                <span className="text-foreground/90 leading-tight">{feature}</span>
              </li>
            ))}
          </ul>

          {/* CTA — adapts to whether this user can actually buy a plan. */}
          <div className="mt-10 flex flex-col items-center gap-3">
            {canUpgrade ? (
              <Button size="lg" render={<Link to="/org/billing" />}>
                <SparklesIcon />
                {t("conversations.upgrade.cta")}
              </Button>
            ) : (
              <p
                className={cn(
                  "text-muted-foreground bg-muted/50 ring-foreground/10 rounded-full px-4 py-2 text-sm ring-1"
                )}
              >
                {t("conversations.upgrade.askAdmin")}
              </p>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
