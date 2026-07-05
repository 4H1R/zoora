import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { useEffect } from "react"
import { useTranslation } from "react-i18next"

import {
  getGetChangelogStatusQueryKey,
  useGetChangelog,
  usePostChangelogMarkSeen,
} from "@/api/changelog/changelog"
import type { GithubCom4H1RZooraInternalDomainChangelogEntry as Entry } from "@/api/model"
import { ChangelogMarkdown } from "@/components/changelog/markdown"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Card } from "@/components/ui/card"
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

  return (
    <div className="mx-auto w-full max-w-3xl space-y-6">
      <div className="space-y-1">
        <PageHeader title={t("whatsNew.title")} />
        <p className="text-muted-foreground text-sm">{t("whatsNew.subtitle")}</p>
      </div>
      {entries.length === 0 && (
        <p className="text-muted-foreground text-sm">{t("whatsNew.empty")}</p>
      )}
      {entries.map((e) => {
        const title = (isFa && e.title_fa) || e.title_en
        const body = (isFa && e.body_fa) || e.body_en || ""
        return (
          <Card key={e.id} className="space-y-3 p-5">
            <div className="flex items-center gap-2">
              {e.version && <Badge variant="secondary">{e.version}</Badge>}
              {e.is_major && <Badge>{t("whatsNew.major")}</Badge>}
              <span className="text-muted-foreground text-xs">
                {e.published_at
                  ? new Date(e.published_at).toLocaleDateString(i18n.language)
                  : ""}
              </span>
            </div>
            <h2 className="text-lg font-semibold">{title}</h2>
            <ChangelogMarkdown>{body}</ChangelogMarkdown>
          </Card>
        )
      })}
    </div>
  )
}
