import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { PlusIcon } from "lucide-react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetAdminChangelogQueryKey,
  useDeleteAdminChangelogId,
  useGetAdminChangelog,
  usePostAdminChangelog,
} from "@/api/admin-changelog/admin-changelog"
import type { GithubCom4H1RZooraInternalDomainChangelogEntry as Entry } from "@/api/model"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { adminHead } from "@/lib/admin-head"

export const Route = createFileRoute("/_admin/admin/changelog/")({
  head: () => adminHead("admin.changelog.title"),
  component: ChangelogListPage,
})

function ChangelogListPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { data } = useGetAdminChangelog({ page: 1 })
  const entries = ((data?.status === 200 && data.data.data?.items) || []) as Entry[]

  const create = usePostAdminChangelog({
    mutation: {
      onSuccess: (res) => {
        const id = res?.status === 201 && res.data.data?.id
        if (id) navigate({ to: "/admin/changelog/$id", params: { id } })
      },
    },
  })
  const del = useDeleteAdminChangelogId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.changelog.deleted"))
        queryClient.invalidateQueries({ queryKey: getGetAdminChangelogQueryKey() })
      },
    },
  })

  function newDraft() {
    create.mutate({ data: { title_en: t("admin.changelog.untitled"), body_en: "" } })
  }

  return (
    <div className="space-y-4">
      <PageHeader
        title={t("admin.changelog.title")}
        actions={
          <Button onClick={newDraft}>
            <PlusIcon className="size-4" /> {t("admin.changelog.new")}
          </Button>
        }
      />
      <div className="grid gap-3">
        {entries.map((e) => (
          <Card
            key={e.id}
            className="flex cursor-pointer items-center justify-between p-4"
            onClick={() => navigate({ to: "/admin/changelog/$id", params: { id: e.id! } })}
          >
            <div className="flex items-center gap-2">
              {e.version && <Badge variant="secondary">{e.version}</Badge>}
              <span className="font-medium">{e.title_en}</span>
              <Badge variant={e.published_at ? "default" : "outline"}>
                {e.published_at ? t("admin.changelog.published") : t("admin.changelog.draft")}
              </Badge>
              {e.is_major && <Badge>{t("whatsNew.major")}</Badge>}
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={(ev) => {
                ev.stopPropagation()
                if (confirm(t("admin.changelog.confirmDelete"))) del.mutate({ id: e.id! })
              }}
            >
              {t("common.delete")}
            </Button>
          </Card>
        ))}
      </div>
    </div>
  )
}
