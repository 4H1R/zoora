import { createFileRoute, redirect } from "@tanstack/react-router"

import { currentSlug } from "@/lib/tenant"

import { FinalCta } from "./-landing/cta"
import { Faq } from "./-landing/faq"
import { Features } from "./-landing/features"
import { LandingFooter } from "./-landing/footer"
import { Hero } from "./-landing/hero"
import { LandingNav } from "./-landing/nav"
import { Pricing } from "./-landing/pricing"
import { Stats } from "./-landing/stats"
import { Workflow } from "./-landing/workflow"

export const Route = createFileRoute("/")({
  beforeLoad: () => {
    if (currentSlug() !== "") throw redirect({ to: "/login" })
  },
  component: RouteComponent,
})

function RouteComponent() {
  return (
    <main className="bg-background text-foreground relative overflow-x-clip">
      <LandingNav />
      <Hero />
      <Stats />
      <Features />
      <Workflow />
      <Pricing />
      <Faq />
      <FinalCta />
      <LandingFooter />
    </main>
  )
}
