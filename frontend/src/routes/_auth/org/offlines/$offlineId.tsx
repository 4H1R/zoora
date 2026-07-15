import type { GithubCom4H1RZooraInternalDomainMedia as Media } from "@/api/model"

import { createFileRoute, Link } from "@tanstack/react-router"
import { ArrowLeftIcon, DownloadIcon, EyeIcon, FileIcon, Loader2Icon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useGetClassesSessionsSessionId } from "@/api/classes/classes"
import { useGetMedia, useGetMediaIdDownloadUrl } from "@/api/media/media"
import { useGetOfflinesId } from "@/api/offlines/offlines"
import { Eyebrow } from "@/components/eyebrow"
import { useBreadcrumb } from "@/components/layout/breadcrumb-context"
import { OFFLINE_ATTACHMENTS_COLLECTION, OFFLINE_MODEL_TYPE } from "@/components/org/offlines/attachments"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { useOrgGuard } from "@/lib/access"
import { orgHead } from "@/lib/org-head"
import { formatSessionDate } from "@/lib/session-status"

export const Route = createFileRoute("/_auth/org/offlines/$offlineId")({
  head: () => orgHead("org.offline.title"),
  component: RouteComponent,
})

function AttachmentViewer({ media }: { media: Media }) {
  const { t } = useTranslation()
  const downloadQuery = useGetMediaIdDownloadUrl(media.id ?? "", undefined, {
    query: { enabled: !!media.id, staleTime: 30 * 60 * 1000 },
  })
  const url = (downloadQuery.data?.status === 200 && downloadQuery.data.data.data?.url) || null
  const mime = media.mime_type ?? ""
  const label = media.name || media.file_name || ""

  return (
    <div className="border-border bg-card flex flex-col gap-3 rounded-2xl border p-4">
      <div className="flex items-center gap-2">
        <FileIcon className="text-muted-foreground size-4 shrink-0" />
        <span className="min-w-0 flex-1 truncate text-sm font-medium">{label}</span>
        {url && (
          <Button variant="outline" size="sm" render={<a href={url} download={label} />}>
            <DownloadIcon className="size-4" />
            {t("org.offline.download")}
          </Button>
        )}
      </div>
      {!url ? (
        <div className="bg-muted flex h-40 items-center justify-center rounded-xl">
          <Loader2Icon className="size-5 animate-spin opacity-60" />
        </div>
      ) : mime.startsWith("video/") ? (
        <video controls src={url} className="max-h-[70vh] w-full rounded-xl" />
      ) : mime.startsWith("image/") ? (
        <img src={url} alt={label} className="max-h-[70vh] w-full rounded-xl object-contain" />
      ) : mime.startsWith("audio/") ? (
        <audio controls src={url} className="w-full" />
      ) : null}
    </div>
  )
}

function RouteComponent() {
  const { t, i18n } = useTranslation()
  const { offlineId } = Route.useParams()
  const allowed = useOrgGuard(["offlines:view", "offlines:view_any"])

  const { data, isPending, isError } = useGetOfflinesId(offlineId)
  const room = (data?.status === 200 && data.data.data) || undefined

  const sessionId = room?.class_session_id
  const { data: sessionData } = useGetClassesSessionsSessionId(sessionId ?? "", {
    query: { enabled: !!sessionId },
  })
  const session = (sessionData?.status === 200 && sessionData.data.data) || undefined

  useBreadcrumb([
    { label: t("org.nav.classes"), to: "/org/classes" },
    {
      label: session?.name ?? null,
      to: "/org/classes/class-sessions/$classSessionId",
      params: { classSessionId: sessionId ?? "" },
      loading: !session,
    },
    { label: room?.title ?? null, loading: !room },
  ])

  const mediaQuery = useGetMedia(
    { model_type: OFFLINE_MODEL_TYPE, model_id: offlineId, collection: OFFLINE_ATTACHMENTS_COLLECTION },
    { query: { enabled: allowed } }
  )
  const attachments = (mediaQuery.data?.status === 200 && mediaQuery.data.data.data) || []

  if (!allowed) return null

  if (isPending) {
    return (
      <div className="flex flex-col gap-8 py-10">
        <Skeleton className="h-5 w-40" />
        <Skeleton className="h-10 w-2/3" />
        <Skeleton className="h-40 w-full" />
      </div>
    )
  }

  if (isError || !room) {
    return (
      <div className="flex flex-col items-start gap-4 py-16">
        <h1 className="text-2xl font-semibold tracking-tight">{t("org.offline.notFound")}</h1>
        <Button variant="outline" render={<Link to="/org/classes" />}>
          <ArrowLeftIcon className="size-4" />
          {t("org.offline.back")}
        </Button>
      </div>
    )
  }

  const createdStr = formatSessionDate(room.created_at, i18n.language, "long")
  const publishedStr = room.published_at ? formatSessionDate(room.published_at, i18n.language, "long") : null

  return (
    <div className="flex flex-col gap-10 pb-16">
      <div className="pt-6">
        <Link
          to="/org/classes/class-sessions/$classSessionId"
          params={{ classSessionId: room.class_session_id ?? "" }}
          className="text-muted-foreground hover:text-foreground inline-flex items-center gap-2 font-mono text-xs tracking-[0.25em] uppercase transition-colors"
        >
          <ArrowLeftIcon className="size-3.5" />
          {t("org.offline.back")}
        </Link>
      </div>

      <header className="flex flex-col gap-4">
        <Eyebrow>{t("org.offline.title")}</Eyebrow>
        <h1 className="text-3xl font-semibold tracking-tight md:text-4xl">{room.title}</h1>
        {room.description && (
          <p className="text-muted-foreground max-w-2xl text-base leading-relaxed">{room.description}</p>
        )}
        <div className="text-muted-foreground flex flex-wrap items-center gap-4 text-sm">
          <span className="inline-flex items-center gap-1.5 tabular-nums">
            <EyeIcon className="size-4" />
            {room.view_count ?? 0} {t("org.offline.views")}
          </span>
          {publishedStr && (
            <span>
              {t("org.offline.published")}: {publishedStr}
            </span>
          )}
          <span>
            {t("org.offline.created")}: {createdStr}
          </span>
        </div>
      </header>

      <section className="flex flex-col gap-4">
        <Eyebrow>{t("org.offline.attachments")}</Eyebrow>
        {attachments.length === 0 ? (
          <p className="text-muted-foreground text-sm">{t("org.offline.noAttachments")}</p>
        ) : (
          <div className="flex flex-col gap-4">
            {attachments.map((m) => (
              <AttachmentViewer key={m.id} media={m} />
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
