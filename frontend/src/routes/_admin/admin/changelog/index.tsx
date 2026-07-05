import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, useNavigate } from "@tanstack/react-router"
import {
  ChevronRightIcon,
  FileTextIcon,
  PlusIcon,
  ScrollTextIcon,
  SparklesIcon,
  SendIcon,
  Trash2Icon,
} from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetAdminChangelogQueryKey,
  useDeleteAdminChangelogId,
  useGetAdminChangelog,
  usePostAdminChangelog,
} from "@/api/admin-changelog/admin-changelog"
import type { GithubCom4H1RZooraInternalDomainChangelogEntry as Entry } from "@/api/model"
import { StatCards } from "@/components/data-table/stat-cards"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { adminHead } from "@/lib/admin-head"
import { useFormatDate } from "@/lib/format-date"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_admin/admin/changelog/")({
  head: () => adminHead("admin.changelog.title"),
  component: ChangelogListPage,
})

function ChangelogListPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const formatDate = useFormatDate()
  const { data, isLoading } = useGetAdminChangelog({ page: 1 })
  const entries = ((data?.status === 200 && data.data.data?.items) || []) as Entry[]

  const [deleteTarget, setDeleteTarget] = useState<Entry | null>(null)

  const create = usePostAdminChangelog({
    mutation: {
      onSuccess: (res) => {
        const id = res?.status === 201 && res.data.data?.id
        if (id) navigate({ to: "/admin/changelog/$id", params: { id } })
      },
      onError: () => toast.error(t("admin.changelog.createFailed")),
    },
  })
  const del = useDeleteAdminChangelogId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.changelog.deleted"))
        queryClient.invalidateQueries({ queryKey: getGetAdminChangelogQueryKey() })
        setDeleteTarget(null)
      },
    },
  })

  function newDraft() {
    // body_en is `required` on the backend DTO — seed a starter template so the
    // draft-create call passes validation instead of silently 400-ing.
    create.mutate({
      data: { title_en: t("admin.changelog.untitled"), body_en: t("admin.changelog.starterBody") },
    })
  }

  const publishedCount = entries.filter((e) => e.published_at).length
  const stats = [
    {
      icon: <ScrollTextIcon />,
      label: t("admin.changelog.stats.total"),
      value: entries.length,
      loading: isLoading,
    },
    {
      icon: <SendIcon />,
      label: t("admin.changelog.stats.published"),
      value: publishedCount,
      loading: isLoading,
    },
    {
      icon: <FileTextIcon />,
      label: t("admin.changelog.stats.drafts"),
      value: entries.length - publishedCount,
      loading: isLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("admin.changelog.title")}
        actions={
          <Button size="sm" onClick={newDraft} disabled={create.isPending}>
            <PlusIcon data-icon="inline-start" />
            {t("common.create")}
          </Button>
        }
      />
      <p className="text-muted-foreground -mt-4 text-sm">{t("admin.changelog.subtitle")}</p>

      <StatCards stats={stats} />

      {isLoading ? (
        <div className="grid gap-3">
          {[0, 1, 2].map((i) => (
            <Skeleton key={i} className="h-[4.5rem] w-full rounded-xl" />
          ))}
        </div>
      ) : entries.length === 0 ? (
        <Card className="flex flex-col items-center gap-3 border-dashed py-14 text-center">
          <div className="bg-muted text-muted-foreground flex size-12 items-center justify-center rounded-2xl">
            <ScrollTextIcon className="size-6" />
          </div>
          <div className="space-y-1">
            <p className="font-medium">{t("admin.changelog.empty")}</p>
            <p className="text-muted-foreground mx-auto max-w-sm text-sm">
              {t("admin.changelog.emptyHint")}
            </p>
          </div>
          <Button size="sm" className="mt-1" onClick={newDraft} disabled={create.isPending}>
            <PlusIcon data-icon="inline-start" />
            {t("common.create")}
          </Button>
        </Card>
      ) : (
        <div className="grid gap-3">
          {entries.map((e) => {
            const isPublished = !!e.published_at
            return (
              <Card
                key={e.id}
                onClick={() => navigate({ to: "/admin/changelog/$id", params: { id: e.id! } })}
                className={cn(
                  "group relative flex-row items-center gap-4 overflow-hidden p-4 pe-3",
                  "cursor-pointer transition-all duration-[--dur-slow] ease-[--ease-out]",
                  "hover:border-foreground/20 hover:shadow-md"
                )}
              >
                {/* status accent rail */}
                <div
                  className={cn(
                    "absolute inset-y-0 start-0 w-1",
                    isPublished ? "bg-primary" : "bg-muted-foreground/25"
                  )}
                />

                <Badge
                  variant="outline"
                  className="shrink-0 font-mono text-xs tabular-nums"
                  title={t("admin.changelog.version")}
                >
                  {e.version || "—"}
                </Badge>

                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <span className="truncate font-medium">
                      {e.title_en || t("admin.changelog.untitled")}
                    </span>
                    {e.is_major && (
                      <Badge className="gap-1">
                        <SparklesIcon className="size-3" />
                        {t("whatsNew.major")}
                      </Badge>
                    )}
                  </div>
                  <p className="text-muted-foreground mt-0.5 truncate text-xs">
                    {isPublished
                      ? t("admin.changelog.publishedOn", { date: formatDate(e.published_at) })
                      : t("admin.changelog.createdOn", { date: formatDate(e.created_at) })}
                  </p>
                </div>

                <Badge
                  variant={isPublished ? "default" : "secondary"}
                  className="shrink-0 gap-1.5"
                >
                  <span
                    className={cn(
                      "size-1.5 rounded-full",
                      isPublished ? "bg-primary-foreground/90" : "bg-muted-foreground"
                    )}
                  />
                  {isPublished ? t("admin.changelog.published") : t("admin.changelog.draft")}
                </Badge>

                <Button
                  variant="ghost"
                  size="icon"
                  className="text-muted-foreground hover:text-destructive shrink-0 opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100"
                  onClick={(ev) => {
                    ev.stopPropagation()
                    setDeleteTarget(e)
                  }}
                >
                  <Trash2Icon className="size-4" />
                </Button>

                <ChevronRightIcon className="text-muted-foreground/50 size-4 shrink-0 transition-transform group-hover:translate-x-0.5 rtl:rotate-180" />
              </Card>
            )
          })}
        </div>
      )}

      <DeleteConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null)
        }}
        resourceName={deleteTarget?.title_en ?? t("admin.changelog.untitled")}
        onConfirm={() => {
          if (deleteTarget?.id) del.mutate({ id: deleteTarget.id })
        }}
        isLoading={del.isPending}
      />
    </div>
  )
}
