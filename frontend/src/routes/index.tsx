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
  // The landing page only belongs on the apex / canonical www host. On any
  // tenant or admin subdomain, send `/` to `/login` — the `_guest` layout bounces an
  // already-authenticated user on to `/org/dashboard` or `/admin/dashboard`.
  //
  // `/` is SSR'd per request (not prerendered) so currentSlug resolves the
  // tenant from the request Host header on the server: tenant/admin hosts are
  // redirected here during SSR (302, before any HTML paints), and only the apex
  // renders the landing. On client navigations currentSlug reads window instead.
  beforeLoad: () => {
    if (currentSlug() !== "") {
      throw redirect({ to: "/login" })
    }
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
