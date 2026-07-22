import type { GithubCom4H1RZooraInternalDomainTutorial as Tutorial } from "@/api/model"

import { createFileRoute } from "@tanstack/react-router"
import { GraduationCapIcon, PlayIcon, SearchIcon, VideoIcon } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { useGetTutorials } from "@/api/tutorials/tutorials"
import { ChangelogMarkdown } from "@/components/changelog/markdown"
import { Eyebrow } from "@/components/eyebrow"
import { aparatEmbedUrl } from "@/components/tutorials/aparat"
import { Dialog, DialogContent } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { orgHead } from "@/lib/org-head"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_auth/org/tutorials")({
  head: () => orgHead("tutorials.title"),
  component: TutorialsPage,
})

function TutorialsPage() {
  const { t, i18n } = useTranslation()
  const isFa = i18n.language.startsWith("fa")
  const { data, isLoading } = useGetTutorials()
  const tutorials = ((data?.status === 200 && data.data.data) || []) as Tutorial[]

  const [query, setQuery] = useState("")
  const [active, setActive] = useState<Tutorial | null>(null)

  // Client-side search over every language's title + description, so a term in
  // either locale finds the video regardless of the current UI language.
  const q = query.trim().toLowerCase()
  const filtered = q
    ? tutorials.filter((tu) =>
        [tu.title_en, tu.title_fa, tu.description_en, tu.description_fa]
          .filter(Boolean)
          .some((field) => field!.toLowerCase().includes(q))
      )
    : tutorials

  const localize = (tu: Tutorial) => ({
    title: (isFa && tu.title_fa) || tu.title_en || "",
    description: (isFa && tu.description_fa) || tu.description_en || "",
  })

  return (
    <div className="mx-auto w-full max-w-6xl">
      <header className="flex flex-col gap-1.5 pb-8">
        <div className="flex items-center gap-2">
          <GraduationCapIcon className="text-primary size-4" />
          <Eyebrow>{t("tutorials.eyebrow")}</Eyebrow>
        </div>

        <h1 className="text-2xl font-semibold tracking-tight">{t("tutorials.title")}</h1>
        <p className="text-muted-foreground text-sm">{t("tutorials.subtitle")}</p>

        <div className="relative mt-4 max-w-md">
          <SearchIcon className="text-muted-foreground pointer-events-none absolute start-3 top-1/2 size-4 -translate-y-1/2" />
          <Input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={t("tutorials.searchPh")}
            className="ps-9"
            aria-label={t("tutorials.searchPh")}
          />
        </div>
      </header>

      {isLoading ? (
        <TutorialGridSkeleton />
      ) : tutorials.length === 0 ? (
        <EmptyState title={t("tutorials.empty")} hint={t("tutorials.emptyHint")} />
      ) : filtered.length === 0 ? (
        <EmptyState title={t("tutorials.noResults", { query })} hint={t("tutorials.noResultsHint")} search />
      ) : (
        <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((tu, i) => {
            const { title, description } = localize(tu)
            return (
              <button
                key={tu.id}
                type="button"
                onClick={() => setActive(tu)}
                className={cn(
                  // No overflow-hidden here: it clips the title's glyph
                  // side-bearing (tracking-tight) at the start edge. The poster
                  // below clips itself, so the card doesn't need it.
                  "group animate-reveal focus-visible:ring-ring/60 flex flex-col rounded-2xl text-start outline-none focus-visible:ring-2"
                )}
                style={{ animationDelay: `${Math.min(i, 9) * 60}ms` }}
              >
                {/* Poster — 16:9, thumbnail with a play affordance on hover. */}
                <div className="bg-muted ring-border/60 group-hover:ring-foreground/20 relative aspect-video overflow-hidden rounded-2xl ring-1 transition-[box-shadow,transform] duration-[--dur-slow] ease-[--ease-out] group-hover:-translate-y-0.5">
                  {tu.thumbnail_url ? (
                    <img
                      src={tu.thumbnail_url}
                      alt=""
                      loading="lazy"
                      className="size-full object-cover transition-transform duration-[--dur-slow] ease-[--ease-out] group-hover:scale-[1.03]"
                    />
                  ) : (
                    <div className="text-muted-foreground/40 grid size-full place-items-center">
                      <VideoIcon className="size-10" />
                    </div>
                  )}

                  {/* Legibility scrim + play button. */}
                  <div className="absolute inset-0 bg-gradient-to-t from-black/45 via-transparent to-transparent opacity-0 transition-opacity duration-[--dur-slow] group-hover:opacity-100" />
                  <div className="absolute inset-0 grid place-items-center">
                    <span className="bg-background/85 text-primary ring-foreground/10 grid size-12 translate-y-1 place-items-center rounded-full opacity-0 shadow-lg ring-1 backdrop-blur-sm transition-all duration-[--dur-slow] ease-[--ease-out] group-hover:translate-y-0 group-hover:opacity-100">
                      <PlayIcon className="size-5 translate-x-px fill-current rtl:-translate-x-px" />
                    </span>
                  </div>
                </div>

                <div className="px-1 pt-3">
                  <h2 className="group-hover:text-primary line-clamp-2 text-sm font-semibold tracking-tight transition-colors">
                    {title || t("tutorials.untitled")}
                  </h2>
                  {description && (
                    <p className="text-muted-foreground mt-1 line-clamp-2 text-xs leading-relaxed">
                      {description
                        .replace(/[#*_`>[\]()!-]/g, " ")
                        .replace(/\s+/g, " ")
                        .trim()}
                    </p>
                  )}
                </div>
              </button>
            )
          })}
        </div>
      )}

      <PlayerDialog
        tutorial={active}
        onClose={() => setActive(null)}
        localize={localize}
        watchLabel={t("tutorials.nowPlaying")}
      />
    </div>
  )
}

function PlayerDialog({
  tutorial,
  onClose,
  localize,
  watchLabel,
}: {
  tutorial: Tutorial | null
  onClose: () => void
  localize: (tu: Tutorial) => { title: string; description: string }
  watchLabel: string
}) {
  const open = !!tutorial
  const view = tutorial ? localize(tutorial) : { title: "", description: "" }

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onClose()}>
      <DialogContent showCloseButton={false} className="max-w-3xl gap-0 overflow-hidden p-0 sm:max-w-3xl">
        {tutorial && (
          <>
            <div className="bg-black">
              <iframe
                key={tutorial.id}
                src={aparatEmbedUrl(tutorial.aparat_hash ?? "")}
                title={view.title}
                sandbox="allow-scripts allow-presentation allow-popups"
                allow="autoplay; fullscreen; picture-in-picture"
                allowFullScreen
                className="aspect-video w-full border-0"
              />
            </div>
            <div className="max-h-[40vh] space-y-2 overflow-y-auto p-5">
              <div className="flex items-center gap-2">
                <span className="bg-primary animate-pulse-dot size-1.5 rounded-full" />
                <span className="text-muted-foreground font-mono text-[0.65rem] tracking-[0.2em] uppercase">
                  {watchLabel}
                </span>
              </div>
              <h2 className="text-lg font-semibold tracking-tight text-balance">{view.title}</h2>
              {view.description && <ChangelogMarkdown>{view.description}</ChangelogMarkdown>}
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}

function TutorialGridSkeleton() {
  return (
    <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: 6 }).map((_, i) => (
        <div key={i} className="flex flex-col">
          <Skeleton className="aspect-video w-full rounded-2xl" />
          <Skeleton className="mt-3 h-4 w-3/4 rounded" />
          <Skeleton className="mt-2 h-3 w-full rounded" />
        </div>
      ))}
    </div>
  )
}

function EmptyState({ title, hint, search }: { title: string; hint: string; search?: boolean }) {
  return (
    <div className="relative isolate overflow-hidden rounded-2xl border px-6 py-20 text-center">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 -z-10 [mask-image:radial-gradient(60%_50%_at_50%_30%,black,transparent)] opacity-70"
        style={{
          background:
            "radial-gradient(circle at 50% 25%, color-mix(in oklch, var(--primary) 16%, transparent), transparent 60%)",
        }}
      />
      <div className="bg-primary/10 text-primary ring-primary/20 mx-auto grid size-14 place-items-center rounded-2xl ring-1">
        {search ? <SearchIcon className="size-6" /> : <GraduationCapIcon className="size-6" />}
      </div>
      <p className="mt-5 text-lg font-semibold tracking-tight text-balance">{title}</p>
      <p className="text-muted-foreground mx-auto mt-1 max-w-xs text-sm text-pretty">{hint}</p>
    </div>
  )
}
