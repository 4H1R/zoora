import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { SparklesIcon } from "lucide-react"
import { useEffect } from "react"
import { useTranslation } from "react-i18next"

import {
  getGetChangelogStatusQueryKey,
  useGetChangelog,
  usePostChangelogMarkSeen,
} from "@/api/changelog/changelog"
import type { GithubCom4H1RZooraInternalDomainChangelogEntry as Entry } from "@/api/model"
import { ChangelogMarkdown } from "@/components/changelog/markdown"
import { Eyebrow } from "@/components/eyebrow"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { orgHead } from "@/lib/org-head"

export const Route = createFileRoute("/_auth/org/whats-new")({
  head: () => orgHead("whatsNew.title"),
  component: WhatsNewPage,
})

function WhatsNewPage() {
  const { t, i18n } = useTranslation()
  const isFa = i18n.language.startsWith("fa")
  const queryClient = useQueryClient()
  const { data } = useGetChangelog({ page: 1 })
  const markSeen = usePostChangelogMarkSeen({
    mutation: {
      onSuccess: () =>
        queryClient.invalidateQueries({ queryKey: getGetChangelogStatusQueryKey() }),
    },
  })

  // Opening the feed acknowledges all current entries.
  useEffect(() => {
    markSeen.mutate()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const page = (data?.status === 200 && data.data.data) || undefined
  const entries = (page?.items as Entry[] | undefined) ?? []

  const fmtDate = (iso?: string) =>
    iso
      ? new Date(iso).toLocaleDateString(i18n.language, {
          day: "numeric",
          month: "short",
          year: "numeric",
        })
      : ""

  return (
    <div className="mx-auto w-full max-w-3xl">
      <header className="flex flex-col gap-1.5 pb-8">
        <div className="flex items-center gap-2">
          <SparklesIcon className="text-primary size-4" />
          <Eyebrow>{t("whatsNew.eyebrow")}</Eyebrow>
        </div>

        <h1 className="text-2xl font-semibold tracking-tight">{t("whatsNew.title")}</h1>
        <p className="text-muted-foreground text-sm">{t("whatsNew.subtitle")}</p>

        {entries.length > 0 && (
          <div className="mt-2 flex items-center gap-2">
            <span className="bg-primary size-1.5 animate-pulse-dot rounded-full" />
            <span className="text-muted-foreground font-mono text-xs tracking-wide">
              {t("whatsNew.count", { count: entries.length })}
            </span>
          </div>
        )}
      </header>

      {entries.length === 0 ? (
        <EmptyState hint={t("whatsNew.emptyHint")} title={t("whatsNew.empty")} />
      ) : (
        <ol className="relative">
          {/* The spine — a hairline the release nodes hang from, fading at the tail. */}
          <div
            aria-hidden
            className="absolute inset-y-2 start-[6px] w-0.5 rounded-full bg-gradient-to-b from-border via-border to-transparent"
          />

          {entries.map((e, i) => {
            const title = (isFa && e.title_fa) || e.title_en
            const body = (isFa && e.body_fa) || e.body_en || ""
            const major = !!e.is_major
            const newest = i === 0

            return (
              <li
                key={e.id}
                className="animate-reveal relative ps-8 pb-10 last:pb-0"
                style={{ animationDelay: `${Math.min(i, 8) * 70}ms` }}
              >
                {/* Node — filled + glowing for majors, hollow for routine releases. */}
                <span
                  aria-hidden
                  className={cn(
                    "absolute start-0 top-1.5 grid size-3.5 place-items-center rounded-full transition-transform",
                    major
                      ? "bg-primary ring-primary/25 shadow-[0_0_0_4px_var(--background)] ring-4"
                      : "bg-background border-border border-2",
                  )}
                >
                  {major && (
                    <span className="bg-primary size-3.5 animate-ping rounded-full opacity-40" />
                  )}
                </span>

                <time className="text-muted-foreground font-mono text-xs tracking-wide">
                  {fmtDate(e.published_at)}
                </time>

                <article
                  className={cn(
                    "group bg-card border-border mt-2 rounded-xl border p-5 transition-colors duration-200",
                    "hover:border-foreground/25",
                    major && "ring-primary/15 ring-1",
                  )}
                >
                  <div className="flex flex-wrap items-center gap-2">
                    {e.version && (
                      <Badge
                        variant="outline"
                        className="font-mono text-xs font-medium tabular-nums"
                      >
                        {e.version}
                      </Badge>
                    )}
                    {major && (
                      <Badge className="gap-1">
                        <SparklesIcon className="size-3" />
                        {t("whatsNew.major")}
                      </Badge>
                    )}
                    {newest && !major && (
                      <Badge variant="secondary">{t("whatsNew.latest")}</Badge>
                    )}
                  </div>

                  <h2 className="mt-3 text-xl font-semibold tracking-tight text-balance">
                    {title}
                  </h2>

                  {body && (
                    <div className="mt-2">
                      <ChangelogMarkdown>{body}</ChangelogMarkdown>
                    </div>
                  )}
                </article>
              </li>
            )
          })}
        </ol>
      )}
    </div>
  )
}

function EmptyState({ hint, title }: { hint: string; title: string }) {
  return (
    <div className="relative isolate overflow-hidden rounded-2xl border px-6 py-20 text-center">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 opacity-70 [mask-image:radial-gradient(60%_50%_at_50%_30%,black,transparent)]"
        style={{
          background:
            "radial-gradient(circle at 50% 25%, color-mix(in oklch, var(--primary) 16%, transparent), transparent 60%)",
        }}
      />
      <div className="bg-primary/10 text-primary ring-primary/20 mx-auto grid size-14 place-items-center rounded-2xl ring-1">
        <SparklesIcon className="size-6" />
      </div>
      <p className="mt-5 text-lg font-semibold tracking-tight">{title}</p>
      <p className="text-muted-foreground mx-auto mt-1 max-w-xs text-sm text-pretty">
        {hint}
      </p>
    </div>
  )
}
