import type { GithubCom4H1RZooraInternalDomainTutorial as Tutorial } from "@/api/model"

import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, useNavigate } from "@tanstack/react-router"
import {
  ChevronRightIcon,
  FileVideoIcon,
  GraduationCapIcon,
  GripVerticalIcon,
  PlusIcon,
  SendIcon,
  Trash2Icon,
  VideoIcon,
} from "lucide-react"
import { Reorder, useDragControls } from "motion/react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getGetAdminTutorialsQueryKey,
  useDeleteAdminTutorialsId,
  useGetAdminTutorials,
  usePostAdminTutorials,
  usePutAdminTutorialsReorder,
} from "@/api/admin-tutorials/admin-tutorials"
import { StatCards } from "@/components/data-table/stat-cards"
import { DeleteConfirmDialog } from "@/components/form/delete-confirm-dialog"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { adminHead } from "@/lib/admin-head"
import { cn } from "@/lib/utils"

export const Route = createFileRoute("/_admin/admin/tutorials/")({
  head: () => adminHead("admin.tutorials.title"),
  component: TutorialsListPage,
})

function TutorialsListPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { data, isLoading } = useGetAdminTutorials()
  const server = ((data?.status === 200 && data.data.data) || []) as Tutorial[]

  // Local ordering mirror for optimistic drag; kept in sync with the server
  // list except while a drag is in flight.
  const [items, setItems] = useState<Tutorial[]>(server)
  const draggingRef = useRef(false)
  useEffect(() => {
    if (!draggingRef.current) setItems(server)
  }, [server])

  const [deleteTarget, setDeleteTarget] = useState<Tutorial | null>(null)

  const reorder = usePutAdminTutorialsReorder({
    mutation: {
      onError: () => {
        toast.error(t("admin.tutorials.reorderFailed"))
        queryClient.invalidateQueries({ queryKey: getGetAdminTutorialsQueryKey() })
      },
    },
  })
  const create = usePostAdminTutorials({
    mutation: {
      onSuccess: (res) => {
        const id = res?.status === 201 && res.data.data?.id
        if (id) navigate({ to: "/admin/tutorials/$id", params: { id } })
      },
      onError: () => toast.error(t("admin.tutorials.createFailed")),
    },
  })
  const del = useDeleteAdminTutorialsId({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.tutorials.deleted"))
        queryClient.invalidateQueries({ queryKey: getGetAdminTutorialsQueryKey() })
        setDeleteTarget(null)
      },
    },
  })

  function newDraft() {
    // aparat_hash is `required` on the backend DTO — seed a placeholder so the
    // draft-create passes validation; the editor prompts for the real link.
    create.mutate({ data: { title_en: t("admin.tutorials.untitled"), aparat_hash: "placeholder" } })
  }

  function persistOrder(next: Tutorial[]) {
    draggingRef.current = false
    const ids = next.map((tu) => tu.id!).filter(Boolean)
    // Skip the write when nothing moved.
    if (ids.join() === server.map((tu) => tu.id).join()) return
    reorder.mutate({ data: { ids } })
  }

  const publishedCount = items.filter((tu) => tu.published_at).length
  const stats = [
    { icon: <FileVideoIcon />, label: t("admin.tutorials.stats.total"), value: items.length, loading: isLoading },
    { icon: <SendIcon />, label: t("admin.tutorials.stats.published"), value: publishedCount, loading: isLoading },
    {
      icon: <VideoIcon />,
      label: t("admin.tutorials.stats.drafts"),
      value: items.length - publishedCount,
      loading: isLoading,
    },
  ]

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={t("admin.tutorials.title")}
        actions={
          <Button size="sm" onClick={newDraft} disabled={create.isPending}>
            <PlusIcon data-icon="inline-start" />
            {t("common.create")}
          </Button>
        }
      />
      <p className="text-muted-foreground -mt-4 text-sm">{t("admin.tutorials.subtitle")}</p>

      <StatCards stats={stats} />

      {isLoading ? (
        <div className="grid gap-3">
          {[0, 1, 2].map((i) => (
            <Skeleton key={i} className="h-[4.5rem] w-full rounded-xl" />
          ))}
        </div>
      ) : items.length === 0 ? (
        <Card className="flex flex-col items-center gap-3 border-dashed py-14 text-center">
          <div className="bg-muted text-muted-foreground flex size-12 items-center justify-center rounded-2xl">
            <GraduationCapIcon className="size-6" />
          </div>
          <div className="space-y-1">
            <p className="font-medium">{t("admin.tutorials.empty")}</p>
            <p className="text-muted-foreground mx-auto max-w-sm text-sm">{t("admin.tutorials.emptyHint")}</p>
          </div>
          <Button size="sm" className="mt-1" onClick={newDraft} disabled={create.isPending}>
            <PlusIcon data-icon="inline-start" />
            {t("common.create")}
          </Button>
        </Card>
      ) : (
        <Reorder.Group
          axis="y"
          values={items}
          onReorder={(next) => {
            draggingRef.current = true
            setItems(next)
          }}
          className="grid gap-3"
        >
          {items.map((tu) => (
            <TutorialRow
              key={tu.id}
              tutorial={tu}
              onEdit={() => navigate({ to: "/admin/tutorials/$id", params: { id: tu.id! } })}
              onDelete={() => setDeleteTarget(tu)}
              onSettle={() => persistOrder(items)}
            />
          ))}
        </Reorder.Group>
      )}

      <DeleteConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null)
        }}
        resourceName={deleteTarget?.title_en ?? t("admin.tutorials.untitled")}
        onConfirm={() => {
          if (deleteTarget?.id) del.mutate({ id: deleteTarget.id })
        }}
        isLoading={del.isPending}
      />
    </div>
  )
}

function TutorialRow({
  tutorial: tu,
  onEdit,
  onDelete,
  onSettle,
}: {
  tutorial: Tutorial
  onEdit: () => void
  onDelete: () => void
  onSettle: () => void
}) {
  const { t } = useTranslation()
  const controls = useDragControls()
  const isPublished = !!tu.published_at

  return (
    <Reorder.Item value={tu} dragListener={false} dragControls={controls} onDragEnd={onSettle} className="list-none">
      <Card
        className={cn(
          "group relative flex-row items-center gap-3 overflow-hidden p-3 pe-3",
          "hover:border-foreground/20 transition-all duration-[--dur-slow] ease-[--ease-out] hover:shadow-md"
        )}
      >
        {/* status accent rail */}
        <div className={cn("absolute inset-y-0 start-0 w-1", isPublished ? "bg-primary" : "bg-muted-foreground/25")} />

        {/* drag handle — only this initiates a reorder */}
        <button
          type="button"
          aria-label={t("admin.tutorials.dragHint")}
          onPointerDown={(e) => controls.start(e)}
          className="text-muted-foreground/50 hover:text-foreground ms-1 cursor-grab touch-none active:cursor-grabbing"
        >
          <GripVerticalIcon className="size-4" />
        </button>

        {/* thumbnail */}
        <div className="bg-muted ring-border/60 relative aspect-video w-24 shrink-0 overflow-hidden rounded-md ring-1">
          {tu.thumbnail_url ? (
            <img src={tu.thumbnail_url} alt="" loading="lazy" className="size-full object-cover" />
          ) : (
            <div className="text-muted-foreground/40 grid size-full place-items-center">
              <VideoIcon className="size-4" />
            </div>
          )}
        </div>

        <button type="button" onClick={onEdit} className="min-w-0 flex-1 text-start">
          <span className="truncate font-medium">{tu.title_en || t("admin.tutorials.untitled")}</span>
          <p className="text-muted-foreground mt-0.5 line-clamp-1 text-xs">
            {tu.description_en?.replace(/[#*_`>[\]()!-]/g, " ").trim() || t("admin.tutorials.noDescription")}
          </p>
        </button>

        <Badge variant={isPublished ? "default" : "secondary"} className="shrink-0 gap-1.5">
          <span
            className={cn("size-1.5 rounded-full", isPublished ? "bg-primary-foreground/90" : "bg-muted-foreground")}
          />
          {isPublished ? t("admin.tutorials.published") : t("admin.tutorials.draft")}
        </Badge>

        <Button
          variant="ghost"
          size="icon"
          className="text-muted-foreground hover:text-destructive shrink-0 opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100"
          onClick={onDelete}
        >
          <Trash2Icon className="size-4" />
        </Button>

        <ChevronRightIcon
          className="text-muted-foreground/50 size-4 shrink-0 cursor-pointer transition-transform group-hover:translate-x-0.5 rtl:rotate-180"
          onClick={onEdit}
        />
      </Card>
    </Reorder.Item>
  )
}
